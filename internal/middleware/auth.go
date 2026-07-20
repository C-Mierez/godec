// Package middleware provides Echo middleware for authentication and request validation.
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

// Auth error codes.
const (
	MissingAPIKey = "MISSING_API_KEY"
	InvalidAPIKey = "INVALID_API_KEY"
	ExpiredAPIKey = "EXPIRED_API_KEY"
)

// NewMissingKeyError returns an AuthError for missing API keys.
func NewMissingKeyError() *AuthError {
	return &AuthError{Code: MissingAPIKey, Message: "API key is missing", Status: 401}
}

// NewInvalidKeyError returns an AuthError for invalid API keys.
func NewInvalidKeyError() *AuthError {
	return &AuthError{Code: InvalidAPIKey, Message: "API key is invalid", Status: 401}
}

// NewExpiredKeyError returns an AuthError for expired API keys.
func NewExpiredKeyError() *AuthError {
	return &AuthError{Code: ExpiredAPIKey, Message: "API key is expired", Status: 403}
}

// Context key for storing the validated APIKey in request context
type contextKey string

// ContextKeyAPIKey is the context key used to store the validated API key.
const ContextKeyAPIKey = contextKey("apiKey")

// APIKeyAuthenticator returns an OpenAPI authentication function that validates
// the BasicAuth scheme using the existing API key validator abstraction.
func APIKeyAuthenticator(validator APIKeyValidator) openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		if input == nil || input.RequestValidationInput == nil || input.RequestValidationInput.Request == nil {
			return fmt.Errorf("invalid authentication input")
		}

		if input.SecuritySchemeName != "BasicAuth" {
			return nil
		}

		if validator == nil {
			return fmt.Errorf("api key validator is required")
		}

		tokenID, tokenSecret, ok := input.RequestValidationInput.Request.BasicAuth()
		if !ok {
			return NewMissingKeyError()
		}

		if tokenID == "" || tokenSecret == "" {
			return NewInvalidKeyError()
		}

		apiKey, err := validator.ValidateAPIKey(ctx, tokenID, tokenSecret)
		if err != nil {
			return err
		}

		requestWithKey := input.RequestValidationInput.Request.WithContext(
			context.WithValue(input.RequestValidationInput.Request.Context(), ContextKeyAPIKey, apiKey),
		)
		input.RequestValidationInput.Request = requestWithKey

		return nil
	}
}
