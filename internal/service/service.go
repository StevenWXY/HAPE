package service

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/store"
)

type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string { return e.Code }

func apiError(status int, code, message string) error {
	return &APIError{Status: status, Code: code, Message: message}
}

type Service struct {
	store           *store.Memory
	now             func() time.Time
	platform        ExternalPlatformClient
	externalMu      sync.Mutex
	airdropVerifier func(domain.AirdropRecord, AirdropResultInput) error
	bnbExecution    *BNBExecution
}

func New(memory *store.Memory, platforms ...ExternalPlatformClient) *Service {
	var platform ExternalPlatformClient
	if len(platforms) > 0 {
		platform = platforms[0]
	}
	return NewWithExternalPlatform(memory, platform)
}

// NewWithExternalPlatform builds a service with an explicitly configured
// partner adapter. Keeping the adapter behind this boundary lets integration
// testing use a fake client and production use the HTTP adapter without
// changing handlers or business rules.
func NewWithExternalPlatform(memory *store.Memory, platform ExternalPlatformClient) *Service {
	if platform == nil {
		platform = NewHTTPExternalPlatform("", nil)
	}
	return &Service{store: memory, now: time.Now, platform: platform}
}

// SetExternalPlatform is intended for startup wiring and integration tests.
// It does not mutate any persisted Clipli state.
func (s *Service) SetExternalPlatform(platform ExternalPlatformClient) {
	if platform != nil {
		s.externalMu.Lock()
		defer s.externalMu.Unlock()
		s.platform = platform
	}
}

func (s *Service) Snapshot() store.State {
	state := s.store.Snapshot()
	state.CLIP.AvailableBalance = state.CLIP.Balance - state.CLIP.Reserved
	syncGenerationAccount(&state)
	return state
}

const anonymousUserRef = "anonymous-demo"

func distributeCLIP(state *store.State, requestID, assetID string, amount int, now time.Time) error {
	if amount <= 0 {
		return nil
	}
	if state.CLIPTreasury.TreasuryBalance < amount {
		return apiError(http.StatusServiceUnavailable, "clip_treasury_insufficient", "The platform treasury cannot satisfy this CLIP distribution")
	}
	userBefore := state.CLIP.Balance
	treasuryBefore := state.CLIPTreasury.TreasuryBalance
	state.CLIP.Balance += amount
	state.CLIPTreasury.TreasuryBalance -= amount
	state.CLIPTreasury.LedgerOutstanding += amount
	state.CLIPTreasury.TotalDistributed += amount
	distribution := domain.CLIPDistribution{
		ID: uniqueID("clip-distribution", now), RequestID: requestID, RuleCode: "hapw_redemption", UserRef: anonymousUserRef,
		AssetID: assetID, Amount: amount, UserBalanceBefore: userBefore, UserBalanceAfter: state.CLIP.Balance,
		TreasuryBefore: treasuryBefore, TreasuryAfter: state.CLIPTreasury.TreasuryBalance, Status: "completed", CreatedAt: formatMinute(now),
	}
	state.Distributions = prepend(distribution, state.Distributions)
	return nil
}

func reclaimCLIP(state *store.State, amount int) {
	if amount <= 0 {
		return
	}
	state.CLIPTreasury.TreasuryBalance += amount
	state.CLIPTreasury.LedgerOutstanding -= amount
	if state.CLIPTreasury.LedgerOutstanding < 0 {
		state.CLIPTreasury.LedgerOutstanding = 0
	}
	state.CLIPTreasury.TotalReclaimed += amount
}

func (s *Service) ListWorks() []domain.Work {
	return s.Snapshot().Works
}

func (s *Service) GetWork(id string) (domain.Work, bool) {
	for _, work := range s.Snapshot().Works {
		if work.ID == id {
			return work, true
		}
	}
	return domain.Work{}, false
}

type AssetListOptions struct {
	Search            string
	Status            string
	Owner             string
	RightsHolder      string
	Redeemable        *bool
	ExchangeAvailable *bool
	Page              int
	PageSize          int
	Sort              string
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

type AssetListResult struct {
	Items      []domain.HAPWAsset `json:"items"`
	Pagination Pagination         `json:"pagination"`
}

func (s *Service) ListAssets(options AssetListOptions) AssetListResult {
	assets := make([]domain.HAPWAsset, 0)
	search := strings.ToLower(strings.TrimSpace(options.Search))
	status := strings.ToLower(strings.TrimSpace(options.Status))
	owner := strings.ToLower(strings.TrimSpace(options.Owner))
	holder := strings.ToLower(strings.TrimSpace(options.RightsHolder))
	for _, asset := range s.PortfolioSnapshot().Assets {
		if asset.Kind != "HAPW" {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(strings.Join([]string{asset.ID, asset.TokenID, asset.Name, asset.NameEn, asset.RightsHolder, asset.RightsHolderEn}, " ")), search) {
			continue
		}
		if status != "" && strings.ToLower(asset.RedemptionStatus) != status {
			continue
		}
		if owner != "" && strings.ToLower(asset.Owner) != owner {
			continue
		}
		if holder != "" && !strings.Contains(strings.ToLower(asset.RightsHolder+" "+asset.RightsHolderEn), holder) {
			continue
		}
		if options.Redeemable != nil && (asset.RedemptionStatus == "available") != *options.Redeemable {
			continue
		}
		if options.ExchangeAvailable != nil && asset.ExchangeAvailable != *options.ExchangeAvailable {
			continue
		}
		assets = append(assets, asset)
	}

	switch options.Sort {
	case "value":
		sort.SliceStable(assets, func(i, j int) bool { return assets[i].Value < assets[j].Value })
	case "-value":
		sort.SliceStable(assets, func(i, j int) bool { return assets[i].Value > assets[j].Value })
	case "issuedAt":
		sort.SliceStable(assets, func(i, j int) bool { return assets[i].Provenance.IssuedAt < assets[j].Provenance.IssuedAt })
	case "-issuedAt":
		sort.SliceStable(assets, func(i, j int) bool { return assets[i].Provenance.IssuedAt > assets[j].Provenance.IssuedAt })
	default:
		sort.SliceStable(assets, func(i, j int) bool { return tokenNumber(assets[i].TokenID) < tokenNumber(assets[j].TokenID) })
	}

	page, pageSize := options.Page, options.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	total := len(assets)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return AssetListResult{Items: assets[start:end], Pagination: Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}}
}

func tokenNumber(tokenID string) int {
	value, _ := strconv.Atoi(strings.TrimPrefix(tokenID, "#"))
	return value
}

func (s *Service) GetAsset(id string) (domain.HAPWAsset, bool) {
	for _, asset := range s.PortfolioSnapshot().Assets {
		if asset.ID == id || asset.TokenID == id || strings.TrimPrefix(asset.TokenID, "#") == strings.TrimPrefix(id, "#") {
			return asset, true
		}
	}
	return domain.HAPWAsset{}, false
}

type HAPWSummary struct {
	TotalAssets       int `json:"totalAssets"`
	RedeemableAssets  int `json:"redeemableAssets"`
	RedeemedAssets    int `json:"redeemedAssets"`
	RestrictedAssets  int `json:"restrictedAssets"`
	ExerciseEligible  int `json:"exerciseEligible"`
	ExchangeAvailable int `json:"exchangeAvailable"`
	TotalValue        int `json:"totalValue"`
	PotentialCredits  int `json:"potentialCredits"`
}

func (s *Service) HAPWSummary() HAPWSummary {
	var summary HAPWSummary
	for _, asset := range s.PortfolioSnapshot().Assets {
		if asset.Kind != "HAPW" {
			continue
		}
		summary.TotalAssets++
		summary.TotalValue += asset.Value
		summary.PotentialCredits += asset.CreditYield
		if asset.Transferable {
			summary.ExerciseEligible++
		}
		if asset.ExchangeAvailable {
			summary.ExchangeAvailable++
		}
		switch asset.RedemptionStatus {
		case "available":
			summary.RedeemableAssets++
		case "redeemed":
			summary.RedeemedAssets++
		default:
			summary.RestrictedAssets++
		}
	}
	return summary
}

func (s *Service) AssetHistory(assetID string) ([]domain.AssetEvent, error) {
	state := s.Snapshot()
	asset, ok := findAsset(state.Assets, assetID)
	if !ok {
		return nil, apiError(http.StatusNotFound, "hapw_not_found", "HAPW asset not found")
	}
	events := []domain.AssetEvent{{ID: "issue-" + asset.ID, AssetID: asset.ID, Type: "issued", Status: asset.Provenance.VerificationStatus, CreatedAt: asset.Provenance.IssuedAt, Details: map[string]any{"issuer": asset.Provenance.Issuer, "certificateId": asset.Provenance.CertificateID}}}
	for _, item := range state.Redemptions {
		if item.AssetID == asset.ID {
			events = append(events, domain.AssetEvent{ID: item.ID, AssetID: asset.ID, Type: "redemption", Status: item.StatusEn, CreatedAt: item.CreatedAt, Details: map[string]any{"receipt": item.Receipt, "creditsGranted": item.CreditsGranted, "clipGranted": item.ClipGranted}})
		}
	}
	for _, item := range state.Exercises {
		if item.AssetID == asset.ID {
			events = append(events, domain.AssetEvent{ID: item.ID, AssetID: asset.ID, Type: "exercise", Status: item.StatusCode, CreatedAt: item.CreatedAt, Details: map[string]any{"platformCode": item.PlatformCode}})
		}
	}
	for _, item := range state.Generations {
		if item.AssetID == asset.ID {
			events = append(events, domain.AssetEvent{ID: item.ID, AssetID: asset.ID, Type: "generation", Status: item.StatusEn, CreatedAt: item.CreatedAt, Details: map[string]any{"duration": item.Duration, "quality": item.Quality, "creditsUsed": item.CreditsUsed, "clipCost": item.ClipCost}})
		}
	}
	for _, item := range state.HAPWExchanges {
		if item.AssetID == asset.ID {
			events = append(events, domain.AssetEvent{ID: item.ID, AssetID: asset.ID, Type: "exchange", Status: item.StatusCode, CreatedAt: item.CreatedAt, Details: map[string]any{"price": item.Price, "fee": item.Fee, "total": item.Total}})
		}
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].CreatedAt > events[j].CreatedAt })
	return events, nil
}

type RedemptionInput struct {
	RequestID string `json:"requestId"`
	AssetID   string `json:"assetId"`
	Accepted  bool   `json:"accepted"`
}

type WalletConnectInput struct {
	Provider string `json:"provider"`
	Address  string `json:"address"`
	ChainID  string `json:"chainId"`
}

type AirdropInput struct {
	RequestID     string `json:"requestId"`
	RuleCode      string `json:"ruleCode"`
	WalletAddress string `json:"walletAddress"`
	ChainID       string `json:"chainId"`
	AssetID       string `json:"assetId"`
	Amount        int    `json:"amount"`
	Token         string `json:"token"`
}

type AirdropResultInput struct {
	Status        string `json:"status"`
	TxHash        string `json:"txHash"`
	ExecutorRef   string `json:"executorRef"`
	FailureReason string `json:"failureReason"`
}

// BNBChainConfig describes the networks accepted by the Clipli airdrop
// workflow while the BEP-20 contract is still being prepared.
type BNBChainConfig struct {
	ChainID          string `json:"chainId"`
	Name             string `json:"name"`
	Network          string `json:"network"`
	NativeCurrency   string `json:"nativeCurrency"`
	ContractDeployed bool   `json:"contractDeployed"`
	SimulationOnly   bool   `json:"simulationOnly"`
}

func BNBChainConfigs() []BNBChainConfig {
	return []BNBChainConfig{
		{ChainID: "0x61", Name: "BNB Smart Chain Testnet", Network: "testnet", NativeCurrency: "tBNB", ContractDeployed: false, SimulationOnly: true},
		{ChainID: "0x38", Name: "BNB Smart Chain Mainnet", Network: "mainnet", NativeCurrency: "BNB", ContractDeployed: false, SimulationOnly: true},
	}
}

func NormalizeBNBChain(chainID string) string {
	switch strings.ToLower(strings.TrimSpace(chainID)) {
	case "0x61", "97":
		return "0x61"
	case "0x38", "56":
		return "0x38"
	default:
		return ""
	}
}

func IsBNBChain(chainID string) bool { return NormalizeBNBChain(chainID) != "" }

type AirdropEligibility struct {
	Eligible      bool   `json:"eligible"`
	RuleCode      string `json:"ruleCode"`
	WalletAddress string `json:"walletAddress"`
	AssetID       string `json:"assetId,omitempty"`
	Amount        int    `json:"amount"`
	Token         string `json:"token"`
	Reason        string `json:"reason"`
	ExistingID    string `json:"existingAirdropId,omitempty"`
}

type RedemptionResult struct {
	Redemption        domain.HAPWRedemption    `json:"redemption"`
	GenerationAccount domain.GenerationAccount `json:"generationAccount"`
	ClipBalance       int                      `json:"clipBalance"`
}

func (s *Service) Redeem(input RedemptionInput) (RedemptionResult, bool, error) {
	var result RedemptionResult
	idempotent := false
	err := s.store.Update(func(state *store.State) error {
		if !validRequestID(input.RequestID) {
			return apiError(http.StatusBadRequest, "invalid_request_id", "A valid operation id is required")
		}
		for _, item := range state.Redemptions {
			if item.RequestID == input.RequestID {
				if item.AssetID != strings.TrimSpace(input.AssetID) || !input.Accepted {
					return apiError(http.StatusConflict, "request_id_reused", "The requestId is already used for another redemption")
				}
				syncGenerationAccount(state)
				result = RedemptionResult{Redemption: item, GenerationAccount: state.GenerationAccount, ClipBalance: state.CLIP.Balance}
				idempotent = true
				return nil
			}
		}
		index := assetIndex(state.Assets, input.AssetID)
		if index < 0 || state.Assets[index].RedemptionStatus != "available" || state.Assets[index].ExchangeAvailable || state.Assets[index].External.SyncStatus != "migrated" || state.Assets[index].External.TransferID == "" {
			return apiError(http.StatusBadRequest, "hapw_not_redeemable", "This HAPW is not available for redemption")
		}
		if !input.Accepted {
			return apiError(http.StatusBadRequest, "redemption_not_confirmed", "Confirm the irreversible HAPW redemption")
		}
		asset := &state.Assets[index]
		grant, _ := RedemptionCLIPGrant(asset.CreditYield)
		now := s.now()
		createdAt := formatMinute(now)
		redemption := domain.HAPWRedemption{
			ID: uniqueID("redeem", now), AssetID: asset.ID, Receipt: fmt.Sprintf("Clipli-LIC-%s-%04d", strings.TrimPrefix(asset.TokenID, "#"), now.UnixNano()%10000),
			CreditsGranted: asset.CreditYield, CreditsRemaining: asset.CreditYield, ClipGranted: grant, RequestID: input.RequestID,
			Status: "有效", StatusEn: "Active", StatusKo: "유효", CreatedAt: createdAt,
		}
		if err := distributeCLIP(state, input.RequestID, asset.ID, grant, now); err != nil {
			return err
		}
		asset.RedemptionStatus, asset.Transferable, asset.ExchangeAvailable = "redeemed", false, false
		asset.Status, asset.StatusEn, asset.StatusKo = "已核销", "Redeemed", "상각 완료"
		state.Redemptions = prepend(redemption, state.Redemptions)
		if validWalletAddress(state.Profile.Wallet) {
			queueRedemptionAirdrop(state, redemption, state.Profile.Wallet, state.Profile.WalletChainID, now)
		}
		state.CLIPTransactions = prepend(domain.CLIPTransaction{
			ID: uniqueID("clip-grant", now), TypeCode: "redemptionGrant", Type: "HAPW 核销领取", TypeEn: "HAPW redemption grant", TypeKo: "HAPW 상각 지급", Amount: grant,
			Counterparty: fmt.Sprintf("HAPW %s · %s", asset.TokenID, asset.Name), CounterpartyEn: fmt.Sprintf("HAPW %s · %s", asset.TokenID, asset.NameEn), CounterpartyKo: fmt.Sprintf("HAPW %s · %s", asset.TokenID, asset.NameKo),
			StatusCode: "completed", Status: "已完成", StatusEn: "Completed", StatusKo: "완료", TxHash: "", CreatedAt: createdAt,
		}, state.CLIPTransactions)
		syncGenerationAccount(state)
		result = RedemptionResult{Redemption: redemption, GenerationAccount: state.GenerationAccount, ClipBalance: state.CLIP.Balance}
		return nil
	})
	return result, idempotent, err
}

type GenerationInput struct {
	RequestID string `json:"requestId"`
	AssetID   string `json:"assetId"`
	Duration  int    `json:"duration"`
	Quality   string `json:"quality"`
	Accepted  bool   `json:"accepted"`
}

type GenerationResult struct {
	Generation        domain.Generation        `json:"generation"`
	GenerationAccount domain.GenerationAccount `json:"generationAccount"`
	ClipBalance       int                      `json:"clipBalance"`
}

func (s *Service) Generate(input GenerationInput) (GenerationResult, bool, error) {
	var result GenerationResult
	idempotent := false
	err := s.store.Update(func(state *store.State) error {
		if !validRequestID(input.RequestID) {
			return apiError(http.StatusBadRequest, "invalid_request_id", "A valid operation id is required")
		}
		for _, item := range state.Generations {
			if item.RequestID == input.RequestID {
				if item.AssetID != strings.TrimSpace(input.AssetID) || item.Duration != input.Duration || item.Quality != strings.TrimSpace(input.Quality) || !input.Accepted {
					return apiError(http.StatusConflict, "request_id_reused", "The requestId is already used for another generation")
				}
				syncGenerationAccount(state)
				result = GenerationResult{Generation: item, GenerationAccount: state.GenerationAccount, ClipBalance: state.CLIP.Balance}
				idempotent = true
				return nil
			}
		}
		asset, ok := findAsset(state.Assets, input.AssetID)
		redemptionIndex := -1
		for index := range state.Redemptions {
			if state.Redemptions[index].AssetID == input.AssetID && state.Redemptions[index].Status == "有效" {
				redemptionIndex = index
				break
			}
		}
		if !ok || redemptionIndex < 0 {
			return apiError(http.StatusBadRequest, "license_required", "Redeem the matching HAPW before using this material")
		}
		cost, costErr := GenerationCost(input.Duration, input.Quality)
		if costErr != nil || (input.Duration != 15 && input.Duration != 30 && input.Duration != 60) || !input.Accepted {
			return apiError(http.StatusBadRequest, "invalid_generation", "Select duration and quality, then confirm the material license")
		}
		clipCost, _ := GenerationCLIPCost(cost)
		syncGenerationAccount(state)
		if state.GenerationAccount.Balance < cost || state.Redemptions[redemptionIndex].CreditsRemaining < cost {
			return apiError(http.StatusBadRequest, "insufficient_generation_credits", "Insufficient creation credits for this HAPW license")
		}
		if state.CLIP.Balance-state.CLIP.Reserved < clipCost {
			return apiError(http.StatusBadRequest, "insufficient_clip", "Insufficient CLIP for this generation")
		}
		now := s.now()
		createdAt := formatMinute(now)
		jobNumber := len(state.Generations) + 1
		generation := domain.Generation{
			ID: uniqueID("video", now), AssetID: asset.ID, Title: fmt.Sprintf("%s · AI 试片 %d", asset.Name, jobNumber), TitleEn: fmt.Sprintf("%s · AI Cut %d", asset.NameEn, jobNumber), TitleKo: fmt.Sprintf("%s · AI 테스트 %d", asset.NameKo, jobNumber),
			Duration: input.Duration, Quality: input.Quality, CreditsUsed: cost, ClipCost: clipCost, ValidViews: 12400 + len(state.Generations)*3100,
			RequestID: input.RequestID, Status: "已生成", StatusEn: "Generated", StatusKo: "생성 완료", CreatedAt: createdAt,
		}
		state.Redemptions[redemptionIndex].CreditsRemaining -= cost
		state.CLIP.Balance -= clipCost
		reclaimCLIP(state, clipCost)
		state.Generations = prepend(generation, state.Generations)
		state.CLIPTransactions = prepend(domain.CLIPTransaction{
			ID: uniqueID("clip-generation", now), TypeCode: "generationFee", Type: "AI 视频生成费", TypeEn: "AI video generation fee", TypeKo: "AI 영상 생성 수수료", Amount: -clipCost,
			Counterparty: generation.Title, CounterpartyEn: generation.TitleEn, CounterpartyKo: generation.TitleKo,
			StatusCode: "completed", Status: "已完成", StatusEn: "Completed", StatusKo: "완료", TxHash: "", CreatedAt: createdAt,
		}, state.CLIPTransactions)
		syncGenerationAccount(state)
		result = GenerationResult{Generation: generation, GenerationAccount: state.GenerationAccount, ClipBalance: state.CLIP.Balance}
		return nil
	})
	return result, idempotent, err
}

type ConversionInput struct {
	RequestID string `json:"requestId"`
	AssetID   string `json:"assetId"`
	Region    string `json:"region"`
	Days      int    `json:"days"`
	Accepted  bool   `json:"accepted"`
}

func (s *Service) Convert(input ConversionInput) (domain.Conversion, bool, error) {
	var result domain.Conversion
	idempotent := false
	err := s.store.Update(func(state *store.State) error {
		if !validRequestID(input.RequestID) {
			return apiError(http.StatusBadRequest, "invalid_request_id", "A valid operation id is required")
		}
		for _, item := range state.Conversions {
			if item.RequestID == input.RequestID {
				if item.AssetID != strings.TrimSpace(input.AssetID) || item.Region != strings.TrimSpace(input.Region) || item.Days != input.Days || !input.Accepted {
					return apiError(http.StatusConflict, "request_id_reused", "The requestId is already used for another conversion")
				}
				result, idempotent = item, true
				return nil
			}
		}
		index := assetIndex(state.Assets, input.AssetID)
		if index < 0 || !state.Assets[index].Transferable {
			return apiError(http.StatusBadRequest, "asset_unavailable", "Asset is not available for conversion")
		}
		if (input.Days != 30 && input.Days != 90 && input.Days != 180) || strings.TrimSpace(input.Region) == "" || !input.Accepted {
			return apiError(http.StatusBadRequest, "invalid_conversion", "Please select a region, duration and accept the terms")
		}
		if state.CLIP.Balance-state.CLIP.Reserved < 18 {
			return apiError(http.StatusBadRequest, "insufficient_clip", "Insufficient CLIP balance")
		}
		asset := &state.Assets[index]
		now := s.now()
		result = domain.Conversion{ID: uniqueID("conversion", now), RequestID: input.RequestID, AssetID: asset.ID, Asset: asset.Name + " " + asset.TokenID, Region: input.Region, Days: input.Days, Fee: 18, StatusCode: "submitted", Status: "已提交", StatusEn: "Submitted", StatusKo: "제출됨", CreatedAt: now.UTC().Format("2006-01-02")}
		state.Conversions = prepend(result, state.Conversions)
		state.CLIP.Balance -= 18
		reclaimCLIP(state, 18)
		state.CLIPTransactions = prepend(domain.CLIPTransaction{ID: uniqueID("clip-license", now), TypeCode: "license", Type: "授权手续费", TypeEn: "License fee", Amount: -18, Counterparty: fmt.Sprintf("HAPW %s · %s", asset.TokenID, asset.Name), CounterpartyEn: fmt.Sprintf("HAPW %s · %s", asset.TokenID, asset.NameEn), StatusCode: "completed", Status: "已完成", StatusEn: "Completed", TxHash: "", CreatedAt: formatMinute(now)}, state.CLIPTransactions)
		asset.Status, asset.StatusEn = "授权处理中", "License processing"
		return nil
	})
	return result, idempotent, err
}

type ExerciseInput struct {
	RequestID    string `json:"requestId"`
	AssetID      string `json:"assetId"`
	PlatformCode string `json:"platformCode"`
	Accepted     bool   `json:"accepted"`
}

func (s *Service) Exercise(input ExerciseInput) (domain.AssetExercise, bool, error) {
	var result domain.AssetExercise
	idempotent := false
	err := s.store.Update(func(state *store.State) error {
		if !validRequestID(input.RequestID) {
			return apiError(http.StatusBadRequest, "invalid_request_id", "A valid operation id is required")
		}
		for _, item := range state.Exercises {
			if item.RequestID == input.RequestID {
				if item.AssetID != strings.TrimSpace(input.AssetID) || item.PlatformCode != strings.TrimSpace(input.PlatformCode) || !input.Accepted {
					return apiError(http.StatusConflict, "request_id_reused", "The requestId is already used for another exercise")
				}
				result, idempotent = item, true
				return nil
			}
		}
		asset, ok := findAsset(state.Assets, input.AssetID)
		if !ok || !asset.Transferable {
			return apiError(http.StatusBadRequest, "asset_unavailable", "Asset is not available for exercise")
		}
		var platform domain.ExternalPlatform
		for _, item := range state.Platforms {
			if item.Code == input.PlatformCode {
				platform = item
				break
			}
		}
		if platform.Code == "" || !input.Accepted {
			return apiError(http.StatusBadRequest, "invalid_exercise", "Select a third-party platform and accept the exercise confirmation")
		}
		now := s.now()
		result = domain.AssetExercise{ID: uniqueID("exercise", now), RequestID: input.RequestID, AssetID: asset.ID, PlatformCode: platform.Code, Direction: "Clipli → " + platform.Name, DirectionEn: "Clipli → " + platform.Name, DirectionKo: "Clipli → " + platform.Name, Value: asset.Value, StatusCode: "awaitingSignature", Status: "待签名", StatusEn: "Awaiting signature", StatusKo: "서명 대기", CreatedAt: now.UTC().Format("2006-01-02")}
		state.Exercises = prepend(result, state.Exercises)
		return nil
	})
	return result, idempotent, err
}

type ExchangeInput struct {
	RequestID string `json:"requestId"`
	AssetID   string `json:"assetId"`
	Accepted  bool   `json:"accepted"`
}

type ExchangeResult struct {
	Exchange    domain.HAPWExchange `json:"exchange"`
	Asset       domain.HAPWAsset    `json:"asset,omitempty"`
	ClipBalance int                 `json:"clipBalance"`
}

func (s *Service) Exchange(input ExchangeInput) (ExchangeResult, bool, error) {
	var result ExchangeResult
	idempotent := false
	err := s.store.Update(func(state *store.State) error {
		if !validRequestID(input.RequestID) {
			return apiError(http.StatusBadRequest, "invalid_request_id", "A valid operation id is required")
		}
		for _, item := range state.HAPWExchanges {
			if item.RequestID == input.RequestID {
				if item.AssetID != strings.TrimSpace(input.AssetID) || !input.Accepted {
					return apiError(http.StatusConflict, "request_id_reused", "The requestId is already used for another exchange")
				}
				result, idempotent = ExchangeResult{Exchange: item, ClipBalance: state.CLIP.Balance}, true
				return nil
			}
		}
		index := assetIndex(state.Assets, input.AssetID)
		if index < 0 || !state.Assets[index].ExchangeAvailable || state.Assets[index].RedemptionStatus != "available" || state.Assets[index].External.SyncStatus != "migrated" || state.Assets[index].External.TransferID == "" || state.Assets[index].Owner != "Clipli" {
			return apiError(http.StatusBadRequest, "hapw_reserve_unavailable", "This reserve HAPW is unavailable")
		}
		if !input.Accepted {
			return apiError(http.StatusBadRequest, "exchange_not_confirmed", "Confirm the CLIP to HAPW exchange")
		}
		policy := BuildExchangePolicy(state.HAPWExchanges, state.Assets, s.now())
		if policy.Reached {
			return apiError(http.StatusTooManyRequests, "hapw_exchange_daily_limit", "The daily HAPW exchange limit has been reached")
		}
		asset := &state.Assets[index]
		quote, quoteErr := QuoteHAPWExchange(asset.ClipPrice, state.CLIP.HAPWExchangeFeeRate)
		if quoteErr != nil {
			return quoteErr
		}
		if state.CLIP.Balance-state.CLIP.Reserved < quote.Total {
			return apiError(http.StatusBadRequest, "insufficient_clip", "Insufficient CLIP balance")
		}
		now := s.now()
		exchange := domain.HAPWExchange{ID: uniqueID("hapw-exchange", now), RequestID: input.RequestID, AssetID: asset.ID, Price: quote.Price, Fee: quote.Fee, Total: quote.Total, FeeRate: quote.FeeRate, StatusCode: "completed", Status: "已完成", StatusEn: "Completed", StatusKo: "완료", CreatedAt: formatMinute(now)}
		asset.ExchangeAvailable = false
		asset.Owner = migrationOwner(state.Session.UserRef, state.Profile.Wallet)
		state.CLIP.Balance -= quote.Total
		reclaimCLIP(state, quote.Total)
		state.HAPWExchanges = prepend(exchange, state.HAPWExchanges)
		state.CLIPTransactions = prepend(domain.CLIPTransaction{ID: uniqueID("clip-hapw", now), TypeCode: "hapwExchange", Type: "HAPW 储备兑换", TypeEn: "HAPW reserve exchange", TypeKo: "HAPW 준비금 교환", Amount: -quote.Total, Counterparty: fmt.Sprintf("HAPW %s · 本金 %d + 手续费 %d", asset.TokenID, quote.Price, quote.Fee), CounterpartyEn: fmt.Sprintf("HAPW %s · price %d + fee %d", asset.TokenID, quote.Price, quote.Fee), StatusCode: "completed", Status: "已完成", StatusEn: "Completed", StatusKo: "완료", TxHash: "", CreatedAt: formatMinute(now)}, state.CLIPTransactions)
		result = ExchangeResult{Exchange: exchange, Asset: *asset, ClipBalance: state.CLIP.Balance}
		return nil
	})
	return result, idempotent, err
}

func (s *Service) ConnectWallet(input WalletConnectInput) (domain.Profile, error) {
	allowed := map[string]bool{"MetaMask": true, "Coinbase Wallet": true, "WalletConnect": true, "Venly": true}
	if !allowed[input.Provider] {
		return domain.Profile{}, apiError(http.StatusBadRequest, "invalid_provider", "Unsupported wallet provider")
	}
	address := normalizeWalletAddress(input.Address)
	if address == "" {
		var profile domain.Profile
		err := s.store.Update(func(state *store.State) error {
			state.Profile.WalletProvider = input.Provider
			state.Profile.WalletStatus = "provider-selected"
			profile = state.Profile
			return nil
		})
		return profile, err
	}
	if !validWalletAddress(address) {
		return domain.Profile{}, apiError(http.StatusBadRequest, "invalid_wallet_address", "A valid EVM wallet address is required")
	}
	chainID := NormalizeBNBChain(input.ChainID)
	if chainID == "" {
		return domain.Profile{}, apiError(http.StatusBadRequest, "unsupported_chain", "Clipli wallet operations currently support BNB Smart Chain Testnet (0x61) and Mainnet (0x38)")
	}
	var profile domain.Profile
	err := s.store.Update(func(state *store.State) error {
		now := s.now().UTC().Format(time.RFC3339)
		state.Profile.WalletProvider, state.Profile.Wallet = input.Provider, address
		state.Profile.WalletChainID, state.Profile.WalletStatus, state.Profile.WalletConnectedAt = chainID, "connected", now
		profile = state.Profile
		return nil
	})
	return profile, err
}

func (s *Service) DisconnectWallet() domain.Profile {
	var profile domain.Profile
	_ = s.store.Update(func(state *store.State) error {
		state.Profile.Wallet, state.Profile.WalletProvider, state.Profile.WalletChainID = "", "", ""
		state.Profile.WalletStatus, state.Profile.WalletConnectedAt = "disconnected", ""
		profile = state.Profile
		return nil
	})
	return profile
}

func (s *Service) WalletAssets(address string) ([]domain.WalletAsset, error) {
	address = normalizeWalletAddress(address)
	if !validWalletAddress(address) {
		return nil, apiError(http.StatusBadRequest, "invalid_wallet_address", "A valid EVM wallet address is required")
	}
	state := s.PortfolioSnapshot()
	items := make([]domain.WalletAsset, 0)
	seen := map[string]bool{}
	for _, item := range state.WalletAssets {
		if strings.EqualFold(item.WalletAddress, address) {
			items = append(items, item)
			seen[item.AssetID] = true
		}
	}
	for _, asset := range state.Assets {
		if !strings.EqualFold(asset.Owner, address) || seen[asset.ID] {
			continue
		}
		items = append(items, walletAssetFromHAPW(asset, address))
	}
	return items, nil
}

func (s *Service) AirdropRules() []domain.AirdropRule {
	return s.Snapshot().AirdropRules
}

func (s *Service) Airdrops(address, status string) ([]domain.AirdropRecord, error) {
	if address != "" {
		address = normalizeWalletAddress(address)
		if !validWalletAddress(address) {
			return nil, apiError(http.StatusBadRequest, "invalid_wallet_address", "A valid EVM wallet address is required")
		}
	}
	state := s.Snapshot()
	items := make([]domain.AirdropRecord, 0)
	for _, item := range state.Airdrops {
		if address != "" && !strings.EqualFold(item.WalletAddress, address) {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) AirdropEligibility(ruleCode, address, assetID string) (AirdropEligibility, error) {
	address = normalizeWalletAddress(address)
	if !validWalletAddress(address) {
		return AirdropEligibility{}, apiError(http.StatusBadRequest, "invalid_wallet_address", "A valid EVM wallet address is required")
	}
	state := s.Snapshot()
	rule, ok := findAirdropRule(state.AirdropRules, ruleCode)
	if !ok || !rule.Enabled {
		return AirdropEligibility{RuleCode: ruleCode, WalletAddress: address, Token: "CLIP", Reason: "airdrop_rule_disabled"}, nil
	}
	eligibility := AirdropEligibility{RuleCode: rule.Code, WalletAddress: address, AssetID: assetID, Token: rule.Token, Reason: "eligible"}
	if rule.Code == "hapw_redemption" {
		for _, redemption := range state.Redemptions {
			if redemption.AssetID == assetID {
				eligibility.Amount = redemption.ClipGranted
				for _, existing := range state.Airdrops {
					if existing.RuleCode == rule.Code && existing.AssetID == assetID && existing.Status != "failed" && existing.Status != "cancelled" {
						eligibility.ExistingID = existing.ID
						eligibility.Reason = "airdrop_already_created"
						return eligibility, nil
					}
				}
				if !redemptionWalletMatches(state, assetID, address) {
					eligibility.Reason = "airdrop_wallet_mismatch"
					return eligibility, nil
				}
				if state.CLIP.Balance-state.CLIP.Reserved < eligibility.Amount {
					eligibility.Reason = "insufficient_clip"
					return eligibility, nil
				}
				eligibility.Eligible = true
				return eligibility, nil
			}
		}
		eligibility.Reason = "hapw_redemption_required"
		return eligibility, nil
	}
	eligibility.Reason = "rule_requires_manual_amount"
	return eligibility, nil
}

func (s *Service) CreateAirdrop(input AirdropInput) (domain.AirdropRecord, bool, error) {
	var result domain.AirdropRecord
	idempotent := false
	address := normalizeWalletAddress(input.WalletAddress)
	if !validWalletAddress(address) {
		return result, false, apiError(http.StatusBadRequest, "invalid_wallet_address", "A valid EVM wallet address is required")
	}
	if !validRequestID(input.RequestID) {
		return result, false, apiError(http.StatusBadRequest, "invalid_request_id", "A valid operation id is required")
	}
	chainID := NormalizeBNBChain(input.ChainID)
	if chainID == "" {
		return result, false, apiError(http.StatusBadRequest, "unsupported_chain", "Airdrops are currently limited to BNB Smart Chain Testnet (0x61) and Mainnet (0x38)")
	}
	err := s.store.Update(func(state *store.State) error {
		rule, ok := findAirdropRule(state.AirdropRules, input.RuleCode)
		if !ok || !rule.Enabled {
			return apiError(http.StatusBadRequest, "airdrop_rule_disabled", "The airdrop rule is unavailable")
		}
		if input.Token != "" && !strings.EqualFold(strings.TrimSpace(input.Token), rule.Token) {
			return apiError(http.StatusBadRequest, "invalid_airdrop_token", "The airdrop token must match the configured rule")
		}
		amount := input.Amount
		allocationSource := "treasury-reservation"
		if rule.Code == "hapw_redemption" {
			if !redemptionWalletMatches(*state, input.AssetID, address) {
				return apiError(http.StatusConflict, "airdrop_wallet_mismatch", "The redemption does not belong to this wallet")
			}
			amount = 0
			allocationSource = "redemption-entitlement"
			amount = 0
			for _, redemption := range state.Redemptions {
				if redemption.AssetID == input.AssetID {
					amount = redemption.ClipGranted
					break
				}
			}
			if amount <= 0 {
				return apiError(http.StatusBadRequest, "hapw_redemption_required", "The wallet is not eligible until the HAPW is redeemed")
			}
		}
		if amount <= 0 || amount > 1000000000 {
			return apiError(http.StatusBadRequest, "invalid_airdrop_amount", "A positive airdrop amount is required")
		}
		token := rule.Token
		for _, item := range state.Airdrops {
			if item.RequestID == input.RequestID {
				if item.Status == "failed" || item.Status == "cancelled" {
					return apiError(http.StatusConflict, "request_id_reused", "A failed airdrop must be recreated with a new requestId")
				}
				if item.RuleCode != rule.Code || !strings.EqualFold(item.WalletAddress, address) || item.ChainID != chainID || item.AssetID != strings.TrimSpace(input.AssetID) || item.Amount != amount || !strings.EqualFold(item.Token, token) {
					return apiError(http.StatusConflict, "request_id_reused", "The requestId is already used for another airdrop")
				}
				result, idempotent = item, true
				return nil
			}
			if rule.Code == "hapw_redemption" && item.RuleCode == rule.Code && item.AssetID == strings.TrimSpace(input.AssetID) && item.Status != "failed" && item.Status != "cancelled" {
				return apiError(http.StatusConflict, "airdrop_already_created", "An active airdrop already exists for this wallet and redemption")
			}
		}
		if allocationSource == "treasury-reservation" {
			if state.CLIPTreasury.TreasuryBalance < amount {
				return apiError(http.StatusServiceUnavailable, "clip_treasury_insufficient", "The platform treasury cannot reserve this airdrop")
			}
			state.CLIPTreasury.TreasuryBalance -= amount
			state.CLIPTreasury.LedgerOutstanding += amount
			state.CLIPTreasury.TotalDistributed += amount
		} else {
			if state.CLIP.Balance-state.CLIP.Reserved < amount {
				return apiError(http.StatusConflict, "insufficient_clip", "The CLIP grant has already been spent or reserved")
			}
			state.CLIP.Reserved += amount
		}
		now := s.now().UTC().Format(time.RFC3339)
		result = domain.AirdropRecord{ID: uniqueID("airdrop", s.now()), RequestID: input.RequestID, RuleCode: rule.Code, WalletAddress: address, ChainID: chainID, AssetID: strings.TrimSpace(input.AssetID), Token: token, Amount: amount, Status: "queued", Eligibility: "eligible", ExecutionMode: "external-executor", AllocationSource: allocationSource, CreatedAt: now, UpdatedAt: now}
		state.Airdrops = prepend(result, state.Airdrops)
		return nil
	})
	return result, idempotent, err
}

// SimulateAirdrop exercises the queued/confirmed state transition without
// broadcasting a blockchain transaction. The receipt is deliberately not a
// 0x hash so it cannot be mistaken for an on-chain transaction.
func (s *Service) SimulateAirdrop(id, chainID string) (domain.AirdropRecord, error) {
	var result domain.AirdropRecord
	chainID = NormalizeBNBChain(chainID)
	if chainID == "" {
		return result, apiError(http.StatusBadRequest, "unsupported_chain", "Choose BNB Smart Chain Testnet (0x61) or Mainnet (0x38)")
	}
	err := s.store.Update(func(state *store.State) error {
		index := -1
		for i := range state.Airdrops {
			if state.Airdrops[i].ID == id {
				index = i
				break
			}
		}
		if index < 0 {
			return apiError(http.StatusNotFound, "airdrop_not_found", "Airdrop record not found")
		}
		item := &state.Airdrops[index]
		if item.Status == "failed" || item.Status == "cancelled" {
			return apiError(http.StatusConflict, "airdrop_failed", "A failed airdrop must be recreated before simulation")
		}
		if item.Status == "confirmed" && item.Simulated {
			result = *item
			return nil
		}
		if item.Status == "confirmed" && !item.Simulated {
			return apiError(http.StatusConflict, "airdrop_already_confirmed", "A confirmed external transaction cannot be simulated")
		}
		now := s.now().UTC().Format(time.RFC3339)
		item.Status = "confirmed"
		item.ChainID = chainID
		item.Network = chainID
		item.Simulated = true
		item.ExecutionMode = "simulated-bnb"
		item.ExecutorRef = "clipli-simulator"
		item.TxHash = "sim-bnb-" + item.ID
		item.FailureReason = ""
		item.UpdatedAt = now
		result = *item
		return nil
	})
	return result, err
}

func (s *Service) UpdateAirdrop(id string, input AirdropResultInput) (domain.AirdropRecord, error) {
	var result domain.AirdropRecord
	var checked domain.AirdropRecord
	for _, item := range s.Snapshot().Airdrops {
		if item.ID != id {
			continue
		}
		checked = item
		if item.Status == "failed" || item.Status == "cancelled" {
			return result, apiError(http.StatusConflict, "airdrop_failed_terminal", "A terminal airdrop cannot be executed")
		}
		if item.Status == "confirmed" {
			if input.Status == "confirmed" && item.TxHash == input.TxHash {
				return item, nil
			}
			return result, apiError(http.StatusConflict, "airdrop_receipt_conflict", "A confirmed airdrop cannot be changed")
		}
		if input.Status == "confirmed" || input.Status == "submitted" || (input.Status == "failed" && item.TxHash != "") {
			if !validTransactionHash(input.TxHash) {
				return result, apiError(http.StatusBadRequest, "invalid_transaction_hash", "A valid BNB transaction hash is required")
			}
			if s.airdropVerifier == nil {
				return result, apiError(http.StatusServiceUnavailable, "airdrop_executor_not_configured", "Configure the CLIP contract and BNB receipt verifier first")
			}
			if err := s.airdropVerifier(item, input); err != nil {
				return result, err
			}
		}
		break
	}
	err := s.store.Update(func(state *store.State) error {
		index := -1
		for i := range state.Airdrops {
			if state.Airdrops[i].ID == id {
				index = i
				break
			}
		}
		if index < 0 {
			return apiError(http.StatusNotFound, "airdrop_not_found", "Airdrop record not found")
		}
		if input.Status != "submitted" && input.Status != "confirmed" && input.Status != "failed" {
			return apiError(http.StatusBadRequest, "invalid_airdrop_status", "Airdrop status must be submitted, confirmed, or failed")
		}
		item := &state.Airdrops[index]
		if item.Status != checked.Status || item.TxHash != checked.TxHash {
			return apiError(http.StatusConflict, "airdrop_state_changed", "The airdrop changed during verification; retry with the same receipt")
		}
		if item.Status == "confirmed" {
			result = *item
			return nil
		}
		if item.Status == "failed" || item.Status == "cancelled" {
			return apiError(http.StatusConflict, "airdrop_failed_terminal", "A failed airdrop must be recreated with a new requestId")
		}
		if (input.Status == "submitted" || input.Status == "confirmed") && !validTransactionHash(input.TxHash) {
			return apiError(http.StatusBadRequest, "invalid_transaction_hash", "A valid BNB transaction hash is required")
		}
		if item.TxHash != "" && item.TxHash != input.TxHash {
			return apiError(http.StatusConflict, "airdrop_receipt_conflict", "A submitted transaction hash cannot be replaced")
		}
		for _, other := range state.Airdrops {
			if other.ID != item.ID && input.TxHash != "" && strings.EqualFold(other.TxHash, input.TxHash) {
				return apiError(http.StatusConflict, "airdrop_receipt_reused", "The transaction already belongs to another airdrop")
			}
		}
		if input.Status == "failed" && strings.TrimSpace(input.FailureReason) == "" {
			return apiError(http.StatusBadRequest, "failure_reason_required", "A failure reason is required")
		}
		if input.Status == "failed" && item.Status != "failed" && item.AllocationSource == "treasury-reservation" {
			state.CLIPTreasury.TreasuryBalance += item.Amount
			state.CLIPTreasury.LedgerOutstanding -= item.Amount
			if state.CLIPTreasury.LedgerOutstanding < 0 {
				state.CLIPTreasury.LedgerOutstanding = 0
			}
			state.CLIPTreasury.TotalDistributed -= item.Amount
		}
		if item.AllocationSource == "redemption-entitlement" && (input.Status == "failed" || input.Status == "confirmed") {
			state.CLIP.Reserved -= item.Amount
			if input.Status == "confirmed" {
				state.CLIP.Balance -= item.Amount
			}
		}
		if input.Status == "confirmed" {
			state.CLIPTreasury.LedgerOutstanding -= item.Amount
			state.CLIPTreasury.OnchainDistributed += item.Amount
		}
		if input.Status == "confirmed" && s.bnbExecution != nil {
			state.CLIPTreasury.TreasuryBalance = 0
		}
		item.Status, item.TxHash, item.ExecutorRef, item.FailureReason = input.Status, input.TxHash, input.ExecutorRef, input.FailureReason
		item.UpdatedAt = s.now().UTC().Format(time.RFC3339)
		result = *item
		return nil
	})
	if err == nil && input.Status == "confirmed" && s.bnbExecution != nil {
		err = s.SyncBNBTreasury(s.bnbExecution)
	}
	return result, err
}

func queueRedemptionAirdrop(state *store.State, redemption domain.HAPWRedemption, address, chainID string, now time.Time) {
	for _, item := range state.Airdrops {
		if item.RuleCode == "hapw_redemption" && item.AssetID == redemption.AssetID && item.Status != "failed" && item.Status != "cancelled" {
			return
		}
	}
	if NormalizeBNBChain(chainID) == "" || state.CLIP.Balance-state.CLIP.Reserved < redemption.ClipGranted {
		return
	}
	state.CLIP.Reserved += redemption.ClipGranted
	createdAt := now.UTC().Format(time.RFC3339)
	state.Airdrops = prepend(domain.AirdropRecord{ID: uniqueID("airdrop", now), RequestID: "airdrop-" + redemption.RequestID, RuleCode: "hapw_redemption", WalletAddress: address, ChainID: chainID, AssetID: redemption.AssetID, Token: "CLIP", Amount: redemption.ClipGranted, Status: "queued", Eligibility: "eligible", ExecutionMode: "external-executor", AllocationSource: "redemption-entitlement", CreatedAt: createdAt, UpdatedAt: createdAt}, state.Airdrops)
}

func redemptionWalletMatches(state store.State, assetID, address string) bool {
	asset, ok := findAsset(state.Assets, assetID)
	if !ok {
		return false
	}
	if validWalletAddress(asset.Owner) {
		return strings.EqualFold(asset.Owner, address)
	}
	return (asset.Owner == state.Session.UserRef || asset.Owner == "Clipli") && strings.EqualFold(state.Profile.Wallet, address)
}

func walletAssetFromHAPW(asset domain.HAPWAsset, address string) domain.WalletAsset {
	return domain.WalletAsset{ID: "wallet-asset-" + asset.ID, WalletAddress: address, AssetID: asset.ID, TokenID: asset.TokenID, Name: asset.Name, Balance: "1", Standard: asset.Provenance.TokenStandard, SourceCode: asset.External.ProviderCode, ExternalAssetID: asset.External.ProviderAssetID, ExternalURL: asset.External.AssetURL, SyncStatus: asset.External.SyncStatus, LastSyncedAt: asset.External.LastSyncedAt, RedemptionStatus: asset.RedemptionStatus, CanRedeem: asset.RedemptionStatus == "available", CanExercise: asset.Transferable}
}

func findAirdropRule(rules []domain.AirdropRule, code string) (domain.AirdropRule, bool) {
	for _, rule := range rules {
		if rule.Code == code {
			return rule, true
		}
	}
	return domain.AirdropRule{}, false
}

func normalizeWalletAddress(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func validWalletAddress(value string) bool {
	if len(value) != 42 || !strings.HasPrefix(value, "0x") {
		return false
	}
	for _, char := range value[2:] {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

func validTransactionHash(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) != 66 || !strings.HasPrefix(value, "0x") {
		return false
	}
	for _, char := range value[2:] {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

func (s *Service) UpdateSettings(walletSign, expiryReminder *bool) domain.Profile {
	var profile domain.Profile
	_ = s.store.Update(func(state *store.State) error {
		if walletSign != nil {
			state.Profile.Settings.WalletSign = *walletSign
		}
		if expiryReminder != nil {
			state.Profile.Settings.ExpiryReminder = *expiryReminder
		}
		profile = state.Profile
		return nil
	})
	return profile
}

func validRequestID(value string) bool { return len(value) >= 8 && len(value) <= 100 }

func findAsset(assets []domain.HAPWAsset, id string) (domain.HAPWAsset, bool) {
	for _, asset := range assets {
		if asset.ID == id || asset.TokenID == id || strings.TrimPrefix(asset.TokenID, "#") == strings.TrimPrefix(id, "#") {
			return asset, true
		}
	}
	return domain.HAPWAsset{}, false
}

func assetIndex(assets []domain.HAPWAsset, id string) int {
	for index := range assets {
		if assets[index].ID == id || assets[index].TokenID == id || strings.TrimPrefix(assets[index].TokenID, "#") == strings.TrimPrefix(id, "#") {
			return index
		}
	}
	return -1
}

func syncGenerationAccount(state *store.State) {
	account := &state.GenerationAccount
	account.Balance, account.LifetimeGranted, account.LifetimeUsed = 0, 0, 0
	account.LifetimeClipGranted, account.LifetimeClipSpent = 0, 0
	for _, redemption := range state.Redemptions {
		if redemption.Status == "有效" {
			account.Balance += redemption.CreditsRemaining
		}
		account.LifetimeGranted += redemption.CreditsGranted
		account.LifetimeClipGranted += redemption.ClipGranted
	}
	for _, generation := range state.Generations {
		account.LifetimeUsed += generation.CreditsUsed
		account.LifetimeClipSpent += generation.ClipCost
	}
}

func uniqueID(prefix string, now time.Time) string {
	var suffix [16]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		panic("cannot generate operation identifier: " + err.Error())
	}
	return fmt.Sprintf("%s-%d-%x", prefix, now.UnixNano(), suffix)
}
func formatMinute(value time.Time) string { return value.UTC().Format("2006-01-02 15:04") }

func prepend[T any](value T, values []T) []T {
	result := make([]T, 0, len(values)+1)
	result = append(result, value)
	return append(result, values...)
}

// PortfolioSnapshot excludes provisional mirrors and unconfirmed upstream records.
func (s *Service) PortfolioSnapshot() store.State {
	state := s.Snapshot()
	assets := make([]domain.HAPWAsset, 0, len(state.Assets))
	for _, asset := range state.Assets {
		if asset.External.SyncStatus == "migrated" && asset.External.TransferID != "" {
			assets = append(assets, asset)
		}
	}
	state.Assets = assets
	return state
}

func (s *Service) ExternalPlatformConfigured() bool {
	client, ok := s.platform.(*HTTPExternalPlatform)
	if ok {
		return strings.TrimSpace(client.BaseURL) != ""
	}
	return s.platform != nil
}

// Sandbox partner accounts may be used to test binding, but cannot mint a
// portfolio entry that the UI presents as a real external transfer.
func (s *Service) ExternalPlatformSandbox() bool {
	client, ok := s.platform.(*HTTPExternalPlatform)
	if !ok {
		return false
	}
	parsed, err := url.Parse(client.BaseURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "api-test.hnccc.com" || host == "localhost" || host == "127.0.0.1" || host == "::1" || strings.HasPrefix(host, "sandbox.") || strings.HasPrefix(host, "api-test.") || strings.HasPrefix(host, "test.") || strings.HasSuffix(host, ".test")
}
