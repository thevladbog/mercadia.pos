package api

import (
	"errors"
	"net/http"
	"time"

	"mercadia.dev/pos/platform/httpapi"
	"mercadia.dev/pos/services/hardware-agent/internal/app"
	"mercadia.dev/pos/services/hardware-agent/internal/domain"
	"mercadia.dev/pos/services/hardware-agent/internal/infra/memory"
	"mercadia.dev/pos/services/hardware-agent/internal/infra/simulated"
)

const version = "0.1.0"

type StatusResponse struct {
	TerminalID  string    `json:"terminalId"`
	Status      string    `json:"status"`
	DeviceCount int       `json:"deviceCount"`
	GeneratedAt time.Time `json:"generatedAt"`
}

type DeviceResponse struct {
	ID        string              `json:"id"`
	Kind      domain.DeviceKind   `json:"kind"`
	Status    domain.DeviceStatus `json:"status"`
	Model     string              `json:"model,omitempty"`
	UpdatedAt time.Time           `json:"updatedAt"`
}

type DeviceCommandResponse struct {
	ID          string               `json:"id"`
	DeviceID    string               `json:"deviceId"`
	Type        string               `json:"type"`
	Payload     map[string]any       `json:"payload,omitempty"`
	Status      domain.CommandStatus `json:"status"`
	Result      map[string]any       `json:"result,omitempty"`
	Error       string               `json:"error,omitempty"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
	CompletedAt *time.Time           `json:"completedAt,omitempty"`
}

type SendDeviceCommandRequest struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
}

type DeviceCommandAcceptedResponse struct {
	Command DeviceCommandResponse `json:"command"`
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
	store := memory.NewStore()
	devices := app.NewDeviceService(store, store, store, simulated.DefaultRegistry())

	info := httpapi.ServiceInfo{
		Name:        "hardware-agent",
		Title:       "Mercadia Hardware Agent",
		Description: "Local device API for fiscal devices, payment terminals, scanners, scales, drawers, printers, MSR, and iButton readers.",
		Version:     version,
	}

	mux := http.NewServeMux()
	spec := httpapi.NewSpec(info)
	httpapi.MountSystemRoutes(mux, spec, info)
	mountRoutes(mux, spec, devices)

	return mux, spec
}

func mountRoutes(mux *http.ServeMux, spec *httpapi.Spec, devices *app.DeviceService) {
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
		status, err := devices.GetAgentStatus(r.Context())
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, StatusResponse{
			TerminalID:  status.TerminalID,
			Status:      status.Status,
			DeviceCount: status.DeviceCount,
			GeneratedAt: status.GeneratedAt,
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
		result, err := devices.ListDevices(r.Context())
		if err != nil {
			writeAppError(w, err)
			return
		}
		response := make([]DeviceResponse, 0, len(result))
		for _, device := range result {
			response = append(response, deviceResponse(device))
		}
		httpapi.WriteJSON(w, http.StatusOK, response)
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/devices/{deviceId}",
		OperationID: "getDevice",
		Summary:     "Get local device",
		Tags:        []string{"hardware"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Local device", Schema: deviceResponseSchema()},
			"404": {Description: "Device was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := devices.GetDevice(r.Context(), r.PathValue("deviceId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, deviceResponse(result.Device))
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/devices/{deviceId}/commands/{commandId}",
		OperationID: "getDeviceCommand",
		Summary:     "Get device command status",
		Tags:        []string{"hardware"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Device command", Schema: deviceCommandResponseSchema()},
			"404": {Description: "Device command was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := devices.GetCommand(r.Context(), r.PathValue("deviceId"), r.PathValue("commandId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, deviceCommandResponse(result.Command))
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/devices/{deviceId}/commands",
		OperationID:         "sendDeviceCommand",
		Summary:             "Send command to a local device",
		Tags:                []string{"hardware"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Device command",
			Required:    true,
			Schema:      sendDeviceCommandRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Device command accepted", Schema: deviceCommandAcceptedResponseSchema()},
			"400": {Description: "Invalid device command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Device was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Device command or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request SendDeviceCommandRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := devices.SendCommand(r.Context(), app.SendDeviceCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			DeviceID:       r.PathValue("deviceId"),
			Type:           request.Type,
			Payload:        request.Payload,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, DeviceCommandAcceptedResponse{
			Command: deviceCommandResponse(result.Command),
		})
	})
}

func deviceResponse(device domain.Device) DeviceResponse {
	return DeviceResponse{
		ID:        device.ID,
		Kind:      device.Kind,
		Status:    device.Status,
		Model:     device.Model,
		UpdatedAt: device.UpdatedAt,
	}
}

func deviceCommandResponse(command domain.DeviceCommand) DeviceCommandResponse {
	return DeviceCommandResponse{
		ID:          command.ID,
		DeviceID:    command.DeviceID,
		Type:        command.Type,
		Payload:     command.Payload,
		Status:      command.Status,
		Result:      command.Result,
		Error:       command.Error,
		CreatedAt:   command.CreatedAt,
		UpdatedAt:   command.UpdatedAt,
		CompletedAt: command.CompletedAt,
	}
}

func writeAppError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, app.ErrDeviceNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "device_not_found", "Device was not found", err.Error())
	case errors.Is(err, app.ErrCommandNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "command_not_found", "Device command was not found", err.Error())
	case errors.Is(err, app.ErrIdempotencyKeyRequired):
		httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
	case errors.Is(err, app.ErrIdempotencyKeyReused):
		httpapi.WriteProblem(w, http.StatusConflict, "idempotency_key_reused", "Idempotency key was reused", err.Error())
	case errors.Is(err, app.ErrInvalidDeviceCommand):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_device_command", "Invalid device command", err.Error())
	default:
		httpapi.WriteProblem(w, http.StatusInternalServerError, "internal_error", "Internal server error", err.Error())
	}
}

func statusResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"terminalId":  httpapi.StringSchema(),
		"status":      httpapi.StringSchema(),
		"deviceCount": {"type": "integer"},
		"generatedAt": httpapi.DateTimeSchema(),
	}, "terminalId", "status", "deviceCount", "generatedAt")
}

func deviceResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":        httpapi.StringSchema(),
		"kind":      httpapi.StringSchema(),
		"status":    httpapi.StringSchema(),
		"model":     httpapi.StringSchema(),
		"updatedAt": httpapi.DateTimeSchema(),
	}, "id", "kind", "status", "updatedAt")
}

func deviceCommandResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":          httpapi.StringSchema(),
		"deviceId":    httpapi.StringSchema(),
		"type":        httpapi.StringSchema(),
		"payload":     httpapi.ObjectSchema(map[string]httpapi.Schema{}),
		"status":      httpapi.StringSchema(),
		"result":      httpapi.ObjectSchema(map[string]httpapi.Schema{}),
		"error":       httpapi.StringSchema(),
		"createdAt":   httpapi.DateTimeSchema(),
		"updatedAt":   httpapi.DateTimeSchema(),
		"completedAt": httpapi.DateTimeSchema(),
	}, "id", "deviceId", "type", "status", "createdAt", "updatedAt")
}

func sendDeviceCommandRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"type":    httpapi.StringSchema(),
		"payload": httpapi.ObjectSchema(map[string]httpapi.Schema{}),
	}, "type")
}

func deviceCommandAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"command": deviceCommandResponseSchema(),
	}, "command")
}
