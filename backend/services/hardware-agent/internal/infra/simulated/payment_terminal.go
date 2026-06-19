package simulated

import (
	"context"
	"fmt"

	"mercadia.dev/pos/services/hardware-agent/internal/domain"
)

type PaymentTerminalAdapter struct {
	lastAuthCode string
	lastRRN      string
}

func NewPaymentTerminalAdapter() *PaymentTerminalAdapter {
	return &PaymentTerminalAdapter{}
}

func (a *PaymentTerminalAdapter) Kind() domain.DeviceKind {
	return domain.DeviceKindPaymentTerminal
}

func (a *PaymentTerminalAdapter) Execute(_ context.Context, device domain.Device, commandType string, payload map[string]any) (map[string]any, error) {
	switch commandType {
	case "get_status":
		return map[string]any{
			"terminalState": "idle",
			"connection":    "online",
			"model":         device.Model,
			"serialNumber":  device.ID,
			"lastAuthCode":  a.lastAuthCode,
		}, nil
	case "authorize":
		amountMinor := int64(0)
		if value, ok := payload["amountMinor"].(float64); ok {
			amountMinor = int64(value)
		}
		currency := "RUB"
		if value, ok := payload["currency"].(string); ok && value != "" {
			currency = value
		}
		a.lastAuthCode = "A1B2C3"
		a.lastRRN = "SIMRRN000042"
		return map[string]any{
			"status":        "approved",
			"authCode":      a.lastAuthCode,
			"rrn":           a.lastRRN,
			"cardMask":      "****1234",
			"amountMinor":   amountMinor,
			"currency":      currency,
			"terminalState": "idle",
		}, nil
	case "capture":
		return map[string]any{
			"status":        "captured",
			"authCode":      a.lastAuthCode,
			"rrn":           a.lastRRN,
			"terminalState": "idle",
		}, nil
	case "cancel":
		return map[string]any{
			"status":        "cancelled",
			"terminalState": "idle",
		}, nil
	case "refund":
		amountMinor := int64(0)
		if value, ok := payload["amountMinor"].(float64); ok {
			amountMinor = int64(value)
		}
		return map[string]any{
			"status":        "refunded",
			"rrn":           a.lastRRN,
			"authCode":      a.lastAuthCode,
			"amountMinor":   amountMinor,
			"terminalState": "idle",
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCommand, commandType)
	}
}
