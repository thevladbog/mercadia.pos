package api

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"mercadia.dev/pos/services/hardware-agent/internal/domain"
)

// devicesEnvVar names the environment variable that, when set, provides the
// device inventory as a JSON array instead of the built-in simulated seed
// (memory.NewStore's seedDevices). See docs/hardware-support-matrix.md §3
// ("Config-driven device inventory").
const devicesEnvVar = "MERCADIA_HARDWARE_AGENT_DEVICES"

// deviceConfig is the JSON shape of one entry in MERCADIA_HARDWARE_AGENT_DEVICES.
//
// Unknown JSON fields are ignored (json.Unmarshal's default behavior; this
// struct does not use a DisallowUnknownFields decoder). This is a deliberate
// choice to keep the config schema open for extension: a future real driver
// (e.g. the ATOL fiscal adapter) is expected to add fields such as
// "connection" (COM port/IP/driver selection) without breaking existing
// config files that don't set them.
type deviceConfig struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Model  string `json:"model"`
	Status string `json:"status"`
}

// loadConfiguredDevices reads devicesEnvVar and parses it into domain
// devices. It returns (nil, nil) when the env var is unset or empty, which
// signals the caller to fall back to the default built-in simulated seed.
func loadConfiguredDevices() ([]domain.Device, error) {
	raw := os.Getenv(devicesEnvVar)
	if raw == "" {
		return nil, nil
	}
	return parseDeviceConfigJSON(raw)
}

// parseDeviceConfigJSON parses and validates the JSON array format of
// devicesEnvVar. Validation fails fast (at server construction, not at
// first command) on: invalid JSON, an empty array, an empty or duplicate
// device id, or an unknown device kind/status.
func parseDeviceConfigJSON(raw string) ([]domain.Device, error) {
	var configs []deviceConfig
	if err := json.Unmarshal([]byte(raw), &configs); err != nil {
		return nil, fmt.Errorf("%s: invalid JSON: %w", devicesEnvVar, err)
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("%s: must contain at least one device", devicesEnvVar)
	}

	seenIDs := make(map[string]struct{}, len(configs))
	now := time.Now().UTC()
	devices := make([]domain.Device, 0, len(configs))
	for i, cfg := range configs {
		if cfg.ID == "" {
			return nil, fmt.Errorf("%s: device at index %d has an empty id", devicesEnvVar, i)
		}
		if _, exists := seenIDs[cfg.ID]; exists {
			return nil, fmt.Errorf("%s: duplicate device id %q", devicesEnvVar, cfg.ID)
		}
		seenIDs[cfg.ID] = struct{}{}

		kind := domain.DeviceKind(cfg.Kind)
		if !isKnownDeviceKind(kind) {
			return nil, fmt.Errorf("%s: device %q has unknown kind %q", devicesEnvVar, cfg.ID, cfg.Kind)
		}

		status := domain.DeviceStatusSimulated
		if cfg.Status != "" {
			status = domain.DeviceStatus(cfg.Status)
			if !isKnownDeviceStatus(status) {
				return nil, fmt.Errorf("%s: device %q has unknown status %q", devicesEnvVar, cfg.ID, cfg.Status)
			}
		}

		devices = append(devices, domain.Device{
			ID:        cfg.ID,
			Kind:      kind,
			Status:    status,
			Model:     cfg.Model,
			UpdatedAt: now,
		})
	}
	return devices, nil
}

func isKnownDeviceKind(kind domain.DeviceKind) bool {
	switch kind {
	case domain.DeviceKindFiscal,
		domain.DeviceKindPaymentTerminal,
		domain.DeviceKindMSR,
		domain.DeviceKindIButton,
		domain.DeviceKindScanner,
		domain.DeviceKindPrinter:
		return true
	default:
		return false
	}
}

func isKnownDeviceStatus(status domain.DeviceStatus) bool {
	switch status {
	case domain.DeviceStatusReady,
		domain.DeviceStatusBusy,
		domain.DeviceStatusOffline,
		domain.DeviceStatusSimulated,
		domain.DeviceStatusError:
		return true
	default:
		return false
	}
}
