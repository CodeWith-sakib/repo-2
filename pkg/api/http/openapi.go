package http

import (
	"encoding/json"
	"strings"
)

type OpenAPISpec struct {
	OpenAPI string                 `json:"openapi"`
	Info    OpenAPIInfo            `json:"info"`
	Paths   map[string]OpenAPIPath `json:"paths"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

type OpenAPIPath map[string]OpenAPIOperation

type OpenAPIOperation struct {
	Summary     string                  `json:"summary"`
	OperationID string                  `json:"operationId"`
	Tags        []string                `json:"tags,omitempty"`
	Responses   map[string]OpenAPIResp `json:"responses"`
}

type OpenAPIResp struct {
	Description string `json:"description"`
}

type OpenAPIGenerator struct {
	title   string
	version string
}

func NewOpenAPIGenerator(title, version string) *OpenAPIGenerator {
	return &OpenAPIGenerator{title: title, version: version}
}

func (g *OpenAPIGenerator) GenerateSpec() OpenAPISpec {
	spec := OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: OpenAPIInfo{
			Title:       g.title,
			Version:     g.version,
			Description: "KestrelFlow Enterprise Workflow Orchestrator REST API",
		},
		Paths: make(map[string]OpenAPIPath),
	}

	routes := []struct {
		Method  string
		Path    string
		Summary string
		OpID    string
		Tag     string
	}{
		{"GET", "/healthz", "Health check probe", "getHealth", "System"},
		{"GET", "/api/v1/workflows", "List workflows", "listWorkflows", "Workflows"},
		{"POST", "/api/v1/workflows", "Create workflow definition", "createWorkflow", "Workflows"},
		{"GET", "/api/v1/workflows/{id}", "Get workflow definition", "getWorkflow", "Workflows"},
		{"POST", "/api/v1/workflows/{id}/runs", "Trigger workflow run", "triggerRun", "Runs"},
		{"GET", "/api/v1/runs/{id}", "Get workflow run status", "getRun", "Runs"},
		{"POST", "/api/v1/runs/{id}/cancel", "Cancel workflow run", "cancelRun", "Runs"},
	}

	for _, r := range routes {
		pathItem, exists := spec.Paths[r.Path]
		if !exists {
			pathItem = make(OpenAPIPath)
		}
		pathItem[strings.ToLower(r.Method)] = OpenAPIOperation{
			Summary:     r.Summary,
			OperationID: r.OpID,
			Tags:        []string{r.Tag},
			Responses: map[string]OpenAPIResp{
				"200": {Description: "Successful operation"},
				"400": {Description: "Validation failure"},
				"401": {Description: "Unauthorized"},
				"404": {Description: "Resource not found"},
			},
		}
		spec.Paths[r.Path] = pathItem
	}

	return spec
}

func (g *OpenAPIGenerator) ToJSON() ([]byte, error) {
	spec := g.GenerateSpec()
	return json.MarshalIndent(spec, "", "  ")
}
