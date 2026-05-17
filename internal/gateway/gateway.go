// Package gateway provides the business logic layer for API key management.
package gateway

import (
	"context"

	"github.com/c-mierez/godec/pkg/db/stores"
	"github.com/c-mierez/godec/pkg/models"
	"github.com/google/uuid"
)

// Gateway provides API key management operations and acts as a service layer.
// It depends on the ApiKeyStore interface for data access.
type Gateway struct {
	apiKeyStore stores.ApiKeyStore
}

// NewGateway creates a new Gateway instance with the provided store.
func NewGateway(store stores.ApiKeyStore) *Gateway {
	return &Gateway{
		apiKeyStore: store,
	}
}

func (g *Gateway) GenerateApiKey(ctx context.Context, tenantID uuid.UUID, name string, scopes []string) (string, *models.ApiKey, error) {
	// Generate the API Key
	plainKey, hashedKey, err := generateKey()
	if err != nil {
		return "", nil, err
	}

	// Store the hashed key
	apiKey, err := g.apiKeyStore.CreateApiKey(ctx, tenantID, name, hashedKey, scopes)
	if err != nil {
		return "", nil, err
	}

	return plainKey, apiKey, nil
}

func (g *Gateway) ValidateApiKey(ctx context.Context, providedKey string) (bool, *models.ApiKey, error) {
	// Fetch the stored hash for the provided key
	apiKey, err := g.apiKeyStore.GetApiKeyByHashedKey(ctx, providedKey)
	if err != nil {
		return false, nil, err
	}

	// Validate the key
	isValid := validateKey(providedKey, apiKey.HashedKey)
	return isValid, apiKey, nil
}
