package api

import (
	"net/http"
	"time"

	"mercadia.dev/pos/platform/httpapi"
	"mercadia.dev/pos/services/store-edge/internal/app"
)

type StoreAuthSettingsResponse struct {
	StoreID                string     `json:"storeId"`
	FailedAttemptLimit     int        `json:"failedAttemptLimit"`
	LockoutDurationSeconds int        `json:"lockoutDurationSeconds"`
	POSAutoLockSeconds     int        `json:"posAutoLockSeconds"`
	UpdatedByID            string     `json:"updatedById,omitempty"`
	UpdatedAt              *time.Time `json:"updatedAt,omitempty"`
}

type SetStoreAuthSettingsRequest struct {
	FailedAttemptLimit     int `json:"failedAttemptLimit"`
	LockoutDurationSeconds int `json:"lockoutDurationSeconds"`
	POSAutoLockSeconds     int `json:"posAutoLockSeconds"`
}

type StoreAuthSettingsAcceptedResponse struct {
	Settings StoreAuthSettingsResponse `json:"settings"`
}

func mountStoreSettingsRoutes(mux *http.ServeMux, spec *httpapi.Spec, auth *app.AuthService, settings *app.StoreAuthSettingsService) {
	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/auth-settings",
		OperationID: "getStoreAuthSettings",
		Summary:     "Get terminal-safe store authentication settings",
		Tags:        []string{"store-settings"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Store authentication settings", Schema: storeAuthSettingsAcceptedResponseSchema()},
			"400": {Description: "Invalid store auth settings query", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := settings.GetStoreAuthSettings(r.Context(), r.PathValue("storeId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, StoreAuthSettingsAcceptedResponse{Settings: storeAuthSettingsResponse(result)})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPut,
		Path:        "/v1/stores/{storeId}/auth-settings",
		OperationID: "setStoreAuthSettings",
		Summary:     "Set store authentication hardening settings. Requires `X-Session-Token` header.",
		Tags:        []string{"store-settings"},
		HeaderParameters: []httpapi.HeaderParamSpec{
			credentialSessionHeaderParameter(),
		},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Store authentication settings command",
			Required:    true,
			Schema:      setStoreAuthSettingsRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Store authentication settings", Schema: storeAuthSettingsAcceptedResponseSchema()},
			"400": {Description: "Invalid store auth settings command", Schema: httpapi.ProblemSchema()},
			"401": {Description: "Session is missing or invalid", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Idempotency key was reused for another settings command", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		session, ok := storeSettingsSession(w, r, auth)
		if !ok {
			return
		}
		idempotencyKey, err := httpapi.RequireIdempotencyKey(r)
		if err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request SetStoreAuthSettingsRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := settings.SetStoreAuthSettings(r.Context(), app.SetStoreAuthSettingsCommand{
			IdempotencyKey:         idempotencyKey,
			StoreID:                r.PathValue("storeId"),
			ManagerID:              session.ActorID,
			FailedAttemptLimit:     request.FailedAttemptLimit,
			LockoutDurationSeconds: request.LockoutDurationSeconds,
			POSAutoLockSeconds:     request.POSAutoLockSeconds,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, StoreAuthSettingsAcceptedResponse{Settings: storeAuthSettingsResponse(result)})
	})
}

func storeSettingsSession(w http.ResponseWriter, r *http.Request, auth *app.AuthService) (*SessionContext, bool) {
	session, err := OptionalSessionFromRequest(r, auth)
	if err != nil {
		writeAppError(w, err)
		return nil, false
	}
	if session == nil {
		writeAppError(w, app.ErrSessionNotFound)
		return nil, false
	}
	return session, true
}

func storeAuthSettingsResponse(result app.StoreAuthSettingsResult) StoreAuthSettingsResponse {
	var updatedAt *time.Time
	if !result.UpdatedAt.IsZero() {
		updatedAt = &result.UpdatedAt
	}
	return StoreAuthSettingsResponse{
		StoreID:                result.StoreID,
		FailedAttemptLimit:     result.FailedAttemptLimit,
		LockoutDurationSeconds: result.LockoutDurationSeconds,
		POSAutoLockSeconds:     result.POSAutoLockSeconds,
		UpdatedByID:            result.UpdatedByID,
		UpdatedAt:              updatedAt,
	}
}

func storeAuthSettingsAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"settings": storeAuthSettingsResponseSchema(),
	}, "settings")
}

func storeAuthSettingsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":                httpapi.StringSchema(),
		"failedAttemptLimit":     {"type": "integer", "minimum": 1, "maximum": 20},
		"lockoutDurationSeconds": {"type": "integer", "minimum": 60, "maximum": 86400},
		"posAutoLockSeconds":     {"type": "integer", "minimum": 30, "maximum": 86400},
		"updatedById":            httpapi.StringSchema(),
		"updatedAt":              httpapi.StringSchema(),
	}, "storeId", "failedAttemptLimit", "lockoutDurationSeconds", "posAutoLockSeconds")
}

func setStoreAuthSettingsRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"failedAttemptLimit":     {"type": "integer", "minimum": 1, "maximum": 20},
		"lockoutDurationSeconds": {"type": "integer", "minimum": 60, "maximum": 86400},
		"posAutoLockSeconds":     {"type": "integer", "minimum": 30, "maximum": 86400},
	}, "failedAttemptLimit", "lockoutDurationSeconds", "posAutoLockSeconds")
}
