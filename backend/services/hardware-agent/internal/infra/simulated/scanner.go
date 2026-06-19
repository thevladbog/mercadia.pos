package simulated

import (
	"context"
	"fmt"

	"mercadia.dev/pos/services/hardware-agent/internal/domain"
)

type ScannerAdapter struct{}

func NewScannerAdapter() *ScannerAdapter {
	return &ScannerAdapter{}
}

func (a *ScannerAdapter) Kind() domain.DeviceKind {
	return domain.DeviceKindScanner
}

func (a *ScannerAdapter) Execute(_ context.Context, device domain.Device, commandType string, payload map[string]any) (map[string]any, error) {
	switch commandType {
	case "get_status":
		return map[string]any{
			"scannerState": "ready",
			"model":        device.Model,
		}, nil
	case "scan":
		barcode := "4600000000001"
		if value, ok := payload["barcode"].(string); ok && value != "" {
			barcode = value
		}
		return map[string]any{
			"barcode": barcode,
			"symbology": "EAN13",
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCommand, commandType)
	}
}
