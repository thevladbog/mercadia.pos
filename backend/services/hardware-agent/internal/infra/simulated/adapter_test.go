package simulated_test

import (
	"context"
	"errors"
	"testing"

	"mercadia.dev/pos/services/hardware-agent/internal/domain"
	"mercadia.dev/pos/services/hardware-agent/internal/infra/simulated"
)

// These tests pin the request/response contract every simulated adapter
// guarantees today via the public Registry/Adapter surface. A real driver
// (starting with ATOL fiscal, see docs/hardware-support-matrix.md §4c) must
// satisfy the same shape so callers (DeviceService, store-edge's hardware
// agent client) don't need to change when the adapter is swapped.

func TestFiscalAdapterCommandContract(t *testing.T) {
	ctx := context.Background()
	device := domain.Device{
		ID:     "fiscal-1",
		Kind:   domain.DeviceKindFiscal,
		Status: domain.DeviceStatusSimulated,
		Model:  "Test Fiscal",
	}
	registry := simulated.NewRegistry(simulated.NewFiscalAdapter())

	// get_status: documented fields, per fiscal.go's status() helper.
	statusResult, err := registry.Execute(ctx, device, "get_status", nil)
	if err != nil {
		t.Fatalf("get_status: %v", err)
	}
	if statusResult["driverState"] != string(device.Status) {
		t.Fatalf("driverState = %v", statusResult["driverState"])
	}
	if statusResult["fiscalMode"] != true {
		t.Fatalf("fiscalMode = %v", statusResult["fiscalMode"])
	}
	if statusResult["shiftState"] != "opened" {
		t.Fatalf("shiftState = %v", statusResult["shiftState"])
	}
	if statusResult["paperPresent"] != true {
		t.Fatalf("paperPresent = %v", statusResult["paperPresent"])
	}
	if statusResult["coverClosed"] != true {
		t.Fatalf("coverClosed = %v", statusResult["coverClosed"])
	}
	if statusResult["serialNumber"] != device.ID {
		t.Fatalf("serialNumber = %v", statusResult["serialNumber"])
	}
	if _, ok := statusResult["firmwareVersion"].(string); !ok {
		t.Fatalf("firmwareVersion missing/not a string: %v", statusResult["firmwareVersion"])
	}
	if _, ok := statusResult["lastDocumentNumber"]; !ok {
		t.Fatal("lastDocumentNumber missing")
	}
	if _, ok := statusResult["sessionNumber"]; !ok {
		t.Fatal("sessionNumber missing")
	}
	if statusResult["model"] != device.Model {
		t.Fatalf("model = %v", statusResult["model"])
	}

	// print_receipt: fiscalSign must be non-empty (store-edge's client
	// validates this, internal/infra/hardwareagent/client.go:275-278).
	printResult, err := registry.Execute(ctx, device, "print_receipt", map[string]any{"totalMinor": float64(1999)})
	if err != nil {
		t.Fatalf("print_receipt: %v", err)
	}
	fiscalSign, ok := printResult["fiscalSign"].(string)
	if !ok || fiscalSign == "" {
		t.Fatalf("expected non-empty fiscalSign, got %v", printResult["fiscalSign"])
	}
	if printResult["driverState"] != "ready" {
		t.Fatalf("driverState = %v", printResult["driverState"])
	}
	if printResult["shiftState"] != "opened" {
		t.Fatalf("shiftState = %v", printResult["shiftState"])
	}
	if _, ok := printResult["fiscalDocumentNumber"]; !ok {
		t.Fatal("fiscalDocumentNumber missing")
	}
	if _, ok := printResult["qrCode"].(string); !ok {
		t.Fatal("qrCode missing/not a string")
	}
	if _, ok := printResult["printedAt"]; !ok {
		t.Fatal("printedAt missing")
	}

	// cancel_receipt
	cancelResult, err := registry.Execute(ctx, device, "cancel_receipt", nil)
	if err != nil {
		t.Fatalf("cancel_receipt: %v", err)
	}
	if cancelResult["cancelled"] != true {
		t.Fatalf("cancelled = %v", cancelResult["cancelled"])
	}
	if cancelResult["driverState"] != "ready" {
		t.Fatalf("driverState = %v", cancelResult["driverState"])
	}

	// close_shift
	closeResult, err := registry.Execute(ctx, device, "close_shift", nil)
	if err != nil {
		t.Fatalf("close_shift: %v", err)
	}
	if closeResult["shiftState"] != "closed" {
		t.Fatalf("shiftState = %v", closeResult["shiftState"])
	}
	if _, ok := closeResult["zReportNumber"]; !ok {
		t.Fatal("zReportNumber missing")
	}
	if closeResult["driverState"] != "ready" {
		t.Fatalf("driverState = %v", closeResult["driverState"])
	}

	// print_receipt while shift is closed must fail (documents the real
	// contract's shift-lifecycle dependency for a future ATOL adapter).
	if _, err := registry.Execute(ctx, device, "print_receipt", map[string]any{"totalMinor": float64(500)}); err == nil {
		t.Fatal("expected print_receipt to fail while shift is closed")
	}

	// open_shift
	openResult, err := registry.Execute(ctx, device, "open_shift", nil)
	if err != nil {
		t.Fatalf("open_shift: %v", err)
	}
	if openResult["shiftState"] != "opened" {
		t.Fatalf("shiftState = %v", openResult["shiftState"])
	}
	if _, ok := openResult["sessionNumber"]; !ok {
		t.Fatal("sessionNumber missing")
	}
	if openResult["driverState"] != "ready" {
		t.Fatalf("driverState = %v", openResult["driverState"])
	}

	// unknown command type -> error
	if _, err := registry.Execute(ctx, device, "not_a_real_command", nil); !errors.Is(err, simulated.ErrUnsupportedCommand) {
		t.Fatalf("expected ErrUnsupportedCommand, got %v", err)
	}
}

func TestPaymentTerminalAdapterCommandContract(t *testing.T) {
	ctx := context.Background()
	device := domain.Device{
		ID:     "payment-1",
		Kind:   domain.DeviceKindPaymentTerminal,
		Status: domain.DeviceStatusSimulated,
		Model:  "Test Terminal",
	}
	registry := simulated.NewRegistry(simulated.NewPaymentTerminalAdapter())

	statusResult, err := registry.Execute(ctx, device, "get_status", nil)
	if err != nil {
		t.Fatalf("get_status: %v", err)
	}
	if statusResult["terminalState"] != "idle" {
		t.Fatalf("terminalState = %v", statusResult["terminalState"])
	}
	if statusResult["connection"] != "online" {
		t.Fatalf("connection = %v", statusResult["connection"])
	}
	if statusResult["model"] != device.Model {
		t.Fatalf("model = %v", statusResult["model"])
	}
	if statusResult["serialNumber"] != device.ID {
		t.Fatalf("serialNumber = %v", statusResult["serialNumber"])
	}
	if _, ok := statusResult["lastAuthCode"]; !ok {
		t.Fatal("lastAuthCode missing")
	}

	authResult, err := registry.Execute(ctx, device, "authorize", map[string]any{"amountMinor": float64(9900), "currency": "RUB"})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if authResult["status"] != "approved" {
		t.Fatalf("status = %v", authResult["status"])
	}
	authCode, ok := authResult["authCode"].(string)
	if !ok || authCode == "" {
		t.Fatalf("expected non-empty authCode, got %v", authResult["authCode"])
	}
	rrn, ok := authResult["rrn"].(string)
	if !ok || rrn == "" {
		t.Fatalf("expected non-empty rrn, got %v", authResult["rrn"])
	}
	if _, ok := authResult["cardMask"].(string); !ok {
		t.Fatal("cardMask missing/not a string")
	}
	if authResult["amountMinor"] != int64(9900) {
		t.Fatalf("amountMinor = %v", authResult["amountMinor"])
	}
	if authResult["currency"] != "RUB" {
		t.Fatalf("currency = %v", authResult["currency"])
	}
	if authResult["terminalState"] != "idle" {
		t.Fatalf("terminalState = %v", authResult["terminalState"])
	}

	captureResult, err := registry.Execute(ctx, device, "capture", nil)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if captureResult["status"] != "captured" {
		t.Fatalf("status = %v", captureResult["status"])
	}
	if captureResult["authCode"] != authCode {
		t.Fatalf("authCode = %v, want %v", captureResult["authCode"], authCode)
	}
	if captureResult["rrn"] != rrn {
		t.Fatalf("rrn = %v, want %v", captureResult["rrn"], rrn)
	}
	if captureResult["terminalState"] != "idle" {
		t.Fatalf("terminalState = %v", captureResult["terminalState"])
	}

	cancelResult, err := registry.Execute(ctx, device, "cancel", nil)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cancelResult["status"] != "cancelled" {
		t.Fatalf("status = %v", cancelResult["status"])
	}
	if cancelResult["terminalState"] != "idle" {
		t.Fatalf("terminalState = %v", cancelResult["terminalState"])
	}

	refundResult, err := registry.Execute(ctx, device, "refund", map[string]any{"amountMinor": float64(9900)})
	if err != nil {
		t.Fatalf("refund: %v", err)
	}
	if refundResult["status"] != "refunded" {
		t.Fatalf("status = %v", refundResult["status"])
	}
	if _, ok := refundResult["rrn"]; !ok {
		t.Fatal("rrn missing")
	}
	if _, ok := refundResult["authCode"]; !ok {
		t.Fatal("authCode missing")
	}
	if refundResult["amountMinor"] != int64(9900) {
		t.Fatalf("amountMinor = %v", refundResult["amountMinor"])
	}
	if refundResult["terminalState"] != "idle" {
		t.Fatalf("terminalState = %v", refundResult["terminalState"])
	}

	if _, err := registry.Execute(ctx, device, "not_a_real_command", nil); !errors.Is(err, simulated.ErrUnsupportedCommand) {
		t.Fatalf("expected ErrUnsupportedCommand, got %v", err)
	}
}

func TestMSRAdapterCommandContract(t *testing.T) {
	ctx := context.Background()
	device := domain.Device{
		ID:     "msr-1",
		Kind:   domain.DeviceKindMSR,
		Status: domain.DeviceStatusSimulated,
		Model:  "Test MSR",
	}
	registry := simulated.NewRegistry(simulated.NewMSRAdapter())

	statusResult, err := registry.Execute(ctx, device, "get_status", nil)
	if err != nil {
		t.Fatalf("get_status: %v", err)
	}
	if statusResult["readerState"] != "ready" {
		t.Fatalf("readerState = %v", statusResult["readerState"])
	}
	if statusResult["model"] != device.Model {
		t.Fatalf("model = %v", statusResult["model"])
	}

	cardResult, err := registry.Execute(ctx, device, "read_card", nil)
	if err != nil {
		t.Fatalf("read_card: %v", err)
	}
	if track1, ok := cardResult["track1"].(string); !ok || track1 == "" {
		t.Fatalf("track1 = %v", cardResult["track1"])
	}
	if track2, ok := cardResult["track2"].(string); !ok || track2 == "" {
		t.Fatalf("track2 = %v", cardResult["track2"])
	}
	if masked, ok := cardResult["masked"].(string); !ok || masked == "" {
		t.Fatalf("masked = %v", cardResult["masked"])
	}

	staffResult, err := registry.Execute(ctx, device, "read_staff_card", nil)
	if err != nil {
		t.Fatalf("read_staff_card: %v", err)
	}
	if staffToken, ok := staffResult["staffToken"].(string); !ok || staffToken == "" {
		t.Fatalf("staffToken = %v", staffResult["staffToken"])
	}
	if masked, ok := staffResult["masked"].(string); !ok || masked == "" {
		t.Fatalf("masked = %v", staffResult["masked"])
	}

	if _, err := registry.Execute(ctx, device, "not_a_real_command", nil); !errors.Is(err, simulated.ErrUnsupportedCommand) {
		t.Fatalf("expected ErrUnsupportedCommand, got %v", err)
	}
}

func TestIButtonAdapterCommandContract(t *testing.T) {
	ctx := context.Background()
	device := domain.Device{
		ID:     "ibutton-1",
		Kind:   domain.DeviceKindIButton,
		Status: domain.DeviceStatusSimulated,
		Model:  "Test iButton",
	}
	registry := simulated.NewRegistry(simulated.NewIButtonAdapter())

	statusResult, err := registry.Execute(ctx, device, "get_status", nil)
	if err != nil {
		t.Fatalf("get_status: %v", err)
	}
	if statusResult["readerState"] != "ready" {
		t.Fatalf("readerState = %v", statusResult["readerState"])
	}
	if statusResult["model"] != device.Model {
		t.Fatalf("model = %v", statusResult["model"])
	}

	keyResult, err := registry.Execute(ctx, device, "read_key", nil)
	if err != nil {
		t.Fatalf("read_key: %v", err)
	}
	if romID, ok := keyResult["romId"].(string); !ok || romID == "" {
		t.Fatalf("romId = %v", keyResult["romId"])
	}

	if _, err := registry.Execute(ctx, device, "not_a_real_command", nil); !errors.Is(err, simulated.ErrUnsupportedCommand) {
		t.Fatalf("expected ErrUnsupportedCommand, got %v", err)
	}
}

func TestScannerAdapterCommandContract(t *testing.T) {
	ctx := context.Background()
	device := domain.Device{
		ID:     "scanner-1",
		Kind:   domain.DeviceKindScanner,
		Status: domain.DeviceStatusSimulated,
		Model:  "Test Scanner",
	}
	registry := simulated.NewRegistry(simulated.NewScannerAdapter())

	statusResult, err := registry.Execute(ctx, device, "get_status", nil)
	if err != nil {
		t.Fatalf("get_status: %v", err)
	}
	if statusResult["scannerState"] != "ready" {
		t.Fatalf("scannerState = %v", statusResult["scannerState"])
	}
	if statusResult["model"] != device.Model {
		t.Fatalf("model = %v", statusResult["model"])
	}

	scanResult, err := registry.Execute(ctx, device, "scan", nil)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if barcode, ok := scanResult["barcode"].(string); !ok || barcode == "" {
		t.Fatalf("barcode = %v", scanResult["barcode"])
	}
	if scanResult["symbology"] != "EAN13" {
		t.Fatalf("symbology = %v", scanResult["symbology"])
	}

	customScanResult, err := registry.Execute(ctx, device, "scan", map[string]any{"barcode": "1234567890123"})
	if err != nil {
		t.Fatalf("scan with payload: %v", err)
	}
	if customScanResult["barcode"] != "1234567890123" {
		t.Fatalf("barcode = %v", customScanResult["barcode"])
	}

	staffResult, err := registry.Execute(ctx, device, "scan_staff_card", nil)
	if err != nil {
		t.Fatalf("scan_staff_card: %v", err)
	}
	if staffToken, ok := staffResult["staffToken"].(string); !ok || staffToken == "" {
		t.Fatalf("staffToken = %v", staffResult["staffToken"])
	}
	if masked, ok := staffResult["masked"].(string); !ok || masked == "" {
		t.Fatalf("masked = %v", staffResult["masked"])
	}
	if staffResult["symbology"] != "CODE128" {
		t.Fatalf("symbology = %v", staffResult["symbology"])
	}

	if _, err := registry.Execute(ctx, device, "not_a_real_command", nil); !errors.Is(err, simulated.ErrUnsupportedCommand) {
		t.Fatalf("expected ErrUnsupportedCommand, got %v", err)
	}
}

func TestPrinterAdapterCommandContract(t *testing.T) {
	ctx := context.Background()
	device := domain.Device{
		ID:     "printer-1",
		Kind:   domain.DeviceKindPrinter,
		Status: domain.DeviceStatusSimulated,
		Model:  "Test Printer",
	}
	registry := simulated.NewRegistry(simulated.NewPrinterAdapter())

	statusResult, err := registry.Execute(ctx, device, "get_status", nil)
	if err != nil {
		t.Fatalf("get_status: %v", err)
	}
	if statusResult["printerState"] != "ready" {
		t.Fatalf("printerState = %v", statusResult["printerState"])
	}
	if statusResult["paperPresent"] != true {
		t.Fatalf("paperPresent = %v", statusResult["paperPresent"])
	}
	if statusResult["coverClosed"] != true {
		t.Fatalf("coverClosed = %v", statusResult["coverClosed"])
	}
	if statusResult["model"] != device.Model {
		t.Fatalf("model = %v", statusResult["model"])
	}

	printResult, err := registry.Execute(ctx, device, "print", nil)
	if err != nil {
		t.Fatalf("print: %v", err)
	}
	if printResult["printedLines"] != 1 {
		t.Fatalf("printedLines = %v", printResult["printedLines"])
	}
	if printResult["printerState"] != "ready" {
		t.Fatalf("printerState = %v", printResult["printerState"])
	}

	customPrintResult, err := registry.Execute(ctx, device, "print", map[string]any{"lines": float64(5)})
	if err != nil {
		t.Fatalf("print with payload: %v", err)
	}
	if customPrintResult["printedLines"] != 5 {
		t.Fatalf("printedLines = %v", customPrintResult["printedLines"])
	}

	if _, err := registry.Execute(ctx, device, "not_a_real_command", nil); !errors.Is(err, simulated.ErrUnsupportedCommand) {
		t.Fatalf("expected ErrUnsupportedCommand, got %v", err)
	}
}

func TestDefaultRegistryHasAllSixKinds(t *testing.T) {
	registry := simulated.DefaultRegistry()
	ctx := context.Background()

	for _, kind := range []domain.DeviceKind{
		domain.DeviceKindFiscal,
		domain.DeviceKindPaymentTerminal,
		domain.DeviceKindMSR,
		domain.DeviceKindIButton,
		domain.DeviceKindScanner,
		domain.DeviceKindPrinter,
	} {
		device := domain.Device{ID: "probe-" + string(kind), Kind: kind, Status: domain.DeviceStatusSimulated}
		if _, err := registry.Execute(ctx, device, "get_status", nil); err != nil {
			t.Fatalf("kind %s: get_status: %v", kind, err)
		}
	}
}
