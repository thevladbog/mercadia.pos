package api

import (
	"net/http"
	"time"

	"mercadia.dev/pos/platform/httpapi"
	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
)

type CreateCentralSessionRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CentralSessionAcceptedResponse struct {
	Session CentralSessionResponse `json:"session"`
}

type CentralSessionResponse struct {
	Token     string               `json:"token"`
	UserID    string               `json:"userId"`
	Roles     []domain.CentralRole `json:"roles"`
	ExpiresAt time.Time            `json:"expiresAt"`
}

type CentralUserResponse struct {
	ID          string               `json:"id"`
	Email       string               `json:"email"`
	DisplayName string               `json:"displayName"`
	Roles       []domain.CentralRole `json:"roles"`
	Active      bool                 `json:"active"`
	CreatedAt   time.Time            `json:"createdAt"`
}

type CentralUsersResponse struct {
	Users []CentralUserResponse `json:"users"`
}

type CentralUserAcceptedResponse struct {
	User CentralUserResponse `json:"user"`
}

type CreateCentralUserRequest struct {
	UserID      string               `json:"userId"`
	Email       string               `json:"email"`
	DisplayName string               `json:"displayName,omitempty"`
	Password    string               `json:"password"`
	Roles       []domain.CentralRole `json:"roles"`
}

type UpdateCentralUserRequest struct {
	DisplayName *string              `json:"displayName,omitempty"`
	Password    *string              `json:"password,omitempty"`
	Roles       []domain.CentralRole `json:"roles,omitempty"`
	Active      *bool                `json:"active,omitempty"`
}

func mountAuthRoutes(mux *http.ServeMux, spec *httpapi.Spec, services Services) {
	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPost,
		Path:        "/v1/auth/sessions",
		OperationID: "createCentralAuthSession",
		Summary:     "Create central admin session",
		Tags:        []string{"auth"},
		RequestBody: &httpapi.BodySpec{
			Description: "Central session creation command",
			Required:    true,
			Schema:      createCentralSessionRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"201": {Description: "Session created", Schema: centralSessionAcceptedResponseSchema()},
			"400": {Description: "Invalid session command", Schema: httpapi.ProblemSchema()},
			"401": {Description: "Invalid credentials", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		var request CreateCentralSessionRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_auth_command", "Invalid session command", err.Error())
			return
		}
		result, err := services.Auth.CreateSession(r.Context(), app.CreateSessionCommand{
			Email:    request.Email,
			Password: request.Password,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, CentralSessionAcceptedResponse{
			Session: centralSessionResponse(result),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/central/users",
		OperationID: "listCentralUsers",
		Summary:     "List central users",
		Description: sessionProtectedDescription("Returns central admin users."),
		Tags:        []string{"central-users"},
		Responses:   protectedResponseSpecs("200", "Central users", centralUsersResponseSchema()),
	}, RequireSession(services.Auth, app.PermissionUsersManage, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		users, err := services.CentralUsers.ListUsers(r.Context(), session)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, CentralUsersResponse{Users: centralUserResponses(users)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPost,
		Path:        "/v1/central/users",
		OperationID: "createCentralUser",
		Summary:     "Create central user",
		Description: sessionProtectedDescription("Creates a central admin user."),
		Tags:        []string{"central-users"},
		RequestBody: &httpapi.BodySpec{
			Description: "Central user creation command",
			Required:    true,
			Schema:      createCentralUserRequestSchema(),
		},
		Responses: protectedResponseSpecs("201", "Central user created", centralUserAcceptedResponseSchema()),
	}, RequireSession(services.Auth, app.PermissionUsersManage, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		var request CreateCentralUserRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_central_user_command", "Invalid central user command", err.Error())
			return
		}
		user, err := services.CentralUsers.CreateUser(r.Context(), app.CreateCentralUserCommand{
			UserID:      request.UserID,
			Email:       request.Email,
			DisplayName: request.DisplayName,
			Password:    request.Password,
			Roles:       request.Roles,
			Session:     session,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, CentralUserAcceptedResponse{User: centralUserResponse(user)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/central/users/{userId}",
		OperationID: "getCentralUser",
		Summary:     "Get central user",
		Description: sessionProtectedDescription("Returns a central admin user."),
		Tags:        []string{"central-users"},
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Central user", centralUserAcceptedResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"404": {Description: "Central user was not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionUsersManage, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		user, err := services.CentralUsers.GetUser(r.Context(), r.PathValue("userId"), session)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, CentralUserAcceptedResponse{User: centralUserResponse(user)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPatch,
		Path:        "/v1/central/users/{userId}",
		OperationID: "updateCentralUser",
		Summary:     "Update central user",
		Description: sessionProtectedDescription("Updates central user roles, password, or active state."),
		Tags:        []string{"central-users"},
		RequestBody: &httpapi.BodySpec{
			Description: "Central user update command",
			Required:    true,
			Schema:      updateCentralUserRequestSchema(),
		},
		Responses: protectedResponseSpecs("200", "Central user updated", centralUserAcceptedResponseSchema()),
	}, RequireSession(services.Auth, app.PermissionUsersManage, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		var request UpdateCentralUserRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_central_user_command", "Invalid central user command", err.Error())
			return
		}
		user, err := services.CentralUsers.UpdateUser(r.Context(), app.UpdateCentralUserCommand{
			UserID:      r.PathValue("userId"),
			DisplayName: request.DisplayName,
			Password:    request.Password,
			Roles:       request.Roles,
			Active:      request.Active,
			Session:     session,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, CentralUserAcceptedResponse{User: centralUserResponse(user)})
	}))
}

func centralSessionResponse(result app.SessionResult) CentralSessionResponse {
	return CentralSessionResponse{
		Token:     result.Token,
		UserID:    result.UserID,
		Roles:     append([]domain.CentralRole(nil), result.Roles...),
		ExpiresAt: result.ExpiresAt,
	}
}

func centralUserResponse(user domain.CentralUser) CentralUserResponse {
	return CentralUserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Roles:       append([]domain.CentralRole(nil), user.Roles...),
		Active:      user.Active,
		CreatedAt:   user.CreatedAt,
	}
}

func centralUserResponses(users []domain.CentralUser) []CentralUserResponse {
	responses := make([]CentralUserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, centralUserResponse(user))
	}
	return responses
}

func createCentralSessionRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"email":    httpapi.StringSchema(),
		"password": httpapi.StringSchema(),
	}, "email", "password")
}

func centralSessionResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"token":     httpapi.StringSchema(),
		"userId":    httpapi.StringSchema(),
		"roles":     httpapi.ArraySchema(httpapi.StringSchema()),
		"expiresAt": httpapi.DateTimeSchema(),
	}, "token", "userId", "roles", "expiresAt")
}

func centralSessionAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"session": centralSessionResponseSchema(),
	}, "session")
}

func centralUserResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":          httpapi.StringSchema(),
		"email":       httpapi.StringSchema(),
		"displayName": httpapi.StringSchema(),
		"roles":       httpapi.ArraySchema(httpapi.StringSchema()),
		"active":      {"type": "boolean"},
		"createdAt":   httpapi.DateTimeSchema(),
	}, "id", "email", "displayName", "roles", "active", "createdAt")
}

func centralUsersResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"users": httpapi.ArraySchema(centralUserResponseSchema()),
	}, "users")
}

func centralUserAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"user": centralUserResponseSchema(),
	}, "user")
}

func createCentralUserRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"userId":      httpapi.StringSchema(),
		"email":       httpapi.StringSchema(),
		"displayName": httpapi.StringSchema(),
		"password":    httpapi.StringSchema(),
		"roles":       httpapi.ArraySchema(httpapi.StringSchema()),
	}, "userId", "email", "password", "roles")
}

func updateCentralUserRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"displayName": httpapi.StringSchema(),
		"password":    httpapi.StringSchema(),
		"roles":       httpapi.ArraySchema(httpapi.StringSchema()),
		"active":      {"type": "boolean"},
	})
}
