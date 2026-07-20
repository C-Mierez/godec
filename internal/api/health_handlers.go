package api

import (
	"context"

	"github.com/c-mierez/godec/pkg/timeutil"
)

// HealthHandlers implements health check endpoints.
type HealthHandlers struct{}

// NewHealthHandlers creates a HealthHandlers.
func NewHealthHandlers() *HealthHandlers {
	return &HealthHandlers{}
}

// Liveness returns a liveness probe response.
func (h *HealthHandlers) Liveness(_ context.Context, _ LivenessRequestObject) (LivenessResponseObject, error) {
	return Liveness200JSONResponse(HealthResponse{
		Status:    "ok",
		Timestamp: timeutil.TimeNow(),
	}), nil
}

// Readiness returns a readiness probe response.
func (h *HealthHandlers) Readiness(_ context.Context, _ ReadinessRequestObject) (ReadinessResponseObject, error) {
	return Readiness200JSONResponse(HealthResponse{
		Status:    "ok",
		Timestamp: timeutil.TimeNow(),
	}), nil
}
