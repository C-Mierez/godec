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
func (s *APIKeyStore) CreateAPIKey(ctx context.Context, tenantID uuid.UUID, name, tokenID, hashedSecret string, scopes []string) (*apikey.APIKey, error) {
	row, err := s.queries.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		TenantID:     db.UUIDToPGUUID(tenantID),
		Name:         name,
		TokenID:      tokenID,
		HashedSecret: hashedSecret,
		Scopes:       scopes,
	})
	if err != nil {
		return nil, err
	}

	domain := apiKeyFromRow(row)
	return &domain, nil
}

// GetAPIKeyByTokenID looks up an API key by its token identifier.
func (s *APIKeyStore) GetAPIKeyByTokenID(ctx context.Context, tokenID string) (*apikey.APIKey, error) {
	row, err := s.queries.GetAPIKeyByTokenID(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	domain := apiKeyFromRow(row)
	return &domain, nil
}
