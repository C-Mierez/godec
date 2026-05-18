package api

import (
	"bytes"
	"context"
	_ "embed"

	"github.com/c-mierez/godec/internal/apidoc"
)

type DocumentationHandlers struct{}

func NewDocumentationHandlers() *DocumentationHandlers {
	return &DocumentationHandlers{}
}

//go:embed spec.yaml
var specYaml []byte

func (h *DocumentationHandlers) GetOpenAPISpec(ctx context.Context, request GetOpenAPISpecRequestObject) (GetOpenAPISpecResponseObject, error) {
	return GetOpenAPISpec200ApplicationyamlResponse{
		Body:          bytes.NewReader(specYaml),
		ContentLength: int64(len(specYaml)),
	}, nil
}

func (h *DocumentationHandlers) GetAPIDocs(ctx context.Context, request GetAPIDocsRequestObject) (GetAPIDocsResponseObject, error) {
	return GetAPIDocs200TexthtmlResponse{
		Body:          bytes.NewReader(apidoc.ApiDocsHtml),
		ContentLength: int64(len(apidoc.ApiDocsHtml)),
	}, nil
}
