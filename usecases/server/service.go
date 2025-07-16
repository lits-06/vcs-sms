package server

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/lits-06/vcs-sms/entity"
	"github.com/xuri/excelize/v2"
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
func (uc *ServerUsecase) ViewServer(ctx context.Context, req QueryServerRequest) (*QueryServerResponse, error) {
	// // Try cache first
	// cacheKey := fmt.Sprintf("servers:list:%+v", req)
	// var response domain.ServerListResponse
	// err := uc.repos.Cache.Get(ctx, cacheKey, &response)
	// if err == nil {
	// 	return &response, nil
	// }

	// Get servers from repository
	servers, total, err := uc.serverRepo.List(ctx, req.Filter, req.Sort, req.Pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	// Prepare response
	response := &QueryServerResponse{
		Servers: servers,
		Total:   total,
	}

	// Cache the result for 5 minutes
	// uc.repos.Cache.Set(ctx, cacheKey, response, 5*60)

	return response, nil
}

func (uc *ServerUsecase) UpdateServer(ctx context.Context, req UpdateServerRequest) error {
	server, err := uc.serverRepo.GetByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve server: %w", err)
	}
	if server == nil {
		return fmt.Errorf("server with ID %s not found", req.ID)
	}

	// Update fields only if they are provided (non-empty)
	if req.Name != "" {
		// Check if new name already exists (but not for current server)
		existingServer, err := uc.serverRepo.GetByName(ctx, req.Name)
		if err != nil {
			return fmt.Errorf("failed to check if server name exists: %w", err)
		}
		if existingServer != nil && existingServer.ID != req.ID {
			return fmt.Errorf("server with name %s already exists", req.Name)
		}

		server.Name = req.Name
	}

	if req.Status != "" {
		server.Status = req.Status
	}

	if req.IPv4 != "" {
		server.IPv4 = req.IPv4
	}

	// Set updated timestamp
	server.UpdatedAt = time.Now()

	// Update server
	err = uc.serverRepo.Update(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
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

func (uc *ServerUsecase) ImportServersFromExcel(ctx context.Context, file multipart.File) (*ImportRespose, error) {
	// Open Excel file
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel file must contain at least one sheet")
	}

	// Get all rows from Sheet1
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get rows from Excel file: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel file must contain at least headers and one data row")
	}

	result := &ImportRespose{
		SuccessServers: make([]string, 0),
		FailureServers: make([]string, 0),
	}

	// Skip header row and process data rows
	for _, row := range rows[1:] {
		if len(row) < 3 { // At least ID, Name, IPv4 required
			continue
		}

		serverID := row[0]
		serverName := row[1]
		serverIPv4 := row[2]

		if serverID == "" || serverName == "" || serverIPv4 == "" {
			result.FailureCount++
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("%s:%s - missing required fields", serverID, serverName))
			continue
		}

		// Check if server with same ID already exists
		existID, err := uc.serverRepo.ExistsWithID(ctx, serverID)
		if err != nil {
			result.FailureCount++
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("%s:%s - failed to check ID existence", serverID, serverName))
			continue
		}
		if existID {
			result.FailureCount++
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("%s:%s - ID already exists", serverID, serverName))
			continue
		}

		// Check if server with same name already exists
		existName, err := uc.serverRepo.ExistsWithName(ctx, serverName)
		if err != nil {
			result.FailureCount++
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("%s:%s - failed to check name existence", serverID, serverName))
			continue
		}
		if existName {
			result.FailureCount++
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("%s:%s - name already exists", serverID, serverName))
			continue
		}

		// Create server request
		req := CreateServerRequest{
			ID:   serverID,
			Name: serverName,
			IPv4: serverIPv4,
		}

		// Create server
		_, err = uc.CreateServer(ctx, req)
		if err != nil {
			result.FailureCount++
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("%s:%s - %v", serverID, serverName, err))
			continue
		}

		result.SuccessCount++
		result.SuccessServers = append(result.SuccessServers, fmt.Sprintf("%s:%s", serverID, serverName))
	}

	return result, nil
}

func (uc *ServerUsecase) ExportServersToExcel(ctx context.Context, req QueryServerRequest) error {
	// Get servers from repository
	servers, _, err := uc.serverRepo.List(ctx, req.Filter, req.Sort, req.Pagination)
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}

	// Create Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	// Set headers
	headers := []string{"ID", "Name", "IPv4", "Status", "Created At", "Updated At", "Last Checked"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cell, header)
	}

	// Add server data
	for i, server := range servers {
		row := i + 2 // Start from row 2 (after headers)
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), server.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), server.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), server.IPv4)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), string(server.Status))
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), server.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), server.UpdatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", row), server.LastChecked.Format("2006-01-02 15:04:05"))
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("servers_export_%s.xlsx", time.Now().Format("20060102_150405"))

	path := filepath.Join("../../exports", filename)

	// Save the file
	if err := f.SaveAs(path); err != nil {
		return fmt.Errorf("failed to save Excel file: %w", err)
	}

	return nil
}
