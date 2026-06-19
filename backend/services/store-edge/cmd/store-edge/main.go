package main

import (
	"log/slog"
	"net/http"
	"os"

	"mercadia.dev/pos/services/store-edge/internal/api"
)

func main() {
	addr := os.Getenv("MERCADIA_STORE_EDGE_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	server := &http.Server{
		Addr:    addr,
		Handler: api.NewServer(),
	}

	slog.Info("starting store edge", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("store edge stopped", "error", err)
		os.Exit(1)
	}
}
