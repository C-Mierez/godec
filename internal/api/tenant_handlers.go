package api

import (
	"context"
	"math"

	"github.com/c-mierez/godec/internal/tenant"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

// TenantHandlers implements tenant management endpoints.
type TenantHandlers struct {
	service *tenant.Service
}

// NewTenantHandlers creates a TenantHandlers backed by the given service.
func NewTenantHandlers(service *tenant.Service) *TenantHandlers {
	return &TenantHandlers{
		service: service,
	}
}

// CreateTenant handles tenant creation requests.
func (h *TenantHandlers) CreateTenant(ctx context.Context, request CreateTenantRequestObject) (CreateTenantResponseObject, error) {
	if request.Body == nil {
		return CreateTenant400JSONResponse{BadRequestJSONResponse{Error: "missing request body"}}, nil
	}

	domainTenant, err := h.service.CreateTenant(ctx, request.Body.Name, string(request.Body.Email))
	if err != nil {
		return CreateTenant400JSONResponse{BadRequestJSONResponse{Error: err.Error()}}, nil
	}

	return CreateTenant201JSONResponse(h.domainTenantToAPI(domainTenant)), nil
}

// ListTenants handles paginated tenant listing requests.
func (h *TenantHandlers) ListTenants(ctx context.Context, request ListTenantsRequestObject) (ListTenantsResponseObject, error) {
	limit := int32(10)
	offset := int32(0)

	if request.Params.Limit != nil {
		limitVal := *request.Params.Limit
		if limitVal < 0 || limitVal > math.MaxInt32 { //nolint:gosec // G115: bounds checked explicitly
			return ListTenants400JSONResponse{BadRequestJSONResponse{Error: "limit exceeds valid range"}}, nil
		}
		limit = int32(limitVal)
	}
	if request.Params.Offset != nil {
		offsetVal := *request.Params.Offset
		if offsetVal < 0 || offsetVal > math.MaxInt32 { //nolint:gosec // G115: bounds checked explicitly
			return ListTenants400JSONResponse{BadRequestJSONResponse{Error: "offset exceeds valid range"}}, nil
		}
		offset = int32(offsetVal)
	}

	tenants, err := h.service.ListTenants(ctx, limit, offset)
	if err != nil {
		return ListTenants400JSONResponse{BadRequestJSONResponse{Error: err.Error()}}, nil
	}

	total, err := h.service.CountTenants(ctx)
	if err != nil {
		return ListTenants400JSONResponse{BadRequestJSONResponse{Error: err.Error()}}, nil
	}

	apiTenants := make([]Tenant, len(tenants))
	for i, t := range tenants {
		apiTenants[i] = h.domainTenantToAPI(t)
	}

	return ListTenants200JSONResponse(TenantListResponse{
		Items: apiTenants,
		Total: total,
	}), nil
}

// GetTenant handles tenant lookup requests by ID.
func (h *TenantHandlers) GetTenant(ctx context.Context, request GetTenantRequestObject) (GetTenantResponseObject, error) {
	domainTenant, err := h.service.GetTenantByID(ctx, uuid.UUID(request.Id))
	if err != nil {
		return GetTenant404JSONResponse{NotFoundJSONResponse{Error: "tenant not found"}}, nil
	}

	return GetTenant200JSONResponse(h.domainTenantToAPI(domainTenant)), nil
}

// SetTenantStatus handles tenant status update requests.
func (h *TenantHandlers) SetTenantStatus(ctx context.Context, request SetTenantStatusRequestObject) (SetTenantStatusResponseObject, error) {
	if request.Body == nil {
		return SetTenantStatus400JSONResponse{BadRequestJSONResponse{Error: "missing request body"}}, nil
	}

	domainStatus := tenant.Status(request.Body.Status)
	updatedTenant, err := h.service.SetTenantStatus(ctx, uuid.UUID(request.Id), domainStatus)
	if err != nil {
		return SetTenantStatus400JSONResponse{BadRequestJSONResponse{Error: err.Error()}}, nil
	}

	return SetTenantStatus200JSONResponse(h.domainTenantToAPI(updatedTenant)), nil
}

// domainTenantToAPI converts domain tenant.Tenant to API Tenant type
func (h *TenantHandlers) domainTenantToAPI(t *tenant.Tenant) Tenant {
	if t == nil {
		return Tenant{}
	}

	return Tenant{
		Id:        types.UUID(t.ID),
		Name:      t.Name,
		Email:     types.Email(t.Email),
		Status:    TenantStatus(t.Status),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
