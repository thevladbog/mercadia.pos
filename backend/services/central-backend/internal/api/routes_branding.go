package api

import (
	"net/http"
	"time"

	"mercadia.dev/pos/platform/httpapi"
	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
)

type ColorSchemeResponse struct {
	ID              string                   `json:"id"`
	Name            string                   `json:"name"`
	LogoURL         string                   `json:"logoUrl"`
	AccentPreset    domain.AccentPreset      `json:"accentPreset,omitempty"`
	AccentColor     string                   `json:"accentColor,omitempty"`
	BackgroundColor string                   `json:"backgroundColor,omitempty"`
	ResolvedAccent  string                   `json:"resolvedAccentColor"`
	Status          domain.ColorSchemeStatus `json:"status"`
	StoreIDs        []string                 `json:"storeIds"`
	CreatedAt       time.Time                `json:"createdAt"`
	UpdatedAt       time.Time                `json:"updatedAt"`
}

type ColorSchemesResponse struct {
	Schemes []ColorSchemeResponse `json:"schemes"`
}

type ColorSchemeAcceptedResponse struct {
	Scheme ColorSchemeResponse `json:"scheme"`
}

type CreateColorSchemeRequest struct {
	SchemeID        string                   `json:"schemeId"`
	Name            string                   `json:"name"`
	LogoURL         string                   `json:"logoUrl,omitempty"`
	AccentPreset    domain.AccentPreset      `json:"accentPreset,omitempty"`
	AccentColor     string                   `json:"accentColor,omitempty"`
	BackgroundColor string                   `json:"backgroundColor,omitempty"`
	Status          domain.ColorSchemeStatus `json:"status,omitempty"`
	StoreIDs        []string                 `json:"storeIds,omitempty"`
}

type UpdateColorSchemeRequest struct {
	Name            *string                   `json:"name,omitempty"`
	LogoURL         *string                   `json:"logoUrl,omitempty"`
	AccentPreset    *domain.AccentPreset      `json:"accentPreset,omitempty"`
	AccentColor     *string                   `json:"accentColor,omitempty"`
	BackgroundColor *string                   `json:"backgroundColor,omitempty"`
	Status          *domain.ColorSchemeStatus `json:"status,omitempty"`
	StoreIDs        []string                  `json:"storeIds,omitempty"`
}

type LayoutGridCategoryResponse struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type LayoutGridTileResponse struct {
	Label      string `json:"label"`
	Color      string `json:"color,omitempty"`
	ProductID  string `json:"productId,omitempty"`
	Empty      bool   `json:"empty,omitempty"`
	CategoryID string `json:"categoryId,omitempty"`
	IconURL    string `json:"iconUrl,omitempty"`
}

type LayoutGridResponse struct {
	Rows       int                          `json:"rows"`
	Cols       int                          `json:"cols"`
	Categories []LayoutGridCategoryResponse `json:"categories,omitempty"`
	Tiles      []LayoutGridTileResponse     `json:"tiles"`
}

type LayoutTemplateResponse struct {
	ID                   string                      `json:"id"`
	Name                 string                      `json:"name"`
	Kind                 domain.LayoutTemplateKind   `json:"kind"`
	AccentPreset         domain.AccentPreset         `json:"accentPreset,omitempty"`
	AccentColor          string                      `json:"accentColor,omitempty"`
	ColorSchemeID        string                      `json:"colorSchemeId,omitempty"`
	ResolvedAccentPreset domain.AccentPreset         `json:"resolvedAccentPreset"`
	ResolvedAccentColor  string                      `json:"resolvedAccentColor"`
	Grid                 LayoutGridResponse          `json:"grid"`
	StoreID              string                      `json:"storeId,omitempty"`
	TerminalType         string                      `json:"terminalType,omitempty"`
	Status               domain.LayoutTemplateStatus `json:"status"`
	Version              int                         `json:"version"`
	CreatedAt            time.Time                   `json:"createdAt"`
	UpdatedAt            time.Time                   `json:"updatedAt"`
}

type LayoutTemplatesResponse struct {
	Templates []LayoutTemplateResponse `json:"templates"`
}

type LayoutTemplateAcceptedResponse struct {
	Template LayoutTemplateResponse `json:"template"`
}

type CreateLayoutTemplateRequest struct {
	TemplateID    string                      `json:"templateId"`
	Name          string                      `json:"name"`
	Kind          domain.LayoutTemplateKind   `json:"kind"`
	AccentPreset  domain.AccentPreset         `json:"accentPreset,omitempty"`
	AccentColor   string                      `json:"accentColor,omitempty"`
	ColorSchemeID string                      `json:"colorSchemeId,omitempty"`
	Grid          domain.LayoutGrid           `json:"grid,omitempty"`
	StoreID       string                      `json:"storeId,omitempty"`
	TerminalType  string                      `json:"terminalType,omitempty"`
	Status        domain.LayoutTemplateStatus `json:"status,omitempty"`
}

type UpdateLayoutTemplateRequest struct {
	Name          *string                      `json:"name,omitempty"`
	Kind          *domain.LayoutTemplateKind   `json:"kind,omitempty"`
	AccentPreset  *domain.AccentPreset         `json:"accentPreset,omitempty"`
	AccentColor   *string                      `json:"accentColor,omitempty"`
	ColorSchemeID *string                      `json:"colorSchemeId,omitempty"`
	Grid          *domain.LayoutGrid           `json:"grid,omitempty"`
	StoreID       *string                      `json:"storeId,omitempty"`
	TerminalType  *string                      `json:"terminalType,omitempty"`
	Status        *domain.LayoutTemplateStatus `json:"status,omitempty"`
}

func mountBrandingRoutes(mux *http.ServeMux, spec *httpapi.Spec, services Services) {
	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/color-schemes",
		OperationID: "listColorSchemes",
		Summary:     "List color schemes",
		Tags:        []string{"color-schemes"},
		Responses:   protectedResponseSpecs("200", "Color schemes", colorSchemesResponseSchema()),
	}, RequireSyncAPIKeyOrSessionAuth(services.Auth, services.SyncAPIKey, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		schemes, err := services.ColorSchemes.ListColorSchemes(r.Context(), session)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, ColorSchemesResponse{Schemes: colorSchemeResponses(schemes)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPost,
		Path:        "/v1/color-schemes",
		OperationID: "createColorScheme",
		Summary:     "Create color scheme",
		Tags:        []string{"color-schemes"},
		RequestBody: &httpapi.BodySpec{Required: true, Schema: createColorSchemeRequestSchema()},
		Responses:   protectedResponseSpecs("201", "Color scheme created", colorSchemeAcceptedResponseSchema()),
	}, RequireSession(services.Auth, app.PermissionUsersManage, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		var request CreateColorSchemeRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_color_scheme_command", "Invalid color scheme command", err.Error())
			return
		}
		scheme, err := services.ColorSchemes.CreateColorScheme(r.Context(), app.CreateColorSchemeCommand{
			SchemeID:        request.SchemeID,
			Name:            request.Name,
			LogoURL:         request.LogoURL,
			AccentPreset:    request.AccentPreset,
			AccentColor:     request.AccentColor,
			BackgroundColor: request.BackgroundColor,
			Status:          request.Status,
			StoreIDs:        request.StoreIDs,
			Session:         session,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, ColorSchemeAcceptedResponse{Scheme: colorSchemeResponse(scheme)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/color-schemes/{schemeId}",
		OperationID: "getColorScheme",
		Summary:     "Get color scheme",
		Tags:        []string{"color-schemes"},
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Color scheme", colorSchemeAcceptedResponseSchema()),
			map[string]httpapi.ResponseSpec{"404": {Description: "Color scheme was not found", Schema: httpapi.ProblemSchema()}},
		),
	}, RequireSyncAPIKeyOrSessionAuth(services.Auth, services.SyncAPIKey, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		scheme, err := services.ColorSchemes.GetColorScheme(r.Context(), r.PathValue("schemeId"), session)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, ColorSchemeAcceptedResponse{Scheme: colorSchemeResponse(scheme)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPatch,
		Path:        "/v1/color-schemes/{schemeId}",
		OperationID: "updateColorScheme",
		Summary:     "Update color scheme",
		Tags:        []string{"color-schemes"},
		RequestBody: &httpapi.BodySpec{Required: true, Schema: updateColorSchemeRequestSchema()},
		Responses:   protectedResponseSpecs("200", "Color scheme updated", colorSchemeAcceptedResponseSchema()),
	}, RequireSession(services.Auth, app.PermissionUsersManage, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		var request UpdateColorSchemeRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_color_scheme_command", "Invalid color scheme command", err.Error())
			return
		}
		scheme, err := services.ColorSchemes.UpdateColorScheme(r.Context(), app.UpdateColorSchemeCommand{
			SchemeID:        r.PathValue("schemeId"),
			Name:            request.Name,
			LogoURL:         request.LogoURL,
			AccentPreset:    request.AccentPreset,
			AccentColor:     request.AccentColor,
			BackgroundColor: request.BackgroundColor,
			Status:          request.Status,
			StoreIDs:        request.StoreIDs,
			Session:         session,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, ColorSchemeAcceptedResponse{Scheme: colorSchemeResponse(scheme)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/layout-templates",
		OperationID: "listLayoutTemplates",
		Summary:     "List layout templates",
		Tags:        []string{"layout-templates"},
		Responses:   protectedResponseSpecs("200", "Layout templates", layoutTemplatesResponseSchema()),
	}, RequireSyncAPIKeyOrSessionAuth(services.Auth, services.SyncAPIKey, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		results, err := services.LayoutTemplates.ListLayoutTemplates(r.Context(), session, app.LayoutTemplateListFilter{
			StoreID: r.URL.Query().Get("storeId"),
			Kind:    domain.LayoutTemplateKind(r.URL.Query().Get("kind")),
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, LayoutTemplatesResponse{Templates: layoutTemplateResponses(results)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPost,
		Path:        "/v1/layout-templates",
		OperationID: "createLayoutTemplate",
		Summary:     "Create layout template",
		Tags:        []string{"layout-templates"},
		RequestBody: &httpapi.BodySpec{Required: true, Schema: createLayoutTemplateRequestSchema()},
		Responses:   protectedResponseSpecs("201", "Layout template created", layoutTemplateAcceptedResponseSchema()),
	}, RequireSession(services.Auth, app.PermissionUsersManage, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		var request CreateLayoutTemplateRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_layout_template_command", "Invalid layout template command", err.Error())
			return
		}
		result, err := services.LayoutTemplates.CreateLayoutTemplate(r.Context(), app.CreateLayoutTemplateCommand{
			TemplateID:    request.TemplateID,
			Name:          request.Name,
			Kind:          request.Kind,
			AccentPreset:  request.AccentPreset,
			AccentColor:   request.AccentColor,
			ColorSchemeID: request.ColorSchemeID,
			Grid:          request.Grid,
			StoreID:       request.StoreID,
			TerminalType:  request.TerminalType,
			Status:        request.Status,
			Session:       session,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusCreated, LayoutTemplateAcceptedResponse{Template: layoutTemplateResponse(result)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/layout-templates/{templateId}",
		OperationID: "getLayoutTemplate",
		Summary:     "Get layout template",
		Tags:        []string{"layout-templates"},
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Layout template", layoutTemplateAcceptedResponseSchema()),
			map[string]httpapi.ResponseSpec{"404": {Description: "Layout template was not found", Schema: httpapi.ProblemSchema()}},
		),
	}, RequireSyncAPIKeyOrSessionAuth(services.Auth, services.SyncAPIKey, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		result, err := services.LayoutTemplates.GetLayoutTemplate(r.Context(), r.PathValue("templateId"), session)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, LayoutTemplateAcceptedResponse{Template: layoutTemplateResponse(result)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodPatch,
		Path:        "/v1/layout-templates/{templateId}",
		OperationID: "updateLayoutTemplate",
		Summary:     "Update layout template",
		Tags:        []string{"layout-templates"},
		RequestBody: &httpapi.BodySpec{Required: true, Schema: updateLayoutTemplateRequestSchema()},
		Responses:   protectedResponseSpecs("200", "Layout template updated", layoutTemplateAcceptedResponseSchema()),
	}, RequireSession(services.Auth, app.PermissionUsersManage, func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		var request UpdateLayoutTemplateRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_layout_template_command", "Invalid layout template command", err.Error())
			return
		}
		result, err := services.LayoutTemplates.UpdateLayoutTemplate(r.Context(), app.UpdateLayoutTemplateCommand{
			TemplateID:    r.PathValue("templateId"),
			Name:          request.Name,
			Kind:          request.Kind,
			AccentPreset:  request.AccentPreset,
			AccentColor:   request.AccentColor,
			ColorSchemeID: request.ColorSchemeID,
			Grid:          request.Grid,
			StoreID:       request.StoreID,
			TerminalType:  request.TerminalType,
			Status:        request.Status,
			Session:       session,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, LayoutTemplateAcceptedResponse{Template: layoutTemplateResponse(result)})
	}))
}

func colorSchemeResponse(scheme domain.ColorScheme) ColorSchemeResponse {
	return ColorSchemeResponse{
		ID:              scheme.ID,
		Name:            scheme.Name,
		LogoURL:         scheme.LogoURL,
		AccentPreset:    scheme.AccentPreset,
		AccentColor:     scheme.AccentColor,
		BackgroundColor: scheme.BackgroundColor,
		ResolvedAccent:  scheme.ResolvedAccentColor(),
		Status:          scheme.Status,
		StoreIDs:        append([]string(nil), scheme.StoreIDs...),
		CreatedAt:       scheme.CreatedAt,
		UpdatedAt:       scheme.UpdatedAt,
	}
}

func colorSchemeResponses(schemes []domain.ColorScheme) []ColorSchemeResponse {
	responses := make([]ColorSchemeResponse, 0, len(schemes))
	for _, scheme := range schemes {
		responses = append(responses, colorSchemeResponse(scheme))
	}
	return responses
}

func layoutGridResponse(grid domain.LayoutGrid) LayoutGridResponse {
	categories := make([]LayoutGridCategoryResponse, 0, len(grid.Categories))
	for _, category := range grid.Categories {
		categories = append(categories, LayoutGridCategoryResponse{
			ID:    category.ID,
			Label: category.Label,
		})
	}
	tiles := make([]LayoutGridTileResponse, 0, len(grid.Tiles))
	for _, tile := range grid.Tiles {
		tiles = append(tiles, LayoutGridTileResponse{
			Label:      tile.Label,
			Color:      tile.Color,
			ProductID:  tile.ProductID,
			Empty:      tile.Empty,
			CategoryID: tile.CategoryID,
			IconURL:    tile.IconURL,
		})
	}
	return LayoutGridResponse{
		Rows:       grid.Rows,
		Cols:       grid.Cols,
		Categories: categories,
		Tiles:      tiles,
	}
}

func layoutTemplateResponse(result app.LayoutTemplateResult) LayoutTemplateResponse {
	template := result.Template
	return LayoutTemplateResponse{
		ID:                   template.ID,
		Name:                 template.Name,
		Kind:                 template.Kind,
		AccentPreset:         template.AccentPreset,
		AccentColor:          template.AccentColor,
		ColorSchemeID:        template.ColorSchemeID,
		ResolvedAccentPreset: result.ResolvedAccentPreset,
		ResolvedAccentColor:  result.ResolvedAccentColor,
		Grid:                 layoutGridResponse(template.Grid),
		StoreID:              template.StoreID,
		TerminalType:         template.TerminalType,
		Status:               template.Status,
		Version:              template.Version,
		CreatedAt:            template.CreatedAt,
		UpdatedAt:            template.UpdatedAt,
	}
}

func layoutTemplateResponses(results []app.LayoutTemplateResult) []LayoutTemplateResponse {
	responses := make([]LayoutTemplateResponse, 0, len(results))
	for _, result := range results {
		responses = append(responses, layoutTemplateResponse(result))
	}
	return responses
}

func layoutGridSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"rows":       {"type": "integer"},
		"cols":       {"type": "integer"},
		"categories": httpapi.ArraySchema(layoutGridCategorySchema()),
		"tiles":      httpapi.ArraySchema(layoutGridTileSchema()),
	})
}

func layoutGridCategorySchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":    httpapi.StringSchema(),
		"label": httpapi.StringSchema(),
	})
}

func layoutGridTileSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"label":      httpapi.StringSchema(),
		"color":      httpapi.StringSchema(),
		"productId":  httpapi.StringSchema(),
		"empty":      {"type": "boolean"},
		"categoryId": httpapi.StringSchema(),
		"iconUrl":    httpapi.StringSchema(),
	})
}

func colorSchemeResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":                  httpapi.StringSchema(),
		"name":                httpapi.StringSchema(),
		"logoUrl":             httpapi.StringSchema(),
		"accentPreset":        httpapi.StringSchema(),
		"accentColor":         httpapi.StringSchema(),
		"backgroundColor":     httpapi.StringSchema(),
		"resolvedAccentColor": httpapi.StringSchema(),
		"status":              httpapi.StringSchema(),
		"storeIds":            httpapi.ArraySchema(httpapi.StringSchema()),
		"createdAt":           httpapi.DateTimeSchema(),
		"updatedAt":           httpapi.DateTimeSchema(),
	}, "id", "name", "resolvedAccentColor", "status", "storeIds", "createdAt", "updatedAt")
}

func colorSchemesResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"schemes": httpapi.ArraySchema(colorSchemeResponseSchema()),
	}, "schemes")
}

func colorSchemeAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"scheme": colorSchemeResponseSchema(),
	}, "scheme")
}

func createColorSchemeRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"schemeId":        httpapi.StringSchema(),
		"name":            httpapi.StringSchema(),
		"logoUrl":         httpapi.StringSchema(),
		"accentPreset":    httpapi.StringSchema(),
		"accentColor":     httpapi.StringSchema(),
		"backgroundColor": httpapi.StringSchema(),
		"status":          httpapi.StringSchema(),
		"storeIds":        httpapi.ArraySchema(httpapi.StringSchema()),
	}, "schemeId", "name")
}

func updateColorSchemeRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"name":            httpapi.StringSchema(),
		"logoUrl":         httpapi.StringSchema(),
		"accentPreset":    httpapi.StringSchema(),
		"accentColor":     httpapi.StringSchema(),
		"backgroundColor": httpapi.StringSchema(),
		"status":          httpapi.StringSchema(),
		"storeIds":        httpapi.ArraySchema(httpapi.StringSchema()),
	})
}

func layoutTemplateResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":                   httpapi.StringSchema(),
		"name":                 httpapi.StringSchema(),
		"kind":                 httpapi.StringSchema(),
		"accentPreset":         httpapi.StringSchema(),
		"accentColor":          httpapi.StringSchema(),
		"colorSchemeId":        httpapi.StringSchema(),
		"resolvedAccentPreset": httpapi.StringSchema(),
		"resolvedAccentColor":  httpapi.StringSchema(),
		"grid":                 layoutGridSchema(),
		"storeId":              httpapi.StringSchema(),
		"terminalType":         httpapi.StringSchema(),
		"status":               httpapi.StringSchema(),
		"version":              {"type": "integer"},
		"createdAt":            httpapi.DateTimeSchema(),
		"updatedAt":            httpapi.DateTimeSchema(),
	}, "id", "name", "kind", "resolvedAccentPreset", "resolvedAccentColor", "grid", "status", "version", "createdAt", "updatedAt")
}

func layoutTemplatesResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"templates": httpapi.ArraySchema(layoutTemplateResponseSchema()),
	}, "templates")
}

func layoutTemplateAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"template": layoutTemplateResponseSchema(),
	}, "template")
}

func createLayoutTemplateRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"templateId":    httpapi.StringSchema(),
		"name":          httpapi.StringSchema(),
		"kind":          httpapi.StringSchema(),
		"accentPreset":  httpapi.StringSchema(),
		"accentColor":   httpapi.StringSchema(),
		"colorSchemeId": httpapi.StringSchema(),
		"grid":          layoutGridSchema(),
		"storeId":       httpapi.StringSchema(),
		"terminalType":  httpapi.StringSchema(),
		"status":        httpapi.StringSchema(),
	}, "templateId", "name", "kind")
}

func updateLayoutTemplateRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"name":          httpapi.StringSchema(),
		"kind":          httpapi.StringSchema(),
		"accentPreset":  httpapi.StringSchema(),
		"accentColor":   httpapi.StringSchema(),
		"colorSchemeId": httpapi.StringSchema(),
		"grid":          layoutGridSchema(),
		"storeId":       httpapi.StringSchema(),
		"terminalType":  httpapi.StringSchema(),
		"status":        httpapi.StringSchema(),
	})
}
