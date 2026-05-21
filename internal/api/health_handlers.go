package api

import (
	"context"
	"time"
)

type HealthHandlers struct{}

func NewHealthHandlers() *HealthHandlers {
	return &HealthHandlers{}
}

func (h *HealthHandlers) Liveness(ctx context.Context, request LivenessRequestObject) (LivenessResponseObject, error) {
	return Liveness200JSONResponse(HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
	}), nil
}

func (h *HealthHandlers) Readiness(ctx context.Context, request ReadinessRequestObject) (ReadinessResponseObject, error) {
	return Readiness200JSONResponse(HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
	}), nil
}
