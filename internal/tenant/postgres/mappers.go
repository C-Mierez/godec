package postgres

import (
	"github.com/c-mierez/godec/internal/tenant"
	db "github.com/c-mierez/godec/internal/postgres/db"
)

func tenantFromRow(row db.Tenant) tenant.Tenant {
	return tenant.Tenant{
		ID:        db.PgUUIDToUUID(row.ID),
		Name:      row.Name,
		Email:     row.Email,
		Status:    tenant.Status(row.Status),
		CreatedAt: db.PgTimestamptzToTime(row.CreatedAt),
		UpdatedAt: db.PgTimestamptzToTime(row.UpdatedAt),
	}
}

func tenantSliceFromRows(rows []db.Tenant) []*tenant.Tenant {
	tenants := make([]*tenant.Tenant, len(rows))
	for i, row := range rows {
		domain := tenantFromRow(row)
		tenants[i] = &domain
	}

	return tenants
}
