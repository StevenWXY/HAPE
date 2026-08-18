package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StevenWXY/HAPE/internal/service"
	"github.com/StevenWXY/HAPE/internal/store"
)

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(service.New(store.NewMemory(store.SeedState())), t.TempDir(), logger)
}

func TestHAPWAssetAPI(t *testing.T) {
	handler := testHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/hapw/assets/asset-2048", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Asset struct {
				ID            string `json:"id"`
				Authorization struct {
					UsageTypes []string `json:"usageTypes"`
				} `json:"authorization"`
			} `json:"asset"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.Asset.ID != "asset-2048" || len(payload.Data.Asset.Authorization.UsageTypes) == 0 {
		t.Fatalf("payload = %#v", payload)
	}
	if recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("security headers missing")
	}
}

func TestHAPWListFilters(t *testing.T) {
	handler := testHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/hapw/assets?redeemable=true&pageSize=2&sort=-value", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Items      []any `json:"items"`
			Pagination struct {
				TotalItems int `json:"totalItems"`
			} `json:"pagination"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Data.Items) != 2 || payload.Data.Pagination.TotalItems != 4 {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestCompatibilityRedemptionEndpoint(t *testing.T) {
	handler := testHandler(t)
	body := []byte(`{"requestId":"api-redeem-2048","assetId":"asset-2048","accepted":true}`)
	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/api/hapw/redemptions", bytes.NewReader(body)))
	if first.Code != http.StatusCreated {
		t.Fatalf("first status = %d body=%s", first.Code, first.Body.String())
	}
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/api/hapw/redemptions", bytes.NewReader(body)))
	if second.Code != http.StatusOK {
		t.Fatalf("second status = %d body=%s", second.Code, second.Body.String())
	}
}

func TestUnknownAPIUsesErrorEnvelope(t *testing.T) {
	handler := testHandler(t)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d", recorder.Code)
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"error"`)) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestHAPWReadRoutesAndWriteWorkflow(t *testing.T) {
	handler := testHandler(t)
	readRoutes := []string{
		"/api/health",
		"/api/overview",
		"/api/works",
		"/api/works/work-tide",
		"/api/assets",
		"/api/studio",
		"/api/profile",
		"/api/v1/hapw/summary",
		"/api/v1/hapw/assets",
		"/api/v1/hapw/assets/asset-2048",
		"/api/v1/hapw/assets/asset-2048/authorization",
		"/api/v1/hapw/assets/asset-2048/history",
		"/api/v1/hapw/assets/asset-2048/works",
		"/api/v1/hapw/assets/asset-2048/exchange-quote",
		"/api/v1/hapw/redemptions",
		"/api/v1/hapw/exercises",
		"/api/v1/hapw/exchanges",
		"/api/v1/hapw/exchange-policy",
		"/api/v1/platforms",
		"/api/v1/session",
		"/api/v1/clip/treasury",
		"/api/v1/clip/distribution-rules",
		"/api/v1/clip/distributions",
		"/api/v1/integrations/asset-requirements",
		"/api/v1/integrations/asset-sources",
		"/api/v1/integrations/asset-sources/haiwen",
		"/api/v1/integrations/asset-sync-runs",
	}
	for _, route := range readRoutes {
		status, _ := doJSONRequest(t, handler, http.MethodGet, route, nil)
		if status != http.StatusOK {
			t.Fatalf("GET %s status = %d", route, status)
		}
	}

	status, data := doJSONRequest(t, handler, http.MethodPost, "/api/v1/hapw/redemptions", map[string]any{
		"requestId": "workflow-redeem-2048", "assetId": "asset-2048", "accepted": true,
	})
	if status != http.StatusCreated {
		t.Fatalf("redeem status = %d", status)
	}
	var redemptionResult struct {
		Redemption struct {
			ID string `json:"id"`
		} `json:"redemption"`
	}
	if err := json.Unmarshal(data, &redemptionResult); err != nil || redemptionResult.Redemption.ID == "" {
		t.Fatalf("invalid redemption response: %s", string(data))
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/v1/hapw/redemptions", map[string]any{
		"requestId": "workflow-redeem-2048", "assetId": "asset-2048", "accepted": true,
	})
	if status != http.StatusOK {
		t.Fatalf("idempotent redeem status = %d", status)
	}
	status, _ = doJSONRequest(t, handler, http.MethodGet, "/api/v1/hapw/redemptions/"+redemptionResult.Redemption.ID, nil)
	if status != http.StatusOK {
		t.Fatalf("redemption detail status = %d", status)
	}

	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/generations", map[string]any{
		"requestId": "workflow-generation-2048", "assetId": "asset-2048", "duration": 15, "quality": "standard", "accepted": true,
	})
	if status != http.StatusCreated {
		t.Fatalf("generation status = %d", status)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/conversions", map[string]any{
		"requestId": "workflow-conversion-771", "assetId": "asset-771", "region": "EU", "days": 30, "accepted": true,
	})
	if status != http.StatusCreated {
		t.Fatalf("conversion status = %d", status)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/v1/hapw/exercises", map[string]any{
		"requestId": "workflow-exercise-332", "assetId": "asset-332", "platformCode": "foundation", "accepted": true,
	})
	if status != http.StatusCreated {
		t.Fatalf("exercise status = %d", status)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/v1/hapw/exchanges", map[string]any{
		"requestId": "workflow-exchange-528", "assetId": "asset-528", "accepted": true,
	})
	if status != http.StatusCreated {
		t.Fatalf("exchange status = %d", status)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/profile/wallet", map[string]any{"provider": "MetaMask"})
	if status != http.StatusOK {
		t.Fatalf("wallet status = %d", status)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/profile/overseas", map[string]any{"account": "HWF-TEST-2048"})
	if status != http.StatusOK {
		t.Fatalf("account bind status = %d", status)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPatch, "/api/profile/settings", map[string]any{"walletSign": false})
	if status != http.StatusOK {
		t.Fatalf("settings status = %d", status)
	}
	status, _ = doJSONRequest(t, handler, http.MethodDelete, "/api/profile/overseas", nil)
	if status != http.StatusOK {
		t.Fatalf("account unbind status = %d", status)
	}

	for _, route := range []string{"/api/v1/hapw/redemptions", "/api/v1/hapw/exercises", "/api/v1/hapw/exchanges", "/api/v1/hapw/exchange-policy"} {
		status, _ := doJSONRequest(t, handler, http.MethodGet, route, nil)
		if status != http.StatusOK {
			t.Fatalf("post-workflow GET %s status = %d", route, status)
		}
	}

	status, data = doJSONRequest(t, handler, http.MethodGet, "/api/v1/clip/treasury", nil)
	if status != http.StatusOK || !bytes.Contains(data, []byte(`"conserved":true`)) {
		t.Fatalf("treasury response does not prove conservation: status=%d body=%s", status, data)
	}
}

func doJSONRequest(t *testing.T, handler http.Handler, method, route string, payload any) (int, json.RawMessage) {
	t.Helper()
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		body = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, route, body)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	var envelope struct {
		Data  json.RawMessage `json:"data"`
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("%s %s invalid JSON: %v; body=%s", method, route, err, recorder.Body.String())
	}
	return recorder.Code, envelope.Data
}
