package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"mercadia.dev/pos/platform/observability"
	"mercadia.dev/pos/services/central-backend/internal/api"
)

func main() {
	observability.SetupLogging("central-backend")

	ctx := context.Background()
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

	server := &http.Server{
		Addr:    addr,
		Handler: observability.InstrumentHTTP("central-backend", api.NewServer()),
	}

	slog.Info("starting central backend", "addr", addr, "otel", observability.OTELEnabled())
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("central backend stopped", "error", err)
		os.Exit(1)
	}
}
