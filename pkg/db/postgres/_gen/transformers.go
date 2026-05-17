package postgres

import (
	"time"

	"github.com/c-mierez/godec/pkg/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (tenant Tenant) ToDomain() models.Tenant {
	return models.Tenant{
		ID:        PgUUIDToUUID(tenant.ID),
		Name:      tenant.Name,
		Email:     tenant.Email,
		Status:    models.TenantStatus(tenant.Status),
		CreatedAt: PgTimestamptzToTime(tenant.CreatedAt),
		UpdatedAt: PgTimestamptzToTime(tenant.UpdatedAt),
	}
}

func TenantFromDomain(tenant models.Tenant) Tenant {
	return Tenant{
		ID:        UuidToPGUUID(tenant.ID),
		Name:      tenant.Name,
		Email:     tenant.Email,
		Status:    string(tenant.Status),
		CreatedAt: TimeToPGTimestamptz(tenant.CreatedAt),
		UpdatedAt: TimeToPGTimestamptz(tenant.UpdatedAt),
	}
}

func (apiKey ApiKey) ToDomain() models.ApiKey {
	return models.ApiKey{
		ID:         PgUUIDToUUID(apiKey.ID),
		TenantID:   PgUUIDToUUID(apiKey.TenantID),
		Name:       apiKey.Name,
		HashedKey:  apiKey.HashedKey,
		Scopes:     append([]string(nil), apiKey.Scopes...),
		CreatedAt:  PgTimestamptzToTime(apiKey.CreatedAt),
		UpdatedAt:  PgTimestamptzToTime(apiKey.UpdatedAt),
		LastUsedAt: PgTimestamptzToOptionalTime(apiKey.LastUsedAt),
		ExpiresAt:  PgTimestamptzToOptionalTime(apiKey.ExpiresAt),
	}
}

func ApiKeyFromDomain(apiKey models.ApiKey) ApiKey {
	return ApiKey{
		ID:         UuidToPGUUID(apiKey.ID),
		TenantID:   UuidToPGUUID(apiKey.TenantID),
		Name:       apiKey.Name,
		HashedKey:  apiKey.HashedKey,
		Scopes:     append([]string(nil), apiKey.Scopes...),
		CreatedAt:  TimeToPGTimestamptz(apiKey.CreatedAt),
		UpdatedAt:  TimeToPGTimestamptz(apiKey.UpdatedAt),
		LastUsedAt: OptionalTimeToPGTimestamptz(apiKey.LastUsedAt),
		ExpiresAt:  OptionalTimeToPGTimestamptz(apiKey.ExpiresAt),
	}
}

func PgUUIDToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}

	return uuid.UUID(id.Bytes)
}

func UuidToPGUUID(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{}
	}

	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

func PgTimestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}

	return ts.Time
}

func PgTimestamptzToOptionalTime(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}

	t := ts.Time
	return &t
}

func TimeToPGTimestamptz(ts time.Time) pgtype.Timestamptz {
	if ts.IsZero() {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{
		Time:  ts,
		Valid: true,
	}
}

func OptionalTimeToPGTimestamptz(ts *time.Time) pgtype.Timestamptz {
	if ts == nil {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{
		Time:  *ts,
		Valid: true,
	}
}
