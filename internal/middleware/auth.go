package middleware

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3filter"
)

// AuthError represents an authentication error that can be mapped to the
// AuthErrorResponse schema in the OpenAPI spec.
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"error"`
	Status  int    `json:"-"`
}

func (e *AuthError) Error() string { return e.Message }

const (
	MissingAPIKey  = "MISSING_API_KEY"
	InvalidAPIKey  = "INVALID_API_KEY"
	ExpiredAPIKey  = "EXPIRED_API_KEY"
)

func NewMissingKeyError() *AuthError {
	return &AuthError{Code: MissingAPIKey, Message: "API key is missing", Status: 401}
}

func NewInvalidKeyError() *AuthError {
	return &AuthError{Code: InvalidAPIKey, Message: "API key is invalid", Status: 401}
}

func NewExpiredKeyError() *AuthError {
	return &AuthError{Code: ExpiredAPIKey, Message: "API key is expired", Status: 403}
}

// Context key for storing the validated ApiKey in request context
type contextKey string

const ContextKeyApiKey = contextKey("apiKey")

// APIKeyAuthenticator returns an OpenAPI authentication function that validates
// the ApiKeyAuth scheme using the existing API key validator abstraction.
func APIKeyAuthenticator(validator APIKeyValidator) openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		if input == nil || input.RequestValidationInput == nil || input.RequestValidationInput.Request == nil {
			return fmt.Errorf("invalid authentication input")
		}

		if input.SecuritySchemeName != "ApiKeyAuth" {
			return nil
		}

		if validator == nil {
			return fmt.Errorf("api key validator is required")
		}

		key := input.RequestValidationInput.Request.Header.Get("X-API-Key")
		if key == "" {
			return NewMissingKeyError()
		}

		apiKey, err := validator.ValidateAPIKey(ctx, key)
		if err != nil {
			return err
		}

		requestWithKey := input.RequestValidationInput.Request.WithContext(
			context.WithValue(input.RequestValidationInput.Request.Context(), ContextKeyApiKey, apiKey),
		)
		input.RequestValidationInput.Request = requestWithKey

		return nil
	}
}
