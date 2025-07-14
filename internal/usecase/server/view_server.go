package server

import (
	"context"
	"fmt"

	"github.com/lits-06/vcs-sms/internal/domain/entities"
	"github.com/lits-06/vcs-sms/internal/domain/repositories"
)

// ViewServerUseCase handles server viewing operations
type ViewServerUseCase struct {
	serverRepo repositories.ServerRepository
}

// NewViewServerUseCase creates a new view server use case
func NewViewServerUseCase(serverRepo repositories.ServerRepository) *ViewServerUseCase {
	return &ViewServerUseCase{
		serverRepo: serverRepo,
	}
}

type ViewServerRequest struct {
	Filter     entities.ServerFilter     `json:"filter"`
	Pagination entities.ServerPagination `json:"pagination"`
	Sort       entities.ServerSort       `json:"sort"`
}

type ViewServerResponse struct {
	Servers []entities.Server `json:"servers"`
	Total   int               `json:"total"`
}

// ViewServers retrieves servers with filtering, sorting, and pagination
func (uc *ViewServerUseCase) ViewServer(ctx context.Context, req ViewServerRequest) (*ViewServerResponse, error) {
	// // Try cache first
	// cacheKey := fmt.Sprintf("servers:list:%+v", req)
	// var response domain.ServerListResponse
	// err := uc.repos.Cache.Get(ctx, cacheKey, &response)
	// if err == nil {
	// 	return &response, nil
	// }

	// Set default values
	if req.Pagination.Size <= 0 {
		req.Pagination.Size = 10 // default page size
	}
	if req.Pagination.Size > 100 {
		req.Pagination.Size = 100 // max page size
	}

	// Calculate offset from page or use from directly
	if req.Pagination.Page > 0 {
		req.Pagination.From = (req.Pagination.Page - 1) * req.Pagination.Size
	}

	// Get servers from repository
	servers, total, err := uc.serverRepo.List(ctx, req.Filter, req.Sort, req.Pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	// Prepare response
	response := &ViewServerResponse{
		Servers: servers,
		Total:   total,
	}

	// Cache the result for 5 minutes
	// uc.repos.Cache.Set(ctx, cacheKey, response, 5*60)

	return response, nil
}

func (uc *ViewServerUseCase) validateSortField(field string) error {
	validFields := map[string]bool{
		"id":           true,
		"name":         true,
		"status":       true,
		"created_at":   true,
		"updated_at":   true,
		"last_checked": true,
		"ipv4":         true,
	}

	if field != "" && !validFields[field] {
		return fmt.Errorf("invalid sort field: %s", field)
	}

	return nil
}
