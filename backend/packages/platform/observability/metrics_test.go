package observability_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"mercadia.dev/pos/platform/observability"
)

func TestMetricsHandlerReturns200(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	observability.MetricsHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}
