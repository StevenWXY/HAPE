package service

import (
	"context"
	"errors"
	"github.com/StevenWXY/HAPE/internal/testfixture"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/store"
)

type integrationRoundTripper func(*http.Request) (*http.Response, error)

func (f integrationRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestHTTPExternalPlatformAcceptsDataEnvelope(t *testing.T) {
	var requests []*http.Request
	client := &http.Client{Transport: integrationRoundTripper(func(request *http.Request) (*http.Response, error) {
		requests = append(requests, request)
		var body string
		switch {
		case strings.HasSuffix(request.URL.Path, "/bind/sms"):
			body = `{"code":0,"message":"success","data":{},"msg":"success"}`
		case strings.HasSuffix(request.URL.Path, "/bind/status"):
			body = `{"code":0,"data":{"externalUserId":"external-1","bound":true,"boundAt":"2026-08-26T10:00:00+08:00"}}`
		case strings.HasSuffix(request.URL.Path, "/bind"):
			body = `{"code":0,"message":"success","data":{"externalUserId":"external-1","bound":true,"boundAt":"2026-08-26T10:00:00+08:00"},"msg":"success"}`
		case strings.HasSuffix(request.URL.Path, "/assets/count"):
			body = `{"code":0,"message":"success","data":{"externalUserId":"external-1","list":[{"tplId":100001,"count":3}]},"msg":"success"}`
		case strings.HasSuffix(request.URL.Path, "/works"):
			body = `{"code":0,"message":"success","data":{"list":[{"workId":100001,"worksName":"作品名称","showcase":["https://cdn.example.com/work.png"],"authors":[{"id":1,"name":"作者"}],"owners":[],"worksType":1,"worksSubType":null,"worksTypeName":"影像","worksIntroduce":"<p>作品描述</p>","publishNum":10}],"total":1,"pageNum":1,"pageSize":20},"msg":"success"}`
		case strings.HasSuffix(request.URL.Path, "/tpls"):
			body = `{"code":0,"message":"success","data":{"list":[{"tplId":100001,"name":"版权名称","description":"<p>版权描述</p>","image":"https://cdn.example.com/template.png","workId":100001,"worksName":"作品名称","worksType":1,"worksSubType":null,"worksTypeName":"影像","authors":[{"id":1,"name":"作者"}],"owners":[{"id":2,"name":"权利人"}],"publishCount":1000}],"total":1,"pageNum":1,"pageSize":20},"msg":"success"}`
		case strings.HasSuffix(request.URL.Path, "/write-off"):
			body = `{"code":0,"message":"success","data":{"requestNo":"wo-1","externalUserId":"external-1","tplId":100001,"num":1,"status":"SUCCESS","writeOffAt":"2026-08-26T10:00:00+08:00"},"msg":"success"}`
		default:
			body = `{"code":0,"message":"success","data":{},"msg":"success"}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: request}, nil
	})}
	platform := NewHTTPExternalPlatform("https://partner.example/api", client)
	platform.AppID = "100001"
	platform.AppKey = "secret"
	dispatch, err := platform.SendVerificationCode(context.Background(), VerificationCodeRequest{UserID: "u-1", ExternalUserID: "external-1", Phone: "13800138000"})
	if err != nil {
		t.Fatalf("dispatch=%#v err=%v", dispatch, err)
	}
	binding, err := platform.BindUser(context.Background(), BindUserRequest{UserID: "u-1", ExternalUserID: "external-1", Phone: "13800138000", Code: "123456"})
	if err != nil || binding.ExternalUserID != "external-1" {
		t.Fatalf("binding=%#v err=%v", binding, err)
	}
	assets, err := platform.ListAssetCounts(context.Background(), "external-1", []int64{100001})
	if err != nil || len(assets) != 1 || assets[0].TplID != 100001 || assets[0].Count != 3 {
		t.Fatalf("assets=%#v err=%v", assets, err)
	}
	templates, err := platform.ListTemplates(context.Background(), 1, 20, 0)
	if err != nil || len(templates.Items) != 1 || templates.Items[0].TplID != 100001 || templates.Items[0].Owners[0].Name != "权利人" {
		t.Fatalf("templates=%#v err=%v", templates, err)
	}
	works, err := platform.ListWorks(context.Background(), 1, 20)
	if err != nil || len(works.Items) != 1 || works.Items[0].WorkID != 100001 || works.Items[0].PublishNum != 10 {
		t.Fatalf("works=%#v err=%v", works, err)
	}
	status, err := platform.GetBindingStatus(context.Background(), "external-1")
	if err != nil || !status.Bound || status.ExternalUserID != "external-1" {
		t.Fatalf("status=%+v err=%v", status, err)
	}
	redeemed, err := platform.WriteOffTemplate(context.Background(), ExternalTemplateWriteOffRequest{ExternalUserID: "external-1", TplID: 100001, Num: 1, RequestNo: "wo-1"})
	if err != nil || redeemed.RequestNo != "wo-1" || redeemed.TplID != 100001 || redeemed.Num != 1 || len(requests) != 7 {
		t.Fatalf("redeemed=%#v requests=%d err=%v", redeemed, len(requests), err)
	}
	if requests[0].Header.Get("x-app-id") != "100001" || requests[0].Header.Get("x-app-key") != "secret" {
		t.Fatalf("Haiwen credentials missing: appid=%q appkey=%q", requests[0].Header.Get("x-app-id"), requests[0].Header.Get("x-app-key"))
	}
	if requests[2].URL.Query().Get("externalUserId") != "external-1" || requests[2].URL.Query().Get("tplIds") != "100001" {
		t.Fatalf("asset count query=%s", requests[2].URL.RawQuery)
	}
}

type migrationFake struct {
	template domain.ExternalAssetTemplate
	count    int
	redeems  int
	err      error
}

func (f *migrationFake) SendVerificationCode(context.Context, VerificationCodeRequest) (VerificationCodeDispatch, error) {
	return VerificationCodeDispatch{}, nil
}
func (f *migrationFake) BindUser(context.Context, BindUserRequest) (ExternalBindingResult, error) {
	return ExternalBindingResult{}, nil
}
func (f *migrationFake) ListAssets(context.Context, string) ([]domain.ExternalAssetHolding, error) {
	return []domain.ExternalAssetHolding{{AssetID: "100053", Quantity: f.count}}, nil
}
func (f *migrationFake) ListAssetCounts(context.Context, string, []int64) ([]domain.ExternalTemplateCount, error) {
	return []domain.ExternalTemplateCount{{TplID: 100053, Count: f.count}}, nil
}
func (f *migrationFake) ListTemplates(context.Context, int, int, int64) (ExternalTemplatePage, error) {
	return ExternalTemplatePage{Items: []domain.ExternalAssetTemplate{f.template}, Total: 1, Page: 1, PageSize: 20}, nil
}
func (f *migrationFake) RedeemAsset(_ context.Context, input ExternalRedemptionRequest) (ExternalRedemptionResult, error) {
	f.redeems++
	if f.err != nil {
		return ExternalRedemptionResult{}, f.err
	}
	return ExternalRedemptionResult{ExternalTxID: "haiwen-write-off-1", Status: "SUCCESS", TplID: input.TplID, Num: input.Num, Quantity: input.Num, RequestNo: input.RequestNo}, nil
}

func (f *migrationFake) WriteOffTemplate(_ context.Context, input ExternalTemplateWriteOffRequest) (ExternalRedemptionResult, error) {
	f.redeems++
	if f.err != nil {
		return ExternalRedemptionResult{}, f.err
	}
	return ExternalRedemptionResult{ExternalTxID: "haiwen-write-off-1", Status: "SUCCESS", TplID: input.TplID, Num: input.Num, Quantity: input.Num, RequestNo: input.RequestNo}, nil
}

func TestHaiwenTemplateMigrationCreatesMatchingClipliAssetBeforeLocalRedemption(t *testing.T) {
	workType := int64(100002)
	fake := &migrationFake{count: 2, template: domain.ExternalAssetTemplate{
		TplID: 100053, Name: "测试222", Description: "<p>123123</p>", Image: "https://cdn.hnccc.com/template.jpeg",
		WorkID: 100002, WorksName: "海直播百部短剧", WorksType: &workType,
		Authors: []domain.ExternalParty{{ID: 100001, Name: "海直播传媒(海南)有限公司"}}, Owners: []domain.ExternalParty{{ID: 100000, Name: "海直播传媒(海南)有限公司"}}, PublishCount: 10000,
	}}
	state := testfixture.SeedState()
	state.Bindings = []domain.ExternalPlatformBinding{{ID: "binding-migration", UserID: "clip-user-migration", ExternalUserID: "100001", PlatformCode: "haiwen", Status: "bound"}}
	svc := NewWithExternalPlatform(store.NewMemory(state), fake)
	svc.now = func() time.Time { return time.Date(2026, 9, 4, 8, 30, 0, 0, time.UTC) }
	if err := svc.SetExternalAssetMappings([]domain.ExternalAssetMappingRule{{TplID: 100053, Version: "haiwen-2026-09-v1", CreditYield: 150, ClipPrice: 520, Currency: "CNY", Active: true}}); err != nil {
		t.Fatal(err)
	}
	input := MigrateExternalAssetInput{UserID: "clip-user-migration", TplID: 100053, Num: 1, RequestNo: "HW-MIGRATION-100053-01", RequestID: "migration-100053-01", Accepted: true}
	preview, err := svc.PreviewExternalAssetMigration(input)
	if err != nil || !preview.Ready || preview.AvailableQuantity != 2 || len(preview.CandidateAssets) != 1 {
		t.Fatalf("preview=%#v err=%v", preview, err)
	}
	if preview.CandidateAssets[0].RedemptionStatus != "pending_external_write_off" || preview.CandidateAssets[0].SourceTemplate.Description != fake.template.Description {
		t.Fatalf("candidate=%#v", preview.CandidateAssets[0])
	}
	before := svc.Snapshot()
	result, idempotent, err := svc.MigrateExternalAsset(input)
	if err != nil || idempotent || fake.redeems != 1 || len(result.Assets) != 1 {
		t.Fatalf("result=%#v idempotent=%v redeems=%d err=%v", result, idempotent, fake.redeems, err)
	}
	asset := result.Assets[0]
	if result.Migration.Status != "completed_rewards_settled" || result.Migration.CreditsGranted != 150 || result.Migration.ClipGranted != 60 || asset.Name != fake.template.Name || asset.Media.PreviewURL != fake.template.Image || asset.RedemptionStatus != "redeemed" || asset.CreditYield != 150 {
		t.Fatalf("migration=%#v asset=%#v", result.Migration, asset)
	}
	afterMigration := svc.Snapshot()
	if afterMigration.GenerationAccount.Balance != before.GenerationAccount.Balance+150 || afterMigration.CLIP.Balance != before.CLIP.Balance+60 {
		t.Fatalf("migration did not settle rewards: before=%#v after=%#v", before.GenerationAccount, afterMigration.GenerationAccount)
	}
	retry, idempotent, err := svc.MigrateExternalAsset(input)
	if err != nil || !idempotent || retry.Migration.ID != result.Migration.ID || fake.redeems != 1 {
		t.Fatalf("retry=%#v idempotent=%v redeems=%d err=%v", retry, idempotent, fake.redeems, err)
	}
	if _, _, err := svc.Redeem(RedemptionInput{RequestID: "clipli-redeem-100053-01", AssetID: asset.ID, Accepted: true}); apiErrorCode(err) != "hapw_not_redeemable" {
		t.Fatalf("settled migrated asset was redeemable again: err=%v", err)
	}
}

func TestHaiwenMigrationRequiresExplicitMappingAndKeepsFailedAssetInactive(t *testing.T) {
	fake := &migrationFake{count: 1, template: domain.ExternalAssetTemplate{TplID: 100053, Name: "测试222", WorkID: 100002, WorksName: "海直播百部短剧", Owners: []domain.ExternalParty{{ID: 1, Name: "权利人"}}, PublishCount: 10000}}
	state := testfixture.SeedState()
	state.Bindings = []domain.ExternalPlatformBinding{{ID: "binding-migration", UserID: "clip-user-migration", ExternalUserID: "100001", PlatformCode: "haiwen", Status: "bound"}}
	svc := NewWithExternalPlatform(store.NewMemory(state), fake)
	input := MigrateExternalAssetInput{UserID: "clip-user-migration", TplID: 100053, Num: 1, RequestNo: "HW-MIGRATION-100053-FAIL", RequestID: "migration-100053-fail", Accepted: true}
	if _, _, err := svc.MigrateExternalAsset(input); apiErrorCode(err) != "external_asset_mapping_required" {
		t.Fatalf("unmapped migration error=%v", err)
	}
	if err := svc.SetExternalAssetMappings([]domain.ExternalAssetMappingRule{{TplID: 100053, Version: "v1", CreditYield: 150, Active: true}}); err != nil {
		t.Fatal(err)
	}
	fake.err = &PlatformError{Status: http.StatusUnprocessableEntity, Code: "external_assets_insufficient", Message: "write-off failed"}
	if _, _, err := svc.MigrateExternalAsset(input); apiErrorCode(err) != "external_assets_insufficient" {
		t.Fatalf("failed migration error=%v", err)
	}
	migrations, _ := svc.ExternalMigrations("clip-user-migration")
	if len(migrations) != 1 || migrations[0].Status != "external_write_off_failed" {
		t.Fatalf("migrations=%#v", migrations)
	}
	if _, visible := svc.GetAsset(migrations[0].ClipliAssetIDs[0]); visible {
		t.Fatal("failed migration must not be visible in portfolio")
	}
	asset, ok := findAsset(svc.Snapshot().Assets, migrations[0].ClipliAssetIDs[0])
	if !ok || asset.Transferable || asset.RedemptionStatus != "external_write_off_failed" {
		t.Fatalf("failed staged asset=%#v", asset)
	}
}

func TestHaiwenMigrationChecksClipTreasuryBeforeWriteOff(t *testing.T) {
	fake := &migrationFake{count: 1, template: domain.ExternalAssetTemplate{TplID: 100053, Name: "测试222", WorkID: 100002, WorksName: "海直播百部短剧", Owners: []domain.ExternalParty{{ID: 1, Name: "权利人"}}, PublishCount: 10000}}
	state := testfixture.SeedState()
	state.Bindings = []domain.ExternalPlatformBinding{{ID: "binding-migration", UserID: "clip-user-migration", ExternalUserID: "100001", PlatformCode: "haiwen", Status: "bound"}}
	state.CLIPTreasury.TreasuryBalance = 0
	svc := NewWithExternalPlatform(store.NewMemory(state), fake)
	if err := svc.SetExternalAssetMappings([]domain.ExternalAssetMappingRule{{TplID: 100053, Version: "v1", CreditYield: 150, Active: true}}); err != nil {
		t.Fatal(err)
	}
	input := MigrateExternalAssetInput{UserID: "clip-user-migration", TplID: 100053, Num: 1, RequestNo: "HW-MIGRATION-TREASURY-01", RequestID: "migration-treasury-01", Accepted: true}
	if _, _, err := svc.MigrateExternalAsset(input); apiErrorCode(err) != "clip_treasury_insufficient" {
		t.Fatalf("treasury preflight error=%v", err)
	}
	if fake.redeems != 0 || len(svc.Snapshot().ExternalMigrations) != 0 || len(svc.Snapshot().Assets) != len(state.Assets) {
		t.Fatalf("irreversible write-off was attempted before treasury preflight: redeems=%d", fake.redeems)
	}
}

func TestHTTPExternalPlatformMapsHaiwenBusinessErrors(t *testing.T) {
	client := &http.Client{Transport: integrationRoundTripper(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"code":422,"message":"可核销资产不足","data":null,"msg":"可核销资产不足"}`)), Header: make(http.Header), Request: request}, nil
	})}
	platform := NewHTTPExternalPlatform("https://partner.example/api", client)
	platform.AppID, platform.AppKey = "100001", "secret"
	_, err := platform.WriteOffTemplate(context.Background(), ExternalTemplateWriteOffRequest{ExternalUserID: "external-1", TplID: 100001, Num: 1, RequestNo: "wo-2"})
	var platformErr *PlatformError
	if !errors.As(err, &platformErr) || platformErr.Status != http.StatusUnprocessableEntity || platformErr.Code != "external_assets_insufficient" {
		t.Fatalf("error=%#v", err)
	}
}

func TestHTTPExternalPlatformDoesNotInferExternalUserID(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: integrationRoundTripper(func(request *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"code":0,"data":{}}`)), Header: make(http.Header), Request: request}, nil
	})}
	platform := NewHTTPExternalPlatform("https://partner.example/api", client)
	platform.AppID, platform.AppKey = "100001", "secret"
	_, err := platform.SendVerificationCode(context.Background(), VerificationCodeRequest{UserID: "clip-user-1", Phone: "13800138000"})
	var platformErr *PlatformError
	if !errors.As(err, &platformErr) || platformErr.Code != "invalid_external_user_id" {
		t.Fatalf("missing external id error=%#v", err)
	}
	if calls != 0 {
		t.Fatalf("adapter made %d network calls for invalid input", calls)
	}
}

type integrationFake struct {
	sends, binds, lists, redeems int
}

func (f *integrationFake) SendVerificationCode(context.Context, VerificationCodeRequest) (VerificationCodeDispatch, error) {
	f.sends++
	return VerificationCodeDispatch{DeliveryID: "delivery-1", ExpiresAt: time.Now().Add(time.Minute).UTC().Format(time.RFC3339), Status: "sent"}, nil
}

func (f *integrationFake) BindUser(_ context.Context, input BindUserRequest) (ExternalBindingResult, error) {
	f.binds++
	if input.Code != "654321" {
		return ExternalBindingResult{}, &PlatformError{Status: 400, Code: "verification_code_invalid", Message: "bad code"}
	}
	externalID := "partner-user-1"
	if input.UserID == "clip-user-2" {
		externalID = "partner-user-2"
	}
	return ExternalBindingResult{ExternalUserID: externalID, BindingID: "partner-binding-" + externalID, Status: "bound", Bound: true, BoundAt: "2026-08-26T10:00:00Z"}, nil
}

func (f *integrationFake) GetBindingStatus(_ context.Context, externalUserID string) (ExternalBindingStatus, error) {
	return ExternalBindingStatus{ExternalUserID: externalUserID, Bound: true, BoundAt: "2026-08-26T10:00:00Z"}, nil
}

func (f *integrationFake) ListAssets(context.Context, string) ([]domain.ExternalAssetHolding, error) {
	f.lists++
	return []domain.ExternalAssetHolding{{AssetID: "external-asset-1", Quantity: 2}}, nil
}

func (f *integrationFake) RedeemAsset(context.Context, ExternalRedemptionRequest) (ExternalRedemptionResult, error) {
	f.redeems++
	return ExternalRedemptionResult{ExternalTxID: "partner-tx-1", Status: "completed", Quantity: 1}, nil
}

func TestExternalPlatformIntegrationFlow(t *testing.T) {
	fake := &integrationFake{}
	svc := NewWithExternalPlatform(store.NewMemory(testfixture.SeedState()), fake)
	sent, err := svc.SendVerificationCode(SendVerificationInput{UserID: "clip-user-1", ExternalUserID: "partner-user-1", Phone: "138-0013-8000", RequestID: "verify-clip-user-1"})
	if err != nil || sent.Challenge.ID == "" || sent.Challenge.Phone != "13800138000" || sent.Challenge.PhoneMasked != "*******8000" {
		t.Fatalf("send result=%#v err=%v", sent, err)
	}
	binding, idempotent, err := svc.BindUser(BindUserInput{UserID: "clip-user-1", ExternalUserID: "partner-user-1", Phone: "13800138000", Code: "654321", VerificationID: sent.Challenge.ID, RequestID: "bind-clip-user-1"})
	if err != nil || idempotent || binding.Binding.ExternalUserID != "partner-user-1" {
		t.Fatalf("bind result=%#v idempotent=%v err=%v", binding, idempotent, err)
	}
	// The second step may submit only the code; the phone is recovered from
	// the pending verification challenge.
	secondSent, err := svc.SendVerificationCode(SendVerificationInput{UserID: "clip-user-2", ExternalUserID: "partner-user-2", Phone: "13900139000"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.BindUser(BindUserInput{UserID: "clip-user-2", ExternalUserID: "partner-user-2", Code: "654321", VerificationID: secondSent.Challenge.ID}); err != nil {
		t.Fatalf("bind without phone: %v", err)
	}
	holdings, err := svc.UserAssets("clip-user-1")
	if err != nil || holdings.TotalQuantity != 2 || fake.lists != 1 {
		t.Fatalf("holdings=%#v err=%v", holdings, err)
	}
	redemption, idempotent, err := svc.RedeemExternalAsset(RedeemExternalAssetInput{UserID: "clip-user-1", AssetID: "external-asset-1", SerialNumber: "serial-0001", RequestID: "redeem-clip-user-1"})
	if err != nil || idempotent || redemption.Redemption.ExternalTxID != "partner-tx-1" {
		t.Fatalf("redemption=%#v idempotent=%v err=%v", redemption, idempotent, err)
	}
	retry, idempotent, err := svc.RedeemExternalAsset(RedeemExternalAssetInput{UserID: "clip-user-1", AssetID: "external-asset-1", SerialNo: "serial-0001", RequestID: "redeem-clip-user-1"})
	if err != nil || !idempotent || retry.Redemption.ID != redemption.Redemption.ID || fake.redeems != 1 {
		t.Fatalf("retry=%#v idempotent=%v redeems=%d err=%v", retry, idempotent, fake.redeems, err)
	}
	if _, _, err := svc.RedeemExternalAsset(RedeemExternalAssetInput{UserID: "clip-user-1", AssetID: "other-asset", SerialNumber: "serial-0001"}); err == nil {
		t.Fatal("expected serial reuse conflict")
	}
}

func TestVerificationChallengesAreScopedToExternalUser(t *testing.T) {
	fake := &integrationFake{}
	svc := NewWithExternalPlatform(store.NewMemory(testfixture.SeedState()), fake)
	first, err := svc.SendVerificationCode(SendVerificationInput{UserID: "clip-user-1", ExternalUserID: "partner-user-1", Phone: "13800138000", RequestID: "verify-shared"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SendVerificationCode(SendVerificationInput{UserID: "clip-user-1", ExternalUserID: "partner-user-2", Phone: "13800138000", RequestID: "verify-shared"}); apiErrorCode(err) != "request_id_reused" {
		t.Fatalf("request id reused across external users: %v", err)
	}
	if _, _, err := svc.BindUser(BindUserInput{UserID: "clip-user-1", ExternalUserID: "partner-user-2", Code: "654321"}); apiErrorCode(err) != "verification_required" {
		t.Fatalf("cross-user challenge was selected: %v", err)
	}
	if _, _, err := svc.BindUser(BindUserInput{UserID: "clip-user-1", ExternalUserID: "partner-user-1", Code: "654321", VerificationID: first.Challenge.ID}); err != nil {
		t.Fatalf("matching challenge failed: %v", err)
	}
}

func TestDemoExternalPlatformHonorsExplicitExternalUserID(t *testing.T) {
	platform := NewDemoExternalPlatform()
	svc := NewWithExternalPlatform(store.NewMemory(testfixture.SeedState()), platform)
	sent, err := svc.SendVerificationCode(SendVerificationInput{UserID: "clip-demo", ExternalUserID: "haiwen-100001", Phone: "13800138000"})
	if err != nil {
		t.Fatal(err)
	}
	result, _, err := svc.BindUser(BindUserInput{UserID: "clip-demo", ExternalUserID: "haiwen-100001", Code: "123456", VerificationID: sent.Challenge.ID})
	if err != nil {
		t.Fatalf("demo binding failed: %v", err)
	}
	if result.Binding.ExternalUserID != "haiwen-100001" {
		t.Fatalf("external user id=%q", result.Binding.ExternalUserID)
	}
	if _, _, err := svc.GetBinding("clip-demo"); err != nil {
		t.Fatalf("binding status read failed: %v", err)
	}
}

func newTestService() *Service {
	service := New(store.NewMemory(testfixture.SeedState()))
	service.now = func() time.Time { return time.Date(2026, 8, 17, 8, 30, 0, 123, time.UTC) }
	return service
}

func TestRedeemIsIdempotent(t *testing.T) {
	service := newTestService()
	input := RedemptionInput{RequestID: "test-redeem-2048", AssetID: "asset-2048", Accepted: true}
	first, repeated, err := service.Redeem(input)
	if err != nil || repeated {
		t.Fatalf("first redeem repeated=%v err=%v", repeated, err)
	}
	if first.Redemption.CreditsGranted != 150 || first.Redemption.ClipGranted != 60 {
		t.Fatalf("redemption = %#v", first.Redemption)
	}
	second, repeated, err := service.Redeem(input)
	if err != nil || !repeated {
		t.Fatalf("second redeem repeated=%v err=%v", repeated, err)
	}
	if second.Redemption.ID != first.Redemption.ID || second.ClipBalance != first.ClipBalance {
		t.Fatalf("idempotent response changed: first=%#v second=%#v", first, second)
	}
	state := service.Snapshot()
	if len(state.Distributions) != 2 || state.Distributions[0].Amount != 60 {
		t.Fatalf("distribution ledger = %#v", state.Distributions)
	}
	if state.CLIPTreasury.TreasuryBalance+state.CLIPTreasury.LiquidityAllocation+state.CLIPTreasury.LedgerOutstanding != state.CLIPTreasury.MintedSupply {
		t.Fatalf("CLIP supply is not conserved: %#v", state.CLIPTreasury)
	}
}

func TestGenerateConsumesMatchingBalances(t *testing.T) {
	service := newTestService()
	_, _, err := service.Redeem(RedemptionInput{RequestID: "test-redeem-2048", AssetID: "asset-2048", Accepted: true})
	if err != nil {
		t.Fatal(err)
	}
	before := service.Snapshot()
	result, repeated, err := service.Generate(GenerationInput{RequestID: "test-generate-2048", AssetID: "asset-2048", Duration: 15, Quality: "pro", Accepted: true})
	if err != nil || repeated {
		t.Fatalf("generate repeated=%v err=%v", repeated, err)
	}
	if result.Generation.CreditsUsed != 23 || result.Generation.ClipCost != 5 {
		t.Fatalf("generation = %#v", result.Generation)
	}
	if result.GenerationAccount.Balance != before.GenerationAccount.Balance-23 || result.ClipBalance != before.CLIP.Balance-5 {
		t.Fatalf("balances before=%#v result=%#v", before.GenerationAccount, result)
	}
	after := service.Snapshot()
	if after.CLIPTreasury.TotalReclaimed != before.CLIPTreasury.TotalReclaimed+5 || after.CLIPTreasury.TreasuryBalance != before.CLIPTreasury.TreasuryBalance+5 {
		t.Fatalf("treasury did not receive generation fee: before=%#v after=%#v", before.CLIPTreasury, after.CLIPTreasury)
	}
}

func TestAllSeedAssetsHaveExternalReferences(t *testing.T) {
	for _, asset := range newTestService().Snapshot().Assets {
		if asset.External.ProviderCode == "" || asset.External.ProviderAssetID == "" || asset.External.SyncStatus != "migrated" || asset.External.TransferID == "" {
			t.Fatalf("asset lacks external provenance: %#v", asset)
		}
	}
}

func TestListAssetsFiltersAndPaginates(t *testing.T) {
	service := newTestService()
	trueValue := true
	result := service.ListAssets(AssetListOptions{Redeemable: &trueValue, Page: 1, PageSize: 2, Sort: "-value"})
	if result.Pagination.TotalItems != 4 || len(result.Items) != 2 {
		t.Fatalf("result = %#v", result)
	}
	if result.Items[0].ID != "asset-2048" {
		t.Fatalf("first asset = %s", result.Items[0].ID)
	}
}

func TestWalletConnectionAndAssetSnapshot(t *testing.T) {
	service := newTestService()
	address := "0x1111111111111111111111111111111111111111"
	profile, err := service.ConnectWallet(WalletConnectInput{Provider: "MetaMask", Address: address, ChainID: "0x61"})
	if err != nil {
		t.Fatal(err)
	}
	if profile.Wallet != address || profile.WalletStatus != "connected" || profile.WalletChainID != "0x61" {
		t.Fatalf("profile = %#v", profile)
	}
	assets, err := service.WalletAssets(address)
	if err != nil || len(assets) != 1 || assets[0].AssetID != "asset-2048" || !assets[0].CanRedeem {
		t.Fatalf("wallet assets = %#v err=%v", assets, err)
	}
}

func TestRedemptionQueuesWalletAirdrop(t *testing.T) {
	service := newTestService()
	address := "0x1111111111111111111111111111111111111111"
	if _, err := service.ConnectWallet(WalletConnectInput{Provider: "MetaMask", Address: address, ChainID: "0x61"}); err != nil {
		t.Fatal(err)
	}
	result, _, err := service.Redeem(RedemptionInput{RequestID: "airdrop-redeem-2048", AssetID: "asset-2048", Accepted: true})
	if err != nil {
		t.Fatal(err)
	}
	items, err := service.Airdrops(address, "queued")
	if err != nil || len(items) != 1 {
		t.Fatalf("airdrops = %#v err=%v", items, err)
	}
	if items[0].AssetID != result.Redemption.AssetID || items[0].Amount != result.Redemption.ClipGranted || items[0].AllocationSource != "redemption-entitlement" {
		t.Fatalf("airdrop = %#v", items[0])
	}
	eligibility, err := service.AirdropEligibility("hapw_redemption", address, "asset-2048")
	if err != nil || eligibility.Eligible || eligibility.Reason != "airdrop_already_created" {
		t.Fatalf("eligibility = %#v err=%v", eligibility, err)
	}
}

func TestAdminAirdropReservesAndRefundsTreasury(t *testing.T) {
	service := newTestService()
	address := "0x2222222222222222222222222222222222222222"
	before := service.Snapshot().CLIPTreasury
	item, idempotent, err := service.CreateAirdrop(AirdropInput{RequestID: "manual-airdrop-1", RuleCode: "admin_approved", WalletAddress: address, ChainID: "0x61", Amount: 25})
	if err != nil || idempotent || item.Status != "queued" || item.Amount != 25 || item.AllocationSource != "treasury-reservation" {
		t.Fatalf("item=%#v idempotent=%v err=%v", item, idempotent, err)
	}
	afterQueue := service.Snapshot().CLIPTreasury
	if afterQueue.TreasuryBalance != before.TreasuryBalance-25 || afterQueue.LedgerOutstanding != before.LedgerOutstanding+25 {
		t.Fatalf("treasury after queue before=%#v after=%#v", before, afterQueue)
	}
	updated, err := service.UpdateAirdrop(item.ID, AirdropResultInput{Status: "failed", FailureReason: "executor rejected"})
	if err != nil || updated.Status != "failed" {
		t.Fatalf("updated=%#v err=%v", updated, err)
	}
	afterFailure := service.Snapshot().CLIPTreasury
	if afterFailure.TreasuryBalance != before.TreasuryBalance || afterFailure.LedgerOutstanding != before.LedgerOutstanding {
		t.Fatalf("treasury was not refunded: before=%#v after=%#v", before, afterFailure)
	}
	if _, err := service.UpdateAirdrop(item.ID, AirdropResultInput{Status: "confirmed", TxHash: "0x" + strings.Repeat("a", 64)}); apiErrorCode(err) != "airdrop_failed_terminal" {
		t.Fatalf("failed airdrop accepted a terminal-state transition: %v", err)
	}
}

func TestMutationRequestIDsRejectDifferentPayloads(t *testing.T) {
	tests := []struct {
		name string
		run  func(*Service) error
	}{
		{name: "redemption", run: func(s *Service) error {
			if _, _, err := s.Redeem(RedemptionInput{RequestID: "conflict-redeem", AssetID: "asset-2048", Accepted: true}); err != nil {
				return err
			}
			_, _, err := s.Redeem(RedemptionInput{RequestID: "conflict-redeem", AssetID: "asset-771", Accepted: true})
			return err
		}},
		{name: "generation", run: func(s *Service) error {
			if _, _, err := s.Redeem(RedemptionInput{RequestID: "setup-generation", AssetID: "asset-2048", Accepted: true}); err != nil {
				return err
			}
			if _, _, err := s.Generate(GenerationInput{RequestID: "conflict-generate", AssetID: "asset-2048", Duration: 15, Quality: "standard", Accepted: true}); err != nil {
				return err
			}
			_, _, err := s.Generate(GenerationInput{RequestID: "conflict-generate", AssetID: "asset-2048", Duration: 30, Quality: "standard", Accepted: true})
			return err
		}},
		{name: "conversion", run: func(s *Service) error {
			if _, _, err := s.Convert(ConversionInput{RequestID: "conflict-convert", AssetID: "asset-771", Region: "EU", Days: 30, Accepted: true}); err != nil {
				return err
			}
			_, _, err := s.Convert(ConversionInput{RequestID: "conflict-convert", AssetID: "asset-771", Region: "US", Days: 30, Accepted: true})
			return err
		}},
		{name: "exercise", run: func(s *Service) error {
			if _, _, err := s.Exercise(ExerciseInput{RequestID: "conflict-exercise", AssetID: "asset-332", PlatformCode: "foundation", Accepted: true}); err != nil {
				return err
			}
			_, _, err := s.Exercise(ExerciseInput{RequestID: "conflict-exercise", AssetID: "asset-332", PlatformCode: "haiwen", Accepted: true})
			return err
		}},
		{name: "exchange", run: func(s *Service) error {
			if _, _, err := s.Exchange(ExchangeInput{RequestID: "conflict-exchange", AssetID: "asset-528", Accepted: true}); err != nil {
				return err
			}
			_, _, err := s.Exchange(ExchangeInput{RequestID: "conflict-exchange", AssetID: "asset-771", Accepted: true})
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if code := apiErrorCode(test.run(newTestService())); code != "request_id_reused" {
				t.Fatalf("error code = %q, want request_id_reused", code)
			}
		})
	}
}

func TestAirdropRejectsPayloadConflictsAndInvalidReceipts(t *testing.T) {
	s := newTestService()
	address := "0x2222222222222222222222222222222222222222"
	item, _, err := s.CreateAirdrop(AirdropInput{RequestID: "airdrop-conflict", RuleCode: "admin_approved", WalletAddress: address, ChainID: "0x61", Amount: 25})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.CreateAirdrop(AirdropInput{RequestID: "airdrop-conflict", RuleCode: "admin_approved", WalletAddress: address, ChainID: "0x61", Amount: 26}); apiErrorCode(err) != "request_id_reused" {
		t.Fatalf("payload conflict error = %v", err)
	}
	if _, _, err := s.CreateAirdrop(AirdropInput{RequestID: "airdrop-token", RuleCode: "admin_approved", WalletAddress: address, ChainID: "0x61", Amount: 25, Token: "USDT"}); apiErrorCode(err) != "invalid_airdrop_token" {
		t.Fatalf("token mismatch error = %v", err)
	}
	if _, err := s.UpdateAirdrop(item.ID, AirdropResultInput{Status: "confirmed", TxHash: "0xabc"}); apiErrorCode(err) != "invalid_transaction_hash" {
		t.Fatalf("invalid tx hash error = %v", err)
	}
}

func apiErrorCode(err error) string {
	var target *APIError
	if errors.As(err, &target) {
		return target.Code
	}
	return ""
}

func TestSandboxCannotCreateRealTransferredAssets(t *testing.T) {
	svc := NewWithExternalPlatform(store.NewMemory(store.InitialState()), NewHTTPExternalPlatform("https://api-test.hnccc.com/api", nil))
	if !svc.ExternalPlatformSandbox() {
		t.Fatal("Haiwen test host must be marked sandbox")
	}
	_, _, err := svc.MigrateExternalAsset(MigrateExternalAssetInput{})
	if apiErrorCode(err) != "external_sandbox_transfer_disabled" {
		t.Fatalf("error=%v", err)
	}
	if len(svc.Snapshot().Assets) != 0 || len(svc.Snapshot().ExternalMigrations) != 0 {
		t.Fatal("sandbox must not create assets or call write-off")
	}
}
