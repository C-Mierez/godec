package api

import (
	"context"

	"github.com/c-mierez/godec/internal/apikey"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

// APIKeyHandlers implements API key management endpoints.
type APIKeyHandlers struct { //nolint:revive // name is intentional for clarity
	service *apikey.Service
}

// NewAPIKeyHandlers creates an APIKeyHandlers backed by the given service.
func NewAPIKeyHandlers(service *apikey.Service) *APIKeyHandlers { //nolint:revive // name is intentional for clarity
	return &APIKeyHandlers{
		service: service,
	}
}

// CreateApiKey handles API key creation requests.
//
//nolint:revive // var-naming: name matches generated StrictServerInterface
func (h *APIKeyHandlers) CreateApiKey(ctx context.Context, request CreateApiKeyRequestObject) (CreateApiKeyResponseObject, error) {
	if request.Body == nil {
		return CreateApiKey400JSONResponse{BadRequestJSONResponse{Error: "missing request body"}}, nil
	}

	scopes := []string{}
	if request.Body.Scopes != nil {
		scopes = *request.Body.Scopes
	}

	secret, apiKey, err := h.service.GenerateAPIKey(
		ctx,
		uuid.UUID(request.Body.TenantId),
		request.Body.Name,
		scopes,
	)
	if err != nil {
		return CreateApiKey400JSONResponse{BadRequestJSONResponse{Error: err.Error()}}, nil
	}

	return CreateApiKey201JSONResponse(h.domainAPIKeyToResponse(secret, apiKey)), nil
}

// domainAPIKeyToResponse converts domain apikey.APIKey to API CreateAPIKeyResponse.
func (h *APIKeyHandlers) domainAPIKeyToResponse(secret string, ak *apikey.APIKey) CreateAPIKeyResponse {
	if ak == nil {
		return CreateAPIKeyResponse{}
	}

	scopes := ak.Scopes
	if scopes == nil {
		scopes = []string{}
	}

	return CreateAPIKeyResponse{
		Id:       types.UUID(ak.ID),
		TenantId: types.UUID(ak.TenantID),
		Name:     ak.Name,
		Scopes:   &scopes,
		TokenId:  ak.TokenID,
		Secret:   secret,
	}
}
