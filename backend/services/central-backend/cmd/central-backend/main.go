package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mercadia.dev/pos/platform/observability"
	"mercadia.dev/pos/services/central-backend/internal/api"
	centralnats "mercadia.dev/pos/services/central-backend/internal/infra/nats"
)

func main() {
	observability.SetupLogging("central-backend")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := observability.SetupTracing(ctx, "central-backend")
	if err != nil {
		slog.Error("failed to initialize tracing", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			slog.Error("failed to shutdown tracing", "error", err)
		}
	}()

	addr := os.Getenv("MERCADIA_CENTRAL_BACKEND_ADDR")
	if addr == "" {
		addr = ":8082"
	}

	bundle, err := api.NewServerBundle(api.ServerOptions{})
	if err != nil {
		slog.Error("failed to initialize central backend", "error", err)
		os.Exit(1)
	}
	defer bundle.Handle.Close()

	serverOpts := api.ServerOptions{}
	if natsURL, enabled := natsConsumerURL(); enabled {
		consumer, err := centralnats.NewConsumer(natsURL, bundle.Services.Sync)
		if err != nil {
			slog.Error("failed to initialize nats consumer", "error", err)
			os.Exit(1)
		}
		defer consumer.Close()

		serverOpts.ReadinessChecks = append(serverOpts.ReadinessChecks, consumer.HealthCheck)
		go func() {
			if err := consumer.Run(ctx); err != nil && ctx.Err() == nil {
				slog.Error("nats sync consumer stopped", "error", err)
			}
		}()
		slog.Info("nats sync consumer enabled", "url", natsURL)
	}

	handler := api.NewHandler(bundle.Services, serverOpts)
	server := &http.Server{
		Addr:    addr,
		Handler: observability.InstrumentHTTP("central-backend", handler),
	}

	go func() {
		<-ctx.Done()
		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error("central backend shutdown failed", "error", err)
		}
	}()

	slog.Info("starting central backend", "addr", addr, "nats_consumer", len(serverOpts.ReadinessChecks) > 0, "otel", observability.OTELEnabled())
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("central backend stopped", "error", err)
		os.Exit(1)
	}
}

func natsConsumerURL() (string, bool) {
	if value, ok := os.LookupEnv("MERCADIA_CENTRAL_BACKEND_NATS_URL"); ok {
		if value == "" {
			value = centralnats.DefaultURL()
		}
		return value, true
	}
	return "", false
}
