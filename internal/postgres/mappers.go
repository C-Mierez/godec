package postgres

import (
	"github.com/c-mierez/godec/internal/apikey"
	db "github.com/c-mierez/godec/internal/postgres/db"
	"github.com/c-mierez/godec/internal/tenant"
)

func apiKeyFromRow(row db.ApiKey) apikey.ApiKey {
	return apikey.ApiKey{
		ID:         db.PgUUIDToUUID(row.ID),
		TenantID:   db.PgUUIDToUUID(row.TenantID),
		Name:       row.Name,
		HashedKey:  row.HashedKey,
		Scopes:     append([]string(nil), row.Scopes...),
		CreatedAt:  db.PgTimestamptzToTime(row.CreatedAt),
		UpdatedAt:  db.PgTimestamptzToTime(row.UpdatedAt),
		LastUsedAt: db.PgTimestamptzToOptionalTime(row.LastUsedAt),
		ExpiresAt:  db.PgTimestamptzToOptionalTime(row.ExpiresAt),
	}
}

func tenantFromRow(row db.Tenant) tenant.Tenant {
	return tenant.Tenant{
		ID:        db.PgUUIDToUUID(row.ID),
		Name:      row.Name,
		Email:     row.Email,
		Status:    tenant.TenantStatus(row.Status),
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
