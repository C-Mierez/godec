package middleware

import (
	"context"
	"log"
	"time"

	"github.com/c-mierez/godec/internal/apikey"
)

// APIKeyValidator defines the interface used by the OAPI authentication glue.
type APIKeyValidator interface {
	ValidateAPIKey(ctx context.Context, key string) (*apikey.ApiKey, error)
}

type apiKeyValidatorImpl struct {
	svc *apikey.Service
}

// NewAPIKeyValidator adapts the real apikey.Service to the APIKeyValidator
// interface expected by the OAPI AuthenticationFunc.
func NewAPIKeyValidator(svc *apikey.Service) APIKeyValidator {
	return &apiKeyValidatorImpl{svc: svc}
}

func (a *apiKeyValidatorImpl) ValidateAPIKey(ctx context.Context, key string) (*apikey.ApiKey, error) {
	isValid, ak, err := a.svc.ValidateAPIKey(ctx, key)
	if err != nil {
		log.Printf("apikey validation failed: %v", err)
		return nil, NewInvalidKeyError()
	}

	if !isValid || ak == nil {
		return nil, NewInvalidKeyError()
	}

	// Check expiration if present
	if ak.ExpiresAt != nil {
		if time.Now().After(*ak.ExpiresAt) {
			return nil, NewExpiredKeyError()
		}
	}

	return ak, nil
}
