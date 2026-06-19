package api

import (
	"net/http"
	"time"

	"mercadia.dev/pos/platform/httpapi"
	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
)

type TerminalOverviewResponse struct {
	ID                       string                `json:"id"`
	StoreID                  string                `json:"storeId"`
	Kind                     domain.TerminalKind   `json:"kind"`
	Status                   domain.TerminalStatus `json:"status"`
	SoftwareVersion          string                `json:"softwareVersion,omitempty"`
	LastSeenAt               time.Time             `json:"lastSeenAt"`
	UpdatedAt                time.Time             `json:"updatedAt"`
	ShiftID                  string                `json:"shiftId,omitempty"`
	CashierID                string                `json:"cashierId,omitempty"`
	DrawerID                 string                `json:"drawerId,omitempty"`
	ReceiptCount             int                   `json:"receiptCount"`
	RevenueMinor             int64                 `json:"revenueMinor"`
	DrawerBalanceMinor       int64                 `json:"drawerBalanceMinor"`
	CurrentReceiptID         string                `json:"currentReceiptId,omitempty"`
	CurrentReceiptStatus     domain.ReceiptStatus  `json:"currentReceiptStatus,omitempty"`
	CurrentReceiptTotalMinor int64                 `json:"currentReceiptTotalMinor,omitempty"`
	AttentionNeeded          bool                  `json:"attentionNeeded"`
}

type PaginatedTerminalOverviewResponse struct {
	Items      []TerminalOverviewResponse `json:"items"`
	TotalCount int                        `json:"totalCount"`
}

type StoreMonitoringSummaryResponse struct {
	RevenueMinorToday      int64 `json:"revenueMinorToday"`
	DrawerCashMinor        int64 `json:"drawerCashMinor"`
	ActiveTerminalCount    int   `json:"activeTerminalCount"`
	FreeTerminalCount      int   `json:"freeTerminalCount"`
	OfflineTerminalCount   int   `json:"offlineTerminalCount"`
	AttentionTerminalCount int   `json:"attentionTerminalCount"`
	ReceiptCountToday      int   `json:"receiptCountToday"`
	AverageReceiptMinor    int64 `json:"averageReceiptMinor"`
}

func mountMonitoringRoutes(mux *http.ServeMux, spec *httpapi.Spec, monitoring *app.TerminalMonitoringService) {
	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/monitoring/terminals",
		OperationID:     "listStoreMonitoringTerminals",
		Summary:         "List terminal monitoring cards for a store",
		Tags:            []string{"monitoring"},
		QueryParameters: paginationQueryParams(),
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Terminal monitoring cards", Schema: paginatedTerminalOverviewResponseSchema()},
			"400": {Description: "Invalid monitoring query", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := monitoring.ListTerminalOverviews(r.Context(), r.PathValue("storeId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		items := make([]TerminalOverviewResponse, 0, len(result.Items))
		for _, overview := range result.Items {
			items = append(items, terminalOverviewResponse(overview))
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedTerminalOverviewResponse{
			Items:      items,
			TotalCount: result.TotalCount,
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/monitoring/summary",
		OperationID: "getStoreMonitoringSummary",
		Summary:     "Get store monitoring summary",
		Tags:        []string{"monitoring"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Store monitoring summary", Schema: storeMonitoringSummaryResponseSchema()},
			"400": {Description: "Invalid monitoring query", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		summary, err := monitoring.GetStoreMonitoringSummary(r.Context(), r.PathValue("storeId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, storeMonitoringSummaryResponse(summary))
	})
}

func terminalOverviewResponse(overview app.TerminalOverview) TerminalOverviewResponse {
	response := TerminalOverviewResponse{
		ID:                 overview.Terminal.ID,
		StoreID:            overview.Terminal.StoreID,
		Kind:               overview.Terminal.Kind,
		Status:             overview.Terminal.Status,
		SoftwareVersion:    overview.Terminal.SoftwareVersion,
		LastSeenAt:         overview.Terminal.LastSeenAt,
		UpdatedAt:          overview.Terminal.UpdatedAt,
		ShiftID:            overview.ShiftID,
		CashierID:          overview.CashierID,
		DrawerID:           overview.DrawerID,
		ReceiptCount:       overview.ReceiptCount,
		RevenueMinor:       overview.RevenueMinor,
		DrawerBalanceMinor: overview.DrawerBalanceMinor,
		AttentionNeeded:    overview.AttentionNeeded,
	}
	if overview.CurrentReceiptID != "" {
		response.CurrentReceiptID = overview.CurrentReceiptID
		response.CurrentReceiptStatus = overview.CurrentReceiptStatus
		response.CurrentReceiptTotalMinor = overview.CurrentReceiptTotalMinor
	}
	return response
}

func storeMonitoringSummaryResponse(summary app.StoreMonitoringSummary) StoreMonitoringSummaryResponse {
	return StoreMonitoringSummaryResponse{
		RevenueMinorToday:      summary.RevenueMinorToday,
		DrawerCashMinor:        summary.DrawerCashMinor,
		ActiveTerminalCount:    summary.ActiveTerminalCount,
		FreeTerminalCount:      summary.FreeTerminalCount,
		OfflineTerminalCount:   summary.OfflineTerminalCount,
		AttentionTerminalCount: summary.AttentionTerminalCount,
		ReceiptCountToday:      summary.ReceiptCountToday,
		AverageReceiptMinor:    summary.AverageReceiptMinor,
	}
}

func paginatedTerminalOverviewResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(terminalOverviewResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func terminalOverviewResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":                       httpapi.StringSchema(),
		"storeId":                  httpapi.StringSchema(),
		"kind":                     httpapi.StringSchema(),
		"status":                   httpapi.StringSchema(),
		"softwareVersion":          httpapi.StringSchema(),
		"lastSeenAt":               httpapi.DateTimeSchema(),
		"updatedAt":                httpapi.DateTimeSchema(),
		"shiftId":                  httpapi.StringSchema(),
		"cashierId":                httpapi.StringSchema(),
		"drawerId":                 httpapi.StringSchema(),
		"receiptCount":             {"type": "integer"},
		"revenueMinor":             {"type": "integer"},
		"drawerBalanceMinor":       {"type": "integer"},
		"currentReceiptId":         httpapi.StringSchema(),
		"currentReceiptStatus":     httpapi.StringSchema(),
		"currentReceiptTotalMinor": {"type": "integer"},
		"attentionNeeded":          {"type": "boolean"},
	}, "id", "storeId", "kind", "status", "lastSeenAt", "updatedAt", "receiptCount", "revenueMinor", "drawerBalanceMinor", "attentionNeeded")
}

func storeMonitoringSummaryResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"revenueMinorToday":      {"type": "integer"},
		"drawerCashMinor":        {"type": "integer"},
		"activeTerminalCount":    {"type": "integer"},
		"freeTerminalCount":      {"type": "integer"},
		"offlineTerminalCount":   {"type": "integer"},
		"attentionTerminalCount": {"type": "integer"},
		"receiptCountToday":      {"type": "integer"},
		"averageReceiptMinor":    {"type": "integer"},
	}, "revenueMinorToday", "drawerCashMinor", "activeTerminalCount", "freeTerminalCount", "offlineTerminalCount", "attentionTerminalCount", "receiptCountToday", "averageReceiptMinor")
}
