package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/service"
	"github.com/StevenWXY/HAPE/internal/store"
)

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	publicDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(publicDir, "admin.html"), []byte("<html>Clipli admin.js</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	return New(service.New(store.NewMemory(store.SeedState())), publicDir, logger)
}

func testHandlerWithService(t *testing.T, serviceLayer *service.Service) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	publicDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(publicDir, "admin.html"), []byte("<html>Clipli admin.js</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	return New(serviceLayer, publicDir, logger)
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

func TestWalletAndAirdropAPIs(t *testing.T) {
	handler := testHandler(t)
	address := "0x1111111111111111111111111111111111111111"
	status, data := doJSONRequest(t, handler, http.MethodGet, "/api/v1/wallet/assets?address="+address, nil)
	if status != http.StatusOK || !bytes.Contains(data, []byte(`"assetId":"asset-2048"`)) {
		t.Fatalf("wallet assets status=%d body=%s", status, data)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/v1/wallet/connect", map[string]any{"provider": "MetaMask", "address": address, "chainId": "0x61"})
	if status != http.StatusOK {
		t.Fatalf("wallet connect status=%d", status)
	}
	status, data = doJSONRequest(t, handler, http.MethodGet, "/api/v1/airdrop-rules", nil)
	if status != http.StatusOK || !bytes.Contains(data, []byte(`"code":"hapw_redemption"`)) || !bytes.Contains(data, []byte(`"code":"admin_approved"`)) {
		t.Fatalf("airdrop rules status=%d body=%s", status, data)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/v1/hapw/redemptions", map[string]any{"requestId": "wallet-api-redeem", "assetId": "asset-2048", "accepted": true})
	if status != http.StatusCreated {
		t.Fatalf("redemption status=%d", status)
	}
	status, data = doJSONRequest(t, handler, http.MethodGet, "/api/v1/airdrops?address="+address, nil)
	if status != http.StatusOK || !bytes.Contains(data, []byte(`"status":"queued"`)) {
		t.Fatalf("airdrops status=%d body=%s", status, data)
	}
}

func TestAirdropAdminAndExecutorAuth(t *testing.T) {
	t.Setenv("CLIPLI_ADMIN_API_KEY", "admin-test-key")
	t.Setenv("CLIPLI_AIRDROP_EXECUTOR_KEY", "executor-test-key")
	handler := testHandler(t)
	address := "0x2222222222222222222222222222222222222222"
	status, _ := doJSONRequest(t, handler, http.MethodPost, "/api/v1/admin/airdrops", map[string]any{"requestId": "admin-api-airdrop", "ruleCode": "admin_approved", "walletAddress": address, "chainId": "0x61", "amount": 12})
	if status != http.StatusUnauthorized {
		t.Fatalf("missing admin key status=%d", status)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/airdrops", bytes.NewReader([]byte(`{"requestId":"admin-api-airdrop","ruleCode":"admin_approved","walletAddress":"0x2222222222222222222222222222222222222222","chainId":"0x61","amount":12}`)))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Clipli-Admin-Key", "admin-test-key")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("admin create status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil || created.Data.ID == "" {
		t.Fatalf("invalid admin response=%s", recorder.Body.String())
	}
	resultBody := []byte(`{"status":"confirmed","txHash":"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`)
	resultReq := httptest.NewRequest(http.MethodPost, "/api/v1/internal/airdrops/"+created.Data.ID+"/result", bytes.NewReader(resultBody))
	resultReq.Header.Set("Content-Type", "application/json")
	resultReq.Header.Set("X-Clipli-Executor-Key", "executor-test-key")
	resultRecorder := httptest.NewRecorder()
	handler.ServeHTTP(resultRecorder, resultReq)
	if resultRecorder.Code != http.StatusOK || !bytes.Contains(resultRecorder.Body.Bytes(), []byte(`"status":"confirmed"`)) {
		t.Fatalf("executor result status=%d body=%s", resultRecorder.Code, resultRecorder.Body.String())
	}
}

func TestAdminConsoleRoute(t *testing.T) {
	handler := testHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("admin page status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte("Clipli")) || !bytes.Contains(recorder.Body.Bytes(), []byte("admin.js")) {
		t.Fatalf("admin page does not contain its independent shell: %s", recorder.Body.String())
	}
}

func TestBNBAdminSimulation(t *testing.T) {
	t.Setenv("CLIPLI_ADMIN_API_KEY", "admin-test-key")
	handler := testHandler(t)
	address := "0x3333333333333333333333333333333333333333"
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/bnb-networks", nil)
	request.Header.Set("X-Clipli-Admin-Key", "admin-test-key")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !bytes.Contains(recorder.Body.Bytes(), []byte(`"chainId":"0x61"`)) || !bytes.Contains(recorder.Body.Bytes(), []byte(`"simulationOnly":true`)) {
		t.Fatalf("networks status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	create := httptest.NewRequest(http.MethodPost, "/api/v1/admin/airdrops", bytes.NewReader([]byte(`{"requestId":"bnb-sim-0001","ruleCode":"admin_approved","walletAddress":"`+address+`","chainId":"0x61","amount":12}`)))
	create.Header.Set("Content-Type", "application/json")
	create.Header.Set("X-Clipli-Admin-Key", "admin-test-key")
	created := httptest.NewRecorder()
	handler.ServeHTTP(created, create)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var payload struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &payload); err != nil || payload.Data.ID == "" {
		t.Fatalf("create payload=%s", created.Body.String())
	}
	simulate := httptest.NewRequest(http.MethodPost, "/api/v1/admin/airdrops/"+payload.Data.ID+"/simulate", bytes.NewReader([]byte(`{"chainId":"0x61"}`)))
	simulate.Header.Set("Content-Type", "application/json")
	simulate.Header.Set("X-Clipli-Admin-Key", "admin-test-key")
	simulated := httptest.NewRecorder()
	handler.ServeHTTP(simulated, simulate)
	if simulated.Code != http.StatusOK || !bytes.Contains(simulated.Body.Bytes(), []byte(`"simulated":true`)) || !bytes.Contains(simulated.Body.Bytes(), []byte(`"executionMode":"simulated-bnb"`)) {
		t.Fatalf("simulate status=%d body=%s", simulated.Code, simulated.Body.String())
	}
}

func TestExternalPlatformBindingAndRedemptionRoutes(t *testing.T) {
	handler := testHandler(t)
	status, data := doJSONRequest(t, handler, http.MethodPost, "/api/v1/integrations/platform/verification-codes", map[string]any{
		"userId": "clip-user-http", "phone": "13800138000",
	})
	if status != http.StatusAccepted || !bytes.Contains(data, []byte(`"verificationId"`)) || bytes.Contains(data, []byte(`"code"`)) {
		t.Fatalf("verification status=%d body=%s", status, data)
	}
	var challenge struct {
		Challenge struct {
			ID string `json:"id"`
		} `json:"challenge"`
	}
	if err := json.Unmarshal(data, &challenge); err != nil || challenge.Challenge.ID == "" {
		t.Fatalf("verification payload=%s", data)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/v1/integrations/platform/bindings", map[string]any{
		"userId": "clip-user-http", "phone": "13800138000", "code": "123456", "verificationId": challenge.Challenge.ID,
	})
	if status != http.StatusCreated {
		t.Fatalf("binding status=%d", status)
	}
	status, data = doJSONRequest(t, handler, http.MethodGet, "/api/v1/integrations/platform/users/clip-user-http/assets", nil)
	if status != http.StatusOK || !bytes.Contains(data, []byte(`"totalQuantity":6`)) {
		t.Fatalf("assets status=%d body=%s", status, data)
	}
	status, data = doJSONRequest(t, handler, http.MethodGet, "/api/v1/integrations/platform/users/clip-user-http/assets?tplIds=100001,100002", nil)
	if status != http.StatusOK || !bytes.Contains(data, []byte(`"tplId":100001`)) || !bytes.Contains(data, []byte(`"count":1`)) {
		t.Fatalf("asset counts status=%d body=%s", status, data)
	}
	status, data = doJSONRequest(t, handler, http.MethodPost, "/api/v1/integrations/platform/users/clip-user-http/assets/asset-2048/redemptions", map[string]any{
		"serialNo": "serial-http-0001",
	})
	if status != http.StatusCreated || !bytes.Contains(data, []byte(`"status":"completed"`)) {
		t.Fatalf("redemption status=%d body=%s", status, data)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/v1/integrations/platform/users/clip-user-http/assets/asset-2048/redemptions", map[string]any{
		"serialNumber": "serial-http-0001",
	})
	if status != http.StatusOK {
		t.Fatalf("idempotent redemption status=%d", status)
	}
	status, data = doJSONRequest(t, handler, http.MethodPost, "/api/v1/integrations/platform/redemptions", map[string]any{
		"userId": "clip-user-http", "requestNo": "WO-HTTP-0002", "tplId": 100001, "num": 2,
	})
	if status != http.StatusCreated || !bytes.Contains(data, []byte(`"requestNo":"WO-HTTP-0002"`)) || !bytes.Contains(data, []byte(`"num":2`)) {
		t.Fatalf("Haiwen redemption status=%d body=%s", status, data)
	}
}

func TestHaiwenTemplateAndMigrationRoutes(t *testing.T) {
	t.Setenv("CLIPLI_ADMIN_API_KEY", "admin-test-key")
	serviceLayer := service.New(store.NewMemory(store.SeedState()))
	if err := serviceLayer.SetExternalAssetMappings([]domain.ExternalAssetMappingRule{{TplID: 100001, Version: "test-v1", CreditYield: 150, ClipPrice: 520, Currency: "CNY", Active: true}}); err != nil {
		t.Fatal(err)
	}
	handler := testHandlerWithService(t, serviceLayer)
	status, data := doJSONRequest(t, handler, http.MethodGet, "/api/v1/integrations/platform/templates?page=1&pageSize=20", nil)
	if status != http.StatusOK || !bytes.Contains(data, []byte(`"tplId":100001`)) || !bytes.Contains(data, []byte(`"migrationReady":true`)) {
		t.Fatalf("templates status=%d body=%s", status, data)
	}
	status, data = doJSONRequest(t, handler, http.MethodPost, "/api/v1/integrations/platform/verification-codes", map[string]any{"userId": "clip-migration-http", "phone": "13800138000"})
	if status != http.StatusAccepted {
		t.Fatalf("verification status=%d body=%s", status, data)
	}
	var challenge struct {
		Challenge struct {
			ID string `json:"id"`
		} `json:"challenge"`
	}
	if err := json.Unmarshal(data, &challenge); err != nil {
		t.Fatal(err)
	}
	status, data = doJSONRequest(t, handler, http.MethodPost, "/api/v1/integrations/platform/bindings", map[string]any{"userId": "clip-migration-http", "phone": "13800138000", "code": "123456", "verificationId": challenge.Challenge.ID})
	if status != http.StatusCreated {
		t.Fatalf("binding status=%d body=%s", status, data)
	}
	input := map[string]any{"tplId": 100001, "num": 1, "requestNo": "HTTP-MIGRATION-100001", "requestId": "http-migration-100001", "accepted": true}
	status, data = doJSONRequest(t, handler, http.MethodPost, "/api/v1/integrations/platform/users/clip-migration-http/migrations/preview", input)
	if status != http.StatusOK || !bytes.Contains(data, []byte(`"ready":true`)) || !bytes.Contains(data, []byte(`"pending_external_write_off"`)) {
		t.Fatalf("preview status=%d body=%s", status, data)
	}
	status, _ = doJSONRequest(t, handler, http.MethodPost, "/api/v1/integrations/platform/users/clip-migration-http/migrations", input)
	if status != http.StatusUnauthorized {
		t.Fatalf("migration without admin key status=%d", status)
	}
	encoded, _ := json.Marshal(input)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/platform/users/clip-migration-http/migrations", bytes.NewReader(encoded))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Clipli-Admin-Key", "admin-test-key")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated || !bytes.Contains(recorder.Body.Bytes(), []byte(`"status":"completed_rewards_settled"`)) || !bytes.Contains(recorder.Body.Bytes(), []byte(`"sourceTemplate"`)) || !bytes.Contains(recorder.Body.Bytes(), []byte(`"creditsGranted":150`)) || !bytes.Contains(recorder.Body.Bytes(), []byte(`"clipGranted":60`)) {
		t.Fatalf("migration status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	status, data = doJSONRequest(t, handler, http.MethodGet, "/api/v1/integrations/platform/users/clip-migration-http/migrations", nil)
	if status != http.StatusOK || !bytes.Contains(data, []byte(`"mappingVersion":"test-v1"`)) {
		t.Fatalf("migration list status=%d body=%s", status, data)
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
