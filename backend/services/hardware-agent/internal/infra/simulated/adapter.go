package simulated

import (
	"context"
	"errors"
	"fmt"

	"mercadia.dev/pos/services/hardware-agent/internal/domain"
)

var ErrUnsupportedCommand = errors.New("unsupported device command")

type Adapter interface {
	Kind() domain.DeviceKind
	Execute(ctx context.Context, device domain.Device, commandType string, payload map[string]any) (map[string]any, error)
}

type Registry struct {
	byKind map[domain.DeviceKind]Adapter
}

func NewRegistry(adapters ...Adapter) *Registry {
	registry := &Registry{byKind: map[domain.DeviceKind]Adapter{}}
	for _, adapter := range adapters {
		registry.byKind[adapter.Kind()] = adapter
	}
	return registry
}

func (r *Registry) Execute(ctx context.Context, device domain.Device, commandType string, payload map[string]any) (map[string]any, error) {
	adapter, ok := r.byKind[device.Kind]
	if !ok {
		return nil, fmt.Errorf("no adapter for device kind %q", device.Kind)
	}
	return adapter.Execute(ctx, device, commandType, payload)
}

func DefaultRegistry() *Registry {
	return NewRegistry(
		NewFiscalAdapter(),
		NewPaymentTerminalAdapter(),
		NewMSRAdapter(),
		NewIButtonAdapter(),
		NewScannerAdapter(),
		NewPrinterAdapter(),
	)
}
