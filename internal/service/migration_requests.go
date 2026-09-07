package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/store"
)

func (s *Service) MigrationRequests(userID string) []domain.MigrationRequest {
	items := []domain.MigrationRequest{}
	for _, item := range s.Snapshot().MigrationRequests {
		if userID == "" || item.UserID == userID {
			items = append(items, item)
		}
	}
	return items
}

// Requests record user consent without calling Haiwen's irreversible endpoint.
func (s *Service) RequestMigration(input MigrateExternalAssetInput) (domain.MigrationRequest, bool, error) {
	var result domain.MigrationRequest
	if s.ExternalPlatformSandbox() {
		return result, false, apiError(http.StatusConflict, "external_sandbox_transfer_disabled", "Sandbox assets cannot be exchanged for real assets")
	}
	userID, binding, num, requestNo, err := s.validateExternalMigrationInput(input, true)
	if err != nil {
		return result, false, err
	}
	if userID != s.Snapshot().Session.UserRef {
		return result, false, apiError(http.StatusForbidden, "session_user_mismatch", "The request must belong to the current account")
	}
	for _, item := range s.MigrationRequests("") {
		if item.RequestID == input.RequestID || item.RequestNo == requestNo {
			if item.UserID != userID || item.TplID != input.TplID || item.Quantity != num || item.RequestNo != requestNo {
				return result, false, apiError(http.StatusConflict, "migration_request_reused", "The request identifier is already in use")
			}
			return item, true, nil
		}
	}
	preview, err := s.PreviewExternalAssetMigration(input)
	if err != nil {
		return result, false, err
	}
	if !preview.Ready || preview.Mapping == nil {
		return result, false, apiError(http.StatusConflict, preview.Reason, "The requested exchange is not available")
	}
	if input.MappingVersion != "" && (input.MappingVersion != preview.Mapping.Version || input.CreditYield != preview.Mapping.CreditYield) {
		return result, false, apiError(http.StatusConflict, "migration_terms_changed", "Exchange terms changed; review the exchange again")
	}
	grant, err := RedemptionCLIPGrant(preview.Mapping.CreditYield)
	if err != nil {
		return result, false, err
	}
	idempotent := false
	err = s.store.Update(func(state *store.State) error {
		mapping, mapped := findExternalAssetMapping(state.ExternalAssetMappings, input.TplID)
		if !mapped || mapping.Version != preview.Mapping.Version || mapping.CreditYield != preview.Mapping.CreditYield {
			return apiError(http.StatusConflict, "migration_terms_changed", "Exchange terms changed; review the exchange again")
		}
		if input.WalletAddress != nil && (!strings.EqualFold(*input.WalletAddress, state.Profile.Wallet) || input.ChainID != state.Profile.WalletChainID) {
			return apiError(http.StatusConflict, "migration_wallet_changed", "The wallet has changed; review the exchange again")
		}
		for _, item := range state.MigrationRequests {
			if item.RequestID == input.RequestID || item.RequestNo == requestNo {
				if item.UserID != userID || item.TplID != input.TplID || item.Quantity != num || item.RequestNo != requestNo {
					return apiError(http.StatusConflict, "migration_request_reused", "The request identifier is already in use")
				}
				result, idempotent = item, true
				return nil
			}
			if item.UserID == userID && item.TplID == input.TplID && (item.Status == "pending_review" || item.Status == "processing" || item.Status == "attention_required") {
				return apiError(http.StatusConflict, "migration_request_pending", "This template already has an outstanding request")
			}
		}
		now := s.now().UTC()
		result = domain.MigrationRequest{ID: uniqueID("migration-request", now), UserID: userID, ExternalUserID: binding.ExternalUserID, RequestID: input.RequestID, RequestNo: requestNo, TplID: input.TplID, Name: preview.Template.Name, Quantity: num, MappingVersion: preview.Mapping.Version, CreditYield: preview.Mapping.CreditYield, ClipGrant: grant, WalletAddress: state.Profile.Wallet, ChainID: state.Profile.WalletChainID, Status: "pending_review", CreatedAt: now.Format(time.RFC3339), UpdatedAt: now.Format(time.RFC3339)}
		state.MigrationRequests = prepend(result, state.MigrationRequests)
		return nil
	})
	return result, idempotent, err
}

func (s *Service) ReviewMigrationRequest(id, action, reason string, accepted bool) (domain.MigrationRequest, error) {
	var request domain.MigrationRequest
	if !accepted || (action != "approve" && action != "reject") {
		return request, apiError(http.StatusBadRequest, "review_not_confirmed", "Confirm an approval or rejection")
	}
	if action == "reject" && strings.TrimSpace(reason) == "" {
		return request, apiError(http.StatusBadRequest, "reason_required", "A rejection reason is required")
	}
	err := s.store.Update(func(state *store.State) error {
		for i := range state.MigrationRequests {
			item := &state.MigrationRequests[i]
			if item.ID != id {
				continue
			}
			if item.Status == "completed" || item.Status == "rejected" {
				request = *item
				return nil
			}
			if item.Status == "processing" {
				return apiError(http.StatusConflict, "migration_processing", "This request is already processing")
			}
			if action == "reject" && item.Status != "pending_review" {
				return apiError(http.StatusConflict, "migration_reconciliation_required", "An attempted write-off cannot be rejected")
			}
			if action == "approve" {
				mapping, ok := findExternalAssetMapping(state.ExternalAssetMappings, item.TplID)
				if !ok || mapping.Version != item.MappingVersion || mapping.CreditYield != item.CreditYield {
					return apiError(http.StatusConflict, "migration_terms_changed", "Exchange terms have changed; request fresh consent")
				}
				item.Status = "processing"
			} else {
				item.Status, item.Reason = "rejected", strings.TrimSpace(reason)
			}
			item.UpdatedAt = s.now().UTC().Format(time.RFC3339)
			request = *item
			return nil
		}
		return apiError(http.StatusNotFound, "migration_request_not_found", "Exchange request not found")
	})
	if err != nil || request.Status != "processing" {
		return request, err
	}
	result, _, migrationErr := s.MigrateExternalAsset(MigrateExternalAssetInput{UserID: request.UserID, ExternalUserID: request.ExternalUserID, TplID: request.TplID, Num: request.Quantity, RequestID: request.RequestID, RequestNo: request.RequestNo, Accepted: true, MappingVersion: request.MappingVersion, CreditYield: request.CreditYield, recipientWallet: &request.WalletAddress, recipientChain: request.ChainID})
	err = s.store.Update(func(state *store.State) error {
		for i := range state.MigrationRequests {
			item := &state.MigrationRequests[i]
			if item.ID != id {
				continue
			}
			item.Status, item.Reason, item.MigrationID = "completed", "", result.Migration.ID
			if migrationErr != nil {
				item.Status, item.Reason = "pending_review", migrationErr.Error()
				for _, migration := range state.ExternalMigrations {
					if migration.RequestNo == item.RequestNo {
						item.Status = "attention_required"
					}
				}
			}
			item.UpdatedAt = s.now().UTC().Format(time.RFC3339)
			request = *item
		}
		return nil
	})
	if err != nil {
		return request, err
	}
	return request, migrationErr
}
