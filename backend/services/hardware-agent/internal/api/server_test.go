package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOpenAPIExposesHardwareOperations(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)

	NewServer().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}

	var document map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatalf("decode OpenAPI: %v", err)
	}

	paths := document["paths"].(map[string]any)
	for _, path := range []string{
		"/v1/hardware/status",
		"/v1/devices",
		"/v1/devices/{deviceId}",
		"/v1/devices/{deviceId}/commands",
		"/v1/devices/{deviceId}/commands/{commandId}",
	} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("expected %s path", path)
		}
	}
}

func TestListDevicesReturnsSeededSimulatedDevices(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/devices", nil)

	NewServer().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}

	var devices []DeviceResponse
	if err := json.Unmarshal(response.Body.Bytes(), &devices); err != nil {
		t.Fatalf("decode devices: %v", err)
	}
	if len(devices) < 6 {
		t.Fatalf("expected at least 6 devices, got %d", len(devices))
	}
}

func TestSendDeviceCommandRequiresIdempotencyKey(t *testing.T) {
	body := bytes.NewBufferString(`{"type":"get_status"}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/devices/sim-fiscal-1/commands", body)

	NewServer().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestSendDeviceCommandIsIdempotentOverHTTP(t *testing.T) {
	server := NewServer()
	body := bytes.NewBufferString(`{"type":"get_status"}`)

	first := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(http.MethodPost, "/v1/devices/sim-fiscal-1/commands", bytes.NewBufferString(body.String()))
	firstRequest.Header.Set("Idempotency-Key", "http-cmd-1")
	server.ServeHTTP(first, firstRequest)

	second := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(http.MethodPost, "/v1/devices/sim-fiscal-1/commands", bytes.NewBufferString(body.String()))
	secondRequest.Header.Set("Idempotency-Key", "http-cmd-1")
	server.ServeHTTP(second, secondRequest)

	if first.Code != http.StatusAccepted || second.Code != http.StatusAccepted {
		t.Fatalf("status codes = %d and %d", first.Code, second.Code)
	}

	var firstResponse DeviceCommandAcceptedResponse
	var secondResponse DeviceCommandAcceptedResponse
	if err := json.Unmarshal(first.Body.Bytes(), &firstResponse); err != nil {
		t.Fatalf("decode first response: %v", err)
	}
	if err := json.Unmarshal(second.Body.Bytes(), &secondResponse); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if firstResponse.Command.ID != secondResponse.Command.ID {
		t.Fatalf("expected same command id, got %s and %s", firstResponse.Command.ID, secondResponse.Command.ID)
	}
}

func TestGetDeviceCommandEventuallyCompletes(t *testing.T) {
	server := NewServer()
	sendBody := bytes.NewBufferString(`{"type":"authorize","payload":{"amountMinor":1200,"currency":"RUB"}}`)
	sendResponse := httptest.NewRecorder()
	sendRequest := httptest.NewRequest(http.MethodPost, "/v1/devices/sim-payment-1/commands", sendBody)
	sendRequest.Header.Set("Idempotency-Key", "poll-cmd-1")
	server.ServeHTTP(sendResponse, sendRequest)

	if sendResponse.Code != http.StatusAccepted {
		t.Fatalf("send status = %d", sendResponse.Code)
	}

	var accepted DeviceCommandAcceptedResponse
	if err := json.Unmarshal(sendResponse.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode accepted response: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		getResponse := httptest.NewRecorder()
		getRequest := httptest.NewRequest(http.MethodGet, "/v1/devices/sim-payment-1/commands/"+accepted.Command.ID, nil)
		server.ServeHTTP(getResponse, getRequest)

		if getResponse.Code != http.StatusOK {
			t.Fatalf("get status = %d", getResponse.Code)
		}

		var command DeviceCommandResponse
		if err := json.Unmarshal(getResponse.Body.Bytes(), &command); err != nil {
			t.Fatalf("decode command: %v", err)
		}
		if command.Status == "completed" {
			if command.Result["status"] != "approved" {
				t.Fatalf("result status = %v", command.Result["status"])
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("command did not complete in time")
}
