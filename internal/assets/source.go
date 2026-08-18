package assets

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/StevenWXY/HAPE/internal/domain"
)

var ErrInvalidSourceURL = errors.New("invalid_hapw_asset_source_url")

// LoadRemote reads a normalized HAPW list from an upstream read-only API. The
// adapter deliberately accepts only a small, stable contract so a provider
// integration cannot leak provider-specific fields into the UI.
func LoadRemote(ctx context.Context, sourceURL string, fallback []domain.HAPWAsset) ([]domain.HAPWAsset, error) {
	parsed, err := url.Parse(strings.TrimSpace(sourceURL))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return append([]domain.HAPWAsset(nil), fallback...), ErrInvalidSourceURL
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return append([]domain.HAPWAsset(nil), fallback...), err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-Clipli-Consumer", "hapw-read-model-v1")
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		return append([]domain.HAPWAsset(nil), fallback...), err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return append([]domain.HAPWAsset(nil), fallback...), fmt.Errorf("hapw_asset_source_status_%d", response.StatusCode)
	}
	var payload json.RawMessage
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&payload); err != nil {
		return append([]domain.HAPWAsset(nil), fallback...), err
	}
	payload = bytes.TrimSpace(payload)
	var items []domain.HAPWAsset
	if len(payload) > 0 && payload[0] == '[' {
		if err := json.Unmarshal(payload, &items); err != nil {
			return append([]domain.HAPWAsset(nil), fallback...), err
		}
	} else {
		var envelope struct {
			Data []domain.HAPWAsset `json:"data"`
		}
		if err := json.Unmarshal(payload, &envelope); err != nil {
			return append([]domain.HAPWAsset(nil), fallback...), err
		}
		items = envelope.Data
	}
	if len(items) == 0 {
		return append([]domain.HAPWAsset(nil), fallback...), errors.New("empty_hapw_asset_source")
	}
	syncedAt := time.Now().UTC().Format(time.RFC3339)
	for index := range items {
		item := &items[index]
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.TokenID) == "" || strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.RightsHolder) == "" {
			return append([]domain.HAPWAsset(nil), fallback...), errors.New("invalid_hapw_asset_source_record")
		}
		if item.Kind == "" {
			item.Kind = "HAPW"
		}
		item.External.SyncStatus = "synced"
		item.External.LastSyncedAt = syncedAt
		if item.External.ProviderAssetID == "" {
			item.External.ProviderAssetID = item.ID
		}
		if item.External.ProviderCode == "" {
			item.External.ProviderCode = "configured-source"
		}
		if item.External.DataVersion == "" {
			item.External.DataVersion = "remote-v1"
		}
	}
	return items, nil
}
