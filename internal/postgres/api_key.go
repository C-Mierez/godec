package postgres

import (
	"context"

	"github.com/c-mierez/godec/internal/apikey"
	db "github.com/c-mierez/godec/internal/postgres/db"
	"github.com/google/uuid"
)

type ApiKeyStore struct {
	queries *db.Queries
}

func NewApiKeyStore(queries *db.Queries) *ApiKeyStore {
	return &ApiKeyStore{queries: queries}
}

func (s *ApiKeyStore) CreateApiKey(ctx context.Context, tenantID uuid.UUID, name, hashedKey string, scopes []string) (*apikey.ApiKey, error) {
	row, err := s.queries.CreateApiKey(ctx, db.CreateApiKeyParams{
		TenantID:  db.UuidToPGUUID(tenantID),
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

func (s *ApiKeyStore) GetApiKeyByHashedKey(ctx context.Context, hashedKey string) (*apikey.ApiKey, error) {
	row, err := s.queries.GetApiKeyByHashedKey(ctx, hashedKey)
	if err != nil {
		return nil, err
	}

	domain := apiKeyFromRow(row)
	return &domain, nil
}
