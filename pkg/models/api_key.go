// Package models defines domain types used throughout the application.
package models

import (
	"time"

	"github.com/google/uuid"
)

// ApiKey represents an API key in the domain layer.
// It contains both the API key metadata and access control information.
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
