package simulated

import (
	"context"
	"fmt"

	"mercadia.dev/pos/services/hardware-agent/internal/domain"
)

type IButtonAdapter struct{}

func NewIButtonAdapter() *IButtonAdapter {
	return &IButtonAdapter{}
}

func (a *IButtonAdapter) Kind() domain.DeviceKind {
	return domain.DeviceKindIButton
}

func (a *IButtonAdapter) Execute(_ context.Context, device domain.Device, commandType string, _ map[string]any) (map[string]any, error) {
	switch commandType {
	case "get_status":
		return map[string]any{
			"readerState": "ready",
			"model":       device.Model,
		}, nil
	case "read_key":
		return map[string]any{
			"romId": "demo-ibutton-senior-1",
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCommand, commandType)
	}
}
