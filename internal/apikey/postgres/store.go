// Package postgres implements the apikey.Store interface using SQLC-generated queries.
package postgres

import (
	"context"

	"github.com/c-mierez/godec/internal/apikey"
	db "github.com/c-mierez/godec/internal/postgres/db"
	"github.com/google/uuid"
)

// Store implements the apikey.Store interface using SQLC-generated queries.
type Store struct {
	queries *db.Queries
}

// NewStore creates a Store backed by the given SQLC queries.
func NewStore(queries *db.Queries) *Store {
	return &Store{queries: queries}
}

// CreateAPIKey inserts a new API key and returns the persisted record.
func (s *Store) CreateAPIKey(ctx context.Context, tenantID uuid.UUID, name, tokenID, hashedSecret string, scopes []string) (*apikey.APIKey, error) {
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
func (s *Store) GetAPIKeyByTokenID(ctx context.Context, tokenID string) (*apikey.APIKey, error) {
	row, err := s.queries.GetAPIKeyByTokenID(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	domain := apiKeyFromRow(row)
	return &domain, nil
}
