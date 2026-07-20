package apikey

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the persistence operations required by the API key service.
type Store interface {
	CreateAPIKey(ctx context.Context, tenantID uuid.UUID, name, hashedKey string, scopes []string) (*APIKey, error)
	GetAPIKeyByHashedKey(ctx context.Context, hashedKey string) (*APIKey, error)
}

// Service implements API key generation and validation logic.
type Service struct {
	store Store
}

// NewService creates an API key service backed by the given store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// GenerateAPIKey creates a new API key, returning the plain key and the persisted record.
func (s *Service) GenerateAPIKey(ctx context.Context, tenantID uuid.UUID, name string, scopes []string) (string, *APIKey, error) {
	plainKey, hashedKey, err := generateKey()
	if err != nil {
		return "", nil, err
	}

	apiKey, err := s.store.CreateAPIKey(ctx, tenantID, name, hashedKey, scopes)
	if err != nil {
		return "", nil, err
	}

	return plainKey, apiKey, nil
}

// ValidateAPIKey verifies a plain API key by hashing it and looking up the stored record.
func (s *Service) ValidateAPIKey(ctx context.Context, plainKey string) (bool, *APIKey, error) {
	hashedKey := hashKey(plainKey)

	apiKey, err := s.store.GetAPIKeyByHashedKey(ctx, hashedKey)
	if err != nil {
		return false, nil, err
	}

	isValid := validateHashedKey(hashedKey, apiKey.HashedKey)
	return isValid, apiKey, nil
}
