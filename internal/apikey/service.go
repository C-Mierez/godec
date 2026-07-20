package apikey

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the persistence operations required by the API key service.
type Store interface {
	CreateAPIKey(ctx context.Context, tenantID uuid.UUID, name, tokenID, hashedSecret string, scopes []string) (*APIKey, error)
	GetAPIKeyByTokenID(ctx context.Context, tokenID string) (*APIKey, error)
}

// Service implements API key generation and validation logic.
type Service struct {
	store Store
}

// NewService creates an API key service backed by the given store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// GenerateAPIKey creates a new API key with split credentials, returning the
// token ID, secret (shown once), and the persisted record.
func (s *Service) GenerateAPIKey(ctx context.Context, tenantID uuid.UUID, name string, scopes []string) (secret string, apiKey *APIKey, err error) {
	tokenID, secret, hashedSecret, err := generateCredentials()
	if err != nil {
		return "", nil, err
	}

	apiKey, err = s.store.CreateAPIKey(ctx, tenantID, name, tokenID, hashedSecret, scopes)
	if err != nil {
		return "", nil, err
	}

	return secret, apiKey, nil
}

// ValidateAPIKey verifies credentials by looking up the token ID and comparing
// the hashed secret.
func (s *Service) ValidateAPIKey(ctx context.Context, tokenID, plainSecret string) (*APIKey, error) {
	apiKey, err := s.store.GetAPIKeyByTokenID(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	computedHash := hashSecret(plainSecret)
	if !validateSecretHash(computedHash, apiKey.HashedSecret) {
		return nil, nil
	}

	return apiKey, nil
}
