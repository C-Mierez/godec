// Package postgres provides PostgreSQL implementations of repository interfaces.
package postgres

import (
	"context"

	gen "github.com/c-mierez/godec/pkg/db/postgres/_gen"
	"github.com/c-mierez/godec/pkg/db/stores"
	"github.com/c-mierez/godec/pkg/models"
	"github.com/google/uuid"
)

// Store implements the ApiKeyStore interface for PostgreSQL.
type Store struct {
	queries *gen.Queries
}

// NewApiKeyStore creates a new PostgreSQL-backed API key store.
func NewApiKeyStore(queries *gen.Queries) stores.ApiKeyStore {
	return &Store{
		queries: queries,
	}
}

func (s *Store) CreateApiKey(ctx context.Context, tenantID uuid.UUID, name, hashedKey string, scopes []string) (*models.ApiKey, error) {
	params := gen.CreateApiKeyParams{
		TenantID:  gen.UuidToPGUUID(tenantID),
		Name:      name,
		HashedKey: hashedKey,
		Scopes:    scopes,
	}

	apiKey, err := s.queries.CreateApiKey(ctx, params)
	if err != nil {
		return nil, err
	}

	domainKey := apiKey.ToDomain()
	return &domainKey, nil
}

func (s *Store) GetApiKeyByID(ctx context.Context, id uuid.UUID) (*models.ApiKey, error) {
	apiKey, err := s.queries.GetApiKeyByID(ctx, gen.UuidToPGUUID(id))
	if err != nil {
		return nil, err
	}

	domainKey := apiKey.ToDomain()
	return &domainKey, nil
}

func (s *Store) GetApiKeyByHashedKey(ctx context.Context, hashedKey string) (*models.ApiKey, error) {
	apiKey, err := s.queries.GetApiKeyByHashedKey(ctx, hashedKey)
	if err != nil {
		return nil, err
	}

	domainKey := apiKey.ToDomain()
	return &domainKey, nil
}

func (s *Store) ListApiKeysByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.ApiKey, error) {
	apiKeys, err := s.queries.ListApiKeysByTenantID(ctx, gen.UuidToPGUUID(tenantID))
	if err != nil {
		return nil, err
	}

	domainKeys := make([]*models.ApiKey, len(apiKeys))
	for i, ak := range apiKeys {
		domainKey := ak.ToDomain()
		domainKeys[i] = &domainKey
	}

	return domainKeys, nil
}

func (s *Store) ListActiveApiKeysByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.ApiKey, error) {
	apiKeys, err := s.queries.ListActiveApiKeysByTenantID(ctx, gen.UuidToPGUUID(tenantID))
	if err != nil {
		return nil, err
	}

	domainKeys := make([]*models.ApiKey, len(apiKeys))
	for i, ak := range apiKeys {
		domainKey := ak.ToDomain()
		domainKeys[i] = &domainKey
	}

	return domainKeys, nil
}

func (s *Store) DeleteApiKey(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeleteApiKey(ctx, gen.UuidToPGUUID(id))
}

func (s *Store) DeleteApiKeysByTenantID(ctx context.Context, tenantID uuid.UUID) error {
	return s.queries.DeleteApiKeysByTenantID(ctx, gen.UuidToPGUUID(tenantID))
}

func (s *Store) UpdateApiKeyName(ctx context.Context, id uuid.UUID, name string) (*models.ApiKey, error) {
	params := gen.UpdateApiKeyNameParams{
		ID:   gen.UuidToPGUUID(id),
		Name: name,
	}

	apiKey, err := s.queries.UpdateApiKeyName(ctx, params)
	if err != nil {
		return nil, err
	}

	domainKey := apiKey.ToDomain()
	return &domainKey, nil
}

func (s *Store) SetApiKeyExpiration(ctx context.Context, id uuid.UUID, expiresAt *models.ApiKey) (*models.ApiKey, error) {
	params := gen.SetApiKeyExpirationParams{
		ID:        gen.UuidToPGUUID(id),
		ExpiresAt: gen.OptionalTimeToPGTimestamptz(expiresAt.ExpiresAt),
	}

	apiKey, err := s.queries.SetApiKeyExpiration(ctx, params)
	if err != nil {
		return nil, err
	}

	domainKey := apiKey.ToDomain()
	return &domainKey, nil
}
