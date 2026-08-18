package service

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
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
	store *store.Memory
	now   func() time.Time
}

func New(memory *store.Memory) *Service {
	return &Service{store: memory, now: time.Now}
}

func (s *Service) Snapshot() store.State {
	state := s.store.Snapshot()
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
	for _, asset := range s.Snapshot().Assets {
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
	for _, asset := range s.Snapshot().Assets {
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
	for _, asset := range s.Snapshot().Assets {
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
				syncGenerationAccount(state)
				result = RedemptionResult{Redemption: item, GenerationAccount: state.GenerationAccount, ClipBalance: state.CLIP.Balance}
				idempotent = true
				return nil
			}
		}
		index := assetIndex(state.Assets, input.AssetID)
		if index < 0 || state.Assets[index].RedemptionStatus != "available" {
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
		state.CLIPTransactions = prepend(domain.CLIPTransaction{
			ID: uniqueID("clip-grant", now), TypeCode: "redemptionGrant", Type: "HAPW 核销领取", TypeEn: "HAPW redemption grant", TypeKo: "HAPW 상각 지급", Amount: grant,
			Counterparty: fmt.Sprintf("HAPW %s · %s", asset.TokenID, asset.Name), CounterpartyEn: fmt.Sprintf("HAPW %s · %s", asset.TokenID, asset.NameEn), CounterpartyKo: fmt.Sprintf("HAPW %s · %s", asset.TokenID, asset.NameKo),
			StatusCode: "completed", Status: "已完成", StatusEn: "Completed", StatusKo: "완료", TxHash: "0xgrant…demo", CreatedAt: createdAt,
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
		if state.CLIP.Balance < clipCost {
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
			StatusCode: "completed", Status: "已完成", StatusEn: "Completed", StatusKo: "완료", TxHash: "0xgen…demo", CreatedAt: createdAt,
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
		if state.CLIP.Balance < 18 {
			return apiError(http.StatusBadRequest, "insufficient_clip", "Insufficient CLIP balance")
		}
		asset := &state.Assets[index]
		now := s.now()
		result = domain.Conversion{ID: uniqueID("conversion", now), RequestID: input.RequestID, AssetID: asset.ID, Asset: asset.Name + " " + asset.TokenID, Region: input.Region, Days: input.Days, Fee: 18, StatusCode: "submitted", Status: "已提交", StatusEn: "Submitted", StatusKo: "제출됨", CreatedAt: now.UTC().Format("2006-01-02")}
		state.Conversions = prepend(result, state.Conversions)
		state.CLIP.Balance -= 18
		reclaimCLIP(state, 18)
		state.CLIPTransactions = prepend(domain.CLIPTransaction{ID: uniqueID("clip-license", now), TypeCode: "license", Type: "授权手续费", TypeEn: "License fee", Amount: -18, Counterparty: fmt.Sprintf("HAPW %s · %s", asset.TokenID, asset.Name), CounterpartyEn: fmt.Sprintf("HAPW %s · %s", asset.TokenID, asset.NameEn), StatusCode: "completed", Status: "已完成", StatusEn: "Completed", TxHash: "0xdemo…clipli", CreatedAt: formatMinute(now)}, state.CLIPTransactions)
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
				result, idempotent = ExchangeResult{Exchange: item, ClipBalance: state.CLIP.Balance}, true
				return nil
			}
		}
		index := assetIndex(state.Assets, input.AssetID)
		if index < 0 || !state.Assets[index].ExchangeAvailable || state.Assets[index].RedemptionStatus != "available" {
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
		quote, _ := QuoteHAPWExchange(asset.ClipPrice, state.CLIP.HAPWExchangeFeeRate)
		if state.CLIP.Balance < quote.Total {
			return apiError(http.StatusBadRequest, "insufficient_clip", "Insufficient CLIP balance")
		}
		now := s.now()
		exchange := domain.HAPWExchange{ID: uniqueID("hapw-exchange", now), RequestID: input.RequestID, AssetID: asset.ID, Price: quote.Price, Fee: quote.Fee, Total: quote.Total, FeeRate: quote.FeeRate, StatusCode: "completed", Status: "已完成", StatusEn: "Completed", StatusKo: "완료", CreatedAt: formatMinute(now)}
		asset.ExchangeAvailable = false
		state.CLIP.Balance -= quote.Total
		reclaimCLIP(state, quote.Total)
		state.HAPWExchanges = prepend(exchange, state.HAPWExchanges)
		state.CLIPTransactions = prepend(domain.CLIPTransaction{ID: uniqueID("clip-hapw", now), TypeCode: "hapwExchange", Type: "HAPW 储备兑换", TypeEn: "HAPW reserve exchange", TypeKo: "HAPW 준비금 교환", Amount: -quote.Total, Counterparty: fmt.Sprintf("HAPW %s · 本金 %d + 手续费 %d", asset.TokenID, quote.Price, quote.Fee), CounterpartyEn: fmt.Sprintf("HAPW %s · price %d + fee %d", asset.TokenID, quote.Price, quote.Fee), StatusCode: "completed", Status: "已完成", StatusEn: "Completed", StatusKo: "완료", TxHash: "0xhapw…demo", CreatedAt: formatMinute(now)}, state.CLIPTransactions)
		result = ExchangeResult{Exchange: exchange, Asset: *asset, ClipBalance: state.CLIP.Balance}
		return nil
	})
	return result, idempotent, err
}

func (s *Service) ConnectWallet(provider string) (domain.Profile, error) {
	allowed := map[string]bool{"MetaMask": true, "Coinbase Wallet": true, "WalletConnect": true, "Venly": true}
	if !allowed[provider] {
		return domain.Profile{}, apiError(http.StatusBadRequest, "invalid_provider", "Unsupported wallet provider")
	}
	var profile domain.Profile
	err := s.store.Update(func(state *store.State) error {
		state.Profile.WalletProvider, state.Profile.Wallet = provider, "0x7E…4A91"
		profile = state.Profile
		return nil
	})
	return profile, err
}

func (s *Service) SetOverseasAccount(account string) (domain.Profile, error) {
	account = strings.TrimSpace(account)
	if len(account) < 4 || len(account) > 64 {
		return domain.Profile{}, apiError(http.StatusBadRequest, "invalid_account", "Enter a valid overseas account")
	}
	for _, char := range account {
		if !(char >= 'A' && char <= 'Z') && !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9') && char != '_' && char != '-' {
			return domain.Profile{}, apiError(http.StatusBadRequest, "invalid_account", "Enter a valid overseas account")
		}
	}
	var profile domain.Profile
	err := s.store.Update(func(state *store.State) error {
		state.Profile.OverseasAccount = account
		profile = state.Profile
		return nil
	})
	return profile, err
}

func (s *Service) ClearOverseasAccount() domain.Profile {
	var profile domain.Profile
	_ = s.store.Update(func(state *store.State) error {
		state.Profile.OverseasAccount = ""
		profile = state.Profile
		return nil
	})
	return profile
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
	return fmt.Sprintf("%s-%d", prefix, now.UnixNano())
}
func formatMinute(value time.Time) string { return value.UTC().Format("2006-01-02 15:04") }

func prepend[T any](value T, values []T) []T {
	result := make([]T, 0, len(values)+1)
	result = append(result, value)
	return append(result, values...)
}
