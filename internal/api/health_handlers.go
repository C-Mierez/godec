package api

import (
	"context"
	"time"

	"github.com/c-mierez/godec/pkg/timeutil"
)

// Pinger is the minimum interface for a dependency that can be health-checked.
// *pgxpool.Pool satisfies this via its Ping method.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Health check status values.
const (
	healthStatusOK    = "ok"
	healthStatusError = "error"
)

// Well-known checker names. Use these as keys in the checkers map
// passed to NewHealthHandlers.
const (
	CheckerDatabase = "database"
)

// How long each dependency check may take before being declared unhealthy.
const healthCheckTimeout = 2 * time.Second

// HealthHandlers implements health check endpoints.
type HealthHandlers struct {
	checkers map[string]Pinger
}

// NewHealthHandlers creates a HealthHandlers with the given named checkers.
// Each key is the dependency name reported in the checks map.
func NewHealthHandlers(checkers map[string]Pinger) *HealthHandlers {
	return &HealthHandlers{checkers: checkers}
}

// Liveness returns a liveness probe response.
func (h *HealthHandlers) Liveness(_ context.Context, _ LivenessRequestObject) (LivenessResponseObject, error) {
	return Liveness200JSONResponse(HealthResponse{
		Status:    healthStatusOK,
		Timestamp: timeutil.TimeNow(),
	}), nil
}

// Readiness returns a readiness probe response.
func (h *HealthHandlers) Readiness(ctx context.Context, _ ReadinessRequestObject) (ReadinessResponseObject, error) {
	checks := map[string]string{}
	healthy := true

	for name, checker := range h.checkers {
		pingCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
		err := checker.Ping(pingCtx)
		cancel()

		if err != nil {
			checks[name] = err.Error()
			healthy = false
		} else {
			checks[name] = healthStatusOK
		}
	}

	if !healthy {
		return Readiness503JSONResponse(HealthResponse{
			Status:    healthStatusError,
			Timestamp: timeutil.TimeNow(),
			Checks:    &checks,
		}), nil
	}

	return Readiness200JSONResponse(HealthResponse{
		Status:    healthStatusOK,
		Timestamp: timeutil.TimeNow(),
		Checks:    &checks,
	}), nil
}
