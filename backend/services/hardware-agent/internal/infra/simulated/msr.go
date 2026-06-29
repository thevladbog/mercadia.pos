package simulated

import (
	"context"
	"fmt"

	"mercadia.dev/pos/services/hardware-agent/internal/domain"
)

type MSRAdapter struct{}

func NewMSRAdapter() *MSRAdapter {
	return &MSRAdapter{}
}

func (a *MSRAdapter) Kind() domain.DeviceKind {
	return domain.DeviceKindMSR
}

func (a *MSRAdapter) Execute(_ context.Context, device domain.Device, commandType string, _ map[string]any) (map[string]any, error) {
	switch commandType {
	case "get_status":
		return map[string]any{
			"readerState": "ready",
			"model":       device.Model,
		}, nil
	case "read_card":
		return map[string]any{
			"track1": "%B4111111111111111^SIM/TEST^3012?",
			"track2": ";4111111111111111=3012?",
			"masked": "****1111",
		}, nil
	case "read_staff_card":
		return map[string]any{
			"staffToken": "demo-msr-senior-1", // #nosec G101 -- deterministic simulator fixture, not a secret.
			"masked":     "MSR staff demo ****0001",
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedCommand, commandType)
	}
}
