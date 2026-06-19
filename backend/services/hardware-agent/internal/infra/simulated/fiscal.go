package simulated

import (
	"context"
	"fmt"

	"mercadia.dev/pos/services/hardware-agent/internal/domain"
)

type FiscalAdapter struct {
	nextDocumentNumber int
	shiftOpen          bool
}

func NewFiscalAdapter() *FiscalAdapter {
	return &FiscalAdapter{
		nextDocumentNumber: 1042,
		shiftOpen:          true,
	}
}

func (a *FiscalAdapter) Kind() domain.DeviceKind {
	return domain.DeviceKindFiscal
}

func (a *FiscalAdapter) Execute(_ context.Context, device domain.Device, commandType string, payload map[string]any) (map[string]any, error) {
	switch commandType {
	case "get_status":
		return a.status(device), nil
	case "open_shift":
		a.shiftOpen = true
		return map[string]any{
			"driverState":   "ready",
			"shiftState":    "opened",
			"sessionNumber": 7,
		}, nil
	case "close_shift":
		a.shiftOpen = false
		return map[string]any{
			"driverState":   "ready",
			"shiftState":    "closed",
			"zReportNumber": a.nextDocumentNumber,
		}, nil
	case "print_receipt":
		if !a.shiftOpen {
			return nil, fmt.Errorf("fiscal shift is closed")
		}
		documentNumber := a.nextDocumentNumber
		a.nextDocumentNumber++
		totalMinor := int64(0)
		if value, ok := payload["totalMinor"].(float64); ok {
			totalMinor = int64(value)
		}
		return map[string]any{
			"driverState":          "ready",
			"fiscalDocumentNumber": documentNumber,
			"fiscalSign":           fmt.Sprintf("SIM-FS-%06d", documentNumber),
			"qrCode":               fmt.Sprintf("t=%d&fn=9999078900000001&i=%d&s=%.2f", documentNumber, documentNumber, float64(totalMinor)/100),
			"shiftState":           "opened",
			"printedAt":            "simulated",
		}, nil
	case "cancel_receipt":
		return map[string]any{
			"driverState": "ready",
			"cancelled":   true,
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCommand, commandType)
	}
}

func (a *FiscalAdapter) status(device domain.Device) map[string]any {
	shiftState := "closed"
	if a.shiftOpen {
		shiftState = "opened"
	}
	return map[string]any{
		"driverState":        string(device.Status),
		"fiscalMode":         true,
		"shiftState":         shiftState,
		"paperPresent":       true,
		"coverClosed":        true,
		"serialNumber":       device.ID,
		"firmwareVersion":    "10.10.0.0",
		"lastDocumentNumber": a.nextDocumentNumber - 1,
		"sessionNumber":      7,
		"model":              device.Model,
	}
}
