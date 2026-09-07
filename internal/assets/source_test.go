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
		_ = json.NewEncoder(response).Encode(map[string]any{"data": []domain.HAPWAsset{{ID: "remote-1", TokenID: "#1", Name: "Remote", RightsHolder: "Holder", Owner: "Clipli", External: domain.ExternalAssetRef{ProviderCode: "haiwen", ProviderAssetID: "remote-1", TransferID: "receipt-1", SyncStatus: "migrated"}}}})
	}))
	defer server.Close()
	loaded, err := LoadRemote(context.Background(), server.URL, []domain.HAPWAsset{{ID: "fallback"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0].External.SyncStatus != "migrated" || loaded[0].External.ProviderAssetID != "remote-1" {
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

func TestRemoteFeedRequiresConfirmedTransferAndAcceptsEmptyHoldings(t *testing.T) {
	for _, tc := range []struct {
		name, payload string
		wantError     bool
		wantCount     int
	}{
		{"empty", `{"data":[]}`, false, 0},
		{"missing envelope", `{}`, true, 1},
		{"null", `{"data":null}`, true, 1},
		{"unconfirmed", `{"data":[{"id":"a","tokenId":"#1","name":"Asset","rightsHolder":"Holder","owner":"Clipli","external":{"providerCode":"haiwen","providerAssetId":"1","syncStatus":"synced"}}]}`, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(tc.payload)) }))
			defer server.Close()
			result, err := LoadRemote(context.Background(), server.URL, []domain.HAPWAsset{{ID: "previous-confirmed"}})
			if (err != nil) != tc.wantError || len(result) != tc.wantCount {
				t.Fatalf("assets=%#v error=%v", result, err)
			}
		})
	}
}
