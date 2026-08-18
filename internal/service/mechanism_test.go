package service

import (
	"testing"
	"time"

	"github.com/StevenWXY/HAPE/internal/domain"
)

func TestGenerationCosts(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		quality  string
		credits  int
		clip     int
	}{
		{name: "standard 15 seconds", duration: 15, quality: "standard", credits: 15, clip: 3},
		{name: "pro 15 seconds rounds up", duration: 15, quality: "pro", credits: 23, clip: 5},
		{name: "pro 60 seconds", duration: 60, quality: "pro", credits: 90, clip: 18},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			credits, err := GenerationCost(test.duration, test.quality)
			if err != nil {
				t.Fatalf("GenerationCost() error = %v", err)
			}
			if credits != test.credits {
				t.Fatalf("credits = %d, want %d", credits, test.credits)
			}
			clip, err := GenerationCLIPCost(credits)
			if err != nil {
				t.Fatalf("GenerationCLIPCost() error = %v", err)
			}
			if clip != test.clip {
				t.Fatalf("clip = %d, want %d", clip, test.clip)
			}
		})
	}
}

func TestRedemptionGrantAndExchangeQuote(t *testing.T) {
	grant, err := RedemptionCLIPGrant(150)
	if err != nil || grant != 60 {
		t.Fatalf("grant = %d, err = %v, want 60", grant, err)
	}
	quote, err := QuoteHAPWExchange(520, HAPWExchangeFeeRate)
	if err != nil {
		t.Fatalf("QuoteHAPWExchange() error = %v", err)
	}
	if quote.Fee != 26 || quote.Total != 546 {
		t.Fatalf("quote = %#v, want fee 26 total 546", quote)
	}
}

func TestExchangePolicy(t *testing.T) {
	now := time.Date(2026, 8, 17, 8, 0, 0, 0, time.UTC)
	exchanges := []domain.HAPWExchange{{StatusCode: "completed", CreatedAt: "2026-08-17 07:30"}}
	assets := []domain.HAPWAsset{
		{Kind: "HAPW", ClipPrice: 100, ExchangeAvailable: true},
		{Kind: "HAPW", ClipPrice: 200, ExchangeAvailable: false},
	}
	policy := BuildExchangePolicy(exchanges, assets, now)
	if policy.UsedToday != 1 || policy.RemainingToday != 1 || policy.Reached {
		t.Fatalf("policy = %#v", policy)
	}
	if policy.InventoryTotal != 2 || policy.InventoryAvailable != 1 {
		t.Fatalf("inventory = %#v", policy)
	}
}
