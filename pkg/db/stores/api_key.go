// Package stores defines repository interfaces for data access abstraction.
package stores

import (
	"context"

	"github.com/c-mierez/godec/pkg/models"
	"github.com/google/uuid"
)

// ApiKeyStore defines the contract for API key storage operations.
// Implementations handle persistence of API keys in different backends.
type ApiKeyStore interface {
	CreateApiKey(ctx context.Context, tenantID uuid.UUID, name, hashedKey string, scopes []string) (*models.ApiKey, error)
	GetApiKeyByID(ctx context.Context, id uuid.UUID) (*models.ApiKey, error)
	GetApiKeyByHashedKey(ctx context.Context, hashedKey string) (*models.ApiKey, error)
	ListApiKeysByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.ApiKey, error)
	ListActiveApiKeysByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.ApiKey, error)
	DeleteApiKey(ctx context.Context, id uuid.UUID) error
	DeleteApiKeysByTenantID(ctx context.Context, tenantID uuid.UUID) error
	UpdateApiKeyName(ctx context.Context, id uuid.UUID, name string) (*models.ApiKey, error)
	SetApiKeyExpiration(ctx context.Context, id uuid.UUID, expiresAt *models.ApiKey) (*models.ApiKey, error)
}
