package api

import (
	"testing"

	"mercadia.dev/pos/services/hardware-agent/internal/domain"
)

func TestParseDeviceConfigJSONValidInput(t *testing.T) {
	raw := `[
		{"id":"cfg-fiscal-1","kind":"fiscal","model":"Test Fiscal"},
		{"id":"cfg-scanner-1","kind":"scanner","model":"Test Scanner","status":"ready"}
	]`

	devices, err := parseDeviceConfigJSON(raw)
	if err != nil {
		t.Fatalf("parseDeviceConfigJSON: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}

	if devices[0].ID != "cfg-fiscal-1" || devices[0].Kind != domain.DeviceKindFiscal || devices[0].Model != "Test Fiscal" {
		t.Fatalf("unexpected first device: %+v", devices[0])
	}
	if devices[0].Status != domain.DeviceStatusSimulated {
		t.Fatalf("expected default status simulated, got %s", devices[0].Status)
	}
	if devices[0].UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}

	if devices[1].Status != domain.DeviceStatusReady {
		t.Fatalf("expected explicit status ready, got %s", devices[1].Status)
	}
}

func TestParseDeviceConfigJSONIgnoresUnknownFields(t *testing.T) {
	raw := `[{"id":"cfg-fiscal-1","kind":"fiscal","connection":{"port":"COM3"}}]`

	devices, err := parseDeviceConfigJSON(raw)
	if err != nil {
		t.Fatalf("parseDeviceConfigJSON: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
}

func TestParseDeviceConfigJSONRejectsInvalidJSON(t *testing.T) {
	if _, err := parseDeviceConfigJSON(`not json`); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseDeviceConfigJSONRejectsEmptyArray(t *testing.T) {
	if _, err := parseDeviceConfigJSON(`[]`); err == nil {
		t.Fatal("expected error for empty array")
	}
}

func TestParseDeviceConfigJSONRejectsEmptyID(t *testing.T) {
	raw := `[{"id":"","kind":"fiscal"}]`
	if _, err := parseDeviceConfigJSON(raw); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestParseDeviceConfigJSONRejectsDuplicateID(t *testing.T) {
	raw := `[{"id":"dup-1","kind":"fiscal"},{"id":"dup-1","kind":"scanner"}]`
	if _, err := parseDeviceConfigJSON(raw); err == nil {
		t.Fatal("expected error for duplicate id")
	}
}

func TestParseDeviceConfigJSONRejectsUnknownKind(t *testing.T) {
	raw := `[{"id":"cfg-1","kind":"toaster"}]`
	if _, err := parseDeviceConfigJSON(raw); err == nil {
		t.Fatal("expected error for unknown kind")
	}
}

func TestParseDeviceConfigJSONRejectsUnknownStatus(t *testing.T) {
	raw := `[{"id":"cfg-1","kind":"fiscal","status":"on_fire"}]`
	if _, err := parseDeviceConfigJSON(raw); err == nil {
		t.Fatal("expected error for unknown status")
	}
}

func TestLoadConfiguredDevicesUnsetReturnsNil(t *testing.T) {
	t.Setenv(devicesEnvVar, "")

	devices, err := loadConfiguredDevices()
	if err != nil {
		t.Fatalf("loadConfiguredDevices: %v", err)
	}
	if devices != nil {
		t.Fatalf("expected nil devices when env is unset, got %+v", devices)
	}
}

func TestLoadConfiguredDevicesParsesEnv(t *testing.T) {
	t.Setenv(devicesEnvVar, `[{"id":"cfg-fiscal-1","kind":"fiscal"}]`)

	devices, err := loadConfiguredDevices()
	if err != nil {
		t.Fatalf("loadConfiguredDevices: %v", err)
	}
	if len(devices) != 1 || devices[0].ID != "cfg-fiscal-1" {
		t.Fatalf("unexpected devices: %+v", devices)
	}
}

func TestLoadConfiguredDevicesPropagatesParseError(t *testing.T) {
	t.Setenv(devicesEnvVar, `not json`)

	if _, err := loadConfiguredDevices(); err == nil {
		t.Fatal("expected error for invalid JSON in env var")
	}
}
