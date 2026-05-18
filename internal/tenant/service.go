package tenant

import (
	"context"

	"github.com/google/uuid"
)

type Store interface {
	CreateTenant(ctx context.Context, name, email string) (*Tenant, error)
	GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetActiveTenantByEmail(ctx context.Context, email string) (*Tenant, error)
	ListTenants(ctx context.Context, limit, offset int32) ([]*Tenant, error)
	ListTenantsByStatus(ctx context.Context, status TenantStatus, limit, offset int32) ([]*Tenant, error)
	ListTenantsByEmail(ctx context.Context, email string) ([]*Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, name, email string) (*Tenant, error)
	SetTenantStatus(ctx context.Context, id uuid.UUID, status TenantStatus) (*Tenant, error)
	CountTenants(ctx context.Context) (int64, error)
	CountTenantsByStatus(ctx context.Context, status TenantStatus) (int64, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateTenant(ctx context.Context, name, email string) (*Tenant, error) {
	return s.store.CreateTenant(ctx, name, email)
}

func (s *Service) GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	return s.store.GetTenantByID(ctx, id)
}

func (s *Service) GetActiveTenantByEmail(ctx context.Context, email string) (*Tenant, error) {
	return s.store.GetActiveTenantByEmail(ctx, email)
}

func (s *Service) ListTenants(ctx context.Context, limit, offset int32) ([]*Tenant, error) {
	return s.store.ListTenants(ctx, limit, offset)
}

func (s *Service) ListTenantsByStatus(ctx context.Context, status TenantStatus, limit, offset int32) ([]*Tenant, error) {
	return s.store.ListTenantsByStatus(ctx, status, limit, offset)
}

func (s *Service) ListTenantsByEmail(ctx context.Context, email string) ([]*Tenant, error) {
	return s.store.ListTenantsByEmail(ctx, email)
}

func (s *Service) UpdateTenant(ctx context.Context, id uuid.UUID, name, email string) (*Tenant, error) {
	return s.store.UpdateTenant(ctx, id, name, email)
}

func (s *Service) SetTenantStatus(ctx context.Context, id uuid.UUID, status TenantStatus) (*Tenant, error) {
	return s.store.SetTenantStatus(ctx, id, status)
}

func (s *Service) CountTenants(ctx context.Context) (int64, error) {
	return s.store.CountTenants(ctx)
}

func (s *Service) CountTenantsByStatus(ctx context.Context, status TenantStatus) (int64, error) {
	return s.store.CountTenantsByStatus(ctx, status)
}
