package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type ServiceInfo struct {
	Name        string
	Title       string
	Description string
	Version     string
}

type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
	Code   string `json:"code"`
}

type HealthResponse struct {
	Service string    `json:"service"`
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Time    time.Time `json:"time"`
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteProblem(w http.ResponseWriter, status int, code string, title string, detail string) {
	WriteJSON(w, status, Problem{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
		Code:   code,
	})
}

func DecodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func RequireIdempotencyKey(r *http.Request) (string, error) {
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		return "", errors.New("missing Idempotency-Key header")
	}
	return key, nil
}

func MountSystemRoutes(mux *http.ServeMux, spec *Spec, info ServiceInfo) {
	Register(mux, spec, Operation{
		Method:      http.MethodGet,
		Path:        "/healthz",
		OperationID: "getHealth",
		Summary:     "Get service health",
		Tags:        []string{"system"},
		Responses: map[string]ResponseSpec{
			"200": {Description: "Service is alive", Schema: HealthResponseSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, HealthResponse{
			Service: info.Name,
			Status:  "ok",
			Version: info.Version,
			Time:    time.Now().UTC(),
		})
	})

	Register(mux, spec, Operation{
		Method:      http.MethodGet,
		Path:        "/readyz",
		OperationID: "getReadiness",
		Summary:     "Get service readiness",
		Tags:        []string{"system"},
		Responses: map[string]ResponseSpec{
			"200": {Description: "Service is ready", Schema: HealthResponseSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, HealthResponse{
			Service: info.Name,
			Status:  "ready",
			Version: info.Version,
			Time:    time.Now().UTC(),
		})
	})

	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, spec.OpenAPI())
	})

	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(ScalarHTML(info.Title)))
	})
}
