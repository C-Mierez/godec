package api

import (
	"bytes"
	"context"
	_ "embed"

	"github.com/c-mierez/godec/apidoc"
	"github.com/getkin/kin-openapi/openapi3"
)

// DocumentationHandlers implements API documentation endpoints.
type DocumentationHandlers struct{}

// NewDocumentationHandlers creates a DocumentationHandlers.
func NewDocumentationHandlers() *DocumentationHandlers {
	return &DocumentationHandlers{}
}

//go:embed spec.yaml
var specYaml []byte

// GetSwagger loads and returns the OpenAPI specification.
func GetSwagger() (*openapi3.T, error) {
	return openapi3.NewLoader().LoadFromData(specYaml)
}

// GetOpenAPISpec returns the OpenAPI specification as YAML.
func (h *DocumentationHandlers) GetOpenAPISpec(_ context.Context, _ GetOpenAPISpecRequestObject) (GetOpenAPISpecResponseObject, error) {
	return GetOpenAPISpec200ApplicationyamlResponse{
		Body:          bytes.NewReader(specYaml),
		ContentLength: int64(len(specYaml)),
	}, nil
}

// GetAPIDocs returns the API documentation HTML page.
func (h *DocumentationHandlers) GetAPIDocs(_ context.Context, _ GetAPIDocsRequestObject) (GetAPIDocsResponseObject, error) {
	return GetAPIDocs200TexthtmlResponse{
		Body:          bytes.NewReader(apidoc.APIDocsHTML),
		ContentLength: int64(len(apidoc.APIDocsHTML)),
	}, nil
}
