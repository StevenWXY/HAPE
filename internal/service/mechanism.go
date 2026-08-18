package service

import (
	"errors"
	"math"
	"time"

	"github.com/StevenWXY/HAPE/internal/domain"
)

const (
	CLIPGrantPerCredit     = 0.4
	CLIPCostPerCredit      = 0.2
	HAPWExchangeFeeRate    = 0.05
	HAPWExchangeDailyLimit = 2
)

var generationFactors = map[string]float64{"standard": 1, "pro": 1.5}

func GenerationCost(duration int, quality string) (int, error) {
	factor, ok := generationFactors[quality]
	if duration <= 0 || !ok {
		return 0, errors.New("invalid_generation_parameters")
	}
	return int(math.Ceil(float64(duration) * factor)), nil
}

func RedemptionCLIPGrant(credits int) (int, error) {
	if credits <= 0 {
		return 0, errors.New("invalid_credit_grant")
	}
	return int(math.Floor(float64(credits) * CLIPGrantPerCredit)), nil
}

func GenerationCLIPCost(credits int) (int, error) {
	if credits <= 0 {
		return 0, errors.New("invalid_credit_cost")
	}
	return int(math.Ceil(float64(credits) * CLIPCostPerCredit)), nil
}

type ExchangeQuote struct {
	Price   int     `json:"price"`
	Fee     int     `json:"fee"`
	Total   int     `json:"total"`
	FeeRate float64 `json:"feeRate"`
}

func QuoteHAPWExchange(price int, feeRate float64) (ExchangeQuote, error) {
	if price <= 0 || feeRate < 0 || feeRate > 1 {
		return ExchangeQuote{}, errors.New("invalid_exchange_quote")
	}
	fee := int(math.Ceil(float64(price) * feeRate))
	return ExchangeQuote{Price: price, Fee: fee, Total: price + fee, FeeRate: feeRate}, nil
}

type ExchangePolicy struct {
	Date               string `json:"date"`
	Timezone           string `json:"timezone"`
	ResetsAt           string `json:"resetsAt"`
	DailyLimit         int    `json:"dailyLimit"`
	UsedToday          int    `json:"usedToday"`
	RemainingToday     int    `json:"remainingToday"`
	Reached            bool   `json:"reached"`
	InventoryTotal     int    `json:"inventoryTotal"`
	InventoryAvailable int    `json:"inventoryAvailable"`
}

func BuildExchangePolicy(exchanges []domain.HAPWExchange, assets []domain.HAPWAsset, now time.Time) ExchangePolicy {
	day := now.UTC().Format("2006-01-02")
	used := 0
	for _, exchange := range exchanges {
		if exchange.StatusCode == "completed" && len(exchange.CreatedAt) >= 10 && exchange.CreatedAt[:10] == day {
			used++
		}
	}
	total, available := 0, 0
	for _, asset := range assets {
		if asset.Kind == "HAPW" && asset.ClipPrice > 0 {
			total++
			if asset.ExchangeAvailable {
				available++
			}
		}
	}
	remaining := HAPWExchangeDailyLimit - used
	if remaining < 0 {
		remaining = 0
	}
	reset := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day()+1, 0, 0, 0, 0, time.UTC)
	return ExchangePolicy{
		Date: day, Timezone: "UTC", ResetsAt: reset.Format(time.RFC3339), DailyLimit: HAPWExchangeDailyLimit,
		UsedToday: used, RemainingToday: remaining, Reached: used >= HAPWExchangeDailyLimit,
		InventoryTotal: total, InventoryAvailable: available,
	}
}

type DexPoolSnapshot struct {
	ClipReserve        int     `json:"clipReserve"`
	USDTReserve        int     `json:"usdtReserve"`
	ClipPerUSDT        float64 `json:"clipPerUsdt"`
	USDTPerCLIP        float64 `json:"usdtPerClip"`
	TotalLiquidityUSDT int     `json:"totalLiquidityUsdt"`
	UpdatedAt          string  `json:"updatedAt"`
}

func BuildDexPoolSnapshot(pool domain.CLIPPool) (DexPoolSnapshot, error) {
	if pool.ClipReserve <= 0 || pool.USDTReserve <= 0 {
		return DexPoolSnapshot{}, errors.New("invalid_liquidity_pool")
	}
	return DexPoolSnapshot{
		ClipReserve: pool.ClipReserve, USDTReserve: pool.USDTReserve,
		ClipPerUSDT:        float64(pool.ClipReserve) / float64(pool.USDTReserve),
		USDTPerCLIP:        float64(pool.USDTReserve) / float64(pool.ClipReserve),
		TotalLiquidityUSDT: pool.USDTReserve * 2, UpdatedAt: pool.UpdatedAt,
	}, nil
}
