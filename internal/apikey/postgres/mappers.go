package postgres

import (
	"github.com/c-mierez/godec/internal/apikey"
	db "github.com/c-mierez/godec/internal/postgres/db"
)

func apiKeyFromRow(row db.ApiKey) apikey.APIKey {
	return apikey.APIKey{
		ID:           db.PgUUIDToUUID(row.ID),
		TenantID:     db.PgUUIDToUUID(row.TenantID),
		Name:         row.Name,
		TokenID:      row.TokenID,
		HashedSecret: row.HashedSecret,
		Scopes:       append([]string(nil), row.Scopes...),
		CreatedAt:    db.PgTimestamptzToTime(row.CreatedAt),
		UpdatedAt:    db.PgTimestamptzToTime(row.UpdatedAt),
		LastUsedAt:   db.PgTimestamptzToOptionalTime(row.LastUsedAt),
		ExpiresAt:    db.PgTimestamptzToOptionalTime(row.ExpiresAt),
	}
}

func apiKeySliceFromRows(rows []db.ApiKey) []*apikey.APIKey {
	keys := make([]*apikey.APIKey, len(rows))
	for i, row := range rows {
		domain := apiKeyFromRow(row)
		keys[i] = &domain
	}

	return keys
}
