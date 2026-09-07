package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/service"
	"github.com/StevenWXY/HAPE/internal/store"
)

type Handler struct {
	service     *service.Service
	publicDir   string
	logger      *slog.Logger
	adminKey    string
	executorKey string
}

func New(serviceLayer *service.Service, publicDir string, logger *slog.Logger) http.Handler {
	handler := &Handler{service: serviceLayer, publicDir: publicDir, logger: logger, adminKey: os.Getenv("CLIPLI_ADMIN_API_KEY"), executorKey: os.Getenv("CLIPLI_AIRDROP_EXECUTOR_KEY")}
	mux := http.NewServeMux()

	// Operational endpoint.
	mux.HandleFunc("GET /api/health", handler.health)
	mux.HandleFunc("GET /admin", handler.adminPage)
	mux.HandleFunc("GET /admin/", handler.adminPage)

	// Versioned HAPW API.
	mux.HandleFunc("GET /api/v1/session", handler.session)
	mux.HandleFunc("POST /api/v1/wallet/connect", handler.connectWallet)
	mux.HandleFunc("DELETE /api/v1/wallet/connect", handler.disconnectWallet)
	mux.HandleFunc("GET /api/v1/wallet/assets", handler.walletAssets)
	mux.HandleFunc("GET /api/v1/wallet/airdrop-eligibility", handler.walletAirdropEligibility)
	mux.HandleFunc("GET /api/v1/airdrops", handler.airdrops)
	mux.HandleFunc("GET /api/v1/airdrop-rules", handler.airdropRules)
	mux.HandleFunc("GET /api/v1/hapw/summary", handler.hapwSummary)
	mux.HandleFunc("GET /api/v1/hapw/assets", handler.hapwAssets)
	mux.HandleFunc("GET /api/v1/hapw/assets/{assetID}", handler.hapwAsset)
	mux.HandleFunc("GET /api/v1/hapw/assets/{assetID}/authorization", handler.hapwAuthorization)
	mux.HandleFunc("GET /api/v1/hapw/assets/{assetID}/history", handler.hapwHistory)
	mux.HandleFunc("GET /api/v1/hapw/assets/{assetID}/works", handler.hapwWorks)
	mux.HandleFunc("GET /api/v1/hapw/assets/{assetID}/exchange-quote", handler.hapwExchangeQuote)
	mux.HandleFunc("GET /api/v1/hapw/redemptions", handler.hapwRedemptions)
	mux.HandleFunc("GET /api/v1/hapw/redemptions/{redemptionID}", handler.hapwRedemption)
	mux.HandleFunc("POST /api/v1/hapw/redemptions", handler.createRedemption)
	mux.HandleFunc("GET /api/v1/hapw/exercises", handler.hapwExercises)
	mux.HandleFunc("POST /api/v1/hapw/exercises", handler.createExercise)
	mux.HandleFunc("GET /api/v1/hapw/exchanges", handler.hapwExchanges)
	mux.HandleFunc("POST /api/v1/hapw/exchanges", handler.createExchange)
	mux.HandleFunc("GET /api/v1/hapw/exchange-policy", handler.hapwExchangePolicy)
	mux.HandleFunc("GET /api/v1/platforms", handler.platforms)
	mux.HandleFunc("GET /api/v1/clip/treasury", handler.clipTreasury)
	mux.HandleFunc("GET /api/v1/clip/distribution-rules", handler.clipDistributionRules)
	mux.HandleFunc("GET /api/v1/clip/distributions", handler.clipDistributions)
	mux.HandleFunc("GET /api/v1/integrations/asset-requirements", handler.assetRequirements)
	mux.HandleFunc("GET /api/v1/integrations/asset-sources", handler.assetSources)
	mux.HandleFunc("GET /api/v1/integrations/asset-sources/{sourceCode}", handler.assetSource)
	mux.HandleFunc("GET /api/v1/integrations/asset-sync-runs", handler.assetSyncRuns)

	// External-platform account and asset integration. The platform verifies
	// the phone/code pair and owns the source-of-truth asset operations; these
	// routes only orchestrate and audit the Clipli side of that flow.
	for _, prefix := range []string{"/api/v1/integrations/platform", "/api/v1/platform"} {
		mux.HandleFunc("POST "+prefix+"/verification-codes", handler.sendVerificationCode)
		mux.HandleFunc("POST "+prefix+"/verification-code", handler.sendVerificationCode)
		mux.HandleFunc("GET "+prefix+"/bindings", handler.getExternalBinding)
		mux.HandleFunc("POST "+prefix+"/bindings", handler.bindExternalUser)
		mux.HandleFunc("GET "+prefix+"/users/{userID}/assets", handler.userAssets)
		mux.HandleFunc("GET "+prefix+"/users/{userID}/asset-counts", handler.userAssetCounts)
		mux.HandleFunc("GET "+prefix+"/works", handler.externalWorks)
		mux.HandleFunc("GET "+prefix+"/templates", handler.externalTemplates)
		mux.HandleFunc("GET "+prefix+"/users/{userID}/migrations", handler.externalMigrations)
		mux.HandleFunc("POST "+prefix+"/users/{userID}/migrations/preview", handler.previewExternalMigration)
		mux.HandleFunc("POST "+prefix+"/users/{userID}/migrations", handler.migrateExternalAsset)
		mux.HandleFunc("GET "+prefix+"/users/{userID}/migration-requests", handler.migrationRequests)
		mux.HandleFunc("POST "+prefix+"/users/{userID}/migration-requests", handler.requestMigration)
		mux.HandleFunc("POST "+prefix+"/users/{userID}/assets/{assetID}/redemptions", handler.redeemExternalAsset)
	}
	// Short aliases are useful for partner onboarding and preserve the same
	// handlers/contracts as the namespaced integration routes.
	mux.HandleFunc("GET /api/v1/users/{userID}/assets", handler.userAssets)
	mux.HandleFunc("GET /api/v1/users/{userID}/asset-counts", handler.userAssetCounts)
	mux.HandleFunc("GET /api/v1/users/{userID}/holdings", handler.userAssets)
	mux.HandleFunc("POST /api/v1/users/{userID}/assets/{assetID}/redemptions", handler.redeemExternalAsset)
	mux.HandleFunc("POST /api/v1/users/{userID}/assets/{assetID}/redeem", handler.redeemExternalAsset)
	mux.HandleFunc("POST /api/v1/users/{userID}/verification-codes", handler.sendVerificationCode)
	mux.HandleFunc("POST /api/v1/users/{userID}/bindings", handler.bindExternalUser)
	mux.HandleFunc("POST /api/v1/bind/send-code", handler.sendVerificationCode)
	mux.HandleFunc("POST /api/v1/bind/request-code", handler.sendVerificationCode)
	mux.HandleFunc("POST /api/v1/bind", handler.bindExternalUser)
	mux.HandleFunc("GET /api/v1/bind", handler.getExternalBinding)
	mux.HandleFunc("GET /api/v1/assets/holdings", handler.userAssetsByQuery)
	mux.HandleFunc("POST /api/v1/assets/redemptions", handler.redeemExternalAssetBody)
	mux.HandleFunc("POST /api/v1/assets/{assetID}/redeem", handler.redeemExternalAsset)
	mux.HandleFunc("GET /api/v1/integrations/platform/assets", handler.userAssetsByQuery)
	mux.HandleFunc("POST /api/v1/integrations/platform/redemptions", handler.redeemExternalAssetBody)
	mux.HandleFunc("POST /api/v1/integrations/verification-codes", handler.sendVerificationCode)
	mux.HandleFunc("POST /api/v1/integrations/bindings", handler.bindExternalUser)
	mux.HandleFunc("POST /api/v1/integrations/bind", handler.bindExternalUser)
	mux.HandleFunc("GET /api/v1/integrations/users/{userID}/assets", handler.userAssets)
	mux.HandleFunc("POST /api/v1/integrations/users/{userID}/assets/{assetID}/redemptions", handler.redeemExternalAsset)
	mux.HandleFunc("GET /api/v1/integrations/assets", handler.userAssetsByQuery)
	mux.HandleFunc("POST /api/v1/integrations/redemptions", handler.redeemExternalAssetBody)
	mux.HandleFunc("GET /api/v1/admin/airdrops", handler.adminAirdrops)
	mux.HandleFunc("GET /api/v1/admin/migration-requests", handler.adminMigrationRequests)
	mux.HandleFunc("POST /api/v1/admin/migration-requests/{requestID}/review", handler.reviewMigrationRequest)
	mux.HandleFunc("POST /api/v1/wallet/airdrops", handler.walletCreateAirdrop)
	mux.HandleFunc("POST /api/v1/wallet/airdrops/{airdropID}/cancel", handler.cancelWalletAirdrop)
	mux.HandleFunc("POST /api/v1/admin/airdrops", handler.adminCreateAirdrop)
	mux.HandleFunc("POST /api/v1/admin/airdrops/{airdropID}/simulate", handler.simulateAirdrop)
	mux.HandleFunc("GET /api/v1/admin/bnb-networks", handler.adminBNBNetworks)
	mux.HandleFunc("GET /api/v1/admin/airdrop-rules", handler.adminAirdropRules)
	mux.HandleFunc("POST /api/v1/internal/airdrops/{airdropID}/result", handler.updateAirdropResult)

	// Existing browser-client contract.
	mux.HandleFunc("GET /api/overview", handler.overview)
	mux.HandleFunc("GET /api/works", handler.works)
	mux.HandleFunc("GET /api/works/{workID}", handler.work)
	mux.HandleFunc("GET /api/assets", handler.assetsDashboard)
	mux.HandleFunc("GET /api/studio", handler.studio)
	mux.HandleFunc("GET /api/profile", handler.profile)
	mux.HandleFunc("POST /api/hapw/redemptions", handler.createRedemption)
	mux.HandleFunc("POST /api/generations", handler.createGeneration)
	mux.HandleFunc("POST /api/conversions", handler.createConversion)
	mux.HandleFunc("POST /api/exercises", handler.createExercise)
	mux.HandleFunc("POST /api/transfers", handler.createExercise)
	mux.HandleFunc("POST /api/clip/hapw-exchanges", handler.createExchange)
	mux.HandleFunc("POST /api/profile/wallet", handler.connectWallet)
	mux.HandleFunc("DELETE /api/profile/wallet", handler.disconnectWallet)
	mux.HandleFunc("POST /api/profile/overseas", handler.setOverseasAccount)
	mux.HandleFunc("DELETE /api/profile/overseas", handler.clearOverseasAccount)
	mux.HandleFunc("PATCH /api/profile/settings", handler.updateSettings)
	mux.HandleFunc("/api/", handler.apiNotFound)
	mux.HandleFunc("/", handler.serveStatic)

	return handler.security(handler.requestLog(mux))
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"status": "ok", "service": "clipli-api", "runtime": "go", "time": time.Now().UTC().Format(time.RFC3339)}})
}

func (h *Handler) adminPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, filepath.Join(h.publicDir, "admin.html"))
}

func (h *Handler) overview(w http.ResponseWriter, _ *http.Request) {
	state := h.service.PortfolioSnapshot()
	var featured *domain.HAPWAsset
	if len(state.Assets) > 0 {
		featured = &state.Assets[0]
	}
	completed := 0
	for _, item := range state.Exercises {
		if item.StatusCode == "completed" {
			completed++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"counts":        map[string]any{"authorized": len(state.Assets), "overseas": completed, "visible": len(state.Works), "redeemed": len(state.Redemptions), "credits": state.GenerationAccount.Balance, "clipGranted": state.GenerationAccount.LifetimeClipGranted},
		"featuredAsset": featured, "steps": []string{"hold", "redeem", "grant", "generate"},
	}})
}

func (h *Handler) works(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"data": h.service.ListWorks()})
}

func (h *Handler) work(w http.ResponseWriter, r *http.Request) {
	work, ok := h.service.GetWork(r.PathValue("workID"))
	if !ok {
		writeError(w, http.StatusNotFound, "work_not_found", "Work not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": work})
}

func (h *Handler) assetsDashboard(w http.ResponseWriter, _ *http.Request) {
	state := h.service.AccountSnapshot()
	assets := make([]domain.HAPWAsset, 0, len(state.Assets))
	holdings, transferable, totalValue := 0, 0, 0
	for _, asset := range state.Assets {
		if asset.Kind != "HAPW" {
			continue
		}
		assets = append(assets, asset)
		if asset.RedemptionStatus != "redeemed" {
			holdings++
			totalValue += asset.Value
			if asset.Transferable {
				transferable++
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"assets":    assets,
		"stats":     map[string]any{"holdings": holdings, "transferable": transferable, "recent": len(state.Exercises), "totalValue": totalValue},
		"transfers": state.Exercises, "exercises": state.Exercises, "platforms": state.Platforms,
		"clip": clipView(state), "clipTransactions": state.CLIPTransactions,
		"generationAccount": state.GenerationAccount, "hapwRedemptions": state.Redemptions,
		"generations": state.Generations, "hapwExchanges": state.HAPWExchanges,
		"assetSources": state.AssetSources, "assetSyncRuns": state.AssetSyncRuns,
		"clipTreasury": state.CLIPTreasury, "clipDistributionRules": state.DistributionRules, "clipDistributions": state.Distributions,
		"session": state.Session,
	}})
}

func (h *Handler) studio(w http.ResponseWriter, _ *http.Request) {
	state := h.service.AccountSnapshot()
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"assets": state.Assets, "generationAccount": state.GenerationAccount, "redemptions": state.Redemptions,
		"generations": state.Generations, "clip": clipView(state),
	}})
}

func (h *Handler) profile(w http.ResponseWriter, _ *http.Request) {
	state := h.service.Snapshot()
	walletAssets, _ := h.service.WalletAssets(state.Profile.Wallet)
	airdrops, _ := h.service.Airdrops(state.Profile.Wallet, "")
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"wallet": state.Profile.Wallet, "walletProvider": state.Profile.WalletProvider, "walletChainId": state.Profile.WalletChainID, "walletStatus": state.Profile.WalletStatus, "walletConnectedAt": state.Profile.WalletConnectedAt, "overseasAccount": state.Profile.OverseasAccount,
		"phone": state.Profile.Phone, "level": state.Profile.Level, "points": state.Profile.Points,
		"settings": state.Profile.Settings, "clipBalance": state.CLIP.Balance, "session": state.Session,
		"walletAssets": walletAssets, "airdrops": airdrops, "clip": state.CLIP,
		"externalPlatform": map[string]any{"code": "haiwen", "configured": h.service.ExternalPlatformConfigured(), "sandbox": h.service.ExternalPlatformSandbox()},
	}})
}

func (h *Handler) session(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"data": h.service.Snapshot().Session})
}

func (h *Handler) clipTreasury(w http.ResponseWriter, _ *http.Request) {
	state := h.service.Snapshot()
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"treasury": state.CLIPTreasury, "accountBalance": state.CLIP.Balance,
		"distributionRules": state.DistributionRules, "ledgerConservation": map[string]any{
			"mintedSupply":        state.CLIPTreasury.MintedSupply,
			"treasuryBalance":     state.CLIPTreasury.TreasuryBalance,
			"liquidityAllocation": state.CLIPTreasury.LiquidityAllocation,
			"ledgerOutstanding":   state.CLIPTreasury.LedgerOutstanding,
			"conserved":           state.CLIPTreasury.TreasuryBalance+state.CLIPTreasury.LiquidityAllocation+state.CLIPTreasury.LedgerOutstanding == state.CLIPTreasury.MintedSupply,
		},
	}})
}

func (h *Handler) clipDistributionRules(w http.ResponseWriter, _ *http.Request) {
	items := h.service.Snapshot().DistributionRules
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) walletAssets(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		address = h.service.Snapshot().Profile.Wallet
	}
	items, err := h.service.WalletAssets(address)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"walletAddress": address, "items": items, "totalItems": len(items), "sourceOfTruth": "external-platform"}})
}

func (h *Handler) walletAirdropEligibility(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		address = h.service.Snapshot().Profile.Wallet
	}
	eligibility, err := h.service.AirdropEligibility(r.URL.Query().Get("ruleCode"), address, r.URL.Query().Get("assetId"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": eligibility})
}

func (h *Handler) airdrops(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		address = h.service.Snapshot().Profile.Wallet
	}
	items, err := h.service.Airdrops(address, r.URL.Query().Get("status"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) airdropRules(w http.ResponseWriter, _ *http.Request) {
	items := h.service.AirdropRules()
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) adminAirdropRules(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	h.airdropRules(w, r)
}

func (h *Handler) adminAirdrops(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	items, err := h.service.Airdrops(r.URL.Query().Get("walletAddress"), r.URL.Query().Get("status"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) adminCreateAirdrop(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	var input service.AirdropInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, idempotent, err := h.service.CreateAirdrop(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, creationStatus(idempotent), map[string]any{"data": result})
}

func (h *Handler) adminBNBNetworks(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	networks := service.BNBChainConfigs()
	for i := range networks {
		if networks[i].ChainID == "0x38" {
			networks[i].ContractDeployed = h.service.Snapshot().CLIPTreasury.MintStatus == "onchain-verified"
			networks[i].SimulationOnly = false
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"items": networks, "contractDeployed": h.service.Snapshot().CLIPTreasury.MintStatus == "onchain-verified", "simulationOnly": false,
		"note": "Real airdrops require the configured CLIP mainnet contract, funded treasury and verified receipts.",
	}})
}

func (h *Handler) simulateAirdrop(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	var input struct {
		ChainID string `json:"chainId"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	writeError(w, http.StatusGone, "airdrop_simulation_disabled", "Real asset ledgers cannot accept simulated transfers")
}

func (h *Handler) updateAirdropResult(w http.ResponseWriter, r *http.Request) {
	if !h.requireExecutor(w, r) {
		return
	}
	var input service.AirdropResultInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, err := h.service.UpdateAirdrop(r.PathValue("airdropID"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) clipDistributions(w http.ResponseWriter, r *http.Request) {
	items := h.service.Snapshot().Distributions
	assetID, requestID := r.URL.Query().Get("assetId"), r.URL.Query().Get("requestId")
	if assetID != "" || requestID != "" {
		filtered := make([]domain.CLIPDistribution, 0, len(items))
		for _, item := range items {
			if (assetID == "" || item.AssetID == assetID) && (requestID == "" || item.RequestID == requestID) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) assetSources(w http.ResponseWriter, _ *http.Request) {
	items := h.service.Snapshot().AssetSources
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items), "sourceOfTruth": "external-platform"}})
}

func (h *Handler) assetSource(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("sourceCode")
	for _, item := range h.service.Snapshot().AssetSources {
		if item.Code == code {
			writeJSON(w, http.StatusOK, map[string]any{"data": item})
			return
		}
	}
	writeError(w, http.StatusNotFound, "asset_source_not_found", "Asset source not found")
}

func (h *Handler) sendVerificationCode(w http.ResponseWriter, r *http.Request) {
	var input service.SendVerificationInput
	if !decodeBody(w, r, &input) {
		return
	}
	if input.UserID == "" {
		input.UserID = r.PathValue("userID")
	}
	if input.UserID == "" {
		input.UserID = r.URL.Query().Get("userId")
	}
	result, err := h.service.SendVerificationCode(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"data": map[string]any{
		"challenge": result.Challenge, "verificationId": result.Challenge.ID,
		"userId": result.Challenge.UserID, "clipliUserId": result.Challenge.UserID, "phoneMasked": result.Challenge.PhoneMasked,
		"status": result.Challenge.Status, "expiresAt": result.Challenge.ExpiresAt,
	}})
}

func (h *Handler) bindExternalUser(w http.ResponseWriter, r *http.Request) {
	var input service.BindUserInput
	if !decodeBody(w, r, &input) {
		return
	}
	if input.UserID == "" {
		input.UserID = r.PathValue("userID")
	}
	if input.UserID == "" {
		input.UserID = r.URL.Query().Get("userId")
	}
	result, idempotent, err := h.service.BindUser(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, creationStatus(idempotent), map[string]any{"data": map[string]any{
		"binding": result.Binding, "userId": result.Binding.UserID,
		"clipliUserId":   result.Binding.UserID,
		"externalUserId": result.Binding.ExternalUserID, "phoneMasked": result.Binding.PhoneMasked, "status": result.Binding.Status,
	}})
}

func (h *Handler) getExternalBinding(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		userID = r.URL.Query().Get("clipliUserId")
	}
	if userID == "" {
		status, err := h.service.ExternalBindingStatus(r.URL.Query().Get("externalUserId"))
		if err != nil {
			h.writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": status})
		return
	}
	binding, _, err := h.service.GetBinding(userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"binding": binding, "userId": binding.UserID,
		"clipliUserId":   binding.UserID,
		"externalUserId": binding.ExternalUserID, "phoneMasked": binding.PhoneMasked, "status": binding.Status,
	}})
}

func (h *Handler) userAssets(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	if raw, ok := r.URL.Query()["tplIds"]; ok {
		tplIDs, err := parseTemplateIDs(strings.Join(raw, ","))
		if err != nil {
			h.writeServiceError(w, err)
			return
		}
		result, err := h.service.UserAssetCounts(userID, tplIDs)
		if err != nil {
			h.writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": result})
		return
	}
	result, err := h.service.UserAssets(userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) userAssetCounts(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	tplIDs, err := parseTemplateIDs(r.URL.Query().Get("tplIds"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	result, err := h.service.UserAssetCounts(userID, tplIDs)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) externalTemplates(w http.ResponseWriter, r *http.Request) {
	workID := int64(queryInt(r, "workId", 0))
	result, err := h.service.ExternalTemplates(queryInt(r, "page", 1), queryInt(r, "pageSize", 20), workID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) externalWorks(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ExternalWorks(queryInt(r, "page", 1), queryInt(r, "pageSize", 20))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) externalMigrations(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	items, err := h.service.ExternalMigrations(userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) previewExternalMigration(w http.ResponseWriter, r *http.Request) {
	var input service.MigrateExternalAssetInput
	if !decodeBody(w, r, &input) {
		return
	}
	if input.UserID == "" {
		input.UserID = r.PathValue("userID")
	}
	result, err := h.service.PreviewExternalAssetMigration(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) migrateExternalAsset(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	var input service.MigrateExternalAssetInput
	if !decodeBody(w, r, &input) {
		return
	}
	if input.UserID == "" {
		input.UserID = r.PathValue("userID")
	}
	result, idempotent, err := h.service.MigrateExternalAsset(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, creationStatus(idempotent), map[string]any{"data": result})
}

func (h *Handler) userAssetsByQuery(w http.ResponseWriter, r *http.Request) {
	h.userAssets(w, r)
}

func (h *Handler) redeemExternalAsset(w http.ResponseWriter, r *http.Request) {
	var input service.RedeemExternalAssetInput
	if !decodeBody(w, r, &input) {
		return
	}
	if input.UserID == "" {
		input.UserID = r.PathValue("userID")
	}
	if input.UserID == "" {
		input.UserID = r.URL.Query().Get("userId")
	}
	if input.AssetID == "" {
		input.AssetID = r.PathValue("assetID")
	}
	h.writeExternalRedemption(w, input)
}

func (h *Handler) redeemExternalAssetBody(w http.ResponseWriter, r *http.Request) {
	var input service.RedeemExternalAssetInput
	if !decodeBody(w, r, &input) {
		return
	}
	if input.UserID == "" {
		input.UserID = r.PathValue("userID")
	}
	if input.UserID == "" {
		input.UserID = r.URL.Query().Get("userId")
	}
	if input.AssetID == "" {
		input.AssetID = r.PathValue("assetID")
	}
	h.writeExternalRedemption(w, input)
}

func (h *Handler) writeExternalRedemption(w http.ResponseWriter, input service.RedeemExternalAssetInput) {
	result, idempotent, err := h.service.RedeemExternalAsset(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	status := creationStatus(idempotent)
	writeJSON(w, status, map[string]any{"data": map[string]any{
		"redemption": result.Redemption, "id": result.Redemption.ID,
		"assetId": result.Redemption.AssetID, "serialNumber": result.Redemption.SerialNumber,
		"serialNo": result.Redemption.SerialNumber, "uniqueSerialNumber": result.Redemption.SerialNumber,
		"requestNo": result.Redemption.RequestNo, "tplId": result.Redemption.TplID, "num": result.Redemption.Num,
		"externalUserId": result.Redemption.ExternalUserID, "status": result.Redemption.Status,
	}})
}

func parseTemplateIDs(raw string) ([]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, serviceError(http.StatusBadRequest, "invalid_tpl_ids", "Provide between 1 and 100 template IDs")
	}
	parts := strings.Split(raw, ",")
	if len(parts) == 0 || len(parts) > 100 {
		return nil, serviceError(http.StatusBadRequest, "invalid_tpl_ids", "Provide between 1 and 100 template IDs")
	}
	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || value <= 0 {
			return nil, serviceError(http.StatusBadRequest, "invalid_tpl_ids", "Template IDs must be positive integers")
		}
		ids = append(ids, value)
	}
	return ids, nil
}

func serviceError(status int, code, message string) error {
	return &service.APIError{Status: status, Code: code, Message: message}
}

func (h *Handler) assetSyncRuns(w http.ResponseWriter, _ *http.Request) {
	items := h.service.Snapshot().AssetSyncRuns
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) assetRequirements(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"contractVersion": "clipli-haiwen-openapi-v1",
		"sourceOfTruth":   "external-platform",
		"authentication":  "x-app-id/x-app-key; server-side only, never exposed to browser",
		"requiredEndpoints": []map[string]any{
			{"method": "GET", "path": "/openapi/works", "purpose": "published work list", "requiredFields": []string{"workId", "worksName", "publishNum"}},
			{"method": "GET", "path": "/openapi/tpls", "purpose": "published copyright template list", "requiredFields": []string{"tplId", "name", "workId", "publishCount"}},
			{"method": "GET", "path": "/openapi/user/bind/status?externalUserId={externalUserId}", "purpose": "binding status before starting the flow", "requiredFields": []string{"externalUserId", "bound", "boundAt"}},
			{"method": "POST", "path": "/openapi/user/bind/sms", "purpose": "send the binding SMS", "requiredFields": []string{"externalUserId", "phone"}},
			{"method": "POST", "path": "/openapi/user/bind", "purpose": "validate the SMS code and create the binding", "requiredFields": []string{"externalUserId", "phone", "smsCode"}},
			{"method": "GET", "path": "/openapi/user/assets/count?externalUserId={externalUserId}&tplIds={tplIds}", "purpose": "query redeemable counts by template", "requiredFields": []string{"externalUserId", "list[].tplId", "list[].count"}},
			{"method": "POST", "path": "/openapi/asset/write-off", "purpose": "write off assets idempotently", "requiredFields": []string{"requestNo", "externalUserId", "tplId", "num"}},
		},
		"clipliInternalTargetEndpoints": []map[string]any{
			{"method": "GET", "path": "/api/v1/hapw/assets/{assetId}", "purpose": "Clipli normalized asset detail; not provided by Haiwen"},
			{"method": "GET", "path": "/api/v1/hapw/assets/{assetId}/authorization", "purpose": "Clipli rights projection; source fields must be mapped explicitly"},
			{"method": "GET", "path": "/api/v1/hapw/assets/{assetId}/history", "purpose": "Clipli audit timeline; not an upstream endpoint"},
			{"method": "GET", "path": "/api/v1/wallet/assets", "purpose": "Clipli wallet view; Haiwen does not expose wallet assets"},
		},
		"normalization": map[string]any{"timeoutSeconds": 5, "maxResponseBytes": 2097152, "failureMode": "retain-last-known-snapshot", "writeBack": false},
	}})
}

func (h *Handler) hapwSummary(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"data": h.service.HAPWSummary()})
}

func (h *Handler) hapwAssets(w http.ResponseWriter, r *http.Request) {
	options := service.AssetListOptions{
		Search: r.URL.Query().Get("search"), Status: r.URL.Query().Get("status"), Owner: r.URL.Query().Get("owner"),
		RightsHolder: r.URL.Query().Get("rightsHolder"), Page: queryInt(r, "page", 1), PageSize: queryInt(r, "pageSize", 20), Sort: r.URL.Query().Get("sort"),
	}
	var err error
	options.Redeemable, err = optionalBool(r, "redeemable")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	options.ExchangeAvailable, err = optionalBool(r, "exchangeAvailable")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": h.service.ListAssets(options)})
}

func (h *Handler) hapwAsset(w http.ResponseWriter, r *http.Request) {
	asset, ok := h.service.GetAsset(r.PathValue("assetID"))
	if !ok {
		writeError(w, http.StatusNotFound, "hapw_not_found", "HAPW asset not found")
		return
	}
	state := h.service.Snapshot()
	works := relatedWorks(state.Works, asset.ID)
	actions := map[string]bool{"redeem": asset.RedemptionStatus == "available", "exercise": asset.Transferable, "exchange": asset.ExchangeAvailable, "generate": asset.RedemptionStatus == "redeemed"}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"asset": asset, "relatedWorks": works, "availableActions": actions}})
}

func (h *Handler) hapwAuthorization(w http.ResponseWriter, r *http.Request) {
	asset, ok := h.service.GetAsset(r.PathValue("assetID"))
	if !ok {
		writeError(w, http.StatusNotFound, "hapw_not_found", "HAPW asset not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"assetId": asset.ID, "tokenId": asset.TokenID, "authorization": asset.Authorization, "provenance": asset.Provenance}})
}

func (h *Handler) hapwHistory(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.AssetHistory(r.PathValue("assetID"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) hapwWorks(w http.ResponseWriter, r *http.Request) {
	asset, ok := h.service.GetAsset(r.PathValue("assetID"))
	if !ok {
		writeError(w, http.StatusNotFound, "hapw_not_found", "HAPW asset not found")
		return
	}
	works := relatedWorks(h.service.ListWorks(), asset.ID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": works, "totalItems": len(works)}})
}

func (h *Handler) hapwExchangeQuote(w http.ResponseWriter, r *http.Request) {
	asset, ok := h.service.GetAsset(r.PathValue("assetID"))
	if !ok {
		writeError(w, http.StatusNotFound, "hapw_not_found", "HAPW asset not found")
		return
	}
	if !asset.ExchangeAvailable || asset.ClipPrice <= 0 {
		writeError(w, http.StatusBadRequest, "hapw_reserve_unavailable", "This reserve HAPW is unavailable")
		return
	}
	quote, err := service.QuoteHAPWExchange(asset.ClipPrice, service.HAPWExchangeFeeRate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_exchange_quote", "Unable to quote this HAPW")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"assetId": asset.ID, "tokenId": asset.TokenID, "quoteAsset": "CLIP", "quote": quote, "expiresInSeconds": 60, "isIndicative": true}})
}

func (h *Handler) hapwRedemptions(w http.ResponseWriter, r *http.Request) {
	items := h.service.Snapshot().Redemptions
	if assetID := r.URL.Query().Get("assetId"); assetID != "" {
		items = filterRedemptions(items, assetID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) hapwRedemption(w http.ResponseWriter, r *http.Request) {
	for _, item := range h.service.Snapshot().Redemptions {
		if item.ID == r.PathValue("redemptionID") {
			writeJSON(w, http.StatusOK, map[string]any{"data": item})
			return
		}
	}
	writeError(w, http.StatusNotFound, "redemption_not_found", "HAPW redemption not found")
}

func (h *Handler) hapwExercises(w http.ResponseWriter, r *http.Request) {
	items := h.service.Snapshot().Exercises
	if assetID := r.URL.Query().Get("assetId"); assetID != "" {
		filtered := make([]domain.AssetExercise, 0)
		for _, item := range items {
			if item.AssetID == assetID {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) hapwExchanges(w http.ResponseWriter, r *http.Request) {
	items := h.service.Snapshot().HAPWExchanges
	if assetID := r.URL.Query().Get("assetId"); assetID != "" {
		filtered := make([]domain.HAPWExchange, 0)
		for _, item := range items {
			if item.AssetID == assetID {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) hapwExchangePolicy(w http.ResponseWriter, _ *http.Request) {
	state := h.service.Snapshot()
	writeJSON(w, http.StatusOK, map[string]any{"data": service.BuildExchangePolicy(state.HAPWExchanges, state.Assets, time.Now())})
}

func (h *Handler) platforms(w http.ResponseWriter, _ *http.Request) {
	items := h.service.Snapshot().Platforms
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) createRedemption(w http.ResponseWriter, r *http.Request) {
	var input service.RedemptionInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, idempotent, err := h.service.Redeem(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, creationStatus(idempotent), map[string]any{"data": result})
}

func (h *Handler) createGeneration(w http.ResponseWriter, r *http.Request) {
	var input service.GenerationInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, idempotent, err := h.service.Generate(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, creationStatus(idempotent), map[string]any{"data": result})
}

func (h *Handler) createConversion(w http.ResponseWriter, r *http.Request) {
	var input service.ConversionInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, idempotent, err := h.service.Convert(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, creationStatus(idempotent), map[string]any{"data": result})
}

func (h *Handler) createExercise(w http.ResponseWriter, r *http.Request) {
	var input service.ExerciseInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, idempotent, err := h.service.Exercise(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, creationStatus(idempotent), map[string]any{"data": result})
}

func (h *Handler) createExchange(w http.ResponseWriter, r *http.Request) {
	var input service.ExchangeInput
	if !decodeBody(w, r, &input) {
		return
	}
	result, idempotent, err := h.service.Exchange(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, creationStatus(idempotent), map[string]any{"data": result})
}

func (h *Handler) connectWallet(w http.ResponseWriter, r *http.Request) {
	var input service.WalletConnectInput
	if !decodeBody(w, r, &input) {
		return
	}
	profile, err := h.service.ConnectWallet(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": profile})
}

func (h *Handler) disconnectWallet(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"data": h.service.DisconnectWallet()})
}

func (h *Handler) setOverseasAccount(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusGone, "verified_binding_required", "Use the external platform SMS verification and binding endpoints")
}

func (h *Handler) clearOverseasAccount(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusGone, "external_unbinding_unsupported", "Unbinding requires the external platform; local profile edits cannot unlink the account")
}

func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	var input struct {
		WalletSign     *bool `json:"walletSign"`
		ExpiryReminder *bool `json:"expiryReminder"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": h.service.UpdateSettings(input.WalletSign, input.ExpiryReminder)})
}

func (h *Handler) serveStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	path := filepath.Clean("/" + r.URL.Path)
	path = strings.TrimPrefix(path, "/")
	filePath := filepath.Join(h.publicDir, path)
	if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
		http.ServeFile(w, r, filePath)
		return
	}
	http.ServeFile(w, r, filepath.Join(h.publicDir, "index.html"))
}

func (h *Handler) apiNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "API route not found")
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var apiErr *service.APIError
	if errors.As(err, &apiErr) {
		writeError(w, apiErr.Status, apiErr.Code, apiErr.Message)
		return
	}
	h.logger.Error("unhandled service error", "error", err)
	writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	return requireSecret(w, r, h.adminKey, "X-Clipli-Admin-Key", "admin_not_configured", "admin_unauthorized")
}

func (h *Handler) requireExecutor(w http.ResponseWriter, r *http.Request) bool {
	return requireSecret(w, r, h.executorKey, "X-Clipli-Executor-Key", "executor_not_configured", "executor_unauthorized")
}

func requireSecret(w http.ResponseWriter, r *http.Request, configured, header, missingCode, invalidCode string) bool {
	if strings.TrimSpace(configured) == "" {
		writeError(w, http.StatusServiceUnavailable, missingCode, "This protected integration is not configured")
		return false
	}
	received := r.Header.Get(header)
	if subtle.ConstantTimeCompare([]byte(received), []byte(configured)) != 1 {
		writeError(w, http.StatusUnauthorized, invalidCode, "The integration credential is invalid")
		return false
	}
	return true
}

func (h *Handler) security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		h.logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started).String())
	})
}

func decodeBody(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must contain one JSON object")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func creationStatus(idempotent bool) int {
	if idempotent {
		return http.StatusOK
	}
	return http.StatusCreated
}

func queryInt(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func optionalBool(r *http.Request, key string) (*bool, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("%s must be true or false", key)
	}
	return &parsed, nil
}

func relatedWorks(works []domain.Work, assetID string) []domain.Work {
	items := make([]domain.Work, 0)
	for _, work := range works {
		if work.LinkedAssetID == assetID {
			items = append(items, work)
		}
	}
	return items
}

func filterRedemptions(items []domain.HAPWRedemption, assetID string) []domain.HAPWRedemption {
	filtered := make([]domain.HAPWRedemption, 0)
	for _, item := range items {
		if item.AssetID == assetID {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func clipView(state store.State) map[string]any {
	pool, _ := service.BuildDexPoolSnapshot(state.CLIP.DexPool)
	policy := service.BuildExchangePolicy(state.HAPWExchanges, state.Assets, time.Now())
	return map[string]any{
		"symbol": state.CLIP.Symbol, "balance": state.CLIP.Balance, "reserved": state.CLIP.Reserved, "availableBalance": state.CLIP.Balance - state.CLIP.Reserved,
		"supplyPolicy": state.CLIP.SupplyPolicy, "supplyPolicyEn": state.CLIP.SupplyPolicyEn,
		"acquisition": state.CLIP.Acquisition, "acquisitionEn": state.CLIP.AcquisitionEn, "acquisitionKo": state.CLIP.AcquisitionKo,
		"quoteAsset":  state.CLIP.QuoteAsset,
		"pricePolicy": state.CLIP.PricePolicy, "pricePolicyEn": state.CLIP.PricePolicyEn, "pricePolicyKo": state.CLIP.PricePolicyKo,
		"contractStatus": state.CLIP.ContractStatus, "contractStatusEn": state.CLIP.ContractStatusEn, "contractStatusKo": state.CLIP.ContractStatusKo,
		"dexUrl": state.CLIP.DexURL, "dexPool": pool,
		"hapwExchangeFeeRate": state.CLIP.HAPWExchangeFeeRate, "hapwExchangePolicy": policy,
		"treasury": state.CLIPTreasury, "distributionRules": state.DistributionRules,
	}
}
