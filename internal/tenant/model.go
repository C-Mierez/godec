// Package tenant defines the tenant domain model and service interface.
package tenant

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the lifecycle state of a tenant.
type Status string

// Tenant status constants.
const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

// Tenant represents a multi-tenant customer record.
type Tenant struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}
