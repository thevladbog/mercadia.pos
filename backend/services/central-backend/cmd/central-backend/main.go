package main

import (
	"log/slog"
	"net/http"
	"os"

	"mercadia.dev/pos/services/central-backend/internal/api"
)

func main() {
	addr := os.Getenv("MERCADIA_CENTRAL_BACKEND_ADDR")
	if addr == "" {
		addr = ":8082"
	}

	server := &http.Server{
		Addr:    addr,
		Handler: api.NewServer(),
	}

	slog.Info("starting central backend", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("central backend stopped", "error", err)
		os.Exit(1)
	}
}
