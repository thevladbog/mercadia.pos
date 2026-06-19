package api

import (
	"net/http"
	"time"

	"mercadia.dev/pos/platform/httpapi"
)

const version = "0.1.0"

type StatusResponse struct {
	Region      string    `json:"region"`
	Status      string    `json:"status"`
	GeneratedAt time.Time `json:"generatedAt"`
}

type SyncEventsAcceptedResponse struct {
	StoreID  string `json:"storeId"`
	Status   string `json:"status"`
	Accepted int    `json:"accepted"`
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
		Name:        "central-backend",
		Title:       "Mercadia Central Backend",
		Description: "Central API for global administration, cross-store reporting, integrations, and Store Edge synchronization.",
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
		Path:        "/v1/central/status",
		OperationID: "getCentralStatus",
		Summary:     "Get central backend status",
		Tags:        []string{"system"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Central backend status", Schema: statusResponseSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		httpapi.WriteJSON(w, http.StatusOK, StatusResponse{
			Region:      "default",
			Status:      "ok",
			GeneratedAt: time.Now().UTC(),
		})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores/{storeId}/sync-events",
		OperationID:         "acceptStoreSyncEvents",
		Summary:             "Accept synchronized Store Edge events",
		Tags:                []string{"sync"},
		RequiresIdempotency: true,
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Sync batch accepted", Schema: syncEventsAcceptedResponseSchema()},
			"400": {Description: "Invalid sync batch", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, SyncEventsAcceptedResponse{
			StoreID:  r.PathValue("storeId"),
			Status:   "accepted",
			Accepted: 0,
		})
	})
}

func statusResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"region":      httpapi.StringSchema(),
		"status":      httpapi.StringSchema(),
		"generatedAt": httpapi.DateTimeSchema(),
	}, "region", "status", "generatedAt")
}

func syncEventsAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":  httpapi.StringSchema(),
		"status":   httpapi.StringSchema(),
		"accepted": {"type": "integer"},
	}, "storeId", "status", "accepted")
}
