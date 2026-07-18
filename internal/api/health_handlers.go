package api

import (
	"context"

	"github.com/c-mierez/godec/pkg/timeutil"
)

type HealthHandlers struct{}

func NewHealthHandlers() *HealthHandlers {
	return &HealthHandlers{}
}

func (h *HealthHandlers) Liveness(ctx context.Context, request LivenessRequestObject) (LivenessResponseObject, error) {
	return Liveness200JSONResponse(HealthResponse{
		Status:    "ok",
		Timestamp: timeutil.TimeNow(),
	}), nil
}

func (h *HealthHandlers) Readiness(ctx context.Context, request ReadinessRequestObject) (ReadinessResponseObject, error) {
	return Readiness200JSONResponse(HealthResponse{
		Status:    "ok",
		Timestamp: timeutil.TimeNow(),
	}), nil
}
