package httpapi

import (
	"encoding/json"
	"net/http"
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
