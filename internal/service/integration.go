package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	ListAssets(ctx context.Context, userID string) ([]domain.ExternalAssetHolding, error)
	RedeemAsset(ctx context.Context, input ExternalRedemptionRequest) (ExternalRedemptionResult, error)
}

// ExternalAssetCounter is implemented by adapters that support the partner's
// template-count endpoint. Older adapters can keep implementing ListAssets;
// the service falls back to that method when this optional capability is absent.
type ExternalAssetCounter interface {
	ListAssetCounts(ctx context.Context, externalUserID string, tplIDs []int64) ([]domain.ExternalAssetHolding, error)
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
	externalUserID := normalizeExternalUserID(input.ExternalUserID, userID)
	if !validUserID(externalUserID) {
		return result, apiError(http.StatusBadRequest, "invalid_external_user_id", "A valid externalUserId is required")
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
				if item.UserID != userID || item.Phone != phone {
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
	if expiresAt == "" {
		expiresAt = now.Add(5 * time.Minute).Format(time.RFC3339)
	} else if _, parseErr := time.Parse(time.RFC3339, expiresAt); parseErr != nil {
		expiresAt = now.Add(5 * time.Minute).Format(time.RFC3339)
	}
	// Any successful dispatch is a pending challenge from Clipli's point of
	// view; partner-specific delivery states are not used for local matching.
	status := "sent"
	challenge := domain.VerificationChallenge{
		ID: uniqueID("verification", now), UserID: userID, ExternalUserID: externalUserID, Phone: phone, PhoneMasked: maskPhone(phone),
		DeliveryID: dispatch.DeliveryID, Status: status, ExpiresAt: expiresAt, CreatedAt: now.Format(time.RFC3339), RequestID: input.RequestID,
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
	for _, item := range state.VerificationChallenges {
		if item.UserID == userID && (phone == "" || item.Phone == phone) && item.Status == "sent" && (input.VerificationID == "" || item.ID == input.VerificationID) {
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
	externalUserID := normalizeExternalUserID(input.ExternalUserID, challenge.ExternalUserID)
	if externalUserID == "" {
		externalUserID = userID
	}
	if !validUserID(externalUserID) {
		return result, false, apiError(http.StatusBadRequest, "invalid_external_user_id", "A valid externalUserId is required")
	}
	external, err := s.platform.BindUser(context.Background(), BindUserRequest{UserID: userID, ExternalUserID: externalUserID, Phone: phone, Code: code, SMSCode: code, VerificationCode: code, VerificationID: challenge.ID, RequestID: input.RequestID})
	if err != nil {
		return result, false, mapPlatformError(err)
	}
	externalUserID = strings.TrimSpace(external.ExternalUserID)
	if externalUserID == "" {
		externalUserID = strings.TrimSpace(external.UserID)
	}
	if externalUserID == "" {
		return result, false, apiError(http.StatusBadGateway, "external_binding_invalid", "The external platform did not return an external user ID")
	}
	now := s.now().UTC()
	// A successful adapter response establishes the Clipli-side bound state;
	// partner-specific status values are intentionally not used as local state.
	status := "bound"
	binding := domain.ExternalPlatformBinding{ID: external.BindingID, UserID: userID, ExternalUserID: externalUserID, Phone: phone, PhoneMasked: maskPhone(phone), PlatformCode: "external-platform", Status: status, BoundAt: now.Format(time.RFC3339), VerificationID: challenge.ID, RequestID: input.RequestID}
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
	binding, ok := s.findBinding(userID)
	if !ok {
		return result, apiError(http.StatusConflict, "user_not_bound", "Bind the Clipli user to the external platform first")
	}
	items, err := s.platform.ListAssets(context.Background(), binding.ExternalUserID)
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
	binding, ok := s.findBinding(userID)
	if !ok {
		return result, apiError(http.StatusConflict, "user_not_bound", "Bind the Clipli user to the external platform first")
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
		items = make([]domain.ExternalAssetHolding, 0)
	}
	counts := make([]ExternalAssetCount, 0, len(items))
	seen := make(map[int64]bool, len(items))
	for _, item := range items {
		tplID, parseErr := strconv.ParseInt(strings.TrimSpace(item.AssetID), 10, 64)
		if parseErr != nil || tplID <= 0 {
			continue
		}
		counts = append(counts, ExternalAssetCount{TplID: tplID, Count: maxInt(item.Quantity, 0)})
		seen[tplID] = true
	}
	for _, tplID := range tplIDs {
		if !seen[tplID] {
			counts = append(counts, ExternalAssetCount{TplID: tplID, Count: 0})
		}
	}
	result = UserAssetCountsResult{UserID: userID, ExternalUserID: binding.ExternalUserID, Items: counts, RetrievedAt: s.now().UTC().Format(time.RFC3339), SourceOfTruth: "external-platform"}
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

func (s *Service) GetBinding(userID string) (domain.ExternalPlatformBinding, bool, error) {
	userID = normalizeUserID(userID)
	if !validUserID(userID) {
		return domain.ExternalPlatformBinding{}, false, apiError(http.StatusBadRequest, "invalid_user_id", "A valid userId is required")
	}
	binding, ok := s.findBinding(userID)
	if !ok {
		return domain.ExternalPlatformBinding{}, false, apiError(http.StatusNotFound, "binding_not_found", "The Clipli user is not bound to the external platform")
	}
	return binding, true, nil
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
	binding, ok := s.findBinding(userID)
	if !ok {
		return result, false, apiError(http.StatusConflict, "user_not_bound", "Bind the Clipli user to the external platform first")
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
	record := domain.ExternalAssetRedemption{ID: uniqueID("external-redemption", now), UserID: userID, ExternalUserID: binding.ExternalUserID, AssetID: assetID, SerialNumber: serialNumber, RequestNo: requestNo, TplID: tplID, Num: num, ExternalTxID: externalTxID, Quantity: quantity, Status: status, RequestID: input.RequestID, RedeemedAt: now.Format(time.RFC3339)}
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

func normalizeExternalUserID(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

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
	externalID := "demo-user-" + input.UserID
	p.bindings[input.UserID] = externalID
	return ExternalBindingResult{ExternalUserID: externalID, BindingID: "demo-binding-" + input.UserID, Status: "bound"}, nil
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

func (p *DemoExternalPlatform) ListAssetCounts(_ context.Context, externalUserID string, tplIDs []int64) ([]domain.ExternalAssetHolding, error) {
	if strings.TrimSpace(externalUserID) == "" {
		return nil, &PlatformError{Status: http.StatusNotFound, Code: "external_user_not_found", Message: "The external user was not found"}
	}
	items := make([]domain.ExternalAssetHolding, 0, len(tplIDs))
	for _, tplID := range tplIDs {
		items = append(items, domain.ExternalAssetHolding{AssetID: strconv.FormatInt(tplID, 10), Quantity: 1, Status: "available", SourceCode: "demo-platform"})
	}
	return items, nil
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

// HTTPExternalPlatform is a deliberately conservative adapter for the next
// integration phase. Its endpoint paths and JSON are stable defaults that can
// be finalized with the partner without changing Clipli's service contract.
type HTTPExternalPlatform struct {
	BaseURL       string
	Client        *http.Client
	AppID         string
	AppKey        string
	APIKey        string // Deprecated compatibility alias for AppKey.
	DefaultTplIDs []int64
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
	externalUserID := normalizeExternalUserID(input.ExternalUserID, input.UserID)
	payload := struct {
		ExternalUserID string `json:"externalUserId"`
		Phone          string `json:"phone"`
	}{ExternalUserID: externalUserID, Phone: input.Phone}
	err := p.request(ctx, http.MethodPost, "/openapi/user/bind/sms", payload, &struct{}{})
	return result, err
}

func (p *HTTPExternalPlatform) BindUser(ctx context.Context, input BindUserRequest) (ExternalBindingResult, error) {
	var result ExternalBindingResult
	externalUserID := normalizeExternalUserID(input.ExternalUserID, input.UserID)
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
	if err == nil {
		if result.ExternalUserID == "" {
			result.ExternalUserID = externalUserID
		}
		if result.Bound {
			result.Status = "bound"
		}
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
	query := url.Values{"externalUserId": []string{externalUserID}}
	var result ExternalBindingStatus
	err := p.request(ctx, http.MethodGet, "/openapi/user/bind/status?"+query.Encode(), nil, &result)
	return result, err
}

func (p *HTTPExternalPlatform) ListAssetCounts(ctx context.Context, externalUserID string, tplIDs []int64) ([]domain.ExternalAssetHolding, error) {
	if len(tplIDs) == 0 || len(tplIDs) > 100 {
		return nil, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_tpl_ids", Message: "Provide between 1 and 100 template IDs"}
	}
	values := make([]string, 0, len(tplIDs))
	for _, tplID := range tplIDs {
		if tplID <= 0 {
			return nil, &PlatformError{Status: http.StatusBadRequest, Code: "invalid_tpl_ids", Message: "Template IDs must be positive integers"}
		}
		values = append(values, strconv.FormatInt(tplID, 10))
	}
	query := url.Values{"externalUserId": []string{strings.TrimSpace(externalUserID)}, "tplIds": []string{strings.Join(values, ",")}}
	var payload struct {
		ExternalUserID string `json:"externalUserId"`
		List           []struct {
			TplID int64 `json:"tplId"`
			Count int   `json:"count"`
		} `json:"list"`
	}
	if err := p.request(ctx, http.MethodGet, "/openapi/user/assets/count?"+query.Encode(), nil, &payload); err != nil {
		return nil, err
	}
	items := make([]domain.ExternalAssetHolding, 0, len(payload.List))
	for _, item := range payload.List {
		if item.TplID <= 0 {
			continue
		}
		items = append(items, domain.ExternalAssetHolding{AssetID: strconv.FormatInt(item.TplID, 10), Quantity: maxInt(item.Count, 0), Status: "available", SourceCode: "haiwen"})
	}
	return items, nil
}

func (p *HTTPExternalPlatform) ListAssets(ctx context.Context, userID string) ([]domain.ExternalAssetHolding, error) {
	if len(p.DefaultTplIDs) == 0 {
		return nil, &PlatformError{Status: http.StatusBadRequest, Code: "tpl_ids_required", Message: "Configure tplIds or call the template-count endpoint with tplIds"}
	}
	return p.ListAssetCounts(ctx, userID, p.DefaultTplIDs)
}

func (p *HTTPExternalPlatform) RedeemAsset(ctx context.Context, input ExternalRedemptionRequest) (ExternalRedemptionResult, error) {
	var result ExternalRedemptionResult
	tplID := input.TplID
	if tplID <= 0 {
		tplID = parseTemplateID(input.AssetID)
	}
	requestNo := strings.TrimSpace(input.RequestNo)
	if requestNo == "" {
		requestNo = strings.TrimSpace(input.SerialNumber)
	}
	num := input.Num
	if num == 0 {
		num = 1
	}
	payload := struct {
		RequestNo      string `json:"requestNo"`
		ExternalUserID string `json:"externalUserId"`
		TplID          int64  `json:"tplId"`
		Num            int    `json:"num"`
	}{RequestNo: requestNo, ExternalUserID: strings.TrimSpace(input.ExternalUserID), TplID: tplID, Num: num}
	if err := p.request(ctx, http.MethodPost, "/openapi/asset/write-off", payload, &result); err != nil {
		return result, err
	}
	if result.RequestNo == "" {
		result.RequestNo = requestNo
	}
	if result.ExternalUserID == "" {
		result.ExternalUserID = strings.TrimSpace(input.ExternalUserID)
	}
	if result.TplID == 0 {
		result.TplID = tplID
	}
	if result.Num == 0 {
		result.Num = num
	}
	if result.Quantity == 0 {
		result.Quantity = result.Num
	}
	if result.Status == "" {
		result.Status = "SUCCESS"
	}
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
