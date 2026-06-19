package httpapi

import (
	"net/http"
	"regexp"
	"strings"
)

type Schema map[string]any

type BodySpec struct {
	Description string
	Required    bool
	Schema      Schema
}

type ResponseSpec struct {
	Description string
	Schema      Schema
}

type QueryParamSpec struct {
	Name        string
	Description string
	Required    bool
	Schema      Schema
}

type Operation struct {
	Method              string
	Path                string
	OperationID         string
	Summary             string
	Description         string
	Tags                []string
	QueryParameters     []QueryParamSpec
	RequestBody         *BodySpec
	Responses           map[string]ResponseSpec
	RequiresIdempotency bool
}

type Spec struct {
	info       ServiceInfo
	operations []Operation
}

func NewSpec(info ServiceInfo) *Spec {
	return &Spec{info: info}
}

func Register(mux *http.ServeMux, spec *Spec, op Operation, handler http.HandlerFunc) {
	spec.operations = append(spec.operations, op)
	mux.HandleFunc(op.Method+" "+op.Path, handler)
}

func (s *Spec) OpenAPI() map[string]any {
	paths := map[string]any{}

	for _, op := range s.operations {
		pathItem, _ := paths[op.Path].(map[string]any)
		if pathItem == nil {
			pathItem = map[string]any{}
			paths[op.Path] = pathItem
		}

		operation := map[string]any{
			"operationId": op.OperationID,
			"summary":     op.Summary,
			"tags":        op.Tags,
			"responses":   responsesToOpenAPI(op.Responses),
		}
		if op.Description != "" {
			operation["description"] = op.Description
		}

		parameters := pathParameters(op.Path)
		if op.RequiresIdempotency {
			parameters = append(parameters, map[string]any{
				"name":        "Idempotency-Key",
				"in":          "header",
				"required":    true,
				"description": "Unique key used to make command execution idempotent.",
				"schema": map[string]any{
					"type":      "string",
					"minLength": 8,
				},
			})
		}
		for _, query := range op.QueryParameters {
			parameters = append(parameters, map[string]any{
				"name":        query.Name,
				"in":          "query",
				"required":    query.Required,
				"description": query.Description,
				"schema":      query.Schema,
			})
		}
		if len(parameters) > 0 {
			operation["parameters"] = parameters
		}

		if op.RequestBody != nil {
			operation["requestBody"] = map[string]any{
				"description": op.RequestBody.Description,
				"required":    op.RequestBody.Required,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": op.RequestBody.Schema,
					},
				},
			}
		}

		pathItem[strings.ToLower(op.Method)] = operation
	}

	return map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       s.info.Title,
			"description": s.info.Description,
			"version":     s.info.Version,
		},
		"paths": paths,
		"components": map[string]any{
			"schemas": map[string]any{
				"Problem": ProblemSchema(),
			},
		},
	}
}

func responsesToOpenAPI(responses map[string]ResponseSpec) map[string]any {
	if len(responses) == 0 {
		responses = map[string]ResponseSpec{
			"204": {Description: "No content"},
		}
	}

	result := map[string]any{}
	for status, response := range responses {
		item := map[string]any{
			"description": response.Description,
		}
		if response.Schema != nil {
			item["content"] = map[string]any{
				"application/json": map[string]any{
					"schema": response.Schema,
				},
			}
		}
		result[status] = item
	}
	return result
}

var pathParamPattern = regexp.MustCompile(`\{([^}/]+)\}`)

func pathParameters(path string) []map[string]any {
	matches := pathParamPattern.FindAllStringSubmatch(path, -1)
	params := make([]map[string]any, 0, len(matches))
	for _, match := range matches {
		params = append(params, map[string]any{
			"name":     match[1],
			"in":       "path",
			"required": true,
			"schema": map[string]any{
				"type": "string",
			},
		})
	}
	return params
}

func ObjectSchema(properties map[string]Schema, required ...string) Schema {
	props := map[string]any{}
	for name, schema := range properties {
		props[name] = schema
	}
	schema := Schema{
		"type":                 "object",
		"properties":           props,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func ArraySchema(item Schema) Schema {
	return Schema{
		"type":  "array",
		"items": item,
	}
}

func StringSchema() Schema {
	return Schema{"type": "string"}
}

func DateTimeSchema() Schema {
	return Schema{"type": "string", "format": "date-time"}
}

func ProblemSchema() Schema {
	return ObjectSchema(map[string]Schema{
		"type":   StringSchema(),
		"title":  StringSchema(),
		"status": {"type": "integer"},
		"detail": StringSchema(),
		"code":   StringSchema(),
	}, "type", "title", "status", "code")
}

func HealthResponseSchema() Schema {
	return ObjectSchema(map[string]Schema{
		"service": StringSchema(),
		"status":  StringSchema(),
		"version": StringSchema(),
		"time":    DateTimeSchema(),
	}, "service", "status", "version", "time")
}

func ScalarHTML(title string) string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>` + title + ` API</title>
</head>
<body>
  <script id="api-reference" data-url="/openapi.json"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.60.0"></script>
</body>
</html>`
}
