package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/service"
	"github.com/StevenWXY/HAPE/internal/store"
)

func TestInitialPortfolioHasNoFabricatedAssetsOrBalances(t *testing.T) {
	handler := New(service.New(store.NewMemory(store.InitialState())), "../../public", slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, route := range []string{"/api/overview", "/api/assets", "/api/studio", "/api/profile", "/api/works", "/api/v1/hapw/assets", "/api/v1/hapw/summary"} {
		status, body := doJSONRequest(t, handler, http.MethodGet, route, nil)
		if status != http.StatusOK {
			t.Fatalf("%s: status %d", route, status)
		}
		if route == "/api/works" {
			if string(body) != "[]" {
				t.Fatalf("works = %s", body)
			}
			continue
		}
		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			t.Fatal(err)
		}
		if route == "/api/overview" {
			if data["featuredAsset"] != nil {
				t.Fatal("empty homepage must have no featured asset")
			}
			for _, value := range data["counts"].(map[string]any) {
				if value != float64(0) {
					t.Fatal("fabricated homepage count")
				}
			}
		}
		for _, key := range []string{"assets", "redemptions", "generations", "clipTransactions", "hapwRedemptions", "exercises", "assetSources", "assetSyncRuns", "items"} {
			if value, present := data[key]; present {
				items, ok := value.([]any)
				if !ok || len(items) != 0 {
					t.Fatalf("%s: %s = %#v", route, key, value)
				}
			}
		}
		for _, key := range []string{"clip", "generationAccount"} {
			if value, present := data[key]; present && value.(map[string]any)["balance"] != float64(0) {
				t.Fatalf("nonzero %s", key)
			}
		}
		if route == "/api/profile" && data["externalPlatform"].(map[string]any)["configured"] != false {
			t.Fatal("partner must not use demo fallback")
		}
	}
	status, _ := doJSONRequest(t, handler, http.MethodPost, "/api/v1/integrations/platform/verification-codes", map[string]any{"userId": "local-user", "externalUserId": "123456", "phone": "13800138000"})
	if status < 400 {
		t.Fatal("unconfigured SMS must not report success")
	}
}

func TestPortfolioShowsConfirmedUserTransfersOnly(t *testing.T) {
	state := store.InitialState()
	confirmed := domain.HAPWAsset{ID: "real-transfer", Kind: "HAPW", TokenID: "#1", Owner: state.Session.UserRef, RedemptionStatus: "redeemed", External: domain.ExternalAssetRef{ProviderCode: "haiwen", ProviderAssetID: "100053", TransferID: "receipt-1", SyncStatus: "migrated"}}
	pending := confirmed
	pending.ID = "pending"
	pending.External.SyncStatus = "pending-write-off"
	snapshot := confirmed
	snapshot.ID = "catalog-only"
	snapshot.External.TransferID = ""
	state.Assets = []domain.HAPWAsset{confirmed, pending, snapshot}
	handler := New(service.New(store.NewMemory(state)), "../../public", slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, route := range []string{"/api/assets", "/api/studio", "/api/v1/hapw/assets"} {
		status, body := doJSONRequest(t, handler, http.MethodGet, route, nil)
		if status != http.StatusOK {
			t.Fatalf("%s: %d", route, status)
		}
		var response struct {
			Assets []domain.HAPWAsset `json:"assets"`
			Items  []domain.HAPWAsset `json:"items"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatal(err)
		}
		assets := append(response.Assets, response.Items...)
		if len(assets) != 1 || assets[0].ID != confirmed.ID {
			t.Fatalf("%s: %#v", route, assets)
		}
	}
}
func TestMigrationRequestReadAndReviewAccess(t *testing.T) {
	t.Setenv("CLIPLI_ADMIN_API_KEY", "review-only-test-key")
	handler := testHandler(t)
	for _, route := range []string{"/api/v1/integrations/platform/users/anonymous-demo/migration-requests"} {
		status, body := doJSONRequest(t, handler, http.MethodGet, route, nil)
		if status != http.StatusOK || !bytes.Contains(body, []byte(`"items":[]`)) {
			t.Fatalf("request history status=%d body=%s", status, body)
		}
	}
	status, _ := doJSONRequest(t, handler, http.MethodGet, "/api/v1/integrations/platform/users/another-user/migration-requests", nil)
	if status != http.StatusForbidden {
		t.Fatalf("another user's history status=%d", status)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/v1/admin/migration-requests/request-1/review", map[string]any{"action": "approve", "accepted": true})
	if status != http.StatusUnauthorized {
		t.Fatalf("unprotected review status=%d", status)
	}
}
