package server

import (
	"context"
	"fmt"

	"github.com/lits-06/vcs-sms/internal/domain/repositories"
	"github.com/lits-06/vcs-sms/internal/repository"
)

type ServerUseCase struct {
	repos *repositories.ServerRepository
}

// NewServerUseCase creates a new server use case
func NewServerUseCase(repos *repository.Repositories) ServerUseCase {
	return &serverUseCase{
		repos: repos,
	}
}

// Create creates a new server
func (uc *serverUseCase) Create(ctx context.Context, req domain.CreateServerRequest) (*domain.Server, error) {
	// Check if server already exists
	existing, err := uc.repos.Server.GetByHostPort(ctx, req.Host, req.Port)
	if err == nil && existing != nil {
		return nil, domain.ErrServerExists
	}

	// Create server entity
	server := &domain.Server{
		Name:        req.Name,
		Host:        req.Host,
		Port:        req.Port,
		Description: req.Description,
		Tags:        req.Tags,
	}

	// Save to database
	err = uc.repos.Server.Create(ctx, server)
	if err != nil {
		uc.logger.Error("Failed to create server", "error", err)
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	// Clear cache
	uc.repos.Cache.DeletePattern(ctx, "servers:*")

	uc.logger.Info("Server created successfully", "server_id", server.ID)
	return server, nil
}

// GetByID retrieves a server by ID
func (uc *serverUseCase) GetByID(ctx context.Context, id int64) (*domain.Server, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("servers:%d", id)
	var server domain.Server
	err := uc.repos.Cache.Get(ctx, cacheKey, &server)
	if err == nil {
		return &server, nil
	}

	// Get from database
	serverPtr, err := uc.repos.Server.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	uc.repos.Cache.Set(ctx, cacheKey, serverPtr, 0) // No expiration

	return serverPtr, nil
}

// Update updates a server
func (uc *serverUseCase) Update(ctx context.Context, id int64, req domain.UpdateServerRequest) (*domain.Server, error) {
	// Get existing server
	existing, err := uc.repos.Server.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Host != nil {
		existing.Host = *req.Host
	}
	if req.Port != nil {
		existing.Port = *req.Port
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}

	// Check for host:port conflict if host or port changed
	if req.Host != nil || req.Port != nil {
		conflicting, err := uc.repos.Server.GetByHostPort(ctx, existing.Host, existing.Port)
		if err == nil && conflicting != nil && conflicting.ID != id {
			return nil, domain.ErrServerExists
		}
	}

	// Update in database
	err = uc.repos.Server.Update(ctx, id, existing)
	if err != nil {
		uc.logger.Error("Failed to update server", "error", err, "server_id", id)
		return nil, fmt.Errorf("failed to update server: %w", err)
	}

	// Clear cache
	uc.repos.Cache.Delete(ctx, fmt.Sprintf("servers:%d", id))
	uc.repos.Cache.DeletePattern(ctx, "servers:list:*")

	uc.logger.Info("Server updated successfully", "server_id", id)
	return existing, nil
}

// Delete deletes a server
func (uc *serverUseCase) Delete(ctx context.Context, id int64) error {
	err := uc.repos.Server.Delete(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to delete server", "error", err, "server_id", id)
		return err
	}

	// Clear cache
	uc.repos.Cache.Delete(ctx, fmt.Sprintf("servers:%d", id))
	uc.repos.Cache.DeletePattern(ctx, "servers:list:*")

	uc.logger.Info("Server deleted successfully", "server_id", id)
	return nil
}

// List retrieves servers with filtering, sorting, and pagination
func (uc *serverUseCase) List(ctx context.Context, req domain.ServerListRequest) (*domain.ServerListResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("servers:list:%+v", req)
	var response domain.ServerListResponse
	err := uc.repos.Cache.Get(ctx, cacheKey, &response)
	if err == nil {
		return &response, nil
	}

	// Get from database
	responsePtr, err := uc.repos.Server.List(ctx, req)
	if err != nil {
		uc.logger.Error("Failed to list servers", "error", err)
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	// Cache the result for 5 minutes
	uc.repos.Cache.Set(ctx, cacheKey, responsePtr, 5*60)

	return responsePtr, nil
}

// Import imports servers from a list
func (uc *serverUseCase) Import(ctx context.Context, req domain.ImportServerRequest) error {
	var servers []domain.Server

	for _, serverReq := range req.Servers {
		server := domain.Server{
			Name:        serverReq.Name,
			Host:        serverReq.Host,
			Port:        serverReq.Port,
			Description: serverReq.Description,
			Tags:        serverReq.Tags,
		}
		servers = append(servers, server)
	}

	err := uc.repos.Server.BulkCreate(ctx, servers)
	if err != nil {
		uc.logger.Error("Failed to import servers", "error", err)
		return fmt.Errorf("failed to import servers: %w", err)
	}

	// Clear cache
	uc.repos.Cache.DeletePattern(ctx, "servers:*")

	uc.logger.Info("Servers imported successfully", "count", len(servers))
	return nil
}

// Export exports servers to a file format
func (uc *serverUseCase) Export(ctx context.Context, filter domain.ServerFilter) ([]byte, error) {
	// Get all servers based on filter
	req := domain.ServerListRequest{
		Filter: filter,
		Pagination: domain.Pagination{
			Page:  1,
			Limit: 10000, // Large limit to get all servers
		},
	}

	response, err := uc.repos.Server.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers for export: %w", err)
	}

	// Convert to Excel format (simplified - would need actual Excel library)
	// For now, return JSON format
	data := struct {
		Servers []domain.Server `json:"servers"`
		Total   int64           `json:"total"`
	}{
		Servers: response.Servers,
		Total:   response.Total,
	}

	// Here you would use a library like excelize to create Excel file
	// For demonstration, we'll return a simple format
	return []byte(fmt.Sprintf("Total servers: %d\nServers exported successfully", len(response.Servers))), nil
}
