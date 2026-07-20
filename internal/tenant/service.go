package tenant

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the persistence operations required by the tenant service.
type Store interface {
	CreateTenant(ctx context.Context, name, email string) (*Tenant, error)
	GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetActiveTenantByEmail(ctx context.Context, email string) (*Tenant, error)
	ListTenants(ctx context.Context, limit, offset int32) ([]*Tenant, error)
	ListTenantsByStatus(ctx context.Context, status Status, limit, offset int32) ([]*Tenant, error)
	ListTenantsByEmail(ctx context.Context, email string) ([]*Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, name, email string) (*Tenant, error)
	SetTenantStatus(ctx context.Context, id uuid.UUID, status Status) (*Tenant, error)
	CountTenants(ctx context.Context) (int64, error)
	CountTenantsByStatus(ctx context.Context, status Status) (int64, error)
}

// Service implements tenant management business logic.
type Service struct {
	store Store
}

// NewService creates a tenant service backed by the given store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// CreateTenant creates a new tenant with the given name and email.
func (s *Service) CreateTenant(ctx context.Context, name, email string) (*Tenant, error) {
	return s.store.CreateTenant(ctx, name, email)
}

// GetTenantByID retrieves a tenant by its UUID.
func (s *Service) GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	return s.store.GetTenantByID(ctx, id)
}

// GetActiveTenantByEmail retrieves the active tenant with the given email.
func (s *Service) GetActiveTenantByEmail(ctx context.Context, email string) (*Tenant, error) {
	return s.store.GetActiveTenantByEmail(ctx, email)
}

// ListTenants returns a paginated list of tenants.
func (s *Service) ListTenants(ctx context.Context, limit, offset int32) ([]*Tenant, error) {
	return s.store.ListTenants(ctx, limit, offset)
}

// ListTenantsByStatus returns a paginated list of tenants filtered by status.
func (s *Service) ListTenantsByStatus(ctx context.Context, status Status, limit, offset int32) ([]*Tenant, error) {
	return s.store.ListTenantsByStatus(ctx, status, limit, offset)
}

// ListTenantsByEmail returns all tenants with the given email.
func (s *Service) ListTenantsByEmail(ctx context.Context, email string) ([]*Tenant, error) {
	return s.store.ListTenantsByEmail(ctx, email)
}

// UpdateTenant updates the name and email of an existing tenant.
func (s *Service) UpdateTenant(ctx context.Context, id uuid.UUID, name, email string) (*Tenant, error) {
	return s.store.UpdateTenant(ctx, id, name, email)
}

// SetTenantStatus changes the status of a tenant.
func (s *Service) SetTenantStatus(ctx context.Context, id uuid.UUID, status Status) (*Tenant, error) {
	return s.store.SetTenantStatus(ctx, id, status)
}

// CountTenants returns the total number of tenants.
func (s *Service) CountTenants(ctx context.Context) (int64, error) {
	return s.store.CountTenants(ctx)
}

// CountTenantsByStatus returns the number of tenants with the given status.
func (s *Service) CountTenantsByStatus(ctx context.Context, status Status) (int64, error) {
	return s.store.CountTenantsByStatus(ctx, status)
}
