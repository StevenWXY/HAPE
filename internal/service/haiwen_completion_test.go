package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/store"
)

type checkedMigrationPlatform struct {
	migrationFake
	bound       bool
	reads       int
	beforeWrite func()
}

func (f *checkedMigrationPlatform) GetBindingStatus(_ context.Context, id string) (ExternalBindingStatus, error) {
	f.reads++
	return ExternalBindingStatus{ExternalUserID: id, Bound: f.bound, BoundAt: "2026-09-08T00:00:00Z"}, nil
}
func (f *checkedMigrationPlatform) WriteOffTemplate(ctx context.Context, input ExternalTemplateWriteOffRequest) (ExternalRedemptionResult, error) {
	if f.beforeWrite != nil {
		f.beforeWrite()
	}
	return f.migrationFake.WriteOffTemplate(ctx, input)
}

func completionService(t *testing.T) (*Service, *checkedMigrationPlatform) {
	t.Helper()
	state := store.InitialState()
	state.CLIPTreasury.TreasuryBalance = 120
	state.Bindings = []domain.ExternalPlatformBinding{{ID: "binding-test", UserID: state.Session.UserRef, ExternalUserID: "external-1", Status: "bound"}}
	fake := &checkedMigrationPlatform{bound: true, migrationFake: migrationFake{count: 2, template: domain.ExternalAssetTemplate{TplID: 100053, Name: "Fixture copyright", WorkID: 100, WorksName: "Fixture work", Owners: []domain.ExternalParty{{ID: 1, Name: "Fixture owner"}}}}}
	svc := New(store.NewMemory(state), fake)
	if err := svc.SetExternalAssetMappings([]domain.ExternalAssetMappingRule{{TplID: 100053, Version: "v1", CreditYield: 150, Active: true}}); err != nil {
		t.Fatal(err)
	}
	return svc, fake
}

func TestMigrationReviewBatchSettlementAndIdempotency(t *testing.T) {
	svc, fake := completionService(t)
	input := MigrateExternalAssetInput{UserID: svc.Snapshot().Session.UserRef, TplID: 100053, Num: 2, RequestID: "test-exchange-request", RequestNo: "test-exchange-write-off", Accepted: true}
	request, repeated, err := svc.RequestMigration(input)
	if err != nil || repeated || fake.redeems != 0 || len(svc.Snapshot().Assets) != 0 {
		t.Fatalf("request=%+v repeated=%v err=%v", request, repeated, err)
	}
	if _, repeated, err := svc.RequestMigration(input); err != nil || !repeated {
		t.Fatalf("request retry err=%v repeated=%v", err, repeated)
	}
	fake.beforeWrite = func() {
		if svc.Snapshot().CLIPTreasury.TreasuryBalance != 0 {
			t.Error("CLIP was not reserved before irreversible write-off")
		}
		_, _, err := svc.CreateAirdrop(AirdropInput{RuleCode: "admin_approved", RequestID: "competing-airdrop", WalletAddress: "0x1111111111111111111111111111111111111111", ChainID: "0x38", Amount: 1})
		if apiErrorCode(err) != "clip_treasury_insufficient" {
			t.Errorf("competing treasury spend: %v", err)
		}
	}
	completed, err := svc.ReviewMigrationRequest(request.ID, "approve", "", true)
	if err != nil || completed.Status != "completed" {
		t.Fatalf("review=%+v err=%v", completed, err)
	}
	state := svc.Snapshot()
	if len(state.Assets) != 2 || len(state.Redemptions) != 2 || state.GenerationAccount.Balance != 300 || state.CLIP.Balance != 120 {
		t.Fatalf("incorrect settlement assets=%d redemptions=%d credits=%d CLIP=%d", len(state.Assets), len(state.Redemptions), state.GenerationAccount.Balance, state.CLIP.Balance)
	}
	if state.Redemptions[0].ID == state.Redemptions[1].ID || state.Distributions[0].ID == state.Distributions[1].ID || state.CLIPTransactions[0].ID == state.CLIPTransactions[1].ID {
		t.Fatal("batch IDs collide")
	}
	if _, err := svc.ReviewMigrationRequest(request.ID, "approve", "", true); err != nil || fake.redeems != 1 {
		t.Fatalf("duplicate approval err=%v writeoffs=%d", err, fake.redeems)
	}
}

func TestRevokedBindingBlocksCountsPreviewAndWriteOff(t *testing.T) {
	for _, operation := range []string{"counts", "preview", "migrate"} {
		t.Run(operation, func(t *testing.T) {
			svc, fake := completionService(t)
			fake.bound = false
			input := MigrateExternalAssetInput{UserID: svc.Snapshot().Session.UserRef, TplID: 100053, Num: 1, RequestID: "revoked-request", RequestNo: "revoked-write-off", Accepted: true}
			var err error
			switch operation {
			case "counts":
				_, err = svc.UserAssetCounts(input.UserID, []int64{100053})
			case "preview":
				_, err = svc.PreviewExternalAssetMigration(input)
			case "migrate":
				_, _, err = svc.MigrateExternalAsset(input)
			}
			if err == nil || fake.reads != 1 || fake.redeems != 0 || len(svc.Snapshot().Assets) != 0 {
				t.Fatalf("revoked binding accepted: err=%v reads=%d writes=%d", err, fake.reads, fake.redeems)
			}
		})
	}
}

func TestReserveMustBePurchasedBeforeRedemption(t *testing.T) {
	svc := newTestService()
	if _, _, err := svc.Redeem(RedemptionInput{AssetID: "asset-528", RequestID: "free-reserve-redemption", Accepted: true}); apiErrorCode(err) != "hapw_not_redeemable" {
		t.Fatalf("unpaid reserve could be redeemed: %v", err)
	}
	result, _, err := svc.Exchange(ExchangeInput{AssetID: "asset-528", RequestID: "paid-reserve-exchange", Accepted: true})
	if err != nil || result.Asset.Owner != svc.Snapshot().Session.UserRef {
		t.Fatalf("ownership was not transferred: %+v err=%v", result.Asset, err)
	}
	if _, _, err := svc.Redeem(RedemptionInput{AssetID: "asset-528", RequestID: "paid-reserve-redemption", Accepted: true}); err != nil {
		t.Fatal(err)
	}
}

func TestAirdropReservesBalanceAndCannotChangeRecipient(t *testing.T) {
	svc, _ := completionService(t)
	wallet := "0x1111111111111111111111111111111111111111"
	if _, err := svc.ConnectWallet(WalletConnectInput{Provider: "MetaMask", Address: wallet, ChainID: "0x38"}); err != nil {
		t.Fatal(err)
	}
	input := MigrateExternalAssetInput{UserID: svc.Snapshot().Session.UserRef, TplID: 100053, Num: 1, RequestID: "wallet-migration", RequestNo: "wallet-write-off", Accepted: true}
	result, _, err := svc.MigrateExternalAsset(input)
	if err != nil {
		t.Fatal(err)
	}
	state := svc.Snapshot()
	if state.CLIP.Balance != 60 || state.CLIP.Reserved != 60 || state.CLIP.AvailableBalance != 0 || len(state.Airdrops) != 1 {
		t.Fatalf("reserved balance=%+v drops=%d", state.CLIP, len(state.Airdrops))
	}
	svc.DisconnectWallet()
	if len(svc.AccountSnapshot().Assets) != 1 {
		t.Fatal("disconnecting a wallet hid account-owned assets")
	}
	if _, err := svc.ConnectWallet(WalletConnectInput{Provider: "MetaMask", Address: wallet, ChainID: "0x38"}); err != nil {
		t.Fatal(err)
	}
	other := "0x2222222222222222222222222222222222222222"
	if _, _, err := svc.CreateAirdrop(AirdropInput{RequestID: "steal-airdrop", RuleCode: "hapw_redemption", WalletAddress: other, ChainID: "0x38", AssetID: result.Assets[0].ID}); apiErrorCode(err) != "airdrop_wallet_mismatch" {
		t.Fatalf("other wallet accepted: %v", err)
	}
	if _, err := svc.CancelWalletAirdrop(state.Airdrops[0].ID); err != nil {
		t.Fatal(err)
	}
	if svc.Snapshot().CLIP.AvailableBalance != 60 {
		t.Fatal("cancel did not release CLIP")
	}
	if _, err := svc.CancelWalletAirdrop(state.Airdrops[0].ID); err != nil || svc.Snapshot().CLIP.AvailableBalance != 60 {
		t.Fatal("duplicate cancellation refunded twice")
	}
}

func TestHaiwenRejectsIncompleteReceiptsCountsAndCatalog(t *testing.T) {
	for _, scenario := range []string{"empty-receipt", "wrong-user", "wrong-status", "missing-count", "negative-count", "duplicate-count", "missing-catalog", "wrong-page", "missing-envelope"} {
		t.Run(scenario, func(t *testing.T) {
			data := map[string]any{"requestNo": "write-off-1", "externalUserId": "external-1", "tplId": 100053, "num": 1, "status": "SUCCESS", "writeOffAt": "2026-09-08T00:00:00Z"}
			switch scenario {
			case "empty-receipt":
				data = map[string]any{}
			case "wrong-user":
				data["externalUserId"] = "other"
			case "wrong-status":
				data["status"] = "FAILED"
			case "missing-count":
				data["list"] = []any{}
			case "negative-count":
				data["list"] = []any{map[string]any{"tplId": 100053, "count": -1}}
			case "duplicate-count":
				data["list"] = []any{map[string]any{"tplId": 100053, "count": 1}, map[string]any{"tplId": 100053, "count": 1}}
			case "missing-catalog":
				data = map[string]any{}
			case "wrong-page":
				data = map[string]any{"list": []any{}, "total": 0, "pageNum": 2, "pageSize": 20}
			}
			body, _ := json.Marshal(map[string]any{"code": 0, "data": data})
			if scenario == "missing-envelope" {
				body = []byte(`{}`)
			}
			platform := NewHTTPExternalPlatform("https://partner.example/api", &http.Client{Transport: integrationRoundTripper(func(request *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(body))), Header: make(http.Header)}, nil
			})})
			var err error
			if strings.Contains(scenario, "count") {
				_, err = platform.ListAssetCounts(context.Background(), "external-1", []int64{100053})
			} else if strings.Contains(scenario, "catalog") || scenario == "wrong-page" {
				_, err = platform.ListTemplates(context.Background(), 1, 20, 0)
			} else {
				_, err = platform.WriteOffTemplate(context.Background(), ExternalTemplateWriteOffRequest{ExternalUserID: "external-1", TplID: 100053, Num: 1, RequestNo: "write-off-1"})
			}
			if err == nil {
				t.Fatal("accepted incomplete source data")
			}
		})
	}
}
