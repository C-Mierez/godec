package api

import (
	"context"

	"github.com/c-mierez/godec/internal/tenant"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

type TenantHandlers struct {
	service *tenant.Service
}

func NewTenantHandlers(service *tenant.Service) *TenantHandlers {
	return &TenantHandlers{
		service: service,
	}
}

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

func (h *TenantHandlers) ListTenants(ctx context.Context, request ListTenantsRequestObject) (ListTenantsResponseObject, error) {
	limit := int32(10)
	offset := int32(0)

	if request.Params.Limit != nil {
		limit = int32(*request.Params.Limit)
	}
	if request.Params.Offset != nil {
		offset = int32(*request.Params.Offset)
	}

	tenants, err := h.service.ListTenants(ctx, limit, offset)
	if err != nil {
		return ListTenants400JSONResponse{BadRequestJSONResponse{Error: err.Error()}}, nil
	}

	apiTenants := make([]Tenant, len(tenants))
	for i, t := range tenants {
		apiTenants[i] = h.domainTenantToAPI(t)
	}

	return ListTenants200JSONResponse(TenantListResponse{
		Items: apiTenants,
	}), nil
}

func (h *TenantHandlers) GetTenant(ctx context.Context, request GetTenantRequestObject) (GetTenantResponseObject, error) {
	domainTenant, err := h.service.GetTenantByID(ctx, uuid.UUID(request.Id))
	if err != nil {
		return GetTenant404JSONResponse{NotFoundJSONResponse{Error: "tenant not found"}}, nil
	}

	return GetTenant200JSONResponse(h.domainTenantToAPI(domainTenant)), nil
}

func (h *TenantHandlers) SetTenantStatus(ctx context.Context, request SetTenantStatusRequestObject) (SetTenantStatusResponseObject, error) {
	if request.Body == nil {
		return SetTenantStatus400JSONResponse{BadRequestJSONResponse{Error: "missing request body"}}, nil
	}

	domainStatus := tenant.TenantStatus(request.Body.Status)
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
