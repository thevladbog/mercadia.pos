package observability

import (
	"log/slog"
	"os"
)

func SetupLogging(service string) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler).With("service", service))
}
