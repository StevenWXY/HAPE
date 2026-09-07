package store

import "github.com/StevenWXY/HAPE/internal/domain"

// InitialState contains configuration only. Holdings and financial history are
// populated by confirmed external transfers; startup never creates sample assets.
func InitialState() State {
	return State{
		Works: []domain.Work{}, Assets: []domain.HAPWAsset{},
		Exercises: []domain.AssetExercise{}, CLIPTransactions: []domain.CLIPTransaction{},
		Redemptions: []domain.HAPWRedemption{}, Generations: []domain.Generation{},
		HAPWExchanges: []domain.HAPWExchange{}, AssetSources: []domain.AssetSource{},
		AssetSyncRuns: []domain.AssetSyncRun{}, Distributions: []domain.CLIPDistribution{},
		WalletAssets: []domain.WalletAsset{}, Airdrops: []domain.AirdropRecord{},
		CLIP:              domain.CLIPAccount{Symbol: "CLIP", QuoteAsset: "USDT", HAPWExchangeFeeRate: 0.05, ContractStatus: "待配置", ContractStatusEn: "Not configured"},
		GenerationAccount: domain.GenerationAccount{StandardFactor: 1, ProFactor: 1.5, ClipGrantPerCredit: 0.4, ClipCostPerCredit: 0.2},
		Profile:           domain.Profile{WalletStatus: "disconnected", Settings: domain.ProfileSettings{WalletSign: true, ExpiryReminder: true}},
		Platforms: []domain.ExternalPlatform{
			{Code: "haiwen", Name: "海文发", URL: "https://hnccc.hzbcm.com/"},
			{Code: "opensea", Name: "OpenSea", URL: "https://opensea.io/"},
			{Code: "foundation", Name: "Foundation", URL: "https://foundation.app/"},
			{Code: "superrare", Name: "SuperRare", URL: "https://superrare.com/"},
			{Code: "artblocks", Name: "Art Blocks", URL: "https://www.artblocks.io/"},
		},
		CLIPTreasury:      domain.CLIPTreasury{Symbol: "CLIP", MintMode: "genesis-plus-governed-reserve", DistributionMode: "centralized-ledger", MintStatus: "not-configured"},
		DistributionRules: []domain.CLIPDistributionRule{{Code: "hapw_redemption", Trigger: "successful_hapw_redemption", Formula: "floor(creditYield * 0.40)", Rate: 0.4, Enabled: true}},
		AirdropRules: []domain.AirdropRule{
			{Code: "hapw_redemption", Name: "HAPW 核销空投", Trigger: "successful_hapw_redemption", Formula: "floor(creditYield * 0.40)", Token: "CLIP", Enabled: true, RequiresWallet: true},
			{Code: "admin_approved", Name: "指定钱包空投", Trigger: "admin_approved", Formula: "amount", Token: "CLIP", Enabled: true, RequiresWallet: true},
		},
		Session: domain.SessionPolicy{Mode: "local-session", UserRef: "anonymous-demo", WalletOptional: true, Persistence: "process-memory", Notice: "Local process session; account authentication and durable storage are required for a multi-user deployment."},
	}
}
