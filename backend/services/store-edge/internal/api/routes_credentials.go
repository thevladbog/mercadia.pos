package api

import (
	"net/http"

	"mercadia.dev/pos/platform/httpapi"
	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
)

type CredentialManagementResponse struct {
	StoreID     string                    `json:"storeId"`
	StorePolicy CredentialPolicyResponse  `json:"storePolicy"`
	Actors      []ActorCredentialResponse `json:"actors"`
}

type ActorCredentialResponse struct {
	ID                 string                      `json:"id"`
	Roles              []domain.Role               `json:"roles"`
	CredentialPolicy   *CredentialPolicyResponse   `json:"credentialPolicy,omitempty"`
	CredentialBindings []CredentialBindingResponse `json:"credentialBindings"`
}

type CredentialPolicyResponse struct {
	Required     bool                    `json:"required"`
	AllowedKinds []domain.CredentialKind `json:"allowedKinds"`
}

type CredentialBindingResponse struct {
	Kind             domain.CredentialKind `json:"kind"`
	TokenFingerprint string                `json:"tokenFingerprint"`
	MaskedToken      string                `json:"maskedToken"`
	Active           bool                  `json:"active"`
}

type SetCredentialPolicyRequest struct {
	ActorID      string                  `json:"actorId"`
	Required     bool                    `json:"required"`
	AllowedKinds []domain.CredentialKind `json:"allowedKinds"`
}

type SetActorCredentialPolicyRequest struct {
	ActorID            string                  `json:"actorId"`
	InheritStorePolicy bool                    `json:"inheritStorePolicy"`
	Required           bool                    `json:"required"`
	AllowedKinds       []domain.CredentialKind `json:"allowedKinds"`
}

type AddCredentialBindingRequest struct {
	ActorID     string                `json:"actorId"`
	Kind        domain.CredentialKind `json:"kind"`
	Token       string                `json:"token"`
	MaskedToken string                `json:"maskedToken,omitempty"`
}

type RevokeCredentialBindingRequest struct {
	ActorID          string                `json:"actorId"`
	Kind             domain.CredentialKind `json:"kind"`
	TokenFingerprint string                `json:"tokenFingerprint"`
}

type CredentialPolicyAcceptedResponse struct {
	Policy CredentialPolicyResponse `json:"policy"`
}

type ActorCredentialAcceptedResponse struct {
	Actor ActorCredentialResponse `json:"actor"`
}

func mountCredentialRoutes(mux *http.ServeMux, spec *httpapi.Spec, credentials *app.CredentialManagementService) {
	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/credential-management",
		OperationID: "getCredentialManagement",
		Summary:     "Get staff credential policies and bindings",
		Tags:        []string{"auth"},
		QueryParameters: []httpapi.QueryParamSpec{
			{Name: "actorId", Description: "Actor requesting credential management state", Required: true, Schema: httpapi.StringSchema()},
		},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Credential management state", Schema: credentialManagementResponseSchema()},
			"400": {Description: "Invalid credential management query", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := credentials.GetCredentialManagement(r.Context(), app.GetCredentialManagementQuery{
			StoreID:   r.PathValue("storeId"),
			ManagerID: r.URL.Query().Get("actorId"),
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, credentialManagementResponse(result))
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPut,
		Path:                "/v1/stores/{storeId}/credential-policy",
		OperationID:         "setStoreCredentialPolicy",
		Summary:             "Set store staff credential policy",
		Tags:                []string{"auth"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Store credential policy command",
			Required:    true,
			Schema:      setCredentialPolicyRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Store credential policy", Schema: credentialPolicyAcceptedResponseSchema()},
			"400": {Description: "Invalid credential policy command", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request SetCredentialPolicyRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := credentials.SetStoreCredentialPolicy(r.Context(), app.SetStoreCredentialPolicyCommand{
			StoreID:      r.PathValue("storeId"),
			ManagerID:    request.ActorID,
			Required:     request.Required,
			AllowedKinds: request.AllowedKinds,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, CredentialPolicyAcceptedResponse{Policy: credentialPolicyResponse(result)})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPut,
		Path:                "/v1/stores/{storeId}/actors/{actorId}/credential-policy",
		OperationID:         "setActorCredentialPolicy",
		Summary:             "Set actor staff credential policy override",
		Tags:                []string{"auth"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Actor credential policy command",
			Required:    true,
			Schema:      setActorCredentialPolicyRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Actor credential policy", Schema: actorCredentialAcceptedResponseSchema()},
			"400": {Description: "Invalid credential policy command", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Actor was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Separation of duties conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request SetActorCredentialPolicyRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := credentials.SetActorCredentialPolicy(r.Context(), app.SetActorCredentialPolicyCommand{
			TargetActorID:      r.PathValue("actorId"),
			ManagerID:          request.ActorID,
			InheritStorePolicy: request.InheritStorePolicy,
			Required:           request.Required,
			AllowedKinds:       request.AllowedKinds,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, ActorCredentialAcceptedResponse{Actor: actorCredentialResponse(result)})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores/{storeId}/actors/{actorId}/credential-bindings",
		OperationID:         "addActorCredentialBinding",
		Summary:             "Add actor staff credential binding",
		Tags:                []string{"auth"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Actor credential binding command",
			Required:    true,
			Schema:      addCredentialBindingRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Actor credential bindings", Schema: actorCredentialAcceptedResponseSchema()},
			"400": {Description: "Invalid credential binding command", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Actor was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Separation of duties conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request AddCredentialBindingRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := credentials.AddCredentialBinding(r.Context(), app.AddCredentialBindingCommand{
			TargetActorID: r.PathValue("actorId"),
			ManagerID:     request.ActorID,
			Kind:          request.Kind,
			Token:         request.Token,
			MaskedToken:   request.MaskedToken,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, ActorCredentialAcceptedResponse{Actor: actorCredentialResponse(result)})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores/{storeId}/actors/{actorId}/credential-bindings/revoke",
		OperationID:         "revokeActorCredentialBinding",
		Summary:             "Revoke actor staff credential binding",
		Tags:                []string{"auth"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Actor credential binding revoke command",
			Required:    true,
			Schema:      revokeCredentialBindingRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Actor credential bindings", Schema: actorCredentialAcceptedResponseSchema()},
			"400": {Description: "Invalid credential binding command", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Actor or binding was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Separation of duties conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request RevokeCredentialBindingRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := credentials.RevokeCredentialBinding(r.Context(), app.RevokeCredentialBindingCommand{
			TargetActorID:    r.PathValue("actorId"),
			ManagerID:        request.ActorID,
			Kind:             request.Kind,
			TokenFingerprint: request.TokenFingerprint,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, ActorCredentialAcceptedResponse{Actor: actorCredentialResponse(result)})
	})
}

func credentialManagementResponse(result app.CredentialManagementResult) CredentialManagementResponse {
	return CredentialManagementResponse{
		StoreID:     result.StoreID,
		StorePolicy: credentialPolicyResponse(result.StorePolicy),
		Actors:      actorCredentialResponses(result.Actors),
	}
}

func actorCredentialResponses(actors []app.ActorCredentialResult) []ActorCredentialResponse {
	responses := make([]ActorCredentialResponse, 0, len(actors))
	for _, actor := range actors {
		responses = append(responses, actorCredentialResponse(actor))
	}
	return responses
}

func actorCredentialResponse(actor app.ActorCredentialResult) ActorCredentialResponse {
	var policy *CredentialPolicyResponse
	if actor.CredentialPolicy != nil {
		response := credentialPolicyResponse(*actor.CredentialPolicy)
		policy = &response
	}
	return ActorCredentialResponse{
		ID:                 actor.ID,
		Roles:              actor.Roles,
		CredentialPolicy:   policy,
		CredentialBindings: credentialBindingResponses(actor.CredentialBindings),
	}
}

func credentialPolicyResponse(policy app.CredentialPolicyResult) CredentialPolicyResponse {
	return CredentialPolicyResponse{Required: policy.Required, AllowedKinds: policy.AllowedKinds}
}

func credentialBindingResponses(bindings []app.CredentialBindingResult) []CredentialBindingResponse {
	responses := make([]CredentialBindingResponse, 0, len(bindings))
	for _, binding := range bindings {
		responses = append(responses, CredentialBindingResponse{
			Kind:             binding.Kind,
			TokenFingerprint: binding.TokenFingerprint,
			MaskedToken:      binding.MaskedToken,
			Active:           binding.Active,
		})
	}
	return responses
}

func credentialManagementResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":     httpapi.StringSchema(),
		"storePolicy": credentialPolicyResponseSchema(),
		"actors":      httpapi.ArraySchema(actorCredentialResponseSchema()),
	}, "storeId", "storePolicy", "actors")
}

func actorCredentialResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":                 httpapi.StringSchema(),
		"roles":              httpapi.ArraySchema(httpapi.StringSchema()),
		"credentialPolicy":   credentialPolicyResponseSchema(),
		"credentialBindings": httpapi.ArraySchema(credentialBindingResponseSchema()),
	}, "id", "roles", "credentialBindings")
}

func credentialPolicyResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"required":     {"type": "boolean"},
		"allowedKinds": httpapi.ArraySchema(credentialKindSchema()),
	}, "required", "allowedKinds")
}

func credentialBindingResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"kind":             credentialKindSchema(),
		"tokenFingerprint": httpapi.StringSchema(),
		"maskedToken":      httpapi.StringSchema(),
		"active":           {"type": "boolean"},
	}, "kind", "tokenFingerprint", "active")
}

func setCredentialPolicyRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"actorId":      httpapi.StringSchema(),
		"required":     {"type": "boolean"},
		"allowedKinds": httpapi.ArraySchema(credentialKindSchema()),
	}, "actorId", "required", "allowedKinds")
}

func setActorCredentialPolicyRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"actorId":            httpapi.StringSchema(),
		"inheritStorePolicy": {"type": "boolean"},
		"required":           {"type": "boolean"},
		"allowedKinds":       httpapi.ArraySchema(credentialKindSchema()),
	}, "actorId", "inheritStorePolicy", "required", "allowedKinds")
}

func addCredentialBindingRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"actorId":     httpapi.StringSchema(),
		"kind":        credentialKindSchema(),
		"token":       httpapi.StringSchema(),
		"maskedToken": httpapi.StringSchema(),
	}, "actorId", "kind", "token")
}

func revokeCredentialBindingRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"actorId":          httpapi.StringSchema(),
		"kind":             credentialKindSchema(),
		"tokenFingerprint": httpapi.StringSchema(),
	}, "actorId", "kind", "tokenFingerprint")
}

func credentialPolicyAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"policy": credentialPolicyResponseSchema(),
	}, "policy")
}

func actorCredentialAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"actor": actorCredentialResponseSchema(),
	}, "actor")
}

func credentialKindSchema() httpapi.Schema {
	return httpapi.EnumStringSchema("ibutton", "msr_card", "barcode_card")
}
