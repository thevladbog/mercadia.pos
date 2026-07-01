package api

import (
	"net/http"
	"strconv"
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

type ResetAuthLockoutRequest struct {
	Reason string `json:"reason,omitempty"`
}

type StoreAuthSettingsAcceptedResponse struct {
	Settings StoreAuthSettingsResponse `json:"settings"`
}

type AuthAttemptsResponse struct {
	Items      []AuthAttemptResponse `json:"items"`
	TotalCount int                   `json:"totalCount"`
}

type AuthAttemptResponse struct {
	ID                    string    `json:"id"`
	StoreID               string    `json:"storeId"`
	ActorID               string    `json:"actorId"`
	TerminalID            string    `json:"terminalId,omitempty"`
	CredentialKind        string    `json:"credentialKind,omitempty"`
	CredentialFingerprint string    `json:"credentialFingerprint,omitempty"`
	Successful            bool      `json:"successful"`
	FailureReason         string    `json:"failureReason,omitempty"`
	CreatedAt             time.Time `json:"createdAt"`
}

type AuthLockoutResetAcceptedResponse struct {
	Reset AuthLockoutResetResponse `json:"reset"`
}

type AuthLockoutResetResponse struct {
	StoreID   string    `json:"storeId"`
	ActorID   string    `json:"actorId"`
	ResetByID string    `json:"resetById"`
	Reason    string    `json:"reason,omitempty"`
	ResetAt   time.Time `json:"resetAt"`
}

func mountStoreSettingsRoutes(mux *http.ServeMux, spec *httpapi.Spec, auth *app.AuthService, settings *app.StoreAuthSettingsService) {
	authAttemptQueryParams := append(paginationQueryParams(),
		httpapi.QueryParamSpec{Name: "actorId", Description: "Filter by actor ID", Schema: httpapi.StringSchema()},
		httpapi.QueryParamSpec{Name: "terminalId", Description: "Filter by terminal ID", Schema: httpapi.StringSchema()},
		httpapi.QueryParamSpec{Name: "successful", Description: "Filter by successful flag", Schema: httpapi.Schema{"type": "boolean"}},
		httpapi.QueryParamSpec{Name: "since", Description: "Filter attempts created at or after this RFC3339 timestamp", Schema: httpapi.StringSchema()},
		httpapi.QueryParamSpec{Name: "until", Description: "Filter attempts created at or before this RFC3339 timestamp", Schema: httpapi.StringSchema()},
	)
	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/auth-attempts",
		OperationID: "listStoreAuthAttempts",
		Summary:     "List store authentication audit attempts. Requires `X-Session-Token` header.",
		Tags:        []string{"store-settings"},
		HeaderParameters: []httpapi.HeaderParamSpec{
			credentialSessionHeaderParameter(),
		},
		QueryParameters: authAttemptQueryParams,
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Authentication attempt audit list", Schema: authAttemptsResponseSchema()},
			"400": {Description: "Invalid auth attempt query", Schema: httpapi.ProblemSchema()},
			"401": {Description: "Session is missing or invalid", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		session, ok := storeSettingsSession(w, r, auth)
		if !ok {
			return
		}
		query := r.URL.Query()
		var successful *bool
		if raw := query.Get("successful"); raw != "" {
			parsed, err := strconv.ParseBool(raw)
			if err != nil {
				httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_successful_filter", "Invalid successful filter", err.Error())
				return
			}
			successful = &parsed
		}
		since, ok := optionalRFC3339Time(w, query.Get("since"), "since")
		if !ok {
			return
		}
		until, ok := optionalRFC3339Time(w, query.Get("until"), "until")
		if !ok {
			return
		}
		result, err := settings.ListAuthAttempts(r.Context(), app.ListAuthAttemptsQuery{
			StoreID:    r.PathValue("storeId"),
			ManagerID:  session.ActorID,
			ActorID:    query.Get("actorId"),
			TerminalID: query.Get("terminalId"),
			Successful: successful,
			Since:      since,
			Until:      until,
			Page:       app.ParsePageParams(query.Get("limit"), query.Get("offset")),
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, authAttemptsResponse(result))
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPost,
		Path:        "/v1/stores/{storeId}/auth-lockouts/{actorId}/reset",
		OperationID: "resetStoreAuthLockout",
		Summary:     "Reset a store auth lockout for an actor. Requires `X-Session-Token` header.",
		Tags:        []string{"store-settings"},
		HeaderParameters: []httpapi.HeaderParamSpec{
			credentialSessionHeaderParameter(),
		},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Auth lockout reset command",
			Required:    false,
			Schema:      resetAuthLockoutRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Auth lockout reset", Schema: authLockoutResetAcceptedResponseSchema()},
			"400": {Description: "Invalid auth lockout reset command", Schema: httpapi.ProblemSchema()},
			"401": {Description: "Session is missing or invalid", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Actor was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Idempotency key was reused for another reset command", Schema: httpapi.ProblemSchema()},
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
		var request ResetAuthLockoutRequest
		if r.Body != nil && r.ContentLength != 0 {
			if err := httpapi.DecodeJSON(r, &request); err != nil {
				httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
				return
			}
		}
		result, err := settings.ResetAuthLockout(r.Context(), app.ResetAuthLockoutCommand{
			IdempotencyKey: idempotencyKey,
			StoreID:        r.PathValue("storeId"),
			ActorID:        r.PathValue("actorId"),
			ManagerID:      session.ActorID,
			Reason:         request.Reason,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, AuthLockoutResetAcceptedResponse{Reset: authLockoutResetResponse(result)})
	})

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

func optionalRFC3339Time(w http.ResponseWriter, raw string, name string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, true
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_time_filter", "Invalid time filter", name+": "+err.Error())
		return time.Time{}, false
	}
	return parsed, true
}

func authAttemptsResponse(result app.PageResult[app.AuthAttemptResult]) AuthAttemptsResponse {
	items := make([]AuthAttemptResponse, 0, len(result.Items))
	for _, attempt := range result.Items {
		items = append(items, AuthAttemptResponse{
			ID:                    attempt.ID,
			StoreID:               attempt.StoreID,
			ActorID:               attempt.ActorID,
			TerminalID:            attempt.TerminalID,
			CredentialKind:        string(attempt.CredentialKind),
			CredentialFingerprint: attempt.CredentialFingerprint,
			Successful:            attempt.Successful,
			FailureReason:         attempt.FailureReason,
			CreatedAt:             attempt.CreatedAt,
		})
	}
	return AuthAttemptsResponse{Items: items, TotalCount: result.TotalCount}
}

func authLockoutResetResponse(result app.AuthLockoutResetResult) AuthLockoutResetResponse {
	return AuthLockoutResetResponse{
		StoreID:   result.StoreID,
		ActorID:   result.ActorID,
		ResetByID: result.ResetByID,
		Reason:    result.Reason,
		ResetAt:   result.ResetAt,
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

func authAttemptsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(authAttemptResponseSchema()),
		"totalCount": {"type": "integer", "minimum": 0},
	}, "items", "totalCount")
}

func authAttemptResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":                    httpapi.StringSchema(),
		"storeId":               httpapi.StringSchema(),
		"actorId":               httpapi.StringSchema(),
		"terminalId":            httpapi.StringSchema(),
		"credentialKind":        httpapi.StringSchema(),
		"credentialFingerprint": httpapi.StringSchema(),
		"successful":            {"type": "boolean"},
		"failureReason":         httpapi.StringSchema(),
		"createdAt":             httpapi.StringSchema(),
	}, "id", "storeId", "actorId", "successful", "createdAt")
}

func resetAuthLockoutRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"reason": httpapi.StringSchema(),
	})
}

func authLockoutResetAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"reset": authLockoutResetResponseSchema(),
	}, "reset")
}

func authLockoutResetResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":   httpapi.StringSchema(),
		"actorId":   httpapi.StringSchema(),
		"resetById": httpapi.StringSchema(),
		"reason":    httpapi.StringSchema(),
		"resetAt":   httpapi.StringSchema(),
	}, "storeId", "actorId", "resetById", "resetAt")
}
