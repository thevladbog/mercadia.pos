package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mercadia.dev/pos/platform/observability"
	"mercadia.dev/pos/services/store-edge/internal/api"
	"mercadia.dev/pos/services/store-edge/internal/app"
	haclient "mercadia.dev/pos/services/store-edge/internal/infra/hardwareagent"
	storenats "mercadia.dev/pos/services/store-edge/internal/infra/nats"
)

func main() {
	observability.SetupLogging("store-edge")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := observability.SetupTracing(ctx, "store-edge")
	if err != nil {
		slog.Error("failed to initialize tracing", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			slog.Error("failed to shutdown tracing", "error", err)
		}
	}()

	addr := os.Getenv("MERCADIA_STORE_EDGE_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	serverOpts := api.ServerOptions{
		DatabaseURL:           os.Getenv("MERCADIA_STORE_EDGE_DATABASE_URL"),
		MigrationsDir:         os.Getenv("MERCADIA_STORE_EDGE_MIGRATIONS_DIR"),
		CentralBackendURL:     os.Getenv("MERCADIA_CENTRAL_BACKEND_URL"),
		HardwareAgentURL:      os.Getenv("MERCADIA_HARDWARE_AGENT_URL"),
		UseHardwareAgent:      os.Getenv("MERCADIA_STORE_EDGE_USE_HARDWARE_AGENT") == "true",
		HardwareAgentFallback: os.Getenv("MERCADIA_STORE_EDGE_HARDWARE_AGENT_FALLBACK") != "false",
		DefaultStoreID:        os.Getenv("MERCADIA_STORE_EDGE_DEFAULT_STORE_ID"),
	}

	if probeEnabled(serverOpts.UseHardwareAgent) {
		haClient := haclient.NewClient(serverOpts.HardwareAgentURL, nil)
		serverOpts.ReadinessChecks = append(serverOpts.ReadinessChecks, haClient.HealthCheck)
		slog.Info("hardware agent readiness probe enabled")
	}

	var publisher *storenats.Publisher
	if natsURL, configured := os.LookupEnv("MERCADIA_STORE_EDGE_NATS_URL"); configured {
		if natsURL == "" {
			natsURL = "nats://127.0.0.1:4222"
		}

		initialBundle, err := api.NewServerBundle(serverOpts)
		if err != nil {
			slog.Error("failed to initialize store edge", "error", err)
			os.Exit(1)
		}

		publisher, err = storenats.NewPublisher(natsURL, initialBundle.Outbox)
		if err != nil {
			slog.Error("failed to initialize nats publisher", "error", err)
			os.Exit(1)
		}
		defer publisher.Close()

		serverOpts.ReadinessChecks = append(serverOpts.ReadinessChecks, publisher.HealthCheck)
		serverOpts.BrokerConnected = publisher.Connected
		go publisher.Run(ctx)
	}

	bundle, err := api.NewServerBundle(serverOpts)
	if err != nil {
		slog.Error("failed to initialize store edge", "error", err)
		os.Exit(1)
	}

	if intervalRaw := os.Getenv("MERCADIA_STORE_EDGE_CATALOG_SYNC_INTERVAL"); intervalRaw != "" && bundle.CatalogSync != nil {
		interval, err := time.ParseDuration(intervalRaw)
		if err != nil {
			slog.Error("invalid catalog sync interval", "value", intervalRaw, "error", err)
			os.Exit(1)
		}
		storeID := bundle.DefaultStoreID
		go runCatalogSyncLoop(ctx, bundle.CatalogSync, storeID, interval)
		slog.Info("catalog sync background worker enabled", "store_id", storeID, "interval", interval.String())
	}

	server := &http.Server{
		Addr:    addr,
		Handler: observability.InstrumentHTTP("store-edge", bundle.Handler),
	}

	slog.Info("starting store edge",
		"addr", addr,
		"postgres", serverOpts.DatabaseURL != "",
		"nats", publisher != nil,
		"hardware_agent", serverOpts.UseHardwareAgent,
		"otel", observability.OTELEnabled(),
	)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("store edge stopped", "error", err)
		os.Exit(1)
	}
}

func probeEnabled(useHardwareAgent bool) bool {
	if value, ok := os.LookupEnv("MERCADIA_STORE_EDGE_HARDWARE_AGENT_READINESS_PROBE"); ok {
		return value == "true"
	}
	return useHardwareAgent
}

func runCatalogSyncLoop(ctx context.Context, catalogSync *app.CatalogSyncService, storeID string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		result, err := catalogSync.Sync(ctx, app.SyncCatalogCommand{StoreID: storeID})
		if err != nil {
			slog.Warn("background catalog sync failed", "store_id", storeID, "error", err)
		} else {
			slog.Info("background catalog sync completed",
				"store_id", result.StoreID,
				"products_count", result.ProductsCount,
				"synced_at", result.SyncedAt,
			)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
