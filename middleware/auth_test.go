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
	called   bool
	tokenID  string
	secret   string
	apiKey   *apikey.APIKey
	err      error
}

func (s *stubAPIKeyValidator) ValidateAPIKey(_ context.Context, tokenID, secret string) (*apikey.APIKey, error) {
	s.called = true
	s.tokenID = tokenID
	s.secret = secret
	if s.err != nil {
		return nil, s.err
	}
	return s.apiKey, nil
}

func TestAPIKeyAuthenticator_IgnoresOtherSecuritySchemes(t *testing.T) {
	validator := &stubAPIKeyValidator{}
	auth := APIKeyAuthenticator(validator)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/media/upload-url", nil)

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
		t.Fatalf("validator should not be called for non-BasicAuth schemes")
	}
}

func TestAPIKeyAuthenticator_ReturnsMissingKeyError(t *testing.T) {
	validator := &stubAPIKeyValidator{}
	auth := APIKeyAuthenticator(validator)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/media/upload-url", nil)

	err := auth(context.Background(), &openapi3filter.AuthenticationInput{
		SecuritySchemeName: "BasicAuth",
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
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/media/upload-url", nil)
	req.SetBasicAuth("gdk_testtoken", "sk_testsecret")

	err := auth(context.Background(), &openapi3filter.AuthenticationInput{
		SecuritySchemeName: "BasicAuth",
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
	validatedAPIKey := &apikey.APIKey{Name: "test-key"}
	validator := &stubAPIKeyValidator{apiKey: validatedAPIKey}
	auth := APIKeyAuthenticator(validator)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/media/upload-url", nil)
	req.SetBasicAuth("gdk_testtoken", "sk_testsecret")
	input := &openapi3filter.AuthenticationInput{
		SecuritySchemeName: "BasicAuth",
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
	if validator.tokenID != "gdk_testtoken" {
		t.Fatalf("expected validator tokenID gdk_testtoken, got %s", validator.tokenID)
	}
	if validator.secret != "sk_testsecret" {
		t.Fatalf("expected validator secret sk_testsecret, got %s", validator.secret)
	}

	stored, ok := input.RequestValidationInput.Request.Context().Value(ContextKeyAPIKey).(*apikey.APIKey)
	if !ok || stored == nil {
		t.Fatalf("expected API key in request context")
	}
	if stored.Name != validatedAPIKey.Name {
		t.Fatalf("expected context API key name %s, got %s", validatedAPIKey.Name, stored.Name)
	}
}
