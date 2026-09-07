package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/store"
)

// ExternalPlatformClient is the integration seam for the partner platform.
// The partner owns SMS delivery, code validation, user records and asset
// ownership. Clipli only stores the resulting binding and operation audit.
type ExternalPlatformClient interface {
	SendVerificationCode(ctx context.Context, input VerificationCodeRequest) (VerificationCodeDispatch, error)
	BindUser(ctx context.Context, input BindUserRequest) (ExternalBindingResult, error)
	RedeemAsset(ctx context.Context, input ExternalRedemptionRequest) (ExternalRedemptionResult, error)
}

// ExternalAssetPortfolio is optional because Haiwen exposes only template
// counts, not itemized asset records.
type ExternalAssetPortfolio interface {
	ListAssets(ctx context.Context, userID string) ([]domain.ExternalAssetHolding, error)
}

// ExternalAssetCounter is implemented by adapters that support the partner's
// narrow template-count endpoint. The service never substitutes itemized
// holdings for this contract because a count is not an asset identity.
type ExternalAssetCounter interface {
	ListAssetCounts(ctx context.Context, externalUserID string, tplIDs []int64) ([]domain.ExternalTemplateCount, error)
}

// ExternalBindingStatusReader is optional but should be implemented by
// production adapters. Local binding state is never treated as authoritative
// when the partner can confirm it.
type ExternalBindingStatusReader interface {
	GetBindingStatus(ctx context.Context, externalUserID string) (ExternalBindingStatus, error)
}

// ExternalTemplateCatalog is implemented by adapters that expose the source
// platform's copyright-template catalog.
type ExternalTemplateCatalog interface {
	ListTemplates(ctx context.Context, page, pageSize int, workID int64) (ExternalTemplatePage, error)
}

type ExternalWorkCatalog interface {
	ListWorks(ctx context.Context, page, pageSize int) (ExternalWorkPage, error)
}

type ExternalTemplateWriter interface {
	WriteOffTemplate(ctx context.Context, input ExternalTemplateWriteOffRequest) (ExternalRedemptionResult, error)
}

type ExternalItemizedAssetCapability interface {
	SupportsItemizedAssets() bool
}

type ExternalTemplatePage struct {
	Items    []domain.ExternalAssetTemplate `json:"items"`
	Total    int                            `json:"total"`
	Page     int                            `json:"page"`
	PageSize int                            `json:"pageSize"`
}

type ExternalWorkPage struct {
	Items    []domain.ExternalWork `json:"items"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
}

type ExternalTemplateWriteOffRequest struct {
	ExternalUserID string `json:"externalUserId"`
	TplID          int64  `json:"tplId"`
	Num            int    `json:"num"`
	RequestNo      string `json:"requestNo"`
}

type VerificationCodeRequest struct {
	UserID         string `json:"userId,omitempty"`
	ExternalUserID string `json:"externalUserId,omitempty"`
	Phone          string `json:"phone"`
	RequestID      string `json:"requestId,omitempty"`
}

type VerificationCodeDispatch struct {
	DeliveryID string `json:"deliveryId,omitempty"`
	ExpiresAt  string `json:"expiresAt,omitempty"`
	Status     string `json:"status,omitempty"`
}

type BindUserRequest struct {
	UserID           string `json:"userId,omitempty"`
	ExternalUserID   string `json:"externalUserId,omitempty"`
	Phone            string `json:"phone"`
	Code             string `json:"code,omitempty"`
	SMSCode          string `json:"smsCode,omitempty"`
	VerificationCode string `json:"verificationCode,omitempty"`
	VerificationID   string `json:"verificationId,omitempty"`
	RequestID        string `json:"requestId,omitempty"`
}

type ExternalBindingResult struct {
	ExternalUserID string `json:"externalUserId"`
	UserID         string `json:"userId,omitempty"`
	BindingID      string `json:"bindingId,omitempty"`
	Status         string `json:"status,omitempty"`
	Bound          bool   `json:"bound,omitempty"`
	BoundAt        string `json:"boundAt,omitempty"`
}

type ExternalRedemptionRequest struct {
	UserID         string `json:"userId,omitempty"`
	ExternalUserID string `json:"externalUserId"`
	AssetID        string `json:"assetId,omitempty"`
	SerialNumber   string `json:"serialNumber,omitempty"`
	RequestNo      string `json:"requestNo,omitempty"`
	TplID          int64  `json:"tplId,omitempty"`
	Num            int    `json:"num,omitempty"`
	RequestID      string `json:"requestId,omitempty"`
}

type ExternalRedemptionResult struct {
	ExternalTxID   string `json:"externalTxId,omitempty"`
	TransactionID  string `json:"transactionId,omitempty"`
	Status         string `json:"status,omitempty"`
	Quantity       int    `json:"quantity,omitempty"`
	RequestNo      string `json:"requestNo,omitempty"`
	ExternalUserID string `json:"externalUserId,omitempty"`
	TplID          int64  `json:"tplId,omitempty"`
	Num            int    `json:"num,omitempty"`
	WriteOffAt     string `json:"writeOffAt,omitempty"`
}

// PlatformError is returned by an adapter when the partner rejected a
// request or is temporarily unavailable. Status is the partner HTTP status
// when known; Clipli maps server errors to a 502 response.
type PlatformError struct {
	Status  int
	Code    string
	Message string
}

func (e *PlatformError) Error() string {
	if e.Code != "" {
		return e.Code
	}
	return e.Message
}

func mapPlatformError(err error) error {
	if err == nil {
		return nil
	}
	var platformErr *PlatformError
	if errors.As(err, &platformErr) {
		status := platformErr.Status
		if status < 400 || status >= 500 {
			status = http.StatusBadGateway
		}
		code, message := platformErr.Code, platformErr.Message
		if code == "" {
			code = "external_platform_error"
		}
		if message == "" {
			message = "The external platform could not complete the request"
		}
		return apiError(status, code, message)
	}
	return apiError(http.StatusBadGateway, "external_platform_unavailable", "The external platform is temporarily unavailable")
}

type SendVerificationInput struct {
	UserID         string `json:"userId"`
	ClipliUserID   string `json:"clipliUserId,omitempty"`
	ExternalUserID string `json:"externalUserId,omitempty"`
	Phone          string `json:"phone"`
	Mobile         string `json:"mobile,omitempty"`
	PhoneNumber    string `json:"phoneNumber,omitempty"`
	RequestID      string `json:"requestId,omitempty"`
}

type VerificationResult struct {
	Challenge domain.VerificationChallenge `json:"challenge"`
}

// SendVerificationCode asks the partner to deliver a code and records only
// non-sensitive delivery metadata in Clipli.
func (s *Service) SendVerificationCode(input SendVerificationInput) (VerificationResult, error) {
	var result VerificationResult
	userID := normalizeUserID(input.UserID)
	if userID == "" {
		userID = normalizeUserID(input.ClipliUserID)
	}
	phoneValue := input.Phone
	if phoneValue == "" {
		phoneValue = input.Mobile
	}
	if phoneValue == "" {
		phoneValue = input.PhoneNumber
	}
	phone, ok := normalizePhone(phoneValue)
	if !validUserID(userID) {
		return result, apiError(http.StatusBadRequest, "invalid_user_id", "A valid userId is required")
	}
	externalUserID := strings.TrimSpace(input.ExternalUserID)
	if !validUserID(externalUserID) {
		return result, apiError(http.StatusBadRequest, "invalid_external_user_id", "A valid externalUserId is required; it cannot be inferred from a Clipli userId")
	}
	if !ok {
		return result, apiError(http.StatusBadRequest, "invalid_phone", "A valid phone number is required")
	}
	if input.RequestID != "" && !validRequestID(input.RequestID) {
		return result, apiError(http.StatusBadRequest, "invalid_request_id", "A valid operation id is required")
	}
	s.externalMu.Lock()
	defer s.externalMu.Unlock()
	if input.RequestID != "" {
		for _, item := range s.Snapshot().VerificationChallenges {
			if item.RequestID == input.RequestID {
				if item.UserID != userID || item.ExternalUserID != externalUserID || item.Phone != phone {
					return result, apiError(http.StatusConflict, "request_id_reused", "The requestId is already used for another verification request")
				}
				return VerificationResult{Challenge: item}, nil
			}
		}
	}
	dispatch, err := s.platform.SendVerificationCode(context.Background(), VerificationCodeRequest{UserID: userID, ExternalUserID: externalUserID, Phone: phone, RequestID: input.RequestID})
	if err != nil {
		return result, mapPlatformError(err)
	}
	now := s.now().UTC()
	expiresAt := dispatch.ExpiresAt
	expiresAtSource := "clipli-local"
	if expiresAt == "" {
		expiresAt = now.Add(5 * time.Minute).Format(time.RFC3339)
	} else if _, parseErr := time.Parse(time.RFC3339, expiresAt); parseErr != nil {
		expiresAt = now.Add(5 * time.Minute).Format(time.RFC3339)
	} else {
		expiresAtSource = "adapter-reported"
	}
	// Any successful dispatch is a pending challenge from Clipli's point of
	// view; partner-specific delivery states are not used for local matching.
	status := "sent"
	deliveryID := strings.TrimSpace(dispatch.DeliveryID)
	deliveryIDSource := ""
	if deliveryID != "" {
		deliveryIDSource = "adapter-reported"
	}
	challenge := domain.VerificationChallenge{
		ID: uniqueID("verification", now), UserID: userID, ExternalUserID: externalUserID, Phone: phone, PhoneMasked: maskPhone(phone),
		DeliveryID: deliveryID, DeliveryIDSource: deliveryIDSource, Status: status, ExpiresAt: expiresAt, ExpiresAtSource: expiresAtSource, CreatedAt: now.Format(time.RFC3339), RequestID: input.RequestID,
	}
	err = s.store.Update(func(state *store.State) error {
		state.VerificationChallenges = prepend(challenge, state.VerificationChallenges)
		result = VerificationResult{Challenge: challenge}
		return nil
	})
	return result, err
}

type BindUserInput struct {
	UserID           string `json:"userId"`
	ClipliUserID     string `json:"clipliUserId,omitempty"`
	ExternalUserID   string `json:"externalUserId,omitempty"`
	Phone            string `json:"phone"`
	Mobile           string `json:"mobile,omitempty"`
	PhoneNumber      string `json:"phoneNumber,omitempty"`
	Code             string `json:"code,omitempty"`
	SMSCode          string `json:"smsCode,omitempty"`
	VerificationCode string `json:"verificationCode,omitempty"`
	VerificationID   string `json:"verificationId,omitempty"`
	RequestID        string `json:"requestId,omitempty"`
}

type BindingResult struct {
	Binding domain.ExternalPlatformBinding `json:"binding"`
}

// BindUser delegates phone/code verification to the partner and persists the
// resulting Clipli-to-external user relation only after successful validation.
func (s *Service) BindUser(input BindUserInput) (BindingResult, bool, error) {
	var result BindingResult
	idempotent := false
	userID := normalizeUserID(input.UserID)
	if userID == "" {
		userID = normalizeUserID(input.ClipliUserID)
	}
	phoneValue := input.Phone
	if phoneValue == "" {
		phoneValue = input.Mobile
	}
	if phoneValue == "" {
		phoneValue = input.PhoneNumber
	}
	phone := ""
	phoneOK := true
	if phoneValue != "" {
		phone, phoneOK = normalizePhone(phoneValue)
	}
	code := strings.TrimSpace(input.Code)
	if code == "" {
		code = strings.TrimSpace(input.SMSCode)
	}
	if code == "" {
		code = strings.TrimSpace(input.VerificationCode)
	}
	if !validUserID(userID) {
		return result, false, apiError(http.StatusBadRequest, "invalid_user_id", "A valid userId is required")
	}
	if phoneValue != "" && !phoneOK {
		return result, false, apiError(http.StatusBadRequest, "invalid_phone", "A valid phone number is required")
	}
	if !validVerificationCode(code) {
		return result, false, apiError(http.StatusBadRequest, "invalid_verification_code", "A 4 to 8 digit verification code is required")
	}
	if input.RequestID != "" && !validRequestID(input.RequestID) {
		return result, false, apiError(http.StatusBadRequest, "invalid_request_id", "A valid operation id is required")
	}
	s.externalMu.Lock()
	defer s.externalMu.Unlock()
	var challenge domain.VerificationChallenge
	state := s.Snapshot()
	for _, item := range state.Bindings {
		if input.RequestID != "" && item.RequestID == input.RequestID && item.UserID != userID {
			return result, false, apiError(http.StatusConflict, "request_id_reused", "The requestId is already used for another binding")
		}
		if (input.RequestID != "" && item.RequestID == input.RequestID) || (item.UserID == userID && item.Status == "bound") {
			if requestedExternal := strings.TrimSpace(input.ExternalUserID); requestedExternal != "" && requestedExternal != item.ExternalUserID {
				return result, false, apiError(http.StatusConflict, "binding_conflict", "The Clipli user is already bound to another external user")
			}
			result, idempotent = BindingResult{Binding: item}, true
			return result, idempotent, nil
		}
	}
	requestedExternalUserID := strings.TrimSpace(input.ExternalUserID)
	for _, item := range state.VerificationChallenges {
		if item.UserID == userID && (requestedExternalUserID == "" || item.ExternalUserID == requestedExternalUserID) && (phone == "" || item.Phone == phone) && item.Status == "sent" && (input.VerificationID == "" || item.ID == input.VerificationID) {
			challenge = item
			break
		}
	}
	if challenge.ID == "" {
		return result, false, apiError(http.StatusBadRequest, "verification_required", "Request a verification code before binding")
	}
	if expires, err := time.Parse(time.RFC3339, challenge.ExpiresAt); err == nil && !s.now().Before(expires) {
		return result, false, apiError(http.StatusBadRequest, "verification_expired", "The verification code has expired")
	}
	if phone == "" {
		phone = challenge.Phone
	}
	externalUserID := strings.TrimSpace(input.ExternalUserID)
	if externalUserID == "" {
		externalUserID = strings.TrimSpace(challenge.ExternalUserID)
	}
	if !validUserID(externalUserID) {
		return result, false, apiError(http.StatusBadRequest, "invalid_external_user_id", "A valid externalUserId is required")
	}
	external, err := s.platform.BindUser(context.Background(), BindUserRequest{UserID: userID, ExternalUserID: externalUserID, Phone: phone, Code: code, SMSCode: code, VerificationCode: code, VerificationID: challenge.ID, RequestID: input.RequestID})
	if err != nil {
		return result, false, mapPlatformError(err)
	}
	if !external.Bound {
		return result, false, apiError(http.StatusBadGateway, "external_binding_not_confirmed", "The external platform did not confirm bound=true")
	}
	externalUserID = strings.TrimSpace(external.ExternalUserID)
	if externalUserID == "" {
		externalUserID = strings.TrimSpace(external.UserID)
	}
	if externalUserID == "" {
		return result, false, apiError(http.StatusBadGateway, "external_binding_invalid", "The external platform did not return an external user ID")
	}
	expectedExternalUserID := strings.TrimSpace(input.ExternalUserID)
	if expectedExternalUserID == "" {
		expectedExternalUserID = strings.TrimSpace(challenge.ExternalUserID)
	}
	if externalUserID != expectedExternalUserID {
		return result, false, apiError(http.StatusConflict, "external_binding_mismatch", "The external platform returned a different external user ID")
	}
	now := s.now().UTC()
	// A successful adapter response establishes the Clipli-side bound state;
	// partner-specific status values are intentionally not used as local state.
	status := "bound"
	boundAt := strings.TrimSpace(external.BoundAt)
	if parsed, parseErr := time.Parse(time.RFC3339, boundAt); parseErr != nil || parsed.IsZero() {
		boundAt = now.Format(time.RFC3339)
	}
	binding := domain.ExternalPlatformBinding{ID: external.BindingID, UserID: userID, ExternalUserID: externalUserID, Phone: phone, PhoneMasked: maskPhone(phone), PlatformCode: "external-platform", Status: status, BoundAt: boundAt, VerificationID: challenge.ID, RequestID: input.RequestID}
	if binding.ID == "" {
		binding.ID = uniqueID("binding", now)
	}
	err = s.store.Update(func(state *store.State) error {
		// A concurrent/retried bind must not create a second relation.
		for _, item := range state.Bindings {
			if item.UserID == userID && item.Status == "bound" {
				result, idempotent = BindingResult{Binding: item}, true
				return nil
			}
		}
		state.Bindings = prepend(binding, state.Bindings)
		for index := range state.VerificationChallenges {
			if state.VerificationChallenges[index].ID == challenge.ID {
				state.VerificationChallenges[index].Status = "consumed"
				state.VerificationChallenges[index].ConsumedAt = now.Format(time.RFC3339)
				break
			}
		}
		result = BindingResult{Binding: binding}
		return nil
	})
	return result, idempotent, err
}

type UserAssetsResult struct {
	UserID         string                        `json:"userId"`
	ExternalUserID string                        `json:"externalUserId"`
	Items          []domain.ExternalAssetHolding `json:"items"`
	AssetCount     int                           `json:"assetCount"`
	TotalQuantity  int                           `json:"totalQuantity"`
	RetrievedAt    string                        `json:"retrievedAt"`
	SourceOfTruth  string                        `json:"sourceOfTruth"`
}

type ExternalAssetCount struct {
	TplID int64 `json:"tplId"`
	Count int   `json:"count"`
}

type UserAssetCountsResult struct {
	UserID         string               `json:"userId"`
	ExternalUserID string               `json:"externalUserId"`
	Items          []ExternalAssetCount `json:"items"`
	RetrievedAt    string               `json:"retrievedAt"`
	SourceOfTruth  string               `json:"sourceOfTruth"`
}

func (s *Service) UserAssets(userID string) (UserAssetsResult, error) {
	var result UserAssetsResult
	userID = normalizeUserID(userID)
	if !validUserID(userID) {
		return result, apiError(http.StatusBadRequest, "invalid_user_id", "A valid userId is required")
	}
	binding, err := s.verifiedBinding(userID)
	if err != nil {
		return result, err
	}
	portfolio, supported := s.platform.(ExternalAssetPortfolio)
	if !supported {
		return result, apiError(http.StatusNotImplemented, "external_asset_portfolio_unsupported", "Haiwen exposes template counts, not itemized asset holdings")
	}
	items, err := portfolio.ListAssets(context.Background(), binding.ExternalUserID)
	if err != nil {
		return result, mapPlatformError(err)
	}
	if items == nil {
		items = make([]domain.ExternalAssetHolding, 0)
	}
	total := 0
	for _, item := range items {
		if item.Quantity > 0 {
			total += item.Quantity
		}
	}
	result = UserAssetsResult{UserID: userID, ExternalUserID: binding.ExternalUserID, Items: items, AssetCount: len(items), TotalQuantity: total, RetrievedAt: s.now().UTC().Format(time.RFC3339), SourceOfTruth: "external-platform"}
	return result, nil
}

// UserAssetCounts queries the partner's exact template-count contract. The
// caller must provide one to one hundred positive template IDs, matching the
// external API's required tplIds query parameter.
func (s *Service) UserAssetCounts(userID string, tplIDs []int64) (UserAssetCountsResult, error) {
	var result UserAssetCountsResult
	userID = normalizeUserID(userID)
	if !validUserID(userID) {
		return result, apiError(http.StatusBadRequest, "invalid_user_id", "A valid userId is required")
	}
	if len(tplIDs) == 0 || len(tplIDs) > 100 {
		return result, apiError(http.StatusBadRequest, "invalid_tpl_ids", "Provide between 1 and 100 template IDs")
	}
	for _, tplID := range tplIDs {
		if tplID <= 0 {
			return result, apiError(http.StatusBadRequest, "invalid_tpl_ids", "Template IDs must be positive integers")
		}
	}
	binding, err := s.verifiedBinding(userID)
	if err != nil {
		return result, err
	}
	counter, ok := s.platform.(ExternalAssetCounter)
	if !ok {
		return result, apiError(http.StatusNotImplemented, "asset_count_not_supported", "The configured external platform does not support template counts")
	}
	items, err := counter.ListAssetCounts(context.Background(), binding.ExternalUserID, tplIDs)
	if err != nil {
		return result, mapPlatformError(err)
	}
	if items == nil {
		items = make([]domain.ExternalTemplateCount, 0)
	}
	counts := make([]ExternalAssetCount, 0, len(items))
	countsByTemplate := make(map[int64]int, len(items))
	for _, item := range items {
		if item.TplID <= 0 || item.Count < 0 {
			return result, apiError(http.StatusBadGateway, "invalid_external_counts", "Haiwen returned invalid quantities")
		}
		if _, duplicate := countsByTemplate[item.TplID]; duplicate {
			return result, apiError(http.StatusBadGateway, "invalid_external_counts", "Haiwen returned duplicate template quantities")
		}
		countsByTemplate[item.TplID] = item.Count
	}
	for _, tplID := range tplIDs {
		count, present := countsByTemplate[tplID]
		if !present {
			return result, apiError(http.StatusBadGateway, "invalid_external_counts", "Haiwen omitted a requested template quantity")
		}
		counts = append(counts, ExternalAssetCount{TplID: tplID, Count: count})
	}
	result = UserAssetCountsResult{UserID: userID, ExternalUserID: binding.ExternalUserID, Items: counts, RetrievedAt: s.now().UTC().Format(time.RFC3339), SourceOfTruth: "haiwen-/openapi/user/assets/count"}
	return result, nil
}

func (s *Service) findBinding(userID string) (domain.ExternalPlatformBinding, bool) {
	for _, item := range s.Snapshot().Bindings {
		if item.UserID == userID && item.Status == "bound" {
			return item, true
		}
	}
	return domain.ExternalPlatformBinding{}, false
}

func (s *Service) verifiedBinding(userID string) (domain.ExternalPlatformBinding, error) {
	if _, ok := s.findBinding(userID); !ok {
		return domain.ExternalPlatformBinding{}, apiError(http.StatusConflict, "user_not_bound", "Bind the Clipli user to the external platform first")
	}
	binding, _, err := s.GetBinding(userID)
	return binding, err
}

func (s *Service) GetBinding(userID string) (domain.ExternalPlatformBinding, bool, error) {
	userID = normalizeUserID(userID)
	if !validUserID(userID) {
		return domain.ExternalPlatformBinding{}, false, apiError(http.StatusBadRequest, "invalid_user_id", "A valid userId is required")
	}
	binding, ok := s.findBinding(userID)
	if !ok {
		return domain.ExternalPlatformBinding{}, false, apiError(http.StatusNotFound, "binding_not_found", "The Clipli user is not bound to the external platform")
	}
	if reader, supported := s.platform.(ExternalBindingStatusReader); supported {
		status, err := reader.GetBindingStatus(context.Background(), binding.ExternalUserID)
		if err != nil {
			return domain.ExternalPlatformBinding{}, false, mapPlatformError(err)
		}
		if !status.Bound {
			_ = s.store.Update(func(state *store.State) error {
				for index := range state.Bindings {
					if state.Bindings[index].ID == binding.ID {
						state.Bindings[index].Status = "unbound"
					}
				}
				return nil
			})
			return domain.ExternalPlatformBinding{}, false, apiError(http.StatusNotFound, "binding_not_found", "The external platform no longer reports this user as bound")
		}
		if strings.TrimSpace(status.ExternalUserID) != binding.ExternalUserID {
			return domain.ExternalPlatformBinding{}, false, apiError(http.StatusConflict, "external_binding_mismatch", "The external platform returned a different external user ID")
		}
		if boundAt := validExternalBoundAt(status.BoundAt); boundAt != "" && boundAt != binding.BoundAt {
			binding.BoundAt = boundAt
			_ = s.store.Update(func(state *store.State) error {
				for index := range state.Bindings {
					if state.Bindings[index].ID == binding.ID {
						state.Bindings[index].BoundAt = boundAt
					}
				}
				return nil
			})
		}
	}
	return binding, true, nil
}

func (s *Service) ExternalBindingStatus(externalUserID string) (ExternalBindingStatus, error) {
	externalUserID = strings.TrimSpace(externalUserID)
	if !validUserID(externalUserID) {
		return ExternalBindingStatus{}, apiError(http.StatusBadRequest, "invalid_external_user_id", "A valid externalUserId is required")
	}
	reader, ok := s.platform.(ExternalBindingStatusReader)
	if !ok {
		return ExternalBindingStatus{}, apiError(http.StatusNotImplemented, "external_binding_status_unsupported", "The configured external platform does not expose binding status")
	}
	status, err := reader.GetBindingStatus(context.Background(), externalUserID)
	if err != nil {
		return ExternalBindingStatus{}, mapPlatformError(err)
	}
	if strings.TrimSpace(status.ExternalUserID) == "" {
		status.ExternalUserID = externalUserID
	} else if strings.TrimSpace(status.ExternalUserID) != externalUserID {
		return ExternalBindingStatus{}, apiError(http.StatusConflict, "external_binding_mismatch", "The external platform returned a different external user ID")
	}
	return status, nil
}

func validExternalBoundAt(value string) string {
	value = strings.TrimSpace(value)
	if parsed, err := time.Parse(time.RFC3339, value); err == nil && !parsed.IsZero() {
		return value
	}
	return ""
}

type ExternalTemplateView struct {
	Template       domain.ExternalAssetTemplate     `json:"template"`
	Mapping        *domain.ExternalAssetMappingRule `json:"mapping,omitempty"`
	MigrationReady bool                             `json:"migrationReady"`
}

type ExternalTemplatesResult struct {
	Items         []ExternalTemplateView `json:"items"`
	Total         int                    `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"pageSize"`
	RetrievedAt   string                 `json:"retrievedAt"`
	SourceOfTruth string                 `json:"sourceOfTruth"`
}

// ExternalTemplates reads the source catalog without changing ownership or
// triggering a write-off. Each source template is returned together with the
// separately configured Clipli economics mapping, when one exists.
func (s *Service) ExternalTemplates(page, pageSize int, workID int64) (ExternalTemplatesResult, error) {
	var result ExternalTemplatesResult
	catalog, ok := s.platform.(ExternalTemplateCatalog)
	if !ok {
		return result, apiError(http.StatusNotImplemented, "external_template_catalog_unsupported", "The configured external platform does not expose a template catalog")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		return result, apiError(http.StatusBadRequest, "invalid_page_size", "pageSize cannot exceed 100")
	}
	if workID < 0 {
		return result, apiError(http.StatusBadRequest, "invalid_work_id", "workId must be a positive integer")
	}
	remote, err := catalog.ListTemplates(context.Background(), page, pageSize, workID)
	if err != nil {
		return result, mapPlatformError(err)
	}
	s.cacheExternalTemplates(remote.Items)
	state := s.Snapshot()
	views := make([]ExternalTemplateView, 0, len(remote.Items))
	for _, template := range remote.Items {
		mapping, mapped := findExternalAssetMapping(state.ExternalAssetMappings, template.TplID)
		var mappingPointer *domain.ExternalAssetMappingRule
		if mapped {
			copy := mapping
			mappingPointer = &copy
		}
		views = append(views, ExternalTemplateView{Template: template, Mapping: mappingPointer, MigrationReady: mapped && len(template.Owners) > 0})
	}
	return ExternalTemplatesResult{Items: views, Total: remote.Total, Page: remote.Page, PageSize: remote.PageSize, RetrievedAt: s.now().UTC().Format(time.RFC3339), SourceOfTruth: "haiwen-/openapi/tpls"}, nil
}

type ExternalWorksResult struct {
	Items         []domain.ExternalWork `json:"items"`
	Total         int                   `json:"total"`
	Page          int                   `json:"page"`
	PageSize      int                   `json:"pageSize"`
	RetrievedAt   string                `json:"retrievedAt"`
	SourceOfTruth string                `json:"sourceOfTruth"`
}

// ExternalWorks reads Haiwen's published work catalog. publishNum is retained
// as a catalog statistic and is never used as a holding or migration amount.
func (s *Service) ExternalWorks(page, pageSize int) (ExternalWorksResult, error) {
	var result ExternalWorksResult
	catalog, ok := s.platform.(ExternalWorkCatalog)
	if !ok {
		return result, apiError(http.StatusNotImplemented, "external_work_catalog_unsupported", "The configured external platform does not expose a work catalog")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		return result, apiError(http.StatusBadRequest, "invalid_page_size", "pageSize cannot exceed 100")
	}
	remote, err := catalog.ListWorks(context.Background(), page, pageSize)
	if err != nil {
		return result, mapPlatformError(err)
	}
	s.cacheExternalWorks(remote.Items)
	return ExternalWorksResult{Items: remote.Items, Total: remote.Total, Page: remote.Page, PageSize: remote.PageSize, RetrievedAt: s.now().UTC().Format(time.RFC3339), SourceOfTruth: "haiwen-/openapi/works"}, nil
}

// SetExternalAssetMappings replaces the server-managed Haiwen economics
// mapping. It is intended for trusted startup configuration, not public input.
func (s *Service) SetExternalAssetMappings(mappings []domain.ExternalAssetMappingRule) error {
	s.externalMu.Lock()
	defer s.externalMu.Unlock()
	seen := make(map[int64]bool, len(mappings))
	validated := make([]domain.ExternalAssetMappingRule, 0, len(mappings))
	for _, mapping := range mappings {
		mapping.Version = strings.TrimSpace(mapping.Version)
		mapping.Currency = strings.ToUpper(strings.TrimSpace(mapping.Currency))
		if mapping.Currency == "" {
			mapping.Currency = "CNY"
		}
		if mapping.TplID <= 0 || mapping.TplID > 9007199254740991 || mapping.Version == "" || mapping.CreditYield <= 0 || mapping.CreditYield > 1000000000 || mapping.ClipPrice < 0 || mapping.ClipPrice > 1000000000 || seen[mapping.TplID] {
			return apiError(http.StatusBadRequest, "invalid_external_asset_mapping", "Each mapping requires a unique tplId, version, and positive creditYield")
		}
		seen[mapping.TplID] = true
		validated = append(validated, mapping)
	}
	return s.store.Update(func(state *store.State) error {
		state.ExternalAssetMappings = validated
		return nil
	})
}

type MigrateExternalAssetInput struct {
	UserID          string  `json:"userId"`
	ClipliUserID    string  `json:"clipliUserId,omitempty"`
	ExternalUserID  string  `json:"externalUserId,omitempty"`
	TplID           int64   `json:"tplId"`
	Num             int     `json:"num"`
	RequestNo       string  `json:"requestNo"`
	RequestID       string  `json:"requestId"`
	Accepted        bool    `json:"accepted"`
	WalletAddress   *string `json:"walletAddress,omitempty"`
	ChainID         string  `json:"chainId,omitempty"`
	MappingVersion  string  `json:"mappingVersion,omitempty"`
	CreditYield     int     `json:"creditYield,omitempty"`
	recipientWallet *string
	recipientChain  string
}

type ExternalMigrationPreview struct {
	Template          domain.ExternalAssetTemplate     `json:"template"`
	Mapping           *domain.ExternalAssetMappingRule `json:"mapping,omitempty"`
	AvailableQuantity int                              `json:"availableQuantity"`
	RequestedQuantity int                              `json:"requestedQuantity"`
	CandidateAssets   []domain.HAPWAsset               `json:"candidateAssets"`
	Ready             bool                             `json:"ready"`
	Reason            string                           `json:"reason"`
}

type ExternalMigrationResult struct {
	Migration domain.ExternalAssetMigration `json:"migration"`
	Assets    []domain.HAPWAsset            `json:"assets"`
}

// PreviewExternalAssetMigration performs only catalog and holding reads. It
// never creates local assets and never calls Haiwen's write-off endpoint.
func (s *Service) PreviewExternalAssetMigration(input MigrateExternalAssetInput) (ExternalMigrationPreview, error) {
	var preview ExternalMigrationPreview
	userID, binding, num, requestNo, err := s.validateExternalMigrationInput(input, false)
	if err != nil {
		return preview, err
	}
	template, err := s.externalTemplate(input.TplID)
	if err != nil {
		return preview, err
	}
	count, err := s.externalHoldingCount(binding.ExternalUserID, input.TplID)
	if err != nil {
		return preview, err
	}
	mapping, mapped := findExternalAssetMapping(s.Snapshot().ExternalAssetMappings, input.TplID)
	preview = ExternalMigrationPreview{Template: template, CandidateAssets: []domain.HAPWAsset{}, AvailableQuantity: count, RequestedQuantity: num, Reason: "mapping_required"}
	if len(template.Owners) == 0 {
		preview.Reason = "external_rights_owner_missing"
		return preview, nil
	}
	if mapped {
		copy := mapping
		preview.Mapping = &copy
		preview.CandidateAssets = buildMigratedAssets(template, mapping, userID, requestNo, num, s.now().UTC(), false)
		preview.Ready = count >= num
		if preview.Ready {
			preview.Reason = "ready"
		} else {
			preview.Reason = "insufficient_external_assets"
		}
	}
	return preview, nil
}

// MigrateExternalAsset mirrors source metadata into Clipli, writes off the
// same quantity at Haiwen, then atomically activates and settles the new
// Clipli assets with Creation Credits and CLIP.
func (s *Service) MigrateExternalAsset(input MigrateExternalAssetInput) (ExternalMigrationResult, bool, error) {
	if s.ExternalPlatformSandbox() {
		return ExternalMigrationResult{}, false, apiError(http.StatusConflict, "external_sandbox_transfer_disabled", "Sandbox platform transfers cannot create real Clipli assets")
	}
	var result ExternalMigrationResult
	userID, binding, num, requestNo, err := s.validateExternalMigrationInput(input, true)
	if err != nil {
		return result, false, err
	}
	s.externalMu.Lock()
	defer s.externalMu.Unlock()

	state := s.Snapshot()
	var migration domain.ExternalAssetMigration
	for _, existing := range state.ExternalMigrations {
		if existing.RequestID == input.RequestID || existing.RequestNo == requestNo {
			if existing.UserID != userID || existing.TplID != input.TplID || existing.Quantity != num || existing.RequestNo != requestNo {
				return result, false, apiError(http.StatusConflict, "migration_request_reused", "The requestId or requestNo is already used for another migration")
			}
			if existing.Status == "completed_rewards_settled" {
				return migrationResult(state, existing), true, nil
			}
			if existing.Status == "reconciliation_required" {
				return result, false, apiError(http.StatusConflict, "migration_reconciliation_required", "This migration has a mismatched Haiwen response and requires manual reconciliation")
			}
			migration = existing
			break
		}
	}
	if migration.ID == "" {
		for _, redemption := range state.ExternalRedemptions {
			if redemption.UserID == userID && redemption.RequestNo == requestNo {
				return result, false, apiError(http.StatusConflict, "external_write_off_already_recorded", "The requestNo is already used by an external write-off")
			}
		}
	}

	template, err := s.externalTemplate(input.TplID)
	if err != nil {
		return result, false, err
	}
	if len(template.Owners) == 0 {
		return result, false, apiError(http.StatusConflict, "external_rights_owner_missing", "The template does not identify its rights owner")
	}
	mapping, mapped := findExternalAssetMapping(s.Snapshot().ExternalAssetMappings, input.TplID)
	if !mapped {
		return result, false, apiError(http.StatusConflict, "external_asset_mapping_required", "Configure an active versioned mapping before migrating this template")
	}
	if input.MappingVersion != "" && (input.MappingVersion != mapping.Version || input.CreditYield != mapping.CreditYield) {
		return result, false, apiError(http.StatusConflict, "migration_terms_changed", "Exchange terms changed; review the original request")
	}
	if migration.ID == "" {
		available, countErr := s.externalHoldingCount(binding.ExternalUserID, input.TplID)
		if countErr != nil {
			return result, false, countErr
		}
		if available < num {
			return result, false, apiError(http.StatusUnprocessableEntity, "external_assets_insufficient", "The external user does not hold enough assets for this migration")
		}
	}

	plannedClip := 0
	if migration.ID == "" {
		grant, grantErr := RedemptionCLIPGrant(mapping.CreditYield)
		if grantErr != nil {
			return result, false, grantErr
		}
		plannedClip = grant * num
	} else {
		for _, asset := range migrationResult(state, migration).Assets {
			grant, grantErr := RedemptionCLIPGrant(asset.CreditYield)
			if grantErr != nil {
				return result, false, grantErr
			}
			plannedClip += grant
		}
	}
	if state.CLIPTreasury.TreasuryBalance+migration.TreasuryReserved < plannedClip {
		return result, false, apiError(http.StatusServiceUnavailable, "clip_treasury_insufficient", "The platform treasury cannot satisfy this CLIP distribution")
	}

	now := s.now().UTC()
	assets := migrationResult(state, migration).Assets
	if migration.ID == "" {
		wallet, chain := state.Profile.Wallet, state.Profile.WalletChainID
		if input.recipientWallet != nil {
			wallet, chain = *input.recipientWallet, input.recipientChain
		}
		assets = buildMigratedAssets(template, mapping, migrationOwner(userID, wallet), requestNo, num, now, false)
		assetIDs := make([]string, len(assets))
		for index := range assets {
			assetIDs[index] = assets[index].ID
		}
		migration = domain.ExternalAssetMigration{ID: uniqueID("external-migration", now), RequestID: input.RequestID, RequestNo: requestNo, UserID: userID, ExternalUserID: binding.ExternalUserID, TplID: input.TplID, Quantity: num, MappingVersion: mapping.Version, ClipliAssetIDs: assetIDs, Status: "pending_external_write_off", CreatedAt: now.Format(time.RFC3339), UpdatedAt: now.Format(time.RFC3339)}
		migration.TreasuryReserved = plannedClip
		migration.WalletAddress, migration.ChainID = wallet, chain
		if err := s.store.Update(func(current *store.State) error {
			if current.CLIPTreasury.TreasuryBalance < plannedClip {
				return apiError(http.StatusServiceUnavailable, "clip_treasury_insufficient", "The platform treasury cannot reserve this distribution")
			}
			current.CLIPTreasury.TreasuryBalance -= plannedClip
			current.Assets = append(assets, current.Assets...)
			current.ExternalMigrations = prepend(migration, current.ExternalMigrations)
			cacheExternalTemplates(current, []domain.ExternalAssetTemplate{template})
			return nil
		}); err != nil {
			return result, false, err
		}
	} else {
		if len(assets) != num {
			return result, false, apiError(http.StatusConflict, "migration_state_invalid", "The pending migration does not contain the expected Clipli assets")
		}
		_ = s.store.Update(func(current *store.State) error {
			for index := range current.ExternalMigrations {
				if current.ExternalMigrations[index].ID == migration.ID {
					current.ExternalMigrations[index].Status = "pending_external_write_off"
					current.ExternalMigrations[index].FailureReason = ""
					current.ExternalMigrations[index].UpdatedAt = now.Format(time.RFC3339)
				}
			}
			for index := range current.Assets {
				for _, assetID := range migration.ClipliAssetIDs {
					if current.Assets[index].ID == assetID {
						current.Assets[index].Status, current.Assets[index].StatusEn = "等待海文发核销", "Pending Haiwen write-off"
						current.Assets[index].Transferable, current.Assets[index].ExchangeAvailable = false, false
						current.Assets[index].RedemptionStatus = "pending_external_write_off"
					}
				}
			}
			return nil
		})
	}

	markFailed := func(status, reason string) {
		_ = s.store.Update(func(current *store.State) error {
			for index := range current.ExternalMigrations {
				if current.ExternalMigrations[index].ID == migration.ID {
					current.ExternalMigrations[index].Status = status
					current.ExternalMigrations[index].FailureReason = reason
					current.ExternalMigrations[index].UpdatedAt = s.now().UTC().Format(time.RFC3339)
				}
			}
			for index := range current.Assets {
				for _, assetID := range migration.ClipliAssetIDs {
					if current.Assets[index].ID == assetID {
						current.Assets[index].Status, current.Assets[index].StatusEn = "迁移待处理", "Migration requires attention"
						current.Assets[index].Transferable, current.Assets[index].ExchangeAvailable = false, false
						current.Assets[index].RedemptionStatus = status
						current.Assets[index].Provenance.VerificationStatus = status
					}
				}
			}
			return nil
		})
	}

	writer, supported := s.platform.(ExternalTemplateWriter)
	if !supported {
		markFailed("external_write_off_failed", "The configured adapter does not expose Haiwen's template batch write-off contract")
		return result, false, apiError(http.StatusNotImplemented, "external_template_write_off_unsupported", "The configured external platform does not support template batch write-off")
	}
	external, err := writer.WriteOffTemplate(context.Background(), ExternalTemplateWriteOffRequest{ExternalUserID: binding.ExternalUserID, RequestNo: requestNo, TplID: input.TplID, Num: num})
	if err != nil {
		status := "external_write_off_failed"
		var platformErr *PlatformError
		if errors.As(err, &platformErr) && platformErr.Code == "external_migration_response_mismatch" {
			status = "reconciliation_required"
		}
		markFailed(status, err.Error())
		return result, false, mapPlatformError(err)
	}
	if external.RequestNo != requestNo || external.Status != "SUCCESS" || external.TplID != input.TplID || external.Num != num || (external.ExternalUserID != "" && external.ExternalUserID != binding.ExternalUserID) || (external.Quantity > 0 && external.Quantity != num) {
		markFailed("reconciliation_required", "Haiwen response tplId or quantity did not match the migration request")
		return result, false, apiError(http.StatusBadGateway, "external_migration_response_mismatch", "Haiwen accepted the request but returned mismatched migration details; manual reconciliation is required")
	}
	externalTxID := strings.TrimSpace(external.ExternalTxID)
	if externalTxID == "" {
		externalTxID = strings.TrimSpace(external.TransactionID)
	}
	err = s.store.Update(func(current *store.State) error {
		current.CLIPTreasury.TreasuryBalance += migration.TreasuryReserved
		grantByAsset := make(map[string]int, len(migration.ClipliAssetIDs))
		totalGrant := 0
		for _, assetID := range migration.ClipliAssetIDs {
			asset, ok := findAsset(current.Assets, assetID)
			if !ok {
				return apiError(http.StatusConflict, "migration_state_invalid", "The migration references a missing Clipli asset")
			}
			alreadySettled := false
			for _, existingRedemption := range current.Redemptions {
				if existingRedemption.AssetID == assetID && existingRedemption.Status == "有效" {
					alreadySettled = true
					break
				}
			}
			if alreadySettled {
				continue
			}
			grant, grantErr := RedemptionCLIPGrant(asset.CreditYield)
			if grantErr != nil {
				return grantErr
			}
			grantByAsset[assetID] = grant
			totalGrant += grant
		}
		if current.CLIPTreasury.TreasuryBalance < totalGrant {
			return apiError(http.StatusServiceUnavailable, "clip_treasury_insufficient", "The platform treasury cannot satisfy this CLIP distribution")
		}
		redemptionIDs := make([]string, 0, len(migration.RedemptionIDs)+len(grantByAsset))
		redemptionIDs = append(redemptionIDs, migration.RedemptionIDs...)
		creditsGranted, clipGranted := migration.CreditsGranted, migration.ClipGranted
		for index := range current.Assets {
			for _, assetID := range migration.ClipliAssetIDs {
				if current.Assets[index].ID == assetID {
					activateMigratedAsset(&current.Assets[index], now)
					if grant, shouldSettle := grantByAsset[assetID]; shouldSettle {
						redemption := domain.HAPWRedemption{ID: uniqueID("redeem", now), AssetID: assetID, Receipt: fmt.Sprintf("Clipli-LIC-%s-%04d", strings.TrimPrefix(current.Assets[index].TokenID, "#"), now.UnixNano()%10000), CreditsGranted: current.Assets[index].CreditYield, CreditsRemaining: current.Assets[index].CreditYield, ClipGranted: grant, RequestID: input.RequestID + "-redeem-" + strconv.Itoa(len(redemptionIDs)+1), Status: "有效", StatusEn: "Active", StatusKo: "유효", CreatedAt: formatMinute(now)}
						if err := distributeCLIP(current, redemption.RequestID, assetID, grant, now); err != nil {
							return err
						}
						current.Redemptions = prepend(redemption, current.Redemptions)
						if validWalletAddress(migration.WalletAddress) {
							queueRedemptionAirdrop(current, redemption, migration.WalletAddress, migration.ChainID, now)
						}
						current.CLIPTransactions = prepend(domain.CLIPTransaction{ID: uniqueID("clip-grant", now), TypeCode: "redemptionGrant", Type: "HAPW 核销领取", TypeEn: "HAPW redemption grant", TypeKo: "HAPW 상각 지급", Amount: grant, Counterparty: fmt.Sprintf("HAPW %s · %s", current.Assets[index].TokenID, current.Assets[index].Name), CounterpartyEn: fmt.Sprintf("HAPW %s · %s", current.Assets[index].TokenID, current.Assets[index].NameEn), CounterpartyKo: current.Assets[index].NameKo, StatusCode: "completed", Status: "已完成", StatusEn: "Completed", StatusKo: "완료", TxHash: "", CreatedAt: formatMinute(now)}, current.CLIPTransactions)
						redemptionIDs = append(redemptionIDs, redemption.ID)
						creditsGranted += redemption.CreditsGranted
						clipGranted += redemption.ClipGranted
						current.Assets[index].RedemptionStatus = "redeemed"
						current.Assets[index].Transferable = false
						current.Assets[index].Status, current.Assets[index].StatusEn, current.Assets[index].StatusKo = "已核销", "Redeemed", "상각 완료"
					}
				}
			}
		}
		externalRedemption := domain.ExternalAssetRedemption{ID: uniqueID("external-redemption", now), UserID: userID, ExternalUserID: binding.ExternalUserID, RequestNo: requestNo, TplID: input.TplID, Num: num, ExternalTxID: externalTxID, Quantity: num, Mode: "template_batch_write_off", Status: "completed", RequestID: input.RequestID, RedeemedAt: now.Format(time.RFC3339)}
		redemptionExists := false
		for _, existing := range current.ExternalRedemptions {
			if existing.UserID == userID && existing.RequestNo == requestNo {
				externalRedemption, redemptionExists = existing, true
				break
			}
		}
		if !redemptionExists {
			current.ExternalRedemptions = prepend(externalRedemption, current.ExternalRedemptions)
		}
		for index := range current.ExternalMigrations {
			if current.ExternalMigrations[index].ID == migration.ID {
				current.ExternalMigrations[index].ExternalRedemptionID = externalRedemption.ID
				current.ExternalMigrations[index].RedemptionIDs = redemptionIDs
				current.ExternalMigrations[index].CreditsGranted = creditsGranted
				current.ExternalMigrations[index].ClipGranted = clipGranted
				current.ExternalMigrations[index].Status = "completed_rewards_settled"
				current.ExternalMigrations[index].TreasuryReserved = 0
				current.ExternalMigrations[index].FailureReason = ""
				current.ExternalMigrations[index].UpdatedAt = now.Format(time.RFC3339)
				migration = current.ExternalMigrations[index]
				break
			}
		}
		result = migrationResult(*current, migration)
		return nil
	})
	if err != nil {
		// Haiwen has already accepted the irreversible write-off. Do not leave a
		// retryable pending record when local settlement cannot be committed.
		markFailed("reconciliation_required", "Haiwen write-off accepted but Clipli settlement failed: "+err.Error())
		return result, false, apiError(http.StatusBadGateway, "migration_settlement_failed", "Haiwen write-off accepted but Clipli settlement requires manual reconciliation")
	}
	return result, false, err
}

func (s *Service) ExternalMigrations(userID string) ([]domain.ExternalAssetMigration, error) {
	userID = normalizeUserID(userID)
	if !validUserID(userID) {
		return nil, apiError(http.StatusBadRequest, "invalid_user_id", "A valid userId is required")
	}
	items := make([]domain.ExternalAssetMigration, 0)
	for _, migration := range s.Snapshot().ExternalMigrations {
		if migration.UserID == userID {
			items = append(items, migration)
		}
	}
	return items, nil
}

func (s *Service) validateExternalMigrationInput(input MigrateExternalAssetInput, requireConfirmation bool) (string, domain.ExternalPlatformBinding, int, string, error) {
	userID := normalizeUserID(input.UserID)
	if userID == "" {
		userID = normalizeUserID(input.ClipliUserID)
	}
	if !validUserID(userID) {
		return "", domain.ExternalPlatformBinding{}, 0, "", apiError(http.StatusBadRequest, "invalid_user_id", "A valid userId is required")
	}
	if input.TplID <= 0 || input.TplID > 9007199254740991 {
		return "", domain.ExternalPlatformBinding{}, 0, "", apiError(http.StatusBadRequest, "invalid_tpl_id", "A positive tplId is required")
	}
	num := input.Num
	if num == 0 {
		num = 1
	}
	if num < 1 || num > 100 {
		return "", domain.ExternalPlatformBinding{}, 0, "", apiError(http.StatusBadRequest, "invalid_num", "num must be between 1 and 100")
	}
	requestNo := strings.TrimSpace(input.RequestNo)
	if !validRequestNo(requestNo) {
		return "", domain.ExternalPlatformBinding{}, 0, "", apiError(http.StatusBadRequest, "invalid_request_no", "A unique requestNo is required")
	}
	if !validRequestID(input.RequestID) {
		return "", domain.ExternalPlatformBinding{}, 0, "", apiError(http.StatusBadRequest, "invalid_request_id", "A valid operation id is required")
	}
	if requireConfirmation && !input.Accepted {
		return "", domain.ExternalPlatformBinding{}, 0, "", apiError(http.StatusBadRequest, "migration_not_confirmed", "Confirm the irreversible Haiwen write-off")
	}
	binding, err := s.verifiedBinding(userID)
	if err != nil {
		return "", domain.ExternalPlatformBinding{}, 0, "", err
	}
	if externalUserID := strings.TrimSpace(input.ExternalUserID); externalUserID != "" && externalUserID != binding.ExternalUserID {
		return "", domain.ExternalPlatformBinding{}, 0, "", apiError(http.StatusConflict, "binding_conflict", "The Clipli user is bound to a different external user")
	}
	return userID, binding, num, requestNo, nil
}

func (s *Service) externalTemplate(tplID int64) (domain.ExternalAssetTemplate, error) {
	for _, template := range s.Snapshot().ExternalTemplates {
		if template.TplID == tplID {
			return template, nil
		}
	}
	catalog, ok := s.platform.(ExternalTemplateCatalog)
	if !ok {
		return domain.ExternalAssetTemplate{}, apiError(http.StatusNotImplemented, "external_template_catalog_unsupported", "The configured external platform does not expose a template catalog")
	}
	for page := 1; page <= 100; page++ {
		result, err := catalog.ListTemplates(context.Background(), page, 100, 0)
		if err != nil {
			return domain.ExternalAssetTemplate{}, mapPlatformError(err)
		}
		s.cacheExternalTemplates(result.Items)
		for _, template := range result.Items {
			if template.TplID == tplID {
				return template, nil
			}
		}
		pageSize := result.PageSize
		if pageSize < 1 {
			pageSize = 100
		}
		if len(result.Items) == 0 || page*pageSize >= result.Total {
			break
		}
	}
	return domain.ExternalAssetTemplate{}, apiError(http.StatusNotFound, "external_template_not_found", "The Haiwen copyright template was not found")
}

func (s *Service) externalHoldingCount(externalUserID string, tplID int64) (int, error) {
	counter, ok := s.platform.(ExternalAssetCounter)
	if !ok {
		return 0, apiError(http.StatusNotImplemented, "external_asset_count_unsupported", "The configured external platform cannot verify template holdings")
	}
	items, err := counter.ListAssetCounts(context.Background(), externalUserID, []int64{tplID})
	if err != nil {
		return 0, mapPlatformError(err)
	}
	for _, item := range items {
		if item.TplID == tplID && item.Count >= 0 {
			return item.Count, nil
		}
	}
	return 0, apiError(http.StatusBadGateway, "invalid_external_counts", "Haiwen omitted the requested template quantity")
}

func (s *Service) cacheExternalTemplates(templates []domain.ExternalAssetTemplate) {
	_ = s.store.Update(func(state *store.State) error {
		cacheExternalTemplates(state, templates)
		return nil
	})
}

func cacheExternalTemplates(state *store.State, templates []domain.ExternalAssetTemplate) {
	for _, template := range templates {
		replaced := false
		for index := range state.ExternalTemplates {
			if state.ExternalTemplates[index].TplID == template.TplID {
				state.ExternalTemplates[index] = template
				replaced = true
				break
			}
		}
		if !replaced {
			state.ExternalTemplates = append(state.ExternalTemplates, template)
		}
	}
}

func (s *Service) cacheExternalWorks(works []domain.ExternalWork) {
	_ = s.store.Update(func(state *store.State) error {
		cacheExternalWorks(state, works)
		return nil
	})
}

func cacheExternalWorks(state *store.State, works []domain.ExternalWork) {
	for _, work := range works {
		replaced := false
		for index := range state.ExternalWorks {
			if state.ExternalWorks[index].WorkID == work.WorkID {
				state.ExternalWorks[index] = work
				replaced = true
				break
			}
		}
		if !replaced {
			state.ExternalWorks = append(state.ExternalWorks, work)
		}
	}
}

func findExternalAssetMapping(mappings []domain.ExternalAssetMappingRule, tplID int64) (domain.ExternalAssetMappingRule, bool) {
	for _, mapping := range mappings {
		if mapping.TplID == tplID && mapping.Active {
			return mapping, true
		}
	}
	return domain.ExternalAssetMappingRule{}, false
}

func buildMigratedAssets(template domain.ExternalAssetTemplate, mapping domain.ExternalAssetMappingRule, owner, requestNo string, quantity int, now time.Time, active bool) []domain.HAPWAsset {
	assets := make([]domain.HAPWAsset, 0, quantity)
	rightsHolder := externalPartyNames(template.Owners)
	if rightsHolder == "" {
		rightsHolder = "海文发权利人信息未提供"
	}
	for index := 0; index < quantity; index++ {
		digest := sha256.Sum256([]byte(fmt.Sprintf("haiwen|%d|%s|%d", template.TplID, requestNo, index+1)))
		suffix := fmt.Sprintf("%x", digest[:8])
		asset := domain.HAPWAsset{
			ID: "haiwen-" + strconv.FormatInt(template.TplID, 10) + "-" + suffix, Kind: "HAPW", TokenID: "#HW-" + strconv.FormatInt(template.TplID, 10) + "-" + suffix[:6],
			Name: template.Name, NameEn: template.Name, NameKo: template.Name, Currency: mapping.Currency,
			Status: "等待海文发核销", StatusEn: "Pending Haiwen write-off", StatusKo: "Haiwen 상각 대기", Transferable: false, Owner: owner,
			RightsHolder: rightsHolder, RightsHolderEn: rightsHolder, RightsHolderKo: rightsHolder,
			AuthorizationScope: "以海文发原始模板及权利文件为准", AuthorizationScopeEn: "Governed by the original Haiwen template and rights documents", AuthorizationScopeKo: "Haiwen 원본 템플릿 및 권리 문서 기준",
			CreditYield: mapping.CreditYield, ClipPrice: mapping.ClipPrice, ExchangeAvailable: false, RedemptionStatus: "pending_external_write_off",
			Authorization:  domain.Authorization{Holder: domain.LocalizedText{ZH: rightsHolder, EN: rightsHolder, KO: rightsHolder}, Scope: domain.LocalizedText{ZH: "以海文发原始模板及权利文件为准", EN: "Governed by the original Haiwen template and rights documents", KO: "Haiwen 원본 템플릿 및 권리 문서 기준"}, Territories: []string{}, UsageTypes: []string{}, Exclusivity: "source-defined", ValidFrom: now.Format("2006-01-02")},
			Provenance:     domain.Provenance{Issuer: "海文发", CertificateID: "HAIWEN-TPL-" + strconv.FormatInt(template.TplID, 10) + "-" + suffix, Network: "haiwen", TokenStandard: "HAIWEN-TEMPLATE-MIRROR", VerificationStatus: "pending-external-write-off", IssuedAt: now.Format("2006-01-02")},
			Media:          domain.AssetMedia{LinkedWorkIDs: []string{"haiwen-work-" + strconv.FormatInt(template.WorkID, 10)}, Format: "source-template", PreviewURL: template.Image},
			External:       domain.ExternalAssetRef{TransferID: requestNo, ProviderCode: "haiwen", ProviderAssetID: strconv.FormatInt(template.TplID, 10), SyncStatus: "pending-write-off", LastSyncedAt: now.Format(time.RFC3339), DataVersion: mapping.Version},
			SourceTemplate: &template,
		}
		if active {
			activateMigratedAsset(&asset, now)
		}
		assets = append(assets, asset)
	}
	return assets
}

func activateMigratedAsset(asset *domain.HAPWAsset, now time.Time) {
	asset.Status, asset.StatusEn, asset.StatusKo = "可核销", "Redeemable", "상각 가능"
	asset.Transferable, asset.RedemptionStatus = true, "available"
	asset.Provenance.VerificationStatus = "haiwen-write-off-confirmed"
	asset.External.SyncStatus, asset.External.LastSyncedAt = "migrated", now.Format(time.RFC3339)
}

func externalPartyNames(parties []domain.ExternalParty) string {
	names := make([]string, 0, len(parties))
	for _, party := range parties {
		if name := strings.TrimSpace(html.UnescapeString(party.Name)); name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(names, "、")
}

func migrationOwner(userID, wallet string) string {
	if validWalletAddress(wallet) {
		return normalizeWalletAddress(wallet)
	}
	return userID
}

func migrationResult(state store.State, migration domain.ExternalAssetMigration) ExternalMigrationResult {
	assets := make([]domain.HAPWAsset, 0, len(migration.ClipliAssetIDs))
	for _, assetID := range migration.ClipliAssetIDs {
		if asset, ok := findAsset(state.Assets, assetID); ok {
			assets = append(assets, asset)
		}
	}
	return ExternalMigrationResult{Migration: migration, Assets: assets}
}

type RedeemExternalAssetInput struct {
	UserID         string `json:"userId"`
	ClipliUserID   string `json:"clipliUserId,omitempty"`
	ExternalUserID string `json:"externalUserId,omitempty"`
	AssetID        string `json:"assetId,omitempty"`
	TplID          int64  `json:"tplId,omitempty"`
	Num            int    `json:"num,omitempty"`
	SerialNumber   string `json:"serialNumber,omitempty"`
	SerialNo       string `json:"serialNo,omitempty"`
	UniqueSerial   string `json:"uniqueSerialNumber,omitempty"`
	RequestNo      string `json:"requestNo,omitempty"`
	RequestID      string `json:"requestId,omitempty"`
}

type ExternalRedemptionResultEnvelope struct {
	Redemption domain.ExternalAssetRedemption `json:"redemption"`
}

// RedeemExternalAsset forwards a serial-numbered redemption to the partner.
// Existing records are returned before calling the partner, making retries
// safe even when the caller repeats a request after a timeout.
func (s *Service) RedeemExternalAsset(input RedeemExternalAssetInput) (ExternalRedemptionResultEnvelope, bool, error) {
	var result ExternalRedemptionResultEnvelope
	userID := normalizeUserID(input.UserID)
	if userID == "" {
		userID = normalizeUserID(input.ClipliUserID)
	}
	assetID := strings.TrimSpace(input.AssetID)
	serialNumber := strings.TrimSpace(input.SerialNumber)
	if serialNumber == "" {
		serialNumber = strings.TrimSpace(input.SerialNo)
	}
	if serialNumber == "" {
		serialNumber = strings.TrimSpace(input.UniqueSerial)
	}
	requestNo := strings.TrimSpace(input.RequestNo)
	if requestNo == "" {
		requestNo = serialNumber
	}
	if !validUserID(userID) {
		return result, false, apiError(http.StatusBadRequest, "invalid_user_id", "A valid userId is required")
	}
	if (assetID == "" && input.TplID <= 0) || len(assetID) > 200 {
		return result, false, apiError(http.StatusBadRequest, "invalid_asset_id", "A valid assetId or tplId is required")
	}
	if !validSerialNumber(serialNumber) {
		if !validSerialNumber(requestNo) {
			return result, false, apiError(http.StatusBadRequest, "invalid_request_no", "A unique requestNo is required")
		}
		serialNumber = requestNo
	}
	if !validRequestNo(requestNo) {
		return result, false, apiError(http.StatusBadRequest, "invalid_request_no", "A unique requestNo is required")
	}
	tplID := input.TplID
	if tplID <= 0 {
		tplID = parseTemplateID(assetID)
	}
	num := input.Num
	if num == 0 {
		num = 1
	}
	if num < 1 || num > 100 {
		return result, false, apiError(http.StatusBadRequest, "invalid_num", "num must be between 1 and 100")
	}
	if input.RequestID != "" && !validRequestID(input.RequestID) {
		return result, false, apiError(http.StatusBadRequest, "invalid_request_id", "A valid operation id is required")
	}
	binding, err := s.verifiedBinding(userID)
	if err != nil {
		return result, false, err
	}
	if capability, declared := s.platform.(ExternalItemizedAssetCapability); declared && !capability.SupportsItemizedAssets() {
		return result, false, apiError(http.StatusNotImplemented, "external_itemized_asset_unsupported", "Haiwen supports template batch write-off only; no individual asset ID or serial number exists")
	}
	if requestedExternal := strings.TrimSpace(input.ExternalUserID); requestedExternal != "" && requestedExternal != binding.ExternalUserID {
		return result, false, apiError(http.StatusConflict, "binding_conflict", "The Clipli user is bound to a different external user")
	}
	s.externalMu.Lock()
	defer s.externalMu.Unlock()
	state := s.Snapshot()
	for _, item := range state.ExternalRedemptions {
		if input.RequestID != "" && item.RequestID == input.RequestID && (item.UserID != userID || item.AssetID != assetID || item.SerialNumber != serialNumber) {
			return result, false, apiError(http.StatusConflict, "request_id_reused", "The requestId is already used for another redemption")
		}
		itemRequestNo := item.RequestNo
		if itemRequestNo == "" {
			itemRequestNo = item.SerialNumber
		}
		itemTplID := item.TplID
		if itemTplID == 0 {
			itemTplID = parseTemplateID(item.AssetID)
		}
		itemNum := item.Num
		if itemNum == 0 {
			itemNum = item.Quantity
		}
		if item.UserID == userID && itemRequestNo == requestNo {
			if (tplID > 0 && itemTplID > 0 && itemTplID != tplID) || (itemNum > 0 && itemNum != num) || (assetID != "" && item.AssetID != "" && item.AssetID != assetID) {
				return result, false, apiError(http.StatusConflict, "request_no_reused", "The requestNo is already used for another redemption")
			}
			return ExternalRedemptionResultEnvelope{Redemption: item}, true, nil
		}
		if item.SerialNumber == serialNumber && (item.UserID != userID || item.AssetID != assetID) {
			return result, false, apiError(http.StatusConflict, "serial_number_reused", "The serialNumber is already used for another redemption")
		}
	}
	external, err := s.platform.RedeemAsset(context.Background(), ExternalRedemptionRequest{UserID: userID, ExternalUserID: binding.ExternalUserID, AssetID: assetID, SerialNumber: serialNumber, RequestNo: requestNo, TplID: tplID, Num: num, RequestID: input.RequestID})
	if err != nil {
		return result, false, mapPlatformError(err)
	}
	// The adapter returns only after the partner has accepted the redemption;
	// expose one stable local status even if partners use different labels.
	status := "completed"
	quantity := external.Quantity
	if quantity <= 0 {
		quantity = num
	}
	if quantity <= 0 {
		quantity = 1
	}
	now := s.now().UTC()
	externalTxID := strings.TrimSpace(external.ExternalTxID)
	if externalTxID == "" {
		externalTxID = strings.TrimSpace(external.TransactionID)
	}
	if assetID == "" && tplID > 0 {
		assetID = strconv.FormatInt(tplID, 10)
	}
	if external.RequestNo != "" {
		requestNo = strings.TrimSpace(external.RequestNo)
	}
	record := domain.ExternalAssetRedemption{ID: uniqueID("external-redemption", now), UserID: userID, ExternalUserID: binding.ExternalUserID, AssetID: assetID, SerialNumber: serialNumber, RequestNo: requestNo, TplID: tplID, Num: num, ExternalTxID: externalTxID, Quantity: quantity, Mode: "itemized_asset", Status: status, RequestID: input.RequestID, RedeemedAt: now.Format(time.RFC3339)}
	err = s.store.Update(func(state *store.State) error {
		for _, item := range state.ExternalRedemptions {
			if item.UserID == userID && item.AssetID == assetID && item.SerialNumber == serialNumber {
				result = ExternalRedemptionResultEnvelope{Redemption: item}
				return nil
			}
		}
		state.ExternalRedemptions = prepend(record, state.ExternalRedemptions)
		result = ExternalRedemptionResultEnvelope{Redemption: record}
		return nil
	})
	return result, false, err
}

func normalizeUserID(value string) string { return strings.TrimSpace(value) }

func maxInt(value, floor int) int {
	if value < floor {
		return floor
	}
	return value
}

func validUserID(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, char := range value {
		if unicode.IsSpace(char) || unicode.IsControl(char) {
			return false
		}
	}
	return true
}

func normalizePhone(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	var b strings.Builder
	for index, char := range value {
		if char >= '0' && char <= '9' {
			b.WriteRune(char)
			continue
		}
		if char == '+' && index == 0 {
			b.WriteRune(char)
			continue
		}
		if char == ' ' || char == '-' || char == '(' || char == ')' {
			continue
		}
		return "", false
	}
	normalized := b.String()
	digits := strings.TrimPrefix(normalized, "+")
	return normalized, len(digits) >= 7 && len(digits) <= 15 && normalized != "+"
}

func maskPhone(phone string) string {
	digits := strings.TrimPrefix(phone, "+")
	if len(digits) <= 4 {
		return phone
	}
	return strings.Repeat("*", len(digits)-4) + digits[len(digits)-4:]
}

func validVerificationCode(value string) bool {
	if len(value) < 4 || len(value) > 8 {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func validSerialNumber(value string) bool {
	if len(value) < 1 || len(value) > 128 {
		return false
	}
	for _, char := range value {
		if unicode.IsSpace(char) || unicode.IsControl(char) {
			return false
		}
	}
	return true
}

func validRequestNo(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 {
		return false
	}
	for _, char := range value {
		if unicode.IsSpace(char) || unicode.IsControl(char) {
			return false
		}
	}
	return true
}

func parseTemplateID(assetID string) int64 {
	assetID = strings.TrimSpace(assetID)
	if assetID == "" {
		return 0
	}
	if parsed, err := strconv.ParseInt(assetID, 10, 64); err == nil && parsed > 0 {
		return parsed
	}
	if strings.HasPrefix(assetID, "asset-") {
		parsed, err := strconv.ParseInt(strings.TrimPrefix(assetID, "asset-"), 10, 64)
		if err == nil && parsed > 0 {
			return parsed
		}
	}
	return 0
}

// DemoExternalPlatform is deliberately small and deterministic. It exists so
// the Clipli project can exercise the complete flow before partner credentials
// and API details are available. The demo code is 123456 and must not be used
// as a production verifier.
type DemoExternalPlatform struct {
	mu       sync.Mutex
	bindings map[string]string
	redeemed map[string]bool
}

func NewDemoExternalPlatform() *DemoExternalPlatform {
	return &DemoExternalPlatform{bindings: map[string]string{}, redeemed: map[string]bool{}}
}

func (p *DemoExternalPlatform) SendVerificationCode(context.Context, VerificationCodeRequest) (VerificationCodeDispatch, error) {
	return VerificationCodeDispatch{DeliveryID: "demo-sms", ExpiresAt: time.Now().UTC().Add(5 * time.Minute).Format(time.RFC3339), Status: "sent"}, nil
}

func (p *DemoExternalPlatform) BindUser(_ context.Context, input BindUserRequest) (ExternalBindingResult, error) {
	if input.Code != "123456" {
		return ExternalBindingResult{}, &PlatformError{Status: http.StatusBadRequest, Code: "verification_code_invalid", Message: "The verification code is invalid"}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	externalID := strings.TrimSpace(input.ExternalUserID)
	if externalID == "" {
		externalID = "demo-user-" + input.UserID
	}
	p.bindings[input.UserID] = externalID
	return ExternalBindingResult{ExternalUserID: externalID, BindingID: "demo-binding-" + input.UserID, Status: "bound", Bound: true, BoundAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

func (p *DemoExternalPlatform) GetBindingStatus(_ context.Context, externalUserID string) (ExternalBindingStatus, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, boundExternalID := range p.bindings {
		if boundExternalID == strings.TrimSpace(externalUserID) {
			return ExternalBindingStatus{ExternalUserID: boundExternalID, Bound: true}, nil
		}
	}
	return ExternalBindingStatus{ExternalUserID: strings.TrimSpace(externalUserID), Bound: false}, nil
}

func (p *DemoExternalPlatform) ListAssets(_ context.Context, externalUserID string) ([]domain.ExternalAssetHolding, error) {
	if externalUserID == "" {
		return nil, &PlatformError{Status: http.StatusBadRequest, Code: "external_user_not_found", Message: "The external user was not found"}
	}
	ids := []string{"asset-2048", "asset-771", "asset-109", "asset-332", "asset-528", "asset-903"}
	items := make([]domain.ExternalAssetHolding, 0, len(ids))
	for _, id := range ids {
		items = append(items, domain.ExternalAssetHolding{AssetID: id, SerialNumber: "demo-" + id, Quantity: 1, Status: "held", SourceCode: "demo-platform", ExternalAssetID: "ext-" + id})
	}
	return items, nil
}

func (p *DemoExternalPlatform) SupportsItemizedAssets() bool { return true }

func (p *DemoExternalPlatform) ListAssetCounts(_ context.Context, externalUserID string, tplIDs []int64) ([]domain.ExternalTemplateCount, error) {
	if strings.TrimSpace(externalUserID) == "" {
		return nil, &PlatformError{Status: http.StatusNotFound, Code: "external_user_not_found", Message: "The external user was not found"}
	}
	items := make([]domain.ExternalTemplateCount, 0, len(tplIDs))
	for _, tplID := range tplIDs {
		items = append(items, domain.ExternalTemplateCount{TplID: tplID, Count: 1, SourceCode: "demo-platform"})
	}
	return items, nil
}

func (p *DemoExternalPlatform) ListTemplates(_ context.Context, page, pageSize int, workID int64) (ExternalTemplatePage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	workType := int64(100002)
	templates := []domain.ExternalAssetTemplate{{TplID: 100001, Name: "演示版权模板", Description: "演示模板", Image: "https://example.com/haiwen-template.png", WorkID: 100001, WorksName: "演示作品", WorksType: &workType, Authors: []domain.ExternalParty{{ID: 100001, Name: "演示作者"}}, Owners: []domain.ExternalParty{{ID: 100001, Name: "演示权利人"}}, PublishCount: 100}}
	if workID > 0 && workID != 100001 {
		templates = []domain.ExternalAssetTemplate{}
	}
	return ExternalTemplatePage{Items: templates, Total: len(templates), Page: page, PageSize: pageSize}, nil
}

func (p *DemoExternalPlatform) ListWorks(_ context.Context, page, pageSize int) (ExternalWorkPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	workType := int64(100002)
	works := []domain.ExternalWork{{WorkID: 100001, WorksName: "演示作品", Showcase: []string{"https://example.com/demo-work.png"}, Authors: []domain.ExternalParty{{ID: 100001, Name: "演示作者"}}, Owners: []domain.ExternalParty{{ID: 100001, Name: "演示权利人"}}, WorksType: &workType, WorksTypeName: "影视", WorksIntroduce: "演示作品", PublishNum: 100}}
	return ExternalWorkPage{Items: works, Total: len(works), Page: page, PageSize: pageSize}, nil
}

func (p *DemoExternalPlatform) RedeemAsset(_ context.Context, input ExternalRedemptionRequest) (ExternalRedemptionResult, error) {
	assetID := strings.TrimSpace(input.AssetID)
	if assetID == "" && input.TplID > 0 {
		assetID = strconv.FormatInt(input.TplID, 10)
	}
	requestNo := strings.TrimSpace(input.RequestNo)
	if requestNo == "" {
		requestNo = strings.TrimSpace(input.SerialNumber)
	}
	if assetID == "" || input.ExternalUserID == "" || requestNo == "" {
		return ExternalRedemptionResult{}, &PlatformError{Status: http.StatusBadRequest, Code: "asset_not_found", Message: "The external asset was not found"}
	}
	num := input.Num
	if num <= 0 {
		num = 1
	}
	key := input.ExternalUserID + ":" + assetID + ":" + requestNo
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.redeemed[key] {
		return ExternalRedemptionResult{ExternalTxID: "demo-redeem-" + requestNo, RequestNo: requestNo, TplID: input.TplID, Num: num, Status: "SUCCESS", Quantity: num}, nil
	}
	p.redeemed[key] = true
	return ExternalRedemptionResult{ExternalTxID: "demo-redeem-" + requestNo, RequestNo: requestNo, TplID: input.TplID, Num: num, Status: "SUCCESS", Quantity: num}, nil
}

func (p *DemoExternalPlatform) WriteOffTemplate(ctx context.Context, input ExternalTemplateWriteOffRequest) (ExternalRedemptionResult, error) {
	return p.RedeemAsset(ctx, ExternalRedemptionRequest{ExternalUserID: input.ExternalUserID, TplID: input.TplID, Num: input.Num, RequestNo: input.RequestNo})
}

// HTTPExternalPlatform is a deliberately conservative adapter for the next
// integration phase. Its endpoint paths and JSON are stable defaults that can
// be finalized with the partner without changing Clipli's service contract.
type HTTPExternalPlatform struct {
	BaseURL string
	Client  *http.Client
	AppID   string
	AppKey  string
	APIKey  string // Deprecated compatibility alias for AppKey.
}

func NewHTTPExternalPlatform(baseURL string, client *http.Client) *HTTPExternalPlatform {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPExternalPlatform{BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"), Client: client}
}

func (p *HTTPExternalPlatform) request(ctx context.Context, method, path string, input any, output any) error {
	if p.BaseURL == "" {
		return &PlatformError{Status: http.StatusServiceUnavailable, Code: "external_platform_not_configured", Message: "The external platform integration is not configured"}
	}
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, p.BaseURL+path, body)
	if err != nil {
		return err
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(p.AppID) != "" {
		req.Header.Set("x-app-id", strings.TrimSpace(p.AppID))
	}
	appKey := strings.TrimSpace(p.AppKey)
	if appKey == "" {
		appKey = strings.TrimSpace(p.APIKey)
	}
	if appKey != "" {
		req.Header.Set("x-app-key", appKey)
		// Keep the old header while existing non-Haiwen adapters are migrated.
		if strings.TrimSpace(p.AppKey) == "" && strings.TrimSpace(p.APIKey) != "" {
			req.Header.Set("Authorization", "Bearer "+appKey)
		}
	}
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return err
	}
	var envelope struct {
		Code    *int            `json:"code"`
		Message string          `json:"message"`
		Msg     string          `json:"msg"`
		Data    json.RawMessage `json:"data"`
	}
	decodedEnvelope := json.Unmarshal(data, &envelope) == nil && envelope.Code != nil
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || (decodedEnvelope && *envelope.Code != 0) {
		status := resp.StatusCode
		if status >= 200 && status < 300 && decodedEnvelope {
			status = externalBusinessStatus(*envelope.Code)
		}
		var payload struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(data, &payload)
		code, message := payload.Error.Code, payload.Error.Message
		if code == "" {
			code, message = payload.Code, payload.Message
		}
		if decodedEnvelope && *envelope.Code != 0 {
			code = externalBusinessCode(*envelope.Code)
			if envelope.Message != "" {
				message = envelope.Message
			} else if envelope.Msg != "" {
				message = envelope.Msg
			}
		}
		if code == "" {
			code = fmt.Sprintf("external_http_%d", status)
		}
		if message == "" {
			message = http.StatusText(status)
		}
		return &PlatformError{Status: status, Code: code, Message: message}
	}
	if !decodedEnvelope || len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return &PlatformError{Status: http.StatusBadGateway, Code: "invalid_external_response", Message: "Haiwen returned an incomplete response envelope"}
	}
	if output != nil && len(data) > 0 {
		if decodedEnvelope {
			if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
				return nil
			}
			if err := json.Unmarshal(envelope.Data, output); err != nil {
				return fmt.Errorf("decode external platform response: %w", err)
			}
			return nil
		}
		if err := json.Unmarshal(data, output); err != nil {
			return fmt.Errorf("decode external platform response: %w", err)
		}
	}
	return nil
}

func decodeExternalJSON(data []byte, output any) error {
	// Partner APIs commonly use either a bare object or Clipli's {data: ...}
	// envelope. Accept both during onboarding so the adapter remains stable.
	data = bytes.TrimSpace(data)
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if len(data) > 0 && data[0] == '{' && json.Unmarshal(data, &envelope) == nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		return json.Unmarshal(envelope.Data, output)
	}
	return json.Unmarshal(data, output)
}

func (p *HTTPExternalPlatform) SendVerificationCode(ctx context.Context, input VerificationCodeRequest) (VerificationCodeDispatch, error) {
	var result VerificationCodeDispatch
	externalUserID := strings.TrimSpace(input.ExternalUserID)
	if !validUserID(externalUserID) {
		return result, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_external_user_id", Message: "A valid externalUserId is required"}
	}
	payload := struct {
		ExternalUserID string `json:"externalUserId"`
		Phone          string `json:"phone"`
	}{ExternalUserID: externalUserID, Phone: input.Phone}
	err := p.request(ctx, http.MethodPost, "/openapi/user/bind/sms", payload, &struct{}{})
	return result, err
}

func (p *HTTPExternalPlatform) BindUser(ctx context.Context, input BindUserRequest) (ExternalBindingResult, error) {
	var result ExternalBindingResult
	externalUserID := strings.TrimSpace(input.ExternalUserID)
	if !validUserID(externalUserID) {
		return result, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_external_user_id", Message: "A valid externalUserId is required"}
	}
	code := input.SMSCode
	if code == "" {
		code = input.Code
	}
	if code == "" {
		code = input.VerificationCode
	}
	payload := struct {
		ExternalUserID string `json:"externalUserId"`
		Phone          string `json:"phone"`
		SMSCode        string `json:"smsCode"`
	}{ExternalUserID: externalUserID, Phone: input.Phone, SMSCode: code}
	err := p.request(ctx, http.MethodPost, "/openapi/user/bind", payload, &result)
	if err == nil && (result.ExternalUserID != externalUserID || !result.Bound || validExternalBoundAt(result.BoundAt) == "") {
		return result, &PlatformError{Status: http.StatusBadGateway, Code: "external_binding_mismatch", Message: "Haiwen returned an incomplete or mismatched binding"}
	}
	if err == nil && result.Bound {
		result.Status = "bound"
	}
	return result, err
}

type ExternalBindingStatus struct {
	ExternalUserID string `json:"externalUserId"`
	Bound          bool   `json:"bound"`
	BoundAt        string `json:"boundAt,omitempty"`
}

func (p *HTTPExternalPlatform) GetBindingStatus(ctx context.Context, externalUserID string) (ExternalBindingStatus, error) {
	externalUserID = strings.TrimSpace(externalUserID)
	if externalUserID == "" {
		return ExternalBindingStatus{}, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_external_user_id", Message: "A valid externalUserId is required"}
	}
	query := url.Values{"externalUserId": []string{externalUserID}}
	var result ExternalBindingStatus
	err := p.request(ctx, http.MethodGet, "/openapi/user/bind/status?"+query.Encode(), nil, &result)
	if err == nil && (result.ExternalUserID != externalUserID || (result.Bound && validExternalBoundAt(result.BoundAt) == "")) {
		return result, &PlatformError{Status: http.StatusBadGateway, Code: "external_binding_mismatch", Message: "Haiwen returned an incomplete or mismatched binding status"}
	}
	return result, err
}

func (p *HTTPExternalPlatform) ListWorks(ctx context.Context, page, pageSize int) (ExternalWorkPage, error) {
	if page < 1 || pageSize < 1 || pageSize > 100 {
		return ExternalWorkPage{}, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_work_query", Message: "Invalid work page or pageSize"}
	}
	query := url.Values{"pageNum": []string{strconv.Itoa(page)}, "pageSize": []string{strconv.Itoa(pageSize)}}
	var payload struct {
		List     []domain.ExternalWork `json:"list"`
		Total    int                   `json:"total"`
		PageNum  int                   `json:"pageNum"`
		PageSize int                   `json:"pageSize"`
	}
	if err := p.request(ctx, http.MethodGet, "/openapi/works?"+query.Encode(), nil, &payload); err != nil {
		return ExternalWorkPage{}, err
	}
	if payload.List == nil || payload.Total < len(payload.List) || payload.PageNum != page || payload.PageSize != pageSize || len(payload.List) > pageSize {
		return ExternalWorkPage{}, &PlatformError{Status: http.StatusBadGateway, Code: "invalid_external_catalog", Message: "Haiwen returned incomplete work pagination"}
	}
	for index := range payload.List {
		if err := normalizeExternalWork(&payload.List[index]); err != nil {
			return ExternalWorkPage{}, &PlatformError{Status: http.StatusBadGateway, Code: "invalid_external_work", Message: err.Error()}
		}
	}
	if payload.PageNum < 1 {
		payload.PageNum = page
	}
	if payload.PageSize < 1 {
		payload.PageSize = pageSize
	}
	return ExternalWorkPage{Items: payload.List, Total: maxInt(payload.Total, len(payload.List)), Page: payload.PageNum, PageSize: payload.PageSize}, nil
}

func (p *HTTPExternalPlatform) ListTemplates(ctx context.Context, page, pageSize int, workID int64) (ExternalTemplatePage, error) {
	if page < 1 || pageSize < 1 || pageSize > 100 || workID < 0 {
		return ExternalTemplatePage{}, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_template_query", Message: "Invalid template page, pageSize, or workId"}
	}
	query := url.Values{"pageNum": []string{strconv.Itoa(page)}, "pageSize": []string{strconv.Itoa(pageSize)}}
	if workID > 0 {
		query.Set("workId", strconv.FormatInt(workID, 10))
	}
	var payload struct {
		List     []domain.ExternalAssetTemplate `json:"list"`
		Total    int                            `json:"total"`
		PageNum  int                            `json:"pageNum"`
		PageSize int                            `json:"pageSize"`
	}
	if err := p.request(ctx, http.MethodGet, "/openapi/tpls?"+query.Encode(), nil, &payload); err != nil {
		return ExternalTemplatePage{}, err
	}
	if payload.List == nil || payload.Total < len(payload.List) || payload.PageNum != page || payload.PageSize != pageSize || len(payload.List) > pageSize {
		return ExternalTemplatePage{}, &PlatformError{Status: http.StatusBadGateway, Code: "invalid_external_catalog", Message: "Haiwen returned incomplete template pagination"}
	}
	for index := range payload.List {
		if err := normalizeExternalTemplate(&payload.List[index]); err != nil {
			return ExternalTemplatePage{}, &PlatformError{Status: http.StatusBadGateway, Code: "invalid_external_template", Message: err.Error()}
		}
	}
	if payload.PageNum < 1 {
		payload.PageNum = page
	}
	if payload.PageSize < 1 {
		payload.PageSize = pageSize
	}
	return ExternalTemplatePage{Items: payload.List, Total: maxInt(payload.Total, len(payload.List)), Page: payload.PageNum, PageSize: payload.PageSize}, nil
}

func normalizeExternalWork(work *domain.ExternalWork) error {
	work.WorksName = strings.TrimSpace(work.WorksName)
	work.WorksTypeName = strings.TrimSpace(work.WorksTypeName)
	if work.WorkID <= 0 || work.WorkID > 9007199254740991 || work.WorksName == "" || work.PublishNum < 0 {
		return errors.New("Haiwen returned an incomplete work")
	}
	for index := range work.Authors {
		work.Authors[index].Name = strings.TrimSpace(work.Authors[index].Name)
	}
	for index := range work.Owners {
		work.Owners[index].Name = strings.TrimSpace(work.Owners[index].Name)
	}
	for index, showcase := range work.Showcase {
		showcase = strings.TrimSpace(showcase)
		if showcase == "" {
			work.Showcase[index] = ""
			continue
		}
		parsed, err := url.Parse(showcase)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
			work.Showcase[index] = ""
			continue
		}
		work.Showcase[index] = showcase
	}
	return nil
}

func (p *HTTPExternalPlatform) ListAssetCounts(ctx context.Context, externalUserID string, tplIDs []int64) ([]domain.ExternalTemplateCount, error) {
	if len(tplIDs) == 0 || len(tplIDs) > 100 {
		return nil, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_tpl_ids", Message: "Provide between 1 and 100 template IDs"}
	}
	externalUserID = strings.TrimSpace(externalUserID)
	if !validUserID(externalUserID) {
		return nil, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_external_user_id", Message: "A valid externalUserId is required"}
	}
	values := make([]string, 0, len(tplIDs))
	for _, tplID := range tplIDs {
		if tplID <= 0 {
			return nil, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_tpl_ids", Message: "Template IDs must be positive integers"}
		}
		values = append(values, strconv.FormatInt(tplID, 10))
	}
	query := url.Values{"externalUserId": []string{externalUserID}, "tplIds": []string{strings.Join(values, ",")}}
	var payload struct {
		ExternalUserID string `json:"externalUserId"`
		List           []struct {
			TplID int64 `json:"tplId"`
			Count *int  `json:"count"`
		} `json:"list"`
	}
	if err := p.request(ctx, http.MethodGet, "/openapi/user/assets/count?"+query.Encode(), nil, &payload); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.ExternalUserID) != externalUserID {
		return nil, &PlatformError{Status: http.StatusConflict, Code: "external_binding_mismatch", Message: "The external platform returned a different external user ID"}
	}
	items := make([]domain.ExternalTemplateCount, 0, len(payload.List))
	requested := make(map[int64]bool, len(tplIDs))
	for _, id := range tplIDs {
		requested[id] = false
	}
	for _, item := range payload.List {
		seen, expected := requested[item.TplID]
		if !expected || seen || item.Count == nil || *item.Count < 0 {
			return nil, &PlatformError{Status: http.StatusBadGateway, Code: "invalid_external_counts", Message: "Haiwen returned invalid or duplicate template quantities"}
		}
		requested[item.TplID] = true
		items = append(items, domain.ExternalTemplateCount{TplID: item.TplID, Count: *item.Count, SourceCode: "haiwen"})
	}
	for _, present := range requested {
		if !present {
			return nil, &PlatformError{Status: http.StatusBadGateway, Code: "invalid_external_counts", Message: "Haiwen omitted a requested template quantity"}
		}
	}
	return items, nil
}

func normalizeExternalTemplate(template *domain.ExternalAssetTemplate) error {
	template.Name = strings.TrimSpace(template.Name)
	template.WorksName = strings.TrimSpace(template.WorksName)
	template.WorksTypeName = strings.TrimSpace(template.WorksTypeName)
	if template.TplID <= 0 || template.TplID > 9007199254740991 || template.Name == "" || template.WorkID <= 0 || template.WorkID > 9007199254740991 || template.WorksName == "" || template.PublishCount < 0 {
		return errors.New("Haiwen returned an incomplete copyright template")
	}
	for index := range template.Authors {
		template.Authors[index].Name = strings.TrimSpace(template.Authors[index].Name)
	}
	for index := range template.Owners {
		template.Owners[index].Name = strings.TrimSpace(template.Owners[index].Name)
	}
	template.Image = strings.TrimSpace(template.Image)
	if template.Image != "" {
		parsed, err := url.Parse(template.Image)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
			template.Image = ""
		}
	}
	return nil
}

func (p *HTTPExternalPlatform) SupportsItemizedAssets() bool { return false }

func (p *HTTPExternalPlatform) RedeemAsset(_ context.Context, _ ExternalRedemptionRequest) (ExternalRedemptionResult, error) {
	return ExternalRedemptionResult{}, &PlatformError{Status: http.StatusNotImplemented, Code: "external_itemized_asset_unsupported", Message: "Haiwen supports template batch write-off only"}
}

func (p *HTTPExternalPlatform) WriteOffTemplate(ctx context.Context, input ExternalTemplateWriteOffRequest) (ExternalRedemptionResult, error) {
	var result ExternalRedemptionResult
	externalUserID := strings.TrimSpace(input.ExternalUserID)
	tplID := input.TplID
	requestNo := strings.TrimSpace(input.RequestNo)
	num := input.Num
	if !validUserID(externalUserID) {
		return result, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_external_user_id", Message: "A valid externalUserId is required"}
	}
	if tplID <= 0 {
		return result, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_tpl_id", Message: "A positive tplId is required"}
	}
	if num < 1 || num > 100 {
		return result, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_num", Message: "num must be between 1 and 100"}
	}
	if !validRequestNo(requestNo) {
		return result, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_request_no", Message: "A valid requestNo is required"}
	}
	payload := struct {
		RequestNo      string `json:"requestNo"`
		ExternalUserID string `json:"externalUserId"`
		TplID          int64  `json:"tplId"`
		Num            int    `json:"num"`
	}{RequestNo: requestNo, ExternalUserID: externalUserID, TplID: tplID, Num: num}
	if err := p.request(ctx, http.MethodPost, "/openapi/asset/write-off", payload, &result); err != nil {
		return result, err
	}
	if result.RequestNo != requestNo || result.ExternalUserID != externalUserID || result.TplID != tplID || result.Num != num || result.Status != "SUCCESS" || validExternalBoundAt(result.WriteOffAt) == "" {
		return result, &PlatformError{Status: http.StatusBadGateway, Code: "external_migration_response_mismatch", Message: "Haiwen returned an incomplete or mismatched write-off receipt"}
	}
	result.Quantity = result.Num
	return result, nil
}

func externalBusinessStatus(code int) int {
	switch code {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusConflict, http.StatusTooManyRequests, http.StatusUnprocessableEntity:
		return code
	default:
		return http.StatusBadGateway
	}
}

func externalBusinessCode(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "external_invalid_request"
	case http.StatusUnauthorized:
		return "external_credentials_invalid"
	case http.StatusNotFound:
		return "external_not_found"
	case http.StatusConflict:
		return "external_conflict"
	case http.StatusUnprocessableEntity:
		return "external_assets_insufficient"
	case http.StatusTooManyRequests:
		return "external_rate_limited"
	default:
		return fmt.Sprintf("external_code_%d", code)
	}
}
