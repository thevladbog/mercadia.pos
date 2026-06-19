package main

import (
	"log/slog"
	"net/http"
	"os"

	"mercadia.dev/pos/services/hardware-agent/internal/api"
)

func main() {
	addr := os.Getenv("MERCADIA_HARDWARE_AGENT_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8083"
	}

	server := &http.Server{
		Addr:    addr,
		Handler: api.NewServer(),
	}

	slog.Info("starting hardware agent", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("hardware agent stopped", "error", err)
		os.Exit(1)
	}
}
