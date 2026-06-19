package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"mercadia.dev/pos/platform/observability"
	"mercadia.dev/pos/services/hardware-agent/internal/api"
)

func main() {
	observability.SetupLogging("hardware-agent")

	ctx := context.Background()
	shutdownTracing, err := observability.SetupTracing(ctx, "hardware-agent")
	if err != nil {
		slog.Error("failed to initialize tracing", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			slog.Error("failed to shutdown tracing", "error", err)
		}
	}()

	addr := os.Getenv("MERCADIA_HARDWARE_AGENT_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8083"
	}

	server := &http.Server{
		Addr:    addr,
		Handler: observability.InstrumentHTTP("hardware-agent", api.NewServer()),
	}

	slog.Info("starting hardware agent", "addr", addr, "otel", observability.OTELEnabled())
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("hardware agent stopped", "error", err)
		os.Exit(1)
	}
}
