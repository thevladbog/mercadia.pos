package api

import (
	"net/http"
	"time"

	"mercadia.dev/pos/platform/httpapi"
)

const version = "0.1.0"

type StatusResponse struct {
	TerminalID  string    `json:"terminalId"`
	Status      string    `json:"status"`
	GeneratedAt time.Time `json:"generatedAt"`
}

type DeviceResponse struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Status string `json:"status"`
}

type DeviceCommandAcceptedResponse struct {
	DeviceID string `json:"deviceId"`
	Status   string `json:"status"`
}

func NewServer() http.Handler {
	mux, _ := newMuxAndSpec()
	return mux
}

func OpenAPI() map[string]any {
	_, spec := newMuxAndSpec()
	return spec.OpenAPI()
}

func newMuxAndSpec() (*http.ServeMux, *httpapi.Spec) {
	info := httpapi.ServiceInfo{
		Name:        "hardware-agent",
		Title:       "Mercadia Hardware Agent",
		Description: "Local device API for fiscal devices, payment terminals, scanners, scales, drawers, printers, MSR, and iButton readers.",
		Version:     version,
	}

	mux := http.NewServeMux()
	spec := httpapi.NewSpec(info)
	httpapi.MountSystemRoutes(mux, spec, info)
	mountRoutes(mux, spec)

	return mux, spec
}

func mountRoutes(mux *http.ServeMux, spec *httpapi.Spec) {
	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/hardware/status",
		OperationID: "getHardwareAgentStatus",
		Summary:     "Get Hardware Agent status",
		Tags:        []string{"hardware"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Hardware Agent status", Schema: statusResponseSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		httpapi.WriteJSON(w, http.StatusOK, StatusResponse{
			TerminalID:  "local-terminal",
			Status:      "ok",
			GeneratedAt: time.Now().UTC(),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/devices",
		OperationID: "listDevices",
		Summary:     "List local devices",
		Tags:        []string{"hardware"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Local device list", Schema: httpapi.ArraySchema(deviceResponseSchema())},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		httpapi.WriteJSON(w, http.StatusOK, []DeviceResponse{
			{ID: "mock-fiscal-1", Kind: "fiscal", Status: "simulated"},
			{ID: "mock-msr-1", Kind: "msr", Status: "simulated"},
			{ID: "mock-ibutton-1", Kind: "ibutton", Status: "simulated"},
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/devices/{deviceId}/commands",
		OperationID:         "sendDeviceCommand",
		Summary:             "Send command to a local device",
		Tags:                []string{"hardware"},
		RequiresIdempotency: true,
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Device command accepted", Schema: deviceCommandAcceptedResponseSchema()},
			"400": {Description: "Invalid device command", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, DeviceCommandAcceptedResponse{
			DeviceID: r.PathValue("deviceId"),
			Status:   "accepted",
		})
	})
}

func statusResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"terminalId":  httpapi.StringSchema(),
		"status":      httpapi.StringSchema(),
		"generatedAt": httpapi.DateTimeSchema(),
	}, "terminalId", "status", "generatedAt")
}

func deviceResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":     httpapi.StringSchema(),
		"kind":   httpapi.StringSchema(),
		"status": httpapi.StringSchema(),
	}, "id", "kind", "status")
}

func deviceCommandAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"deviceId": httpapi.StringSchema(),
		"status":   httpapi.StringSchema(),
	}, "deviceId", "status")
}
