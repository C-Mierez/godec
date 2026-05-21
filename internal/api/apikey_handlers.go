package api

import (
	"context"

	"github.com/c-mierez/godec/internal/apikey"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

type APIKeyHandlers struct {
	service *apikey.Service
}

func NewAPIKeyHandlers(service *apikey.Service) *APIKeyHandlers {
	return &APIKeyHandlers{
		service: service,
	}
}

func (h *APIKeyHandlers) CreateApiKey(ctx context.Context, request CreateApiKeyRequestObject) (CreateApiKeyResponseObject, error) {
	if request.Body == nil {
		return CreateApiKey400JSONResponse{BadRequestJSONResponse{Error: "missing request body"}}, nil
	}

	scopes := []string{}
	if request.Body.Scopes != nil {
		scopes = *request.Body.Scopes
	}

	plainKey, apiKey, err := h.service.GenerateApiKey(
		ctx,
		uuid.UUID(request.Body.TenantId),
		request.Body.Name,
		scopes,
	)
	if err != nil {
		return CreateApiKey400JSONResponse{BadRequestJSONResponse{Error: err.Error()}}, nil
	}

	return CreateApiKey201JSONResponse(h.domainAPIKeyToResponse(plainKey, apiKey)), nil
}

// domainAPIKeyToResponse converts domain apikey.ApiKey to API CreateAPIKeyResponse
func (h *APIKeyHandlers) domainAPIKeyToResponse(plainKey string, ak *apikey.ApiKey) CreateAPIKeyResponse {
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
		ApiKey:   plainKey,
	}
}
