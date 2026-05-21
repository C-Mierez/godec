package tenant

import (
	"time"

	"github.com/google/uuid"
)

type TenantStatus string

const (
	TenantStatusActive   TenantStatus = "active"
	TenantStatusInactive TenantStatus = "inactive"
)

type Tenant struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Status    TenantStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
