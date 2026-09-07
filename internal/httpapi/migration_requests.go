package httpapi

import (
	"github.com/StevenWXY/HAPE/internal/service"
	"net/http"
)

func (h *Handler) migrationRequests(w http.ResponseWriter, r *http.Request) {
	userID := h.service.Snapshot().Session.UserRef
	if r.PathValue("userID") != userID {
		writeError(w, http.StatusForbidden, "session_user_mismatch", "The request must belong to the current account")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": h.service.MigrationRequests(userID)}})
}

func (h *Handler) requestMigration(w http.ResponseWriter, r *http.Request) {
	var input service.MigrateExternalAssetInput
	if !decodeBody(w, r, &input) {
		return
	}
	input.UserID = r.PathValue("userID")
	result, repeated, err := h.service.RequestMigration(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, creationStatus(repeated), map[string]any{"data": map[string]any{"request": result}})
}

func (h *Handler) adminMigrationRequests(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": h.service.MigrationRequests("")}})
}

func (h *Handler) reviewMigrationRequest(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	var input struct {
		Action   string `json:"action"`
		Reason   string `json:"reason"`
		Accepted bool   `json:"accepted"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	result, err := h.service.ReviewMigrationRequest(r.PathValue("requestID"), input.Action, input.Reason, input.Accepted)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"request": result}})
}

func (h *Handler) walletCreateAirdrop(w http.ResponseWriter, r *http.Request) {
	var input service.AirdropInput
	if !decodeBody(w, r, &input) {
		return
	}
	profile := h.service.Snapshot().Profile
	input.RuleCode, input.WalletAddress, input.ChainID = "hapw_redemption", profile.Wallet, profile.WalletChainID
	result, repeated, err := h.service.CreateAirdrop(input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, creationStatus(repeated), map[string]any{"data": map[string]any{"airdrop": result}})
}

func (h *Handler) cancelWalletAirdrop(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.CancelWalletAirdrop(r.PathValue("airdropID"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"airdrop": result}})
}
