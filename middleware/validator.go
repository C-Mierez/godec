package middleware

import (
	"context"
	"time"

	"github.com/c-mierez/godec/internal/apikey"
)

// APIKeyValidator defines the interface used by the OAPI authentication glue.
type APIKeyValidator interface {
	ValidateAPIKey(ctx context.Context, tokenID, secret string) (*apikey.APIKey, error)
}

type apiKeyValidator struct {
	apikeyService *apikey.Service
}

// NewAPIKeyValidator adapts the real apikey.Service to the APIKeyValidator
// interface expected by the OAPI AuthenticationFunc.
func NewAPIKeyValidator(svc *apikey.Service) APIKeyValidator {
	return &apiKeyValidator{apikeyService: svc}
}

func (a *apiKeyValidator) ValidateAPIKey(ctx context.Context, tokenID, secret string) (*apikey.APIKey, error) {
	ak, err := a.apikeyService.ValidateAPIKey(ctx, tokenID, secret)
	if err != nil {
		return nil, NewInvalidKeyError()
	}

	if ak == nil {
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
