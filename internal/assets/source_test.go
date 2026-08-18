package assets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StevenWXY/HAPE/internal/domain"
)

func TestLoadRemoteEnvelopeAndStampExternalRef(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Clipli-Consumer") == "" {
			t.Fatal("missing consumer header")
		}
		_ = json.NewEncoder(response).Encode(map[string]any{"data": []domain.HAPWAsset{{ID: "remote-1", TokenID: "#1", Name: "Remote", RightsHolder: "Holder"}}})
	}))
	defer server.Close()
	loaded, err := LoadRemote(context.Background(), server.URL, []domain.HAPWAsset{{ID: "fallback"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0].External.SyncStatus != "synced" || loaded[0].External.ProviderAssetID != "remote-1" {
		t.Fatalf("loaded = %#v", loaded)
	}
}

func TestLoadRemoteFallsBackOnInvalidURL(t *testing.T) {
	fallback := []domain.HAPWAsset{{ID: "fallback"}}
	loaded, err := LoadRemote(context.Background(), "file:///tmp/assets.json", fallback)
	if err == nil || len(loaded) != 1 || loaded[0].ID != "fallback" {
		t.Fatalf("loaded=%#v err=%v", loaded, err)
	}
}
