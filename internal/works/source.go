package works

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

var ErrInvalidSourceURL = errors.New("invalid_work_source_url")

// LoadRemote loads a normalized work list from a configured third-party API.
// The API may return either an array or {"data": [...]} and must expose only
// http(s) URLs. Callers decide whether to keep the fallback on error.
func LoadRemote(ctx context.Context, sourceURL string, fallback []domain.Work) ([]domain.Work, error) {
	parsed, err := url.Parse(strings.TrimSpace(sourceURL))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return append([]domain.Work(nil), fallback...), ErrInvalidSourceURL
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return append([]domain.Work(nil), fallback...), err
	}
	request.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return append([]domain.Work(nil), fallback...), err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return append([]domain.Work(nil), fallback...), fmt.Errorf("work_source_status_%d", response.StatusCode)
	}
	var envelope struct {
		Data []domain.Work `json:"data"`
	}
	var payload json.RawMessage
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&payload); err != nil {
		return append([]domain.Work(nil), fallback...), err
	}
	payload = bytes.TrimSpace(payload)
	if len(payload) > 0 && payload[0] == '[' {
		if err := json.Unmarshal(payload, &envelope.Data); err != nil {
			return append([]domain.Work(nil), fallback...), err
		}
	} else if err := json.Unmarshal(payload, &envelope); err != nil {
		return append([]domain.Work(nil), fallback...), err
	}
	if len(envelope.Data) == 0 {
		return append([]domain.Work(nil), fallback...), errors.New("empty_work_source")
	}
	for index := range envelope.Data {
		if strings.TrimSpace(envelope.Data[index].ID) == "" || strings.TrimSpace(envelope.Data[index].Title) == "" || strings.TrimSpace(envelope.Data[index].LinkedAssetID) == "" {
			return append([]domain.Work(nil), fallback...), errors.New("invalid_work_source_record")
		}
		envelope.Data[index].Source = "remote"
	}
	return envelope.Data, nil
}
