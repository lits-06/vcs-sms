package server

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/lits-06/vcs-sms/entity"
	"github.com/xuri/excelize/v2"
)

type ServerUsecase struct {
	serverRepo     Repository
	cacheRepo      CacheRepository
	serverProvider Provider
}

func NewServerUsecase(serverRepo Repository, cacheRepo CacheRepository, serverProvider Provider) *ServerUsecase {
	return &ServerUsecase{
		serverRepo:     serverRepo,
		cacheRepo:      cacheRepo,
		serverProvider: serverProvider,
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
		ID:        req.ID,
		Name:      req.Name,
		Status:    req.Status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IPv4:      req.IPv4,
	}

	// Transaction pattern với rollback
	var (
		providerCreated = false
		dbCreated       = false
		cacheCreated    = false
	)

	// Rollback function
	defer func() {
		if !providerCreated || !dbCreated || !cacheCreated {
			// Rollback provider
			if providerCreated {
				if rollbackErr := uc.serverProvider.DeleteServer(ctx, server.ID); rollbackErr != nil {
					// Log error nhưng không fail toàn bộ transaction
					fmt.Printf("Failed to rollback provider for server %s: %v\n", server.ID, rollbackErr)
				}
			}

			// Rollback database
			if dbCreated {
				if rollbackErr := uc.serverRepo.Delete(ctx, server.ID); rollbackErr != nil {
					fmt.Printf("Failed to rollback database for server %s: %v\n", server.ID, rollbackErr)
				}
			}

			// Rollback cache
			if cacheCreated {
				if rollbackErr := uc.cacheRepo.DeleteServer(ctx, server.ID); rollbackErr != nil {
					fmt.Printf("Failed to rollback cache for server %s: %v\n", server.ID, rollbackErr)
				}
			}
		}
	}()

	err = uc.serverProvider.CreateServer(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}
	providerCreated = true

	// Save to database
	err = uc.serverRepo.Create(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}
	dbCreated = true

	err = uc.cacheRepo.SetServer(ctx, server.ID, server)
	if err != nil {
		return nil, fmt.Errorf("failed to save server to cache: %w", err)
	}
	cacheCreated = true

	return server, nil
}

// ViewServers retrieves servers with filtering, sorting, and pagination
func (uc *ServerUsecase) ViewServer(ctx context.Context, req QueryServerRequest) (*QueryServerResponse, error) {
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

	if req.Status != "" {
		server.Status = req.Status
	}

	err = uc.serverProvider.UpdateServer(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to update server in provider: %w", err)
	}

	// Update server
	err = uc.serverRepo.Update(ctx, server)
	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}

	err = uc.cacheRepo.SetServer(ctx, server.ID, server)
	if err != nil {
		return fmt.Errorf("failed to update server in cache: %w", err)
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

	err = uc.serverProvider.DeleteServer(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to delete server from provider: %w", err)
	}

	// Delete the server
	err = uc.serverRepo.Delete(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	err = uc.cacheRepo.DeleteServer(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to delete server from cache: %w", err)
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
		return nil, fmt.Errorf("excel file must contain at least one sheet")
	}

	// Get all rows from Sheet1
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get rows from Excel file: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("excel file must contain at least headers and one data row")
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
	headers := []string{"ID", "Name", "IPv4", "Status", "Created At", "Updated At"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cell, header)
	}

	// Add server data
	for i, server := range *servers {
		row := i + 2 // Start from row 2 (after headers)
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), server.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), server.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), server.IPv4)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), string(server.Status))
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), server.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), server.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	// Tìm project root (thư mục chứa go.mod)
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Tạo đường dẫn exports từ project root
	exportDir := filepath.Join(projectRoot, "exports")

	// Tạo thư mục nếu chưa tồn tại
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return fmt.Errorf("failed to create exports directory: %w", err)
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("servers_export_%s.xlsx", time.Now().Format("20060102_150405"))

	path := filepath.Join(exportDir, filename)

	// Save the file
	if err := f.SaveAs(path); err != nil {
		return fmt.Errorf("failed to save Excel file: %w", err)
	}

	return nil
}

// Helper function để tìm project root
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Kiểm tra xem có go.mod file không
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Đã đến root của filesystem
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("go.mod not found")
}
