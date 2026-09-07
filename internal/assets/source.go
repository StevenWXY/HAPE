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
		return append([]domain.HAPWAsset{}, fallback...), ErrInvalidSourceURL
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return append([]domain.HAPWAsset{}, fallback...), err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-Clipli-Consumer", "hapw-read-model-v1")
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		return append([]domain.HAPWAsset{}, fallback...), err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return append([]domain.HAPWAsset{}, fallback...), fmt.Errorf("hapw_asset_source_status_%d", response.StatusCode)
	}
	var payload json.RawMessage
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&payload); err != nil {
		return append([]domain.HAPWAsset{}, fallback...), err
	}
	payload = bytes.TrimSpace(payload)
	var items []domain.HAPWAsset
	if len(payload) > 0 && payload[0] == '[' {
		if err := json.Unmarshal(payload, &items); err != nil {
			return append([]domain.HAPWAsset{}, fallback...), err
		}
	} else {
		var envelope struct {
			Data []domain.HAPWAsset `json:"data"`
		}
		if err := json.Unmarshal(payload, &envelope); err != nil {
			return append([]domain.HAPWAsset{}, fallback...), err
		}
		items = envelope.Data
	}
	if items == nil {
		return append([]domain.HAPWAsset{}, fallback...), errors.New("invalid_hapw_asset_source_envelope")
	}
	if len(items) == 0 {
		return []domain.HAPWAsset{}, nil
	}
	seen := map[string]bool{}
	syncedAt := time.Now().UTC().Format(time.RFC3339)
	for index := range items {
		item := &items[index]
		if seen[item.ID] {
			return append([]domain.HAPWAsset{}, fallback...), errors.New("duplicate_hapw_asset")
		}
		seen[item.ID] = true
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.TokenID) == "" || strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.RightsHolder) == "" {
			return append([]domain.HAPWAsset{}, fallback...), errors.New("invalid_hapw_asset_source_record")
		}
		if item.Kind == "" {
			item.Kind = "HAPW"
		}
		if strings.TrimSpace(item.External.ProviderCode) == "" || strings.TrimSpace(item.External.ProviderAssetID) == "" || strings.TrimSpace(item.External.TransferID) == "" || item.External.SyncStatus != "migrated" || strings.TrimSpace(item.Owner) == "" {
			return append([]domain.HAPWAsset{}, fallback...), errors.New("unconfirmed_hapw_transfer")
		}
		item.External.LastSyncedAt = syncedAt

		if item.External.DataVersion == "" {
			item.External.DataVersion = "remote-v1"
		}
	}
	return items, nil
}
