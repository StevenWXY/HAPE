package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	assetsource "github.com/StevenWXY/HAPE/internal/assets"
	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/httpapi"
	"github.com/StevenWXY/HAPE/internal/service"
	"github.com/StevenWXY/HAPE/internal/store"
	"github.com/StevenWXY/HAPE/internal/works"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	root, err := os.Getwd()
	if err != nil {
		logger.Error("resolve working directory", "error", err)
		os.Exit(1)
	}
	publicDir := envOr("PUBLIC_DIR", filepath.Join(root, "public"))
	host, port := envOr("HOST", "127.0.0.1"), envOr("PORT", "4173")

	memory := store.NewMemory(store.SeedState())
	if sourceURL := os.Getenv("HAPW_ASSET_SOURCE_URL"); sourceURL != "" {
		state := memory.Snapshot()
		startedAt := time.Now().UTC()
		loadedAssets, sourceErr := assetsource.LoadRemote(context.Background(), sourceURL, state.Assets)
		if sourceErr != nil {
			logger.Warn("HAPW asset source unavailable; retaining last known snapshot", "error", sourceErr)
			_ = memory.Update(func(current *store.State) error {
				current.AssetSyncRuns = append([]domain.AssetSyncRun{{ID: fmt.Sprintf("sync-%d", startedAt.UnixNano()), SourceCode: "configured-source", Status: "failed", StartedAt: startedAt.Format(time.RFC3339), CompletedAt: time.Now().UTC().Format(time.RFC3339), Error: sourceErr.Error()}}, current.AssetSyncRuns...)
				return nil
			})
		} else {
			completedAt := time.Now().UTC()
			_ = memory.Update(func(current *store.State) error {
				current.Assets = loadedAssets
				current.AssetSources = []domain.AssetSource{{Code: "configured-source", Name: "Configured external HAPW provider", BaseURL: publicSourceURL(sourceURL), Mode: "remote-read-only", Status: "synced", SyncIntervalSeconds: 900, LastSyncedAt: completedAt.Format(time.RFC3339), AssetCount: len(loadedAssets)}}
				current.AssetSyncRuns = append([]domain.AssetSyncRun{{ID: fmt.Sprintf("sync-%d", startedAt.UnixNano()), SourceCode: "configured-source", Status: "completed", StartedAt: startedAt.Format(time.RFC3339), CompletedAt: completedAt.Format(time.RFC3339), RecordsRead: len(loadedAssets), RecordsValid: len(loadedAssets), RecordsSaved: len(loadedAssets)}}, current.AssetSyncRuns...)
				return nil
			})
			logger.Info("remote HAPW asset source loaded", "count", len(loadedAssets))
		}
	}
	if sourceURL := os.Getenv("WORK_SOURCE_URL"); sourceURL != "" {
		state := memory.Snapshot()
		loadedWorks, sourceErr := works.LoadRemote(context.Background(), sourceURL, state.Works)
		if sourceErr != nil {
			logger.Warn("work source unavailable; using seed works", "error", sourceErr)
		} else {
			_ = memory.Update(func(current *store.State) error { current.Works = loadedWorks; return nil })
			logger.Info("remote work source loaded", "count", len(loadedWorks))
		}
	}
	serviceLayer := service.New(memory)
	if rawMappings := strings.TrimSpace(os.Getenv("CLIPLI_EXTERNAL_ASSET_MAPPINGS")); rawMappings != "" {
		var mappings []domain.ExternalAssetMappingRule
		if err := json.Unmarshal([]byte(rawMappings), &mappings); err != nil {
			logger.Warn("external asset mappings are invalid; migrations will remain disabled", "error", err)
		} else if err := serviceLayer.SetExternalAssetMappings(mappings); err != nil {
			logger.Warn("external asset mappings failed validation; migrations will remain disabled", "error", err)
		} else {
			logger.Info("external asset mappings configured", "count", len(mappings))
		}
	}
	if platformURL := os.Getenv("CLIPLI_EXTERNAL_PLATFORM_URL"); platformURL != "" {
		platformClient := service.NewHTTPExternalPlatform(platformURL, nil)
		platformClient.AppID = os.Getenv("CLIPLI_EXTERNAL_PLATFORM_APP_ID")
		platformClient.AppKey = os.Getenv("CLIPLI_EXTERNAL_PLATFORM_APP_KEY")
		if platformClient.AppKey == "" {
			platformClient.APIKey = os.Getenv("CLIPLI_EXTERNAL_PLATFORM_API_KEY")
		}
		platformClient.DefaultTplIDs = parseIntList(os.Getenv("CLIPLI_EXTERNAL_PLATFORM_TPL_IDS"))
		serviceLayer.SetExternalPlatform(platformClient)
		logger.Info("external platform integration configured", "baseURL", publicSourceURL(platformURL))
	}
	server := &http.Server{
		Addr:              host + ":" + port,
		Handler:           httpapi.New(serviceLayer, publicDir, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("Clipli running", "url", fmt.Sprintf("http://%s:%s", host, port), "runtime", "go")
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Error("server stopped", "error", serveErr)
			os.Exit(1)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}

func publicSourceURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "configured"
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseIntList(raw string) []int64 {
	values := make([]int64, 0)
	for _, part := range strings.Split(raw, ",") {
		value, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err == nil && value > 0 {
			values = append(values, value)
		}
	}
	return values
}
