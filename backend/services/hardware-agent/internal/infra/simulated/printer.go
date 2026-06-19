package simulated

import (
	"context"
	"fmt"

	"mercadia.dev/pos/services/hardware-agent/internal/domain"
)

type PrinterAdapter struct{}

func NewPrinterAdapter() *PrinterAdapter {
	return &PrinterAdapter{}
}

func (a *PrinterAdapter) Kind() domain.DeviceKind {
	return domain.DeviceKindPrinter
}

func (a *PrinterAdapter) Execute(_ context.Context, device domain.Device, commandType string, payload map[string]any) (map[string]any, error) {
	switch commandType {
	case "get_status":
		return map[string]any{
			"printerState": "ready",
			"paperPresent": true,
			"coverClosed":  true,
			"model":        device.Model,
		}, nil
	case "print":
		lines := 1
		if value, ok := payload["lines"].(float64); ok {
			lines = int(value)
		}
		return map[string]any{
			"printedLines": lines,
			"printerState": "ready",
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCommand, commandType)
	}
}
