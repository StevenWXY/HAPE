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
	ExternalTemplates      []domain.ExternalAssetTemplate
	ExternalWorks          []domain.ExternalWork
	ExternalAssetMappings  []domain.ExternalAssetMappingRule
	ExternalMigrations     []domain.ExternalAssetMigration
	MigrationRequests      []domain.MigrationRequest
}

type Memory struct {
	mu    sync.RWMutex
	state State
	path  string
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
	next := cloneState(m.state)
	if err := fn(&next); err != nil {
		return err
	}
	if m.path != "" {
		if err := persistState(m.path, next); err != nil {
			return err
		}
	}
	m.state = cloneState(next)
	return nil
}

func cloneState(state State) State {
	clone := state
	clone.Works = append([]domain.Work{}, state.Works...)
	clone.Assets = append([]domain.HAPWAsset{}, state.Assets...)
	for index := range clone.Assets {
		clone.Assets[index].Authorization.Territories = append([]string{}, state.Assets[index].Authorization.Territories...)
		clone.Assets[index].Authorization.UsageTypes = append([]string{}, state.Assets[index].Authorization.UsageTypes...)
		clone.Assets[index].Media.LinkedWorkIDs = append([]string{}, state.Assets[index].Media.LinkedWorkIDs...)
		if state.Assets[index].SourceTemplate != nil {
			template := cloneExternalTemplate(*state.Assets[index].SourceTemplate)
			clone.Assets[index].SourceTemplate = &template
		}
	}
	clone.Exercises = append([]domain.AssetExercise{}, state.Exercises...)
	clone.CLIPTransactions = append([]domain.CLIPTransaction{}, state.CLIPTransactions...)
	clone.Redemptions = append([]domain.HAPWRedemption{}, state.Redemptions...)
	clone.Generations = append([]domain.Generation{}, state.Generations...)
	clone.Conversions = append([]domain.Conversion{}, state.Conversions...)
	clone.HAPWExchanges = append([]domain.HAPWExchange{}, state.HAPWExchanges...)
	clone.Platforms = append([]domain.ExternalPlatform{}, state.Platforms...)
	clone.AssetSources = append([]domain.AssetSource{}, state.AssetSources...)
	clone.AssetSyncRuns = append([]domain.AssetSyncRun{}, state.AssetSyncRuns...)
	clone.DistributionRules = append([]domain.CLIPDistributionRule{}, state.DistributionRules...)
	clone.Distributions = append([]domain.CLIPDistribution{}, state.Distributions...)
	clone.WalletAssets = append([]domain.WalletAsset{}, state.WalletAssets...)
	clone.AirdropRules = append([]domain.AirdropRule{}, state.AirdropRules...)
	clone.Airdrops = append([]domain.AirdropRecord{}, state.Airdrops...)
	clone.VerificationChallenges = append([]domain.VerificationChallenge{}, state.VerificationChallenges...)
	clone.Bindings = append([]domain.ExternalPlatformBinding{}, state.Bindings...)
	clone.ExternalRedemptions = append([]domain.ExternalAssetRedemption{}, state.ExternalRedemptions...)
	clone.ExternalTemplates = make([]domain.ExternalAssetTemplate, len(state.ExternalTemplates))
	for index := range state.ExternalTemplates {
		clone.ExternalTemplates[index] = cloneExternalTemplate(state.ExternalTemplates[index])
	}
	clone.ExternalWorks = make([]domain.ExternalWork, len(state.ExternalWorks))
	for index := range state.ExternalWorks {
		clone.ExternalWorks[index] = cloneExternalWork(state.ExternalWorks[index])
	}
	clone.ExternalAssetMappings = append([]domain.ExternalAssetMappingRule{}, state.ExternalAssetMappings...)
	clone.ExternalMigrations = append([]domain.ExternalAssetMigration{}, state.ExternalMigrations...)
	clone.MigrationRequests = append([]domain.MigrationRequest{}, state.MigrationRequests...)
	for index := range clone.ExternalMigrations {
		clone.ExternalMigrations[index].ClipliAssetIDs = append([]string{}, state.ExternalMigrations[index].ClipliAssetIDs...)
		clone.ExternalMigrations[index].RedemptionIDs = append([]string{}, state.ExternalMigrations[index].RedemptionIDs...)
	}
	return clone
}

func cloneExternalTemplate(value domain.ExternalAssetTemplate) domain.ExternalAssetTemplate {
	clone := value
	if value.WorksType != nil {
		worksType := *value.WorksType
		clone.WorksType = &worksType
	}
	if value.WorksSubType != nil {
		worksSubType := *value.WorksSubType
		clone.WorksSubType = &worksSubType
	}
	clone.Authors = append([]domain.ExternalParty{}, value.Authors...)
	clone.Owners = append([]domain.ExternalParty{}, value.Owners...)
	return clone
}

func cloneExternalWork(value domain.ExternalWork) domain.ExternalWork {
	clone := value
	clone.Showcase = append([]string{}, value.Showcase...)
	clone.Authors = append([]domain.ExternalParty{}, value.Authors...)
	clone.Owners = append([]domain.ExternalParty{}, value.Owners...)
	if value.WorksType != nil {
		worksType := *value.WorksType
		clone.WorksType = &worksType
	}
	if value.WorksSubType != nil {
		worksSubType := *value.WorksSubType
		clone.WorksSubType = &worksSubType
	}
	return clone
}
