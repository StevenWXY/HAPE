package store

import "github.com/StevenWXY/HAPE/internal/domain"

func SeedState() State {
	works := []domain.Work{
		{
			ID: "work-water", Title: "HAPW水浒", TitleEn: "HAPW Water Margin", TitleKo: "HAPW 수호전",
			Category: "国风叙事", CategoryEn: "Folk narrative", CategoryKo: "전통 서사", Duration: "00:58", Views: 128400,
			Creator: "Clipli创作组", CreatorEn: "Clipli Studio", CreatorKo: "Clipli 크리에이티브 팀",
			Summary: "一百零八种数字身份，在霓虹水泊重新集结。", SummaryEn: "108 digital identities regroup in a neon marsh.", SummaryKo: "108개의 디지털 정체성이 네온빛 물가에서 다시 모인다.",
			Description:   "失散的角色从不同城市回到同一片水泊。短片以人物集结为主线，保留后续角色支线和区域发行版本的扩展空间。",
			DescriptionEn: "Scattered characters return from different cities to the same marsh. The short centers on their reunion, leaving room for character stories and regional cuts.",
			DescriptionKo: "흩어졌던 인물들이 서로 다른 도시에서 같은 물가로 돌아온다. 재회를 중심으로 전개되며, 이후 인물별 이야기와 지역별 버전으로 확장할 수 있다.",
			Format:        "AI 动画短片", FormatEn: "AI animated short", FormatKo: "AI 애니메이션 단편", LinkedAssetID: "asset-528", HaiwenURL: "https://hnccc.hzbcm.com/", Accent: "coral", Source: "seed",
		},
		{
			ID: "work-signal", Title: "夜航者：零点信号", TitleEn: "Night Flyers: Zero Signal", TitleKo: "야간 비행자: 0시 신호",
			Category: "科幻短剧", CategoryEn: "Sci-fi short", CategoryKo: "SF 숏드라마", Duration: "01:24", Views: 86200,
			Creator: "MINTLAB", CreatorEn: "MINTLAB", CreatorKo: "MINTLAB",
			Summary: "从一枚失联 NFT 发出的频率，穿过城市上空。", SummaryEn: "A frequency from a lost NFT crosses the city skyline.", SummaryKo: "연결이 끊긴 NFT가 보낸 주파수가 도시의 밤하늘을 가로지른다.",
			Description:   "零点以后，城市上空出现一段无法识别的广播。三位夜航员沿着信号寻找失联资产，也逐渐发现广播来自尚未发生的明天。",
			DescriptionEn: "After midnight, an unidentified broadcast crosses the city. Three night flyers trace a missing asset and discover the signal comes from a tomorrow that has not happened yet.",
			DescriptionKo: "자정 이후 정체불명의 방송이 도시를 가로지른다. 세 명의 야간 비행자는 사라진 자산의 신호를 추적하다가, 그 신호가 아직 오지 않은 내일에서 왔음을 알게 된다.",
			Format:        "科幻微短剧", FormatEn: "Sci-fi micro series", FormatKo: "SF 마이크로 시리즈", LinkedAssetID: "asset-771", HaiwenURL: "https://hnccc.hzbcm.com/", Accent: "amber", Source: "seed",
		},
		{
			ID: "work-tide", Title: "海风计划：登陆", TitleEn: "Sea Breeze: Landing", TitleKo: "바닷바람 프로젝트: 상륙",
			Category: "冒险", CategoryEn: "Adventure", CategoryKo: "모험", Duration: "00:46", Views: 62400,
			Creator: "Mori", CreatorEn: "Mori", CreatorKo: "Mori",
			Summary: "HAPW #2048 的首支资产叙事短片，向未知海岸出发。", SummaryEn: "The first asset story for HAPW #2048 heads for an unknown shore.", SummaryKo: "HAPW #2048의 첫 자산 서사가 미지의 해안으로 향한다.",
			Description:   "海风计划记录一支小队第一次离开母港。作品先在国内完成版权发行，再根据字幕、配音和渠道要求制作不同区域版本。",
			DescriptionEn: "Sea Breeze follows a crew leaving its home port for the first time. The work begins with a domestic rights release, followed by regional subtitle, dubbing and channel versions.",
			DescriptionKo: "바닷바람 프로젝트는 처음으로 모항을 떠나는 팀을 기록한다. 국내 저작권 배급을 먼저 마친 뒤 자막, 더빙, 채널 요구에 맞춰 지역별 버전을 제작한다.",
			Format:        "资产叙事短片", FormatEn: "Asset story short", FormatKo: "자산 서사 단편", LinkedAssetID: "asset-2048", HaiwenURL: "https://hnccc.hzbcm.com/", Accent: "teal", Source: "seed",
		},
		{
			ID: "work-orbit", Title: "轨道之外", TitleEn: "Beyond the Orbit", TitleKo: "궤도 너머",
			Category: "实验影像", CategoryEn: "Experimental", CategoryKo: "실험 영상", Duration: "01:02", Views: 43100,
			Creator: "Kite", CreatorEn: "Kite", CreatorKo: "Kite",
			Summary: "一段关于离开、回望与重新拥有的等距叙事。", SummaryEn: "An isometric story about leaving, looking back and owning again.", SummaryKo: "떠남과 회상, 다시 소유하는 일을 다룬 아이소메트릭 서사.",
			Description:   "一座空间站在最后一次换轨前开放旧物寄存。人们留下的不只是物件，也包括一段可被授权、行权和重新讲述的记忆。",
			DescriptionEn: "Before its final orbital change, a station opens an archive for old belongings. People leave objects and memories that can be licensed, exercised and retold.",
			DescriptionKo: "마지막 궤도 변경을 앞둔 우주 정거장이 기록소를 연다. 사람들은 허가하고 행사하며 다시 이야기할 수 있는 기억을 남긴다.",
			Format:        "实验动画", FormatEn: "Experimental animation", FormatKo: "실험 애니메이션", LinkedAssetID: "asset-332", HaiwenURL: "https://hnccc.hzbcm.com/", Accent: "gold", Source: "seed",
		},
	}

	assets := []domain.HAPWAsset{
		seedAsset("asset-2048", "#2048", "海风计划", "Sea Breeze", "바닷바람 프로젝트", 18800, "可核销", "Redeemable", "상각 가능", true, "海风计划版权方", "Sea Breeze rights holder", "바닷바람 프로젝트 저작권자", "角色、场景与叙事元素的 AI 视频改编", "AI video adaptation of characters, scenes and story elements", "캐릭터, 장면 및 서사 요소의 AI 영상 각색", 150, 520, true, "available", []string{"work-tide"}, "2026-05-12"),
		seedAsset("asset-771", "#771", "夜航者", "Night Flyers", "야간 비행자", 12600, "可核销", "Redeemable", "상각 가능", true, "MINTLAB", "MINTLAB", "MINTLAB", "角色设定、城市空间与非独家短视频改编", "Characters, city environments and non-exclusive short-video adaptations", "캐릭터 설정, 도시 공간 및 비독점 숏폼 영상 각색", 120, 390, true, "available", []string{"work-signal"}, "2026-04-20"),
		seedAsset("asset-109", "#109", "远岸", "Far Shore", "먼 해안", 9900, "权属复核中", "Rights review", "권리 검토 중", false, "远岸版权组", "Far Shore rights group", "먼 해안 권리 그룹", "待版权持有方完成授权复核", "Pending final authorization review by the rights holder", "저작권자의 최종 이용 허가 검토 대기 중", 90, 0, false, "restricted", nil, "2026-06-08"),
		seedAsset("asset-332", "#332", "风暴档案", "Storm Archive", "폭풍 기록", 7200, "可核销", "Redeemable", "상각 가능", true, "Kite Studio", "Kite Studio", "Kite Studio", "场景素材与实验影像再创作", "Scene materials and experimental video derivatives", "장면 소재 및 실험 영상의 2차 창작", 100, 280, true, "available", []string{"work-orbit"}, "2026-03-18"),
		seedAsset("asset-528", "#528", "水浒角色集", "Water Margin Cast", "수호전 인물집", 3800, "可核销", "Redeemable", "상각 가능", true, "Clipli 创作组", "Clipli Studio", "Clipli 크리에이티브 팀", "已登记角色素材的 AI 动画短片生成", "AI animated shorts using the registered character set", "등록된 캐릭터 소재를 사용한 AI 애니메이션 단편 생성", 120, 220, true, "available", []string{"work-water"}, "2026-02-11"),
		seedAsset("asset-903", "#903", "低语岛", "Whisper Island", "속삭임의 섬", 2140, "已核销", "Redeemed", "상각 완료", false, "Mori", "Mori", "Mori", "岛屿场景、声音设定与短视频改编", "Island scenes, sound concepts and short-video adaptations", "섬 장면, 사운드 콘셉트 및 숏폼 영상 각색", 120, 0, false, "redeemed", nil, "2026-01-24"),
	}

	return State{
		Works:  works,
		Assets: assets,
		Exercises: []domain.AssetExercise{
			{ID: "exercise-1", RequestID: "seed-exercise-1", AssetID: "asset-2048", PlatformCode: "haiwen", Direction: "Clipli → 海文发", DirectionEn: "Clipli → HAIWEN", Value: 18800, StatusCode: "completed", Status: "已完成", StatusEn: "Completed", CreatedAt: "2026-08-03"},
			{ID: "exercise-2", RequestID: "seed-exercise-2", AssetID: "asset-771", PlatformCode: "opensea", Direction: "Clipli → OpenSea", DirectionEn: "Clipli → OpenSea", Value: 12600, StatusCode: "pending", Status: "待确认", StatusEn: "Pending", CreatedAt: "2026-07-29"},
			{ID: "exercise-3", RequestID: "seed-exercise-3", AssetID: "asset-109", PlatformCode: "superrare", Direction: "Clipli → SuperRare", DirectionEn: "Clipli → SuperRare", Value: 9900, StatusCode: "archived", Status: "已归档", StatusEn: "Archived", CreatedAt: "2026-06-18"},
		},
		CLIP: domain.CLIPAccount{
			Symbol: "CLIP", Balance: 3460, SupplyPolicy: "一次性铸造，后续不增发", SupplyPolicyEn: "One-time mint with no further issuance",
			Acquisition: "核销 HAPW 时领取，也可通过外部 DEX 购买", AcquisitionEn: "Granted when HAPW is redeemed or purchased through an external DEX", AcquisitionKo: "HAPW 상각 시 지급되거나 외부 DEX에서 구매",
			QuoteAsset: "USDT", PricePolicy: "由外部 DEX 流动性决定，不固定锚定", PricePolicyEn: "Market-priced by external DEX liquidity; no fixed peg", PricePolicyKo: "외부 DEX 유동성에 따른 시장 가격, 고정 페그 없음",
			ContractStatus: "简化合约：固定供应与平台金库", ContractStatusEn: "Minimal contract: fixed supply and platform treasury", ContractStatusKo: "간소화 계약: 고정 공급 및 플랫폼 금고", DexURL: "https://app.uniswap.org/swap/",
			DexPool: domain.CLIPPool{ClipReserve: 250000, USDTReserve: 25000, UpdatedAt: "2026-08-10 16:00"}, HAPWExchangeFeeRate: 0.05,
		},
		CLIPTransactions: []domain.CLIPTransaction{
			{ID: "clip-tx-grant-1", TypeCode: "redemptionGrant", Type: "HAPW 核销领取", TypeEn: "HAPW redemption grant", Amount: 48, Counterparty: "HAPW #903 · 低语岛", CounterpartyEn: "HAPW #903 · Whisper Island", CounterpartyKo: "HAPW #903 · 속삭임의 섬", StatusCode: "completed", Status: "已完成", StatusEn: "Completed", TxHash: "0xgrant…903a", CreatedAt: "2026-08-06 09:30"},
			{ID: "clip-tx-generate-1", TypeCode: "generationFee", Type: "AI 视频生成费", TypeEn: "AI video generation fee", Amount: -5, Counterparty: "《低语岛：潮汐试片》", CounterpartyEn: "Whisper Island: Tide Test", StatusCode: "completed", Status: "已完成", StatusEn: "Completed", TxHash: "0xgen…903a", CreatedAt: "2026-08-06 17:42"},
			{ID: "clip-tx-2", TypeCode: "license", Type: "授权手续费", TypeEn: "License fee", Amount: -18, Counterparty: "HAPW #2048 · 海风计划", CounterpartyEn: "HAPW #2048 · Sea Breeze", StatusCode: "completed", Status: "已完成", StatusEn: "Completed", TxHash: "0xa83d…71b4", CreatedAt: "2026-08-03 09:42"},
			{ID: "clip-tx-3", TypeCode: "dex", Type: "DEX 购入", TypeEn: "DEX purchase", Amount: 1200, Counterparty: "外部 DEX", CounterpartyEn: "External DEX", StatusCode: "completed", Status: "已完成", StatusEn: "Completed", TxHash: "0x08be…d990", CreatedAt: "2026-07-28 19:05"},
			{ID: "clip-tx-4", TypeCode: "localization", Type: "本地化服务", TypeEn: "Localization service", Amount: -60, Counterparty: "海文发本地化服务", CounterpartyEn: "HAIWEN localization", StatusCode: "completed", Status: "已完成", StatusEn: "Completed", TxHash: "0x7dd1…4e18", CreatedAt: "2026-07-22 11:16"},
		},
		GenerationAccount: domain.GenerationAccount{Balance: 97, LifetimeGranted: 120, LifetimeUsed: 23, StandardFactor: 1, ProFactor: 1.5, ClipGrantPerCredit: 0.4, ClipCostPerCredit: 0.2, LifetimeClipGranted: 48, LifetimeClipSpent: 5},
		Redemptions:       []domain.HAPWRedemption{{ID: "redeem-903", AssetID: "asset-903", Receipt: "Clipli-LIC-903-A1", CreditsGranted: 120, CreditsRemaining: 97, ClipGranted: 48, RequestID: "seed-redeem-903", Status: "有效", StatusEn: "Active", StatusKo: "유효", CreatedAt: "2026-08-06 09:30"}},
		Generations:       []domain.Generation{{ID: "video-903-1", AssetID: "asset-903", Title: "低语岛：潮汐试片", TitleEn: "Whisper Island: Tide Test", TitleKo: "속삭임의 섬: 조수 테스트", Duration: 15, Quality: "pro", CreditsUsed: 23, ValidViews: 18600, ClipCost: 5, RequestID: "seed-video-903-1", Status: "已生成", StatusEn: "Generated", StatusKo: "생성 완료", CreatedAt: "2026-08-06 17:42"}},
		Profile:           domain.Profile{OverseasAccount: "HWF-8391", Phone: "138****1024", Level: 3, Points: 3460, Settings: domain.ProfileSettings{WalletSign: true, ExpiryReminder: true}},
		Platforms: []domain.ExternalPlatform{
			{Code: "haiwen", Name: "海文发", URL: "https://hnccc.hzbcm.com/"},
			{Code: "opensea", Name: "OpenSea", URL: "https://opensea.io/"},
			{Code: "foundation", Name: "Foundation", URL: "https://foundation.app/"},
			{Code: "superrare", Name: "SuperRare", URL: "https://superrare.com/"},
			{Code: "artblocks", Name: "Art Blocks", URL: "https://www.artblocks.io/"},
		},
		AssetSources: []domain.AssetSource{
			{Code: "haiwen", Name: "海文发", BaseURL: "https://hnccc.hzbcm.com/", Mode: "external-snapshot", Status: "synced", SyncIntervalSeconds: 900, LastSyncedAt: "2026-08-18T08:30:00Z", AssetCount: 2},
			{Code: "opensea", Name: "OpenSea", BaseURL: "https://opensea.io/", Mode: "external-snapshot", Status: "synced", SyncIntervalSeconds: 900, LastSyncedAt: "2026-08-18T08:30:00Z", AssetCount: 1},
			{Code: "foundation", Name: "Foundation", BaseURL: "https://foundation.app/", Mode: "external-snapshot", Status: "synced", SyncIntervalSeconds: 900, LastSyncedAt: "2026-08-18T08:30:00Z", AssetCount: 1},
			{Code: "superrare", Name: "SuperRare", BaseURL: "https://superrare.com/", Mode: "external-snapshot", Status: "synced", SyncIntervalSeconds: 900, LastSyncedAt: "2026-08-18T08:30:00Z", AssetCount: 1},
			{Code: "artblocks", Name: "Art Blocks", BaseURL: "https://www.artblocks.io/", Mode: "external-snapshot", Status: "synced", SyncIntervalSeconds: 900, LastSyncedAt: "2026-08-18T08:30:00Z", AssetCount: 1},
		},
		AssetSyncRuns: []domain.AssetSyncRun{{ID: "sync-demo-20260818", SourceCode: "external-platforms", Status: "completed", StartedAt: "2026-08-18T08:29:58Z", CompletedAt: "2026-08-18T08:30:00Z", RecordsRead: 6, RecordsValid: 6, RecordsSaved: 6}},
		CLIPTreasury: domain.CLIPTreasury{
			Symbol: "CLIP", MintMode: "one-time-fixed-supply", MintedSupply: 10000000, TreasuryBalance: 9746540,
			LiquidityAllocation: 250000, LedgerOutstanding: 3460, TotalDistributed: 3460, TotalReclaimed: 0,
			PlatformWallet: "0xCLIPLI…TREASURY", ContractAddress: "not-published", Network: "demo",
			DistributionMode: "centralized-ledger", MintStatus: "prototype-snapshot", MintedAt: "2026-01-01T00:00:00Z",
		},
		DistributionRules: []domain.CLIPDistributionRule{
			{Code: "hapw_redemption", Trigger: "successful_hapw_redemption", Formula: "floor(creditYield * 0.40)", Rate: 0.4, Description: "HAPW 核销成功后由平台金库一次性划拨", Enabled: true},
		},
		Distributions: []domain.CLIPDistribution{{ID: "distribution-903", RequestID: "seed-redeem-903", RuleCode: "hapw_redemption", UserRef: "anonymous-demo", AssetID: "asset-903", Amount: 48, UserBalanceBefore: 3412, UserBalanceAfter: 3460, TreasuryBefore: 9746588, TreasuryAfter: 9746540, Status: "completed", CreatedAt: "2026-08-06 09:30"}},
		Session:       domain.SessionPolicy{Mode: "anonymous-demo", UserRef: "anonymous-demo", IdentityVerification: false, KYCRequired: false, WalletOptional: true, Persistence: "process-memory", Notice: "No real identity verification; wallet connection is an optional operation handle."},
	}
}

func seedAsset(id, tokenID, name, nameEn, nameKo string, value int, status, statusEn, statusKo string, transferable bool, holder, holderEn, holderKo, scope, scopeEn, scopeKo string, creditYield, clipPrice int, exchangeAvailable bool, redemptionStatus string, workIDs []string, issuedAt string) domain.HAPWAsset {
	providerCode, providerAssetID, assetURL := seedExternalRef(id)
	return domain.HAPWAsset{
		ID: id, Kind: "HAPW", TokenID: tokenID, Name: name, NameEn: nameEn, NameKo: nameKo,
		Value: value, Currency: "CNY", Status: status, StatusEn: statusEn, StatusKo: statusKo,
		Transferable: transferable, Owner: "Clipli", RightsHolder: holder, RightsHolderEn: holderEn, RightsHolderKo: holderKo,
		AuthorizationScope: scope, AuthorizationScopeEn: scopeEn, AuthorizationScopeKo: scopeKo,
		CreditYield: creditYield, ClipPrice: clipPrice, ExchangeAvailable: exchangeAvailable, RedemptionStatus: redemptionStatus,
		Authorization: domain.Authorization{
			Holder: domain.LocalizedText{ZH: holder, EN: holderEn, KO: holderKo}, Scope: domain.LocalizedText{ZH: scope, EN: scopeEn, KO: scopeKo},
			Territories: []string{"CN", "GLOBAL"}, UsageTypes: []string{"ai-video", "short-video", "digital-display"}, Exclusivity: "non-exclusive",
			CommercialUse: true, DerivativeWorks: true, AITraining: false, Sublicensable: false, ValidFrom: issuedAt,
		},
		Provenance: domain.Provenance{
			Issuer: holder, CertificateID: "CLIPLI-HAPW-" + tokenID[1:], Network: "prototype", TokenStandard: "HAPW-1",
			VerificationStatus: "demo-verified", IssuedAt: issuedAt,
		},
		Media:    domain.AssetMedia{LinkedWorkIDs: workIDs, Format: "mixed-media"},
		External: domain.ExternalAssetRef{ProviderCode: providerCode, ProviderAssetID: providerAssetID, AssetURL: assetURL, SyncStatus: "synced", LastSyncedAt: "2026-08-18T08:30:00Z", DataVersion: "snapshot-2026-08-18"},
	}
}

func seedExternalRef(assetID string) (string, string, string) {
	refs := map[string][3]string{
		"asset-2048": {"haiwen", "HWF-HAPW-2048", "https://hnccc.hzbcm.com/"},
		"asset-771":  {"haiwen", "HWF-HAPW-771", "https://hnccc.hzbcm.com/"},
		"asset-109":  {"opensea", "OPENSEA-109", "https://opensea.io/"},
		"asset-332":  {"foundation", "FOUNDATION-332", "https://foundation.app/"},
		"asset-528":  {"superrare", "SUPERRARE-528", "https://superrare.com/"},
		"asset-903":  {"artblocks", "ARTBLOCKS-903", "https://www.artblocks.io/"},
	}
	ref := refs[assetID]
	return ref[0], ref[1], ref[2]
}
