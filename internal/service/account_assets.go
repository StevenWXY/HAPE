package service

import (
	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/store"
	"strings"
)

// Migration ownership is attached to the Clipli account. Changing or
// disconnecting its payout wallet does not remove its transferred licenses.
func (s *Service) AccountSnapshot() store.State {
	state := s.PortfolioSnapshot()
	owned := map[string]bool{}
	for _, exchange := range state.HAPWExchanges {
		owned[exchange.AssetID] = true
	}
	for _, migration := range state.ExternalMigrations {
		if migration.UserID == state.Session.UserRef {
			for _, id := range migration.ClipliAssetIDs {
				owned[id] = true
			}
		}
	}
	items := []domain.HAPWAsset{}
	for _, asset := range state.Assets {
		if owned[asset.ID] || asset.Owner == state.Session.UserRef || asset.Owner == "Clipli" || (state.Profile.Wallet != "" && strings.EqualFold(asset.Owner, state.Profile.Wallet)) {
			items = append(items, asset)
		}
	}
	state.Assets = items
	return state
}
