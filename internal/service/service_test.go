package service

import (
	"testing"
	"time"

	"github.com/StevenWXY/HAPE/internal/store"
)

func newTestService() *Service {
	service := New(store.NewMemory(store.SeedState()))
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
		if asset.External.ProviderCode == "" || asset.External.ProviderAssetID == "" || asset.External.SyncStatus != "synced" {
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
