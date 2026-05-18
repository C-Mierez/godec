package apikey

import (
	"context"

	"github.com/google/uuid"
)

type Store interface {
	CreateApiKey(ctx context.Context, tenantID uuid.UUID, name, hashedKey string, scopes []string) (*ApiKey, error)
	GetApiKeyByHashedKey(ctx context.Context, hashedKey string) (*ApiKey, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) GenerateApiKey(ctx context.Context, tenantID uuid.UUID, name string, scopes []string) (string, *ApiKey, error) {
	plainKey, hashedKey, err := generateKey()
	if err != nil {
		return "", nil, err
	}

	apiKey, err := s.store.CreateApiKey(ctx, tenantID, name, hashedKey, scopes)
	if err != nil {
		return "", nil, err
	}

	return plainKey, apiKey, nil
}

func (s *Service) ValidateApiKey(ctx context.Context, plainKey string) (bool, *ApiKey, error) {
	hashedKey := hashKey(plainKey)

	apiKey, err := s.store.GetApiKeyByHashedKey(ctx, hashedKey)
	if err != nil {
		return false, nil, err
	}

	isValid := validateHashedKey(hashedKey, apiKey.HashedKey)
	return isValid, apiKey, nil
}
