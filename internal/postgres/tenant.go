package postgres

import (
	"context"

	db "github.com/c-mierez/godec/internal/postgres/db"
	"github.com/c-mierez/godec/internal/tenant"
	"github.com/google/uuid"
)

type TenantStore struct {
	queries *db.Queries
}

func NewTenantStore(queries *db.Queries) *TenantStore {
	return &TenantStore{queries: queries}
}

func (s *TenantStore) CreateTenant(ctx context.Context, name, email string) (*tenant.Tenant, error) {
	row, err := s.queries.CreateTenant(ctx, db.CreateTenantParams{Name: name, Email: email})
	if err != nil {
		return nil, err
	}

	domain := tenantFromRow(row)
	return &domain, nil
}

func (s *TenantStore) GetTenantByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	row, err := s.queries.GetTenantByID(ctx, db.UuidToPGUUID(id))
	if err != nil {
		return nil, err
	}

	domain := tenantFromRow(row)
	return &domain, nil
}

func (s *TenantStore) GetActiveTenantByEmail(ctx context.Context, email string) (*tenant.Tenant, error) {
	row, err := s.queries.GetActiveTenantByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	domain := tenantFromRow(row)
	return &domain, nil
}

func (s *TenantStore) ListTenants(ctx context.Context, limit, offset int32) ([]*tenant.Tenant, error) {
	rows, err := s.queries.ListTenants(ctx, db.ListTenantsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	return tenantSliceFromRows(rows), nil
}

func (s *TenantStore) ListTenantsByStatus(ctx context.Context, status tenant.TenantStatus, limit, offset int32) ([]*tenant.Tenant, error) {
	rows, err := s.queries.ListTenantsByStatus(ctx, db.ListTenantsByStatusParams{Status: string(status), Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	return tenantSliceFromRows(rows), nil
}

func (s *TenantStore) ListTenantsByEmail(ctx context.Context, email string) ([]*tenant.Tenant, error) {
	rows, err := s.queries.ListTenantsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return tenantSliceFromRows(rows), nil
}

func (s *TenantStore) UpdateTenant(ctx context.Context, id uuid.UUID, name, email string) (*tenant.Tenant, error) {
	row, err := s.queries.UpdateTenant(ctx, db.UpdateTenantParams{ID: db.UuidToPGUUID(id), Name: name, Email: email})
	if err != nil {
		return nil, err
	}

	domain := tenantFromRow(row)
	return &domain, nil
}

func (s *TenantStore) SetTenantStatus(ctx context.Context, id uuid.UUID, status tenant.TenantStatus) (*tenant.Tenant, error) {
	row, err := s.queries.SetTenantStatus(ctx, db.SetTenantStatusParams{ID: db.UuidToPGUUID(id), Status: string(status)})
	if err != nil {
		return nil, err
	}

	domain := tenantFromRow(row)
	return &domain, nil
}

func (s *TenantStore) CountTenants(ctx context.Context) (int64, error) {
	return s.queries.CountTenants(ctx)
}

func (s *TenantStore) CountTenantsByStatus(ctx context.Context, status tenant.TenantStatus) (int64, error) {
	return s.queries.CountTenantsByStatus(ctx, string(status))
}
