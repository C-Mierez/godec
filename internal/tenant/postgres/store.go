// Package postgres implements the tenant.Store interface using SQLC-generated queries.
package postgres

import (
	"context"

	db "github.com/c-mierez/godec/internal/postgres/db"
	"github.com/c-mierez/godec/internal/tenant"
	"github.com/google/uuid"
)

// Store implements the tenant.Store interface using SQLC-generated queries.
type Store struct {
	queries *db.Queries
}

// NewStore creates a Store backed by the given SQLC queries.
func NewStore(queries *db.Queries) *Store {
	return &Store{queries: queries}
}

// CreateTenant inserts a new tenant and returns the persisted record.
func (s *Store) CreateTenant(ctx context.Context, name, email string) (*tenant.Tenant, error) {
	row, err := s.queries.CreateTenant(ctx, db.CreateTenantParams{Name: name, Email: email})
	if err != nil {
		return nil, err
	}

	domain := tenantFromRow(row)
	return &domain, nil
}

// GetTenantByID looks up a tenant by its UUID.
func (s *Store) GetTenantByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	row, err := s.queries.GetTenantByID(ctx, db.UUIDToPGUUID(id))
	if err != nil {
		return nil, err
	}

	domain := tenantFromRow(row)
	return &domain, nil
}

// GetActiveTenantByEmail looks up the active tenant with the given email.
func (s *Store) GetActiveTenantByEmail(ctx context.Context, email string) (*tenant.Tenant, error) {
	row, err := s.queries.GetActiveTenantByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	domain := tenantFromRow(row)
	return &domain, nil
}

// ListTenants returns a paginated list of tenants ordered by creation time.
func (s *Store) ListTenants(ctx context.Context, limit, offset int32) ([]*tenant.Tenant, error) {
	rows, err := s.queries.ListTenants(ctx, db.ListTenantsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	return tenantSliceFromRows(rows), nil
}

// ListTenantsByStatus returns a paginated list of tenants filtered by status.
func (s *Store) ListTenantsByStatus(ctx context.Context, status tenant.Status, limit, offset int32) ([]*tenant.Tenant, error) {
	rows, err := s.queries.ListTenantsByStatus(ctx, db.ListTenantsByStatusParams{Status: string(status), Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	return tenantSliceFromRows(rows), nil
}

// ListTenantsByEmail returns all tenants with the given email.
func (s *Store) ListTenantsByEmail(ctx context.Context, email string) ([]*tenant.Tenant, error) {
	rows, err := s.queries.ListTenantsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return tenantSliceFromRows(rows), nil
}

// UpdateTenant updates the name and email of an existing tenant.
func (s *Store) UpdateTenant(ctx context.Context, id uuid.UUID, name, email string) (*tenant.Tenant, error) {
	row, err := s.queries.UpdateTenant(ctx, db.UpdateTenantParams{ID: db.UUIDToPGUUID(id), Name: name, Email: email})
	if err != nil {
		return nil, err
	}

	domain := tenantFromRow(row)
	return &domain, nil
}

// SetTenantStatus changes the status of a tenant.
func (s *Store) SetTenantStatus(ctx context.Context, id uuid.UUID, status tenant.Status) (*tenant.Tenant, error) {
	row, err := s.queries.SetTenantStatus(ctx, db.SetTenantStatusParams{ID: db.UUIDToPGUUID(id), Status: string(status)})
	if err != nil {
		return nil, err
	}

	domain := tenantFromRow(row)
	return &domain, nil
}

// CountTenants returns the total number of tenants.
func (s *Store) CountTenants(ctx context.Context) (int64, error) {
	return s.queries.CountTenants(ctx)
}

// CountTenantsByStatus returns the number of tenants with the given status.
func (s *Store) CountTenantsByStatus(ctx context.Context, status tenant.Status) (int64, error) {
	return s.queries.CountTenantsByStatus(ctx, string(status))
}
