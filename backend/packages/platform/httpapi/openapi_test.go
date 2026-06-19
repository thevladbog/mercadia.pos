package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAPIIncludesOperationAndIdempotencyHeader(t *testing.T) {
	spec := NewSpec(ServiceInfo{
		Name:    "test",
		Title:   "Test API",
		Version: "0.1.0",
	})

	spec.operations = append(spec.operations, Operation{
		Method:              http.MethodPost,
		Path:                "/v1/things/{thingId}/commands",
		OperationID:         "runThingCommand",
		Summary:             "Run command",
		Tags:                []string{"things"},
		RequiresIdempotency: true,
		Responses: map[string]ResponseSpec{
			"202": {Description: "Accepted"},
		},
	})

	document := spec.OpenAPI()
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal OpenAPI: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode OpenAPI: %v", err)
	}

	paths := decoded["paths"].(map[string]any)
	path := paths["/v1/things/{thingId}/commands"].(map[string]any)
	post := path["post"].(map[string]any)

	if got := post["operationId"]; got != "runThingCommand" {
		t.Fatalf("operationId = %v", got)
	}

	parameters := post["parameters"].([]any)
	names := map[string]bool{}
	for _, parameter := range parameters {
		item := parameter.(map[string]any)
		names[item["name"].(string)] = true
	}

	if !names["thingId"] {
		t.Fatal("expected path parameter thingId")
	}
	if !names["Idempotency-Key"] {
		t.Fatal("expected Idempotency-Key header parameter")
	}
}

func TestOpenAPIUsesVersion31AndRequiresOperationIDs(t *testing.T) {
	spec := NewSpec(ServiceInfo{
		Name:    "test",
		Title:   "Test API",
		Version: "0.1.0",
	})

	spec.operations = []Operation{
		{
			Method:      http.MethodGet,
			Path:        "/v1/alpha",
			OperationID: "getAlpha",
			Summary:     "Alpha",
			Responses:   map[string]ResponseSpec{"200": {Description: "OK"}},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/beta",
			OperationID: "getBeta",
			Summary:     "Beta",
			Responses:   map[string]ResponseSpec{"200": {Description: "OK"}},
		},
	}

	document := spec.OpenAPI()
	if document["openapi"] != "3.1.0" {
		t.Fatalf("openapi version = %v", document["openapi"])
	}

	paths := document["paths"].(map[string]any)
	for path, item := range paths {
		pathItem := item.(map[string]any)
		for method, operation := range pathItem {
			op := operation.(map[string]any)
			if op["operationId"] == "" {
				t.Fatalf("missing operationId on %s %s", method, path)
			}
		}
	}
}

func TestScalarHTMLPinsVersionedCDN(t *testing.T) {
	html := ScalarHTML("Mercadia Test")
	if !strings.Contains(html, "@scalar/api-reference@1.60.0") {
		t.Fatalf("expected pinned Scalar CDN, got: %s", html)
	}
	if !strings.Contains(html, `data-url="/openapi.json"`) {
		t.Fatal("expected openapi.json data-url")
	}
}

func TestMountMetricsRouteReturns200(t *testing.T) {
	mux := http.NewServeMux()
	MountMetricsRoute(mux)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}
