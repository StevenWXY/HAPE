package httpapi

import (
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
	service   *service.Service
	publicDir string
	logger    *slog.Logger
}

func New(serviceLayer *service.Service, publicDir string, logger *slog.Logger) http.Handler {
	handler := &Handler{service: serviceLayer, publicDir: publicDir, logger: logger}
	mux := http.NewServeMux()

	// Operational endpoint.
	mux.HandleFunc("GET /api/health", handler.health)

	// Versioned HAPW API.
	mux.HandleFunc("GET /api/v1/session", handler.session)
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

func (h *Handler) overview(w http.ResponseWriter, _ *http.Request) {
	state := h.service.Snapshot()
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"counts":        map[string]any{"authorized": 18420, "overseas": 2908, "visible": 86, "redeemed": len(state.Redemptions), "credits": state.GenerationAccount.Balance, "clipGranted": state.GenerationAccount.LifetimeClipGranted},
		"featuredAsset": state.Assets[0], "steps": []string{"hold", "redeem", "grant", "generate"},
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
	state := h.service.Snapshot()
	assets := make([]domain.HAPWAsset, 0, len(state.Assets))
	holdings, transferable, totalValue := 0, 0, 0
	for _, asset := range state.Assets {
		if asset.Kind != "HAPW" || asset.Owner != "Clipli" {
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
	state := h.service.Snapshot()
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"assets": state.Assets, "generationAccount": state.GenerationAccount, "redemptions": state.Redemptions,
		"generations": state.Generations, "clip": clipView(state),
	}})
}

func (h *Handler) profile(w http.ResponseWriter, _ *http.Request) {
	state := h.service.Snapshot()
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"wallet": state.Profile.Wallet, "walletProvider": state.Profile.WalletProvider, "overseasAccount": state.Profile.OverseasAccount,
		"phone": state.Profile.Phone, "level": state.Profile.Level, "points": state.Profile.Points,
		"settings": state.Profile.Settings, "clipBalance": state.CLIP.Balance, "session": state.Session,
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

func (h *Handler) assetSyncRuns(w http.ResponseWriter, _ *http.Request) {
	items := h.service.Snapshot().AssetSyncRuns
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items, "totalItems": len(items)}})
}

func (h *Handler) assetRequirements(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"contractVersion": "clipli-external-asset-v1",
		"sourceOfTruth":   "external-platform",
		"authentication":  "provider-api-key-or-signed-service-account; never exposed to browser",
		"requiredEndpoints": []map[string]any{
			{"method": "GET", "path": "/assets", "purpose": "full or cursor-based HAPW asset list", "requiredFields": []string{"id", "tokenId", "name", "rightsHolder", "status", "authorization", "provenance"}},
			{"method": "GET", "path": "/assets/{assetId}", "purpose": "single asset detail and current state", "requiredFields": []string{"id", "tokenId", "status", "owner", "externalUrl", "updatedAt"}},
			{"method": "GET", "path": "/assets/{assetId}/authorization", "purpose": "rights scope and territories", "requiredFields": []string{"holder", "scope", "territories", "usageTypes", "validFrom"}},
			{"method": "GET", "path": "/assets/{assetId}/media", "purpose": "linked work and preview metadata", "requiredFields": []string{"linkedWorkIds", "format"}},
			{"method": "GET", "path": "/assets/{assetId}/events", "purpose": "issue, transfer and rights-state history", "requiredFields": []string{"eventId", "type", "status", "createdAt"}},
			{"method": "GET", "path": "/works?assetId={assetId}", "purpose": "works linked to an external HAPW", "requiredFields": []string{"id", "title", "linkedAssetId", "externalUrl"}},
			{"method": "POST", "path": "/webhooks/asset-events", "purpose": "optional near-real-time invalidation", "requiredFields": []string{"eventId", "assetId", "eventType", "occurredAt", "signature"}},
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
	var input struct {
		Provider string `json:"provider"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	profile, err := h.service.ConnectWallet(input.Provider)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": profile})
}

func (h *Handler) setOverseasAccount(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Account string `json:"account"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	profile, err := h.service.SetOverseasAccount(input.Account)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": profile})
}

func (h *Handler) clearOverseasAccount(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"data": h.service.ClearOverseasAccount()})
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

func (h *Handler) security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'")
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
		"symbol": state.CLIP.Symbol, "balance": state.CLIP.Balance,
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
