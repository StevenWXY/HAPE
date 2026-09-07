package service

import (
	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/store"
	"net/http"
	"strings"
	"time"
)

func (s *Service) CancelWalletAirdrop(id string) (domain.AirdropRecord, error) {
	var result domain.AirdropRecord
	err := s.store.Update(func(state *store.State) error {
		for i := range state.Airdrops {
			item := &state.Airdrops[i]
			if item.ID != id {
				continue
			}
			if item.AllocationSource != "redemption-entitlement" || !strings.EqualFold(item.WalletAddress, state.Profile.Wallet) {
				return apiError(http.StatusForbidden, "airdrop_wallet_mismatch", "This task belongs to another wallet")
			}
			if item.Status == "cancelled" {
				result = *item
				return nil
			}
			if item.Status != "queued" || item.TxHash != "" {
				return apiError(http.StatusConflict, "airdrop_already_submitted", "An already submitted transfer cannot be cancelled")
			}
			state.CLIP.Reserved -= item.Amount
			item.Status = "cancelled"
			item.UpdatedAt = s.now().UTC().Format(time.RFC3339)
			result = *item
			return nil
		}
		return apiError(http.StatusNotFound, "airdrop_not_found", "Airdrop record not found")
	})
	return result, err
}
