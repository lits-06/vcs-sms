package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/server_service/internal/domain"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
)

type serverUsecase struct {
	serverRepo domain.Repository
	cacheRepo  domain.CacheRepository
}

func NewServerUsecase(serverRepo domain.Repository, cacheRepo domain.CacheRepository) domain.UseCase {
	return &serverUsecase{
		serverRepo: serverRepo,
		cacheRepo:  cacheRepo,
	}
}

func (uc *serverUsecase) CreateServer(ctx context.Context, req *domain.CreateServerRequest) (*domain.Server, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.CreateServer")
	defer span.Finish()

	// Check if server with same ID already exists
	exist, err := uc.serverRepo.ExistsWithID(ctx, req.ID)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to check if server ID exists: %w", err))
	}
	if exist {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("server with ID %s: %w", req.ID, domain.ErrServerExists))
	}

	// Check if server with same name already exists
	exist, err = uc.serverRepo.ExistsWithName(ctx, req.Name)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to check if server name exists: %w", err))
	}
	if exist {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("server with name %s: %w", req.Name, domain.ErrServerExists))
	}

	// Create server entity
	server := &domain.Server{
		ID:        req.ID,
		Name:      req.Name,
		Status:    req.Status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IPv4:      req.IPv4,
	}

	// Save to database
	err = uc.serverRepo.Create(ctx, server)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to create server: %w", err))
	}

	err = uc.cacheRepo.SetServer(ctx, server.ID, server)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to save server to cache: %w", err))
	}

	return server, nil
}

// ViewServers retrieves servers with filtering, sorting, and pagination
func (uc *serverUsecase) ViewServer(ctx context.Context, req *domain.QueryServerRequest) (*domain.QueryServerResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.ViewServer")
	defer span.Finish()

	// Get servers from repository
	servers, total, err := uc.serverRepo.List(ctx, req.Filter, req.Sort, req.Pagination)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to list servers: %w", err))
	}

	// Prepare response
	response := &domain.QueryServerResponse{
		Servers: servers,
		Total:   total,
	}

	return response, nil
}

func (uc *serverUsecase) UpdateServer(ctx context.Context, req *domain.UpdateServerRequest) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.UpdateServer")
	defer span.Finish()

	server, err := uc.serverRepo.GetByID(ctx, req.ID)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to retrieve server: %w", err))
	}
	if server == nil {
		return tracing.TraceWithErr(span, fmt.Errorf("server with ID %s: %w", req.ID, domain.ErrServerNotFound))
	}

	// Update fields only if they are provided (non-empty)
	if req.Name != "" {
		// Check if new name already exists (but not for current server)
		existingServer, err := uc.serverRepo.GetByName(ctx, req.Name)
		if err != nil {
			return tracing.TraceWithErr(span, fmt.Errorf("failed to check if server name exists: %w", err))
		}
		if existingServer != nil && existingServer.ID != req.ID {
			return tracing.TraceWithErr(span, fmt.Errorf("server with name %s: %w", req.Name, domain.ErrServerExists))
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

	// Update server
	err = uc.serverRepo.Update(ctx, server)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to update server: %w", err))
	}

	err = uc.cacheRepo.SetServer(ctx, server.ID, server)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to update server in cache: %w", err))
	}

	return nil
}

func (uc *serverUsecase) DeleteServer(ctx context.Context, serverID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.DeleteServer")
	defer span.Finish()

	if serverID == "" {
		return tracing.TraceWithErr(span, fmt.Errorf("server ID cannot be empty"))
	}

	// Check if server exists
	exist, err := uc.serverRepo.ExistsWithID(ctx, serverID)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("server not found: %w", err))
	}
	if !exist {
		return tracing.TraceWithErr(span, fmt.Errorf("server with ID %s does not exist: %w", serverID, domain.ErrServerNotFound))
	}

	// Delete the server
	err = uc.serverRepo.Delete(ctx, serverID)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to delete server: %w", err))
	}

	err = uc.cacheRepo.DeleteServer(ctx, serverID)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to delete server from cache: %w", err))
	}

	return nil
}

func (uc *serverUsecase) ImportServersFromExcel(ctx context.Context, file multipart.File) (*domain.ImportResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.ImportServersFromExcel")
	defer span.Finish()

	// Open Excel file
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to open Excel file: %w", err))
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("excel file must contain at least one sheet"))
	}

	// Get all rows from Sheet1
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to get rows from Excel file: %w", err))
	}

	if len(rows) < 2 {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("excel file must contain at least headers and one data row"))
	}

	result := &domain.ImportResponse{
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
		status := row[3]

		if serverID == "" || serverName == "" || serverIPv4 == "" || status == "" {
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
		req := &domain.CreateServerRequest{
			ID:     serverID,
			Name:   serverName,
			IPv4:   serverIPv4,
			Status: domain.ServerStatus(status),
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

func (uc *serverUsecase) ExportServersToExcel(ctx context.Context, req *domain.QueryServerRequest) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.ExportServersToExcel")
	defer span.Finish()

	// Get servers from repository
	servers, _, err := uc.serverRepo.List(ctx, req.Filter, req.Sort, req.Pagination)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to list servers: %w", err))
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

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to write Excel file to buffer: %w", err))
	}

	return buffer.Bytes(), nil
}

func (uc *serverUsecase) UpdateServerStatus(ctx context.Context, serverID string, status domain.ServerStatus) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.UpdateServerStatus")
	defer span.Finish()

	err := uc.serverRepo.UpdateStatus(ctx, serverID, status)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to update server status: %w", err))
	}

	return nil
}
