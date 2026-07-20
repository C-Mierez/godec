package apikey

import (
	"time"

	"github.com/google/uuid"
)

// APIKey represents an API key with its metadata, scopes, and expiration.
type APIKey struct {
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
