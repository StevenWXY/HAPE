package works

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StevenWXY/HAPE/internal/domain"
)

func TestLoadRemoteEnvelopeAndNormalizeSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(response).Encode(map[string]any{"data": []domain.Work{{ID: "remote-1", Title: "Remote work", LinkedAssetID: "asset-2048"}}})
	}))
	defer server.Close()
	loaded, err := LoadRemote(context.Background(), server.URL, []domain.Work{{ID: "fallback"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0].Source != "remote" || loaded[0].ID != "remote-1" {
		t.Fatalf("loaded = %#v", loaded)
	}
}

func TestLoadRemoteFallsBackOnInvalidURL(t *testing.T) {
	fallback := []domain.Work{{ID: "fallback", Title: "Seed", LinkedAssetID: "asset-2048"}}
	loaded, err := LoadRemote(context.Background(), "file:///tmp/works.json", fallback)
	if err == nil || len(loaded) != 1 || loaded[0].ID != "fallback" {
		t.Fatalf("loaded=%#v err=%v", loaded, err)
	}
}
