// Package postgres implements persistence layer repositories using SQLC-generated queries.
package postgres

import (
	"context"

	"github.com/c-mierez/godec/internal/apikey"
	db "github.com/c-mierez/godec/internal/postgres/db"
	"github.com/google/uuid"
)

// APIKeyStore implements the apikey.Store interface using SQLC-generated queries.
type APIKeyStore struct {
	queries *db.Queries
}

// NewAPIKeyStore creates an APIKeyStore backed by the given SQLC queries.
func NewAPIKeyStore(queries *db.Queries) *APIKeyStore {
	return &APIKeyStore{queries: queries}
}

// CreateAPIKey inserts a new API key and returns the persisted record.
func (s *APIKeyStore) CreateAPIKey(ctx context.Context, tenantID uuid.UUID, name, hashedKey string, scopes []string) (*apikey.APIKey, error) {
	row, err := s.queries.CreateApiKey(ctx, db.CreateApiKeyParams{
		TenantID:  db.UUIDToPGUUID(tenantID),
		Name:      name,
		HashedKey: hashedKey,
		Scopes:    scopes,
	})
	if err != nil {
		return nil, err
	}

	domain := apiKeyFromRow(row)
	return &domain, nil
}

// GetAPIKeyByHashedKey looks up an API key by its SHA-256 hash.
func (s *APIKeyStore) GetAPIKeyByHashedKey(ctx context.Context, hashedKey string) (*apikey.APIKey, error) {
	row, err := s.queries.GetApiKeyByHashedKey(ctx, hashedKey)
	if err != nil {
		return nil, err
	}

	domain := apiKeyFromRow(row)
	return &domain, nil
}
