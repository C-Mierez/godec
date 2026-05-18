package apikey

import (
	"time"

	"github.com/google/uuid"
)

type ApiKey struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Name       string
	HashedKey  string
	Scopes     []string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
}
