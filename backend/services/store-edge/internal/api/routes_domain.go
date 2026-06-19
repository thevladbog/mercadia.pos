package api

import (
	"net/http"
	"time"

	"mercadia.dev/pos/platform/httpapi"
	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
)

const sessionTokenHeader = "X-Session-Token"

type SessionContext struct {
	Token     string
	ActorID   string
	Roles     []domain.Role
	ExpiresAt time.Time
}

func OptionalSessionFromRequest(r *http.Request, auth *app.AuthService) (*SessionContext, error) {
	token := r.Header.Get(sessionTokenHeader)
	if token == "" {
		return nil, nil
	}
	session, err := auth.ResolveSession(r.Context(), token)
	if err != nil {
		return nil, err
	}
	return &SessionContext{
		Token:     session.Token,
		ActorID:   session.ActorID,
		Roles:     session.Roles,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

type CreateSessionRequest struct {
	ActorID string `json:"actorId"`
	PIN     string `json:"pin"`
}

type SessionAcceptedResponse struct {
	Session SessionResponse `json:"session"`
}

type SessionResponse struct {
	Token     string         `json:"token"`
	ActorID   string         `json:"actorId"`
	Roles     []domain.Role  `json:"roles"`
	ExpiresAt time.Time      `json:"expiresAt"`
}

type PaginatedReceiptsResponse struct {
	Items      []ReceiptResponse `json:"items"`
	TotalCount int               `json:"totalCount"`
}

type PaginatedShiftsResponse struct {
	Items      []ShiftResponse `json:"items"`
	TotalCount int             `json:"totalCount"`
}

type PaginatedCashMovementsResponse struct {
	Items      []CashMovementResponse `json:"items"`
	TotalCount int                    `json:"totalCount"`
}

type PaginatedCashRecountsResponse struct {
	Items      []CashRecountResponse `json:"items"`
	TotalCount int                   `json:"totalCount"`
}

type PaginatedOperationJournalResponse struct {
	Items      []OperationJournalEntryResponse `json:"items"`
	TotalCount int                             `json:"totalCount"`
}

type OperationJournalEntryResponse struct {
	ID            string    `json:"id"`
	StoreID       string    `json:"storeId"`
	OperationType string    `json:"operationType"`
	ActorID       string    `json:"actorId"`
	ReferenceID   string    `json:"referenceId,omitempty"`
	Summary       string    `json:"summary,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type CreateReceiptReturnRequest struct {
	Lines   []ReturnLineRequest `json:"lines"`
	Reason  string              `json:"reason"`
	ActorID string              `json:"actorId"`
}

type CreateNoReceiptReturnRequest struct {
	Lines        []ReturnLineRequest `json:"lines"`
	Reason       string              `json:"reason"`
	ActorID      string              `json:"actorId"`
	ApprovedByID string              `json:"approvedById"`
}

type ReturnLineRequest struct {
	LineID         string `json:"lineId,omitempty"`
	ProductID      string `json:"productId,omitempty"`
	Name           string `json:"name,omitempty"`
	Quantity       int64  `json:"quantity"`
	UnitPriceMinor int64  `json:"unitPriceMinor,omitempty"`
}

type ReturnAcceptedResponse struct {
	Return ReturnResponse `json:"return"`
}

type PaginatedReturnsResponse struct {
	Items      []ReturnResponse `json:"items"`
	TotalCount int              `json:"totalCount"`
}

type SettleReturnRequest struct {
	ActorID  string `json:"actorId,omitempty"`
	Reason   string `json:"reason,omitempty"`
	DrawerID string `json:"drawerId,omitempty"`
}

type ReturnSettledResponse struct {
	Return   ReturnResponse    `json:"return"`
	Payments []PaymentResponse `json:"payments"`
}

type ReturnResponse struct {
	ID           string               `json:"id"`
	StoreID      string               `json:"storeId"`
	ReceiptID    string               `json:"receiptId,omitempty"`
	Kind         domain.ReturnKind    `json:"kind"`
	Lines        []ReturnLineResponse `json:"lines"`
	Reason       string               `json:"reason"`
	ActorID      string               `json:"actorId"`
	ApprovedByID string               `json:"approvedById,omitempty"`
	TotalMinor   int64                `json:"totalMinor"`
	Status       domain.ReturnStatus  `json:"status"`
	CreatedAt    time.Time            `json:"createdAt"`
}

type ReturnLineResponse struct {
	LineID         string `json:"lineId,omitempty"`
	ProductID      string `json:"productId,omitempty"`
	Name           string `json:"name"`
	Quantity       int64  `json:"quantity"`
	UnitPriceMinor int64  `json:"unitPriceMinor"`
	TotalMinor     int64  `json:"totalMinor"`
}

type ApplyLineDiscountRequest struct {
	AmountMinor int64  `json:"amountMinor"`
	Reason      string `json:"reason"`
	ActorID     string `json:"actorId"`
}

type ValidateMarkingRequest struct {
	Code string `json:"code"`
}

type MarkingValidationResponse struct {
	Valid     bool   `json:"valid"`
	Code      string `json:"code"`
	ProductID string `json:"productId,omitempty"`
	Message   string `json:"message,omitempty"`
}

func paginationQueryParams() []httpapi.QueryParamSpec {
	return []httpapi.QueryParamSpec{
		{Name: "limit", Description: "Maximum number of items to return", Schema: httpapi.Schema{"type": "integer", "minimum": 1, "maximum": app.MaxPageLimit}},
		{Name: "offset", Description: "Number of items to skip", Schema: httpapi.Schema{"type": "integer", "minimum": 0}},
	}
}

func mountDomainRoutes(
	mux *http.ServeMux,
	spec *httpapi.Spec,
	auth *app.AuthService,
	returns *app.ReturnsService,
	returnSettlement *app.ReturnSettlementService,
	fiscalization *app.FiscalizationService,
	discounts *app.DiscountService,
	marking *app.MarkingService,
	journal *app.OperationJournalService,
) {
	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPost,
		Path:        "/v1/auth/sessions",
		OperationID: "createAuthSession",
		Summary:     "Create cashier session",
		Tags:        []string{"auth"},
		RequestBody: &httpapi.BodySpec{
			Description: "Session creation command",
			Required:    true,
			Schema:      createSessionRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"201": {Description: "Session created", Schema: sessionAcceptedResponseSchema()},
			"400": {Description: "Invalid session command", Schema: httpapi.ProblemSchema()},
			"401": {Description: "Invalid credentials", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		var request CreateSessionRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := auth.CreateSession(r.Context(), app.CreateSessionCommand{
			ActorID: request.ActorID,
			PIN:     request.PIN,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, SessionAcceptedResponse{
			Session: sessionResponse(result),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/returns/{returnId}",
		OperationID:     "getReturn",
		Summary:         "Get return",
		Tags:            []string{"returns"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Return", Schema: returnAcceptedResponseSchema()},
			"404": {Description: "Return was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := returns.GetReturn(r.Context(), r.PathValue("returnId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, ReturnAcceptedResponse{Return: returnResponse(result.Return)})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/receipts/{receiptId}/returns",
		OperationID:     "listReceiptReturns",
		Summary:         "List returns for receipt",
		Tags:            []string{"returns"},
		QueryParameters: paginationQueryParams(),
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Returns for receipt", Schema: paginatedReturnsResponseSchema()},
			"404": {Description: "Receipt was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := returns.ListReturnsByReceipt(r.Context(), r.PathValue("receiptId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedReturnsResponse{
			Items:      returnResponses(result.Items),
			TotalCount: result.TotalCount,
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/receipts/{receiptId}/returns",
		OperationID:         "createReceiptReturn",
		Summary:             "Create return against fiscalized receipt",
		Tags:                []string{"returns"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Receipt return command",
			Required:    true,
			Schema:      createReceiptReturnRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Return accepted", Schema: returnAcceptedResponseSchema()},
			"400": {Description: "Invalid return command", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Receipt was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Return or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request CreateReceiptReturnRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := returns.CreateReceiptReturn(r.Context(), app.CreateReceiptReturnCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			ReceiptID:      r.PathValue("receiptId"),
			Lines:          toReturnLineCommands(request.Lines),
			Reason:         request.Reason,
			ActorID:        request.ActorID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, ReturnAcceptedResponse{Return: returnResponse(result.Return)})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/returns/{returnId}/settle",
		OperationID:         "settleReturn",
		Summary:             "Settle return by refunding payments or disbursing cash for no-receipt returns",
		Tags:                []string{"returns"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Return settlement command",
			Required:    false,
			Schema:      settleReturnRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Return settled", Schema: returnSettledResponseSchema()},
			"400": {Description: "Invalid settlement command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Return was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Return settlement conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request SettleReturnRequest
		if r.ContentLength > 0 {
			if err := httpapi.DecodeJSON(r, &request); err != nil {
				httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
				return
			}
		}
		result, err := returnSettlement.SettleReturn(r.Context(), app.SettleReturnCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			ReturnID:       r.PathValue("returnId"),
			ActorID:        request.ActorID,
			Reason:         request.Reason,
			DrawerID:       request.DrawerID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, ReturnSettledResponse{
			Return:   returnResponse(result.Return),
			Payments: paymentResponses(result.Payments),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/returns/{returnId}/fiscal-documents",
		OperationID:         "createReturnFiscalDocument",
		Summary:             "Create fiscal return document for settled with-receipt return",
		Tags:                []string{"fiscalization", "returns"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Return fiscalization command",
			Required:    true,
			Schema:      createFiscalDocumentRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Return fiscal document command accepted", Schema: fiscalDocumentAcceptedResponseSchema()},
			"400": {Description: "Invalid fiscalization command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Return was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Return fiscalization or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request CreateFiscalDocumentRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := fiscalization.CreateReturnFiscalDocument(r.Context(), app.CreateReturnFiscalDocumentCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			ReturnID:       r.PathValue("returnId"),
			DeviceID:       request.DeviceID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, FiscalDocumentAcceptedResponse{
			Document: fiscalDocumentResponse(result.Document),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/returns/{returnId}/fiscal-documents",
		OperationID: "listReturnFiscalDocuments",
		Summary:     "List fiscal documents for return",
		Tags:        []string{"fiscalization", "returns"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Return fiscal documents", Schema: fiscalDocumentsResponseSchema()},
			"404": {Description: "Return was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := fiscalization.ListReturnFiscalDocuments(r.Context(), r.PathValue("returnId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, FiscalDocumentsResponse{
			Documents: fiscalDocumentResponses(result),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores/{storeId}/returns/no-receipt",
		OperationID:         "createNoReceiptReturn",
		Summary:             "Create no-receipt return with approval",
		Tags:                []string{"returns"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "No-receipt return command",
			Required:    true,
			Schema:      createNoReceiptReturnRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Return accepted", Schema: returnAcceptedResponseSchema()},
			"400": {Description: "Invalid return command", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Return or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request CreateNoReceiptReturnRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := returns.CreateNoReceiptReturn(r.Context(), app.CreateNoReceiptReturnCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			StoreID:        r.PathValue("storeId"),
			Lines:          toReturnLineCommands(request.Lines),
			Reason:         request.Reason,
			ActorID:        request.ActorID,
			ApprovedByID:   request.ApprovedByID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, ReturnAcceptedResponse{Return: returnResponse(result.Return)})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/receipts/{receiptId}/lines/{lineId}/discount",
		OperationID:         "applyReceiptLineDiscount",
		Summary:             "Apply line discount with permission and reason",
		Tags:                []string{"checkout", "discounts"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Line discount command",
			Required:    true,
			Schema:      applyLineDiscountRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Discount applied", Schema: receiptAcceptedResponseSchema()},
			"400": {Description: "Invalid discount command", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Receipt was not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Discount or idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request ApplyLineDiscountRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := discounts.ApplyLineDiscount(r.Context(), app.ApplyLineDiscountCommand{
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			ReceiptID:      r.PathValue("receiptId"),
			LineID:         r.PathValue("lineId"),
			AmountMinor:    request.AmountMinor,
			Reason:         request.Reason,
			ActorID:        request.ActorID,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, ReceiptAcceptedResponse{Receipt: receiptResponse(result.Receipt)})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPost,
		Path:        "/v1/receipts/{receiptId}/marking/validate",
		OperationID: "validateReceiptMarking",
		Summary:     "Validate DataMatrix marking code",
		Tags:        []string{"marking"},
		RequestBody: &httpapi.BodySpec{
			Description: "Marking validation command",
			Required:    true,
			Schema:      validateMarkingRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Marking validation result", Schema: markingValidationResponseSchema()},
			"400": {Description: "Invalid marking command", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Receipt was not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		var request ValidateMarkingRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_json", "Invalid JSON", err.Error())
			return
		}
		result, err := marking.ValidateMarking(r.Context(), app.ValidateMarkingCommand{
			ReceiptID: r.PathValue("receiptId"),
			Code:      request.Code,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, MarkingValidationResponse{
			Valid:     result.Validation.Valid,
			Code:      result.Validation.Code,
			ProductID: result.Validation.ProductID,
			Message:   result.Validation.Message,
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/operation-journal",
		OperationID:     "listOperationJournal",
		Summary:         "List store operation journal",
		Tags:            []string{"store-operations"},
		QueryParameters: paginationQueryParams(),
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Operation journal entries", Schema: paginatedOperationJournalResponseSchema()},
			"400": {Description: "Invalid journal query", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := journal.ListOperationJournal(r.Context(), r.PathValue("storeId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedOperationJournalResponse{
			Items:      operationJournalEntryResponses(result.Items),
			TotalCount: result.TotalCount,
		})
	})
}

func toReturnLineCommands(lines []ReturnLineRequest) []app.ReturnLineCommand {
	result := make([]app.ReturnLineCommand, 0, len(lines))
	for _, line := range lines {
		result = append(result, app.ReturnLineCommand{
			LineID:         line.LineID,
			ProductID:      line.ProductID,
			Name:           line.Name,
			Quantity:       line.Quantity,
			UnitPriceMinor: line.UnitPriceMinor,
		})
	}
	return result
}

func sessionResponse(result app.SessionResult) SessionResponse {
	return SessionResponse{
		Token:     result.Token,
		ActorID:   result.ActorID,
		Roles:     result.Roles,
		ExpiresAt: result.ExpiresAt,
	}
}

func returnResponse(ret domain.Return) ReturnResponse {
	lines := make([]ReturnLineResponse, 0, len(ret.Lines))
	for _, line := range ret.Lines {
		lines = append(lines, ReturnLineResponse{
			LineID:         line.LineID,
			ProductID:      line.ProductID,
			Name:           line.Name,
			Quantity:       line.Quantity,
			UnitPriceMinor: line.UnitPriceMinor,
			TotalMinor:     line.TotalMinor,
		})
	}
	return ReturnResponse{
		ID:           ret.ID,
		StoreID:      ret.StoreID,
		ReceiptID:    ret.ReceiptID,
		Kind:         ret.Kind,
		Lines:        lines,
		Reason:       ret.Reason,
		ActorID:      ret.ActorID,
		ApprovedByID: ret.ApprovedByID,
		TotalMinor:   ret.TotalMinor,
		Status:       ret.Status,
		CreatedAt:    ret.CreatedAt,
	}
}

func returnResponses(returns []domain.Return) []ReturnResponse {
	result := make([]ReturnResponse, 0, len(returns))
	for _, ret := range returns {
		result = append(result, returnResponse(ret))
	}
	return result
}

func operationJournalEntryResponses(entries []domain.OperationJournalEntry) []OperationJournalEntryResponse {
	result := make([]OperationJournalEntryResponse, 0, len(entries))
	for _, entry := range entries {
		result = append(result, OperationJournalEntryResponse{
			ID:            entry.ID,
			StoreID:       entry.StoreID,
			OperationType: entry.OperationType,
			ActorID:       entry.ActorID,
			ReferenceID:   entry.ReferenceID,
			Summary:       entry.Summary,
			CreatedAt:     entry.CreatedAt,
		})
	}
	return result
}

func createSessionRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"actorId": httpapi.StringSchema(),
		"pin":     httpapi.StringSchema(),
	}, "actorId", "pin")
}

func sessionAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"session": sessionResponseSchema(),
	}, "session")
}

func sessionResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"token":     httpapi.StringSchema(),
		"actorId":   httpapi.StringSchema(),
		"roles":     httpapi.ArraySchema(httpapi.StringSchema()),
		"expiresAt": httpapi.DateTimeSchema(),
	}, "token", "actorId", "roles", "expiresAt")
}

func createReceiptReturnRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"lines":   httpapi.ArraySchema(returnLineRequestSchema()),
		"reason":  httpapi.StringSchema(),
		"actorId": httpapi.StringSchema(),
	}, "lines", "reason", "actorId")
}

func createNoReceiptReturnRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"lines":        httpapi.ArraySchema(returnLineRequestSchema()),
		"reason":       httpapi.StringSchema(),
		"actorId":      httpapi.StringSchema(),
		"approvedById": httpapi.StringSchema(),
	}, "lines", "reason", "actorId", "approvedById")
}

func returnLineRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"lineId":         httpapi.StringSchema(),
		"productId":      httpapi.StringSchema(),
		"name":           httpapi.StringSchema(),
		"quantity":       {"type": "integer", "minimum": 1},
		"unitPriceMinor": {"type": "integer", "minimum": 0},
	}, "quantity")
}

func returnAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"return": returnResponseSchema(),
	}, "return")
}

func paginatedReturnsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(returnResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func settleReturnRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"actorId":  httpapi.StringSchema(),
		"reason":   httpapi.StringSchema(),
		"drawerId": httpapi.StringSchema(),
	})
}

func returnSettledResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"return":   returnResponseSchema(),
		"payments": httpapi.ArraySchema(paymentResponseSchema()),
	}, "return", "payments")
}

func returnResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":           httpapi.StringSchema(),
		"storeId":      httpapi.StringSchema(),
		"receiptId":    httpapi.StringSchema(),
		"kind":         httpapi.StringSchema(),
		"lines":        httpapi.ArraySchema(returnLineResponseSchema()),
		"reason":       httpapi.StringSchema(),
		"actorId":      httpapi.StringSchema(),
		"approvedById": httpapi.StringSchema(),
		"totalMinor":   {"type": "integer"},
		"status":       httpapi.StringSchema(),
		"createdAt":    httpapi.DateTimeSchema(),
	}, "id", "storeId", "kind", "lines", "reason", "actorId", "totalMinor", "status", "createdAt")
}

func returnLineResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"lineId":         httpapi.StringSchema(),
		"productId":      httpapi.StringSchema(),
		"name":           httpapi.StringSchema(),
		"quantity":       {"type": "integer"},
		"unitPriceMinor": {"type": "integer"},
		"totalMinor":     {"type": "integer"},
	}, "name", "quantity", "unitPriceMinor", "totalMinor")
}

func applyLineDiscountRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"amountMinor": {"type": "integer", "minimum": 1},
		"reason":      httpapi.StringSchema(),
		"actorId":     httpapi.StringSchema(),
	}, "amountMinor", "reason", "actorId")
}

func validateMarkingRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"code": httpapi.StringSchema(),
	}, "code")
}

func markingValidationResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"valid":     {"type": "boolean"},
		"code":      httpapi.StringSchema(),
		"productId": httpapi.StringSchema(),
		"message":   httpapi.StringSchema(),
	}, "valid", "code")
}

func paginatedOperationJournalResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(operationJournalEntryResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func operationJournalEntryResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":            httpapi.StringSchema(),
		"storeId":       httpapi.StringSchema(),
		"operationType": httpapi.StringSchema(),
		"actorId":       httpapi.StringSchema(),
		"referenceId":   httpapi.StringSchema(),
		"summary":       httpapi.StringSchema(),
		"createdAt":     httpapi.DateTimeSchema(),
	}, "id", "storeId", "operationType", "actorId", "createdAt")
}

func paginatedReceiptsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(receiptResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func paginatedShiftsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(shiftResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func paginatedCashMovementsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(cashMovementResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func paginatedCashRecountsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(cashRecountResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}
