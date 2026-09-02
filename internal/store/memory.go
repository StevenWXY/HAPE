package store

import (
	"sync"

	"github.com/StevenWXY/HAPE/internal/domain"
)

type State struct {
	Works                  []domain.Work
	Assets                 []domain.HAPWAsset
	Exercises              []domain.AssetExercise
	CLIP                   domain.CLIPAccount
	CLIPTransactions       []domain.CLIPTransaction
	GenerationAccount      domain.GenerationAccount
	Redemptions            []domain.HAPWRedemption
	Generations            []domain.Generation
	Profile                domain.Profile
	Conversions            []domain.Conversion
	HAPWExchanges          []domain.HAPWExchange
	Platforms              []domain.ExternalPlatform
	AssetSources           []domain.AssetSource
	AssetSyncRuns          []domain.AssetSyncRun
	CLIPTreasury           domain.CLIPTreasury
	DistributionRules      []domain.CLIPDistributionRule
	Distributions          []domain.CLIPDistribution
	WalletAssets           []domain.WalletAsset
	AirdropRules           []domain.AirdropRule
	Airdrops               []domain.AirdropRecord
	Session                domain.SessionPolicy
	VerificationChallenges []domain.VerificationChallenge
	Bindings               []domain.ExternalPlatformBinding
	ExternalRedemptions    []domain.ExternalAssetRedemption
}

type Memory struct {
	mu    sync.RWMutex
	state State
}

func NewMemory(state State) *Memory {
	return &Memory{state: cloneState(state)}
}

func (m *Memory) Snapshot() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneState(m.state)
}

func (m *Memory) Update(fn func(*State) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fn(&m.state)
}

func cloneState(state State) State {
	clone := state
	clone.Works = append([]domain.Work(nil), state.Works...)
	clone.Assets = append([]domain.HAPWAsset(nil), state.Assets...)
	for index := range clone.Assets {
		clone.Assets[index].Authorization.Territories = append([]string(nil), state.Assets[index].Authorization.Territories...)
		clone.Assets[index].Authorization.UsageTypes = append([]string(nil), state.Assets[index].Authorization.UsageTypes...)
		clone.Assets[index].Media.LinkedWorkIDs = append([]string(nil), state.Assets[index].Media.LinkedWorkIDs...)
	}
	clone.Exercises = append([]domain.AssetExercise(nil), state.Exercises...)
	clone.CLIPTransactions = append([]domain.CLIPTransaction(nil), state.CLIPTransactions...)
	clone.Redemptions = append([]domain.HAPWRedemption(nil), state.Redemptions...)
	clone.Generations = append([]domain.Generation(nil), state.Generations...)
	clone.Conversions = append([]domain.Conversion(nil), state.Conversions...)
	clone.HAPWExchanges = append([]domain.HAPWExchange(nil), state.HAPWExchanges...)
	clone.Platforms = append([]domain.ExternalPlatform(nil), state.Platforms...)
	clone.AssetSources = append([]domain.AssetSource(nil), state.AssetSources...)
	clone.AssetSyncRuns = append([]domain.AssetSyncRun(nil), state.AssetSyncRuns...)
	clone.DistributionRules = append([]domain.CLIPDistributionRule(nil), state.DistributionRules...)
	clone.Distributions = append([]domain.CLIPDistribution(nil), state.Distributions...)
	clone.WalletAssets = append([]domain.WalletAsset(nil), state.WalletAssets...)
	clone.AirdropRules = append([]domain.AirdropRule(nil), state.AirdropRules...)
	clone.Airdrops = append([]domain.AirdropRecord(nil), state.Airdrops...)
	clone.VerificationChallenges = append([]domain.VerificationChallenge(nil), state.VerificationChallenges...)
	clone.Bindings = append([]domain.ExternalPlatformBinding(nil), state.Bindings...)
	clone.ExternalRedemptions = append([]domain.ExternalAssetRedemption(nil), state.ExternalRedemptions...)
	return clone
}
