package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/c-mierez/godec/internal/apikey"
	"github.com/getkin/kin-openapi/openapi3filter"
)

type stubAPIKeyValidator struct {
	called bool
	key    string
	apiKey *apikey.ApiKey
	err    error
}

func (s *stubAPIKeyValidator) ValidateAPIKey(_ context.Context, key string) (*apikey.ApiKey, error) {
	s.called = true
	s.key = key
	if s.err != nil {
		return nil, s.err
	}
	return s.apiKey, nil
}

func TestAPIKeyAuthenticator_IgnoresOtherSecuritySchemes(t *testing.T) {
	validator := &stubAPIKeyValidator{}
	auth := APIKeyAuthenticator(validator)
	req := httptest.NewRequest("GET", "/media/upload-url", nil)

	err := auth(context.Background(), &openapi3filter.AuthenticationInput{
		SecuritySchemeName: "OtherScheme",
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: req,
		},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if validator.called {
		t.Fatalf("validator should not be called for non-ApiKeyAuth schemes")
	}
}

func TestAPIKeyAuthenticator_ReturnsMissingKeyError(t *testing.T) {
	validator := &stubAPIKeyValidator{}
	auth := APIKeyAuthenticator(validator)
	req := httptest.NewRequest("GET", "/media/upload-url", nil)

	err := auth(context.Background(), &openapi3filter.AuthenticationInput{
		SecuritySchemeName: "ApiKeyAuth",
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: req,
		},
	})

	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %v", err)
	}
	if authErr.Code != MissingAPIKey {
		t.Fatalf("expected %s, got %s", MissingAPIKey, authErr.Code)
	}
}

func TestAPIKeyAuthenticator_ReturnsExpiredKeyError(t *testing.T) {
	validator := &stubAPIKeyValidator{err: NewExpiredKeyError()}
	auth := APIKeyAuthenticator(validator)
	req := httptest.NewRequest("GET", "/media/upload-url", nil)
	req.Header.Set("X-API-Key", "some-key")

	err := auth(context.Background(), &openapi3filter.AuthenticationInput{
		SecuritySchemeName: "ApiKeyAuth",
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: req,
		},
	})

	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got %v", err)
	}
	if authErr.Code != ExpiredAPIKey {
		t.Fatalf("expected code %s, got %s", ExpiredAPIKey, authErr.Code)
	}
	if authErr.Status != 403 {
		t.Fatalf("expected status 403, got %d", authErr.Status)
	}
}

func TestAPIKeyAuthenticator_SetsAPIKeyInRequestContext(t *testing.T) {
	validatedAPIKey := &apikey.ApiKey{Name: "test-key"}
	validator := &stubAPIKeyValidator{apiKey: validatedAPIKey}
	auth := APIKeyAuthenticator(validator)
	req := httptest.NewRequest("GET", "/media/upload-url", nil)
	req.Header.Set("X-API-Key", "plain-key")
	input := &openapi3filter.AuthenticationInput{
		SecuritySchemeName: "ApiKeyAuth",
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: req,
		},
	}

	err := auth(context.Background(), input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !validator.called {
		t.Fatalf("expected validator to be called")
	}
	if validator.key != "plain-key" {
		t.Fatalf("expected validator key plain-key, got %s", validator.key)
	}

	stored, ok := input.RequestValidationInput.Request.Context().Value(ContextKeyApiKey).(*apikey.ApiKey)
	if !ok || stored == nil {
		t.Fatalf("expected API key in request context")
	}
	if stored.Name != validatedAPIKey.Name {
		t.Fatalf("expected context API key name %s, got %s", validatedAPIKey.Name, stored.Name)
	}
}
