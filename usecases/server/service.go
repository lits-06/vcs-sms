package server

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/lits-06/vcs-sms/entity"
)

type ServerUsecase struct {
	serverRepo Repository
}

func NewServerUsecase(serverRepo Repository) *ServerUsecase {
	return &ServerUsecase{
		serverRepo: serverRepo,
	}
}

func (uc *ServerUsecase) CreateServer(ctx context.Context, req CreateServerRequest) (*entity.Server, error) {
	// Check if server with same ID already exists
	exist, err := uc.serverRepo.ExistsWithID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if server ID exists: %w", err)
	}
	if exist {
		return nil, fmt.Errorf("server with ID %s already exists", req.ID)
	}

	// Check if server with same name already exists
	exist, err = uc.serverRepo.ExistsWithName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check if server name exists: %w", err)
	}
	if exist {
		return nil, fmt.Errorf("server with name %s already exists", req.Name)
	}

	// Create server entity
	server := &entity.Server{
		ID:          req.ID,
		Name:        req.Name,
		Status:      entity.StatusOffline,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		LastChecked: time.Now(),
		IPv4:        req.IPv4,
	}

	// Save to database
	err = uc.serverRepo.Create(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	return server, nil
}

// ViewServers retrieves servers with filtering, sorting, and pagination
func (uc *ServerUsecase) ViewServer(ctx context.Context, req ViewServerRequest) (*ViewServerResponse, error) {
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

func (uc *ServerUsecase) validateSortField(field string) error {
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

func (uc *ServerUsecase) UpdateServer(ctx context.Context, req UpdateServerRequest) error {
	// Validate input
	if req.ServerID == "" {
		return fmt.Errorf("server ID is required")
	}

	if len(req.UpdateData) == 0 {
		return fmt.Errorf("update data is required")
	}

	// Validate update data
	if err := uc.validateUpdateData(req.UpdateData); err != nil {
		return fmt.Errorf("invalid update data: %w", err)
	}

	server, err := uc.serverRepo.GetByID(ctx, req.ServerID)
	if err != nil {
		return fmt.Errorf("failed to retrieve server: %w", err)
	}
	if server == nil {
		return fmt.Errorf("server with ID %s not found", req.ServerID)
	}

	for key, value := range req.UpdateData {
		switch key {
		case "name":
			server.Name = value.(string)
		case "status":
			server.Status = value.(entity.ServerStatus)
		case "ipv4":
			server.IPv4 = value.(string)
		default:
		}
	}

	// Update server
	err = uc.serverRepo.Update(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}

	return nil
}

func (uc *ServerUsecase) validateUpdateData(updateData map[string]interface{}) error {
	allowedFields := map[string]bool{
		"name":   true,
		"status": true,
		"ipv4":   true,
	}

	for field := range updateData {
		if !allowedFields[field] {
			return fmt.Errorf("invalid field for update: %s", field)
		}
	}

	// Validate specific field values
	if name, exists := updateData["name"]; exists {
		if err := validateNameField(name); err != nil {
			return err
		}
		exists, err := uc.serverRepo.ExistsWithName(context.Background(), name.(string))
		if err != nil {
			return fmt.Errorf("failed to check server name existence: %w", err)
		}
		if exists {
			return fmt.Errorf("server with name %s already exists", name)
		}
	}

	if status, exists := updateData["status"]; exists {
		if err := validateStatusField(status); err != nil {
			return err
		}
	}

	if ipv4, exists := updateData["ipv4"]; exists {
		if err := validateIPv4Field(ipv4); err != nil {
			return err
		}
	}

	return nil
}

func validateNameField(name interface{}) error {
	nameStr, ok := name.(string)
	if !ok {
		return fmt.Errorf("name must be a string")
	}
	if nameStr == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

func validateStatusField(status interface{}) error {
	statusStr, ok := status.(string)
	if !ok {
		return fmt.Errorf("status must be a string")
	}
	validStatuses := map[string]bool{
		string(entity.StatusOnline):  true,
		string(entity.StatusOffline): true,
		string(entity.StatusUnknown): true,
	}
	if !validStatuses[statusStr] {
		return fmt.Errorf("invalid status: %s. Valid statuses are: online, offline, unknown", statusStr)
	}
	return nil
}

// validateIPv4 validates IPv4 address format
func validateIPv4Field(ip interface{}) error {
	ipStr, ok := ip.(string)
	if !ok {
		return fmt.Errorf("ipv4 must be a string")
	}

	ipv4Regex := regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)
	if !ipv4Regex.MatchString(ipStr) {
		return fmt.Errorf("invalid IPv4 format")
	}
	return nil
}

func (uc *ServerUsecase) DeleteServer(ctx context.Context, serverID string) error {
	if serverID == "" {
		return fmt.Errorf("server ID cannot be empty")
	}

	// Check if server exists
	exist, err := uc.serverRepo.ExistsWithID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}
	if !exist {
		return fmt.Errorf("server with ID %s does not exist", serverID)
	}

	// Delete the server
	err = uc.serverRepo.Delete(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	return nil
}
