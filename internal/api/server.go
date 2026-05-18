package api

import (
	"context"

	"github.com/c-mierez/godec/internal/apikey"
	"github.com/c-mierez/godec/internal/tenant"
)

type Server struct {
	tenants *TenantHandlers
	apikeys *APIKeyHandlers
	health  *HealthHandlers
	media   *MediaHandlers
}

func NewServer(tenantService *tenant.Service, apiKeyService *apikey.Service) *Server {
	return &Server{
		tenants: NewTenantHandlers(tenantService),
		apikeys: NewAPIKeyHandlers(apiKeyService),
		health:  NewHealthHandlers(),
		media:   NewMediaHandlers(),
	}
}

// Liveness delegates to health handlers
func (s *Server) Liveness(ctx context.Context, request LivenessRequestObject) (LivenessResponseObject, error) {
	return s.health.Liveness(ctx, request)
}

// Readiness delegates to health handlers
func (s *Server) Readiness(ctx context.Context, request ReadinessRequestObject) (ReadinessResponseObject, error) {
	return s.health.Readiness(ctx, request)
}

// CreateApiKey delegates to apikey handlers
func (s *Server) CreateApiKey(ctx context.Context, request CreateApiKeyRequestObject) (CreateApiKeyResponseObject, error) {
	return s.apikeys.CreateApiKey(ctx, request)
}

// GetMediaUploadURL delegates to media handlers
func (s *Server) GetMediaUploadURL(ctx context.Context, request GetMediaUploadURLRequestObject) (GetMediaUploadURLResponseObject, error) {
	return s.media.GetMediaUploadURL(ctx, request)
}

// ListTenants delegates to tenant handlers
func (s *Server) ListTenants(ctx context.Context, request ListTenantsRequestObject) (ListTenantsResponseObject, error) {
	return s.tenants.ListTenants(ctx, request)
}

// CreateTenant delegates to tenant handlers
func (s *Server) CreateTenant(ctx context.Context, request CreateTenantRequestObject) (CreateTenantResponseObject, error) {
	return s.tenants.CreateTenant(ctx, request)
}

// GetTenant delegates to tenant handlers
func (s *Server) GetTenant(ctx context.Context, request GetTenantRequestObject) (GetTenantResponseObject, error) {
	return s.tenants.GetTenant(ctx, request)
}

// SetTenantStatus delegates to tenant handlers
func (s *Server) SetTenantStatus(ctx context.Context, request SetTenantStatusRequestObject) (SetTenantStatusResponseObject, error) {
	return s.tenants.SetTenantStatus(ctx, request)
}
