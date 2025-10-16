package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/server_service/internal/domain"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
)

type serverUsecase struct {
	serverRepo domain.Repository
	publisher  domain.EventPublisher
	log        logger.Logger
}

func NewServerUsecase(serverRepo domain.Repository, publisher domain.EventPublisher, log logger.Logger) domain.UseCase {
	return &serverUsecase{
		serverRepo: serverRepo,
		publisher:  publisher,
		log:        log,
	}
}

func (uc *serverUsecase) CreateServer(ctx context.Context, name, ipv4 string, port int) (*domain.Server, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.CreateServer")
	defer span.Finish()

	// Check if server with same name already exists
	exist, err := uc.serverRepo.ExistsWithName(ctx, name)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("serverRepo.ExistsWithName: %w", err))
	}
	if exist {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("server with name %s: %w", name, domain.ErrServerExists))
	}

	// Create server entity
	server := &domain.Server{
		Name:      name,
		Status:    domain.StatusOffline, // Default status is OFF
		Port:      port,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IPv4:      ipv4,
	}

	// Save to database
	createdServer, err := uc.serverRepo.Create(ctx, server)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("serverRepo.Create: %w", err))
	}

	// Publish event
	err = uc.publisher.PublishServerCreate(ctx, []domain.Server{*createdServer})
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("publisher.PublishServerCreate: %w", err))
	}

	return createdServer, nil
}

// ViewServers retrieves servers with filtering, sorting, and pagination
func (uc *serverUsecase) ViewServer(ctx context.Context, name, status, ipv4 string, from, to int, sort, order string) (*[]domain.Server, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.ViewServer")
	defer span.Finish()

	// Get servers from repository
	servers, total, err := uc.serverRepo.List(ctx, name, status, ipv4, from, to, sort, order)
	if err != nil {
		return nil, 0, tracing.TraceWithErr(span, fmt.Errorf("serverRepo.List: %w", err))
	}

	return servers, total, nil
}

func (uc *serverUsecase) UpdateServer(ctx context.Context, id, name, ipv4 string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.UpdateServer")
	defer span.Finish()

	// Check if server exists
	exist, err := uc.serverRepo.ExistsWithID(ctx, id)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("serverRepo.ExistsWithID: %w", err))
	}
	if !exist {
		return tracing.TraceWithErr(span, fmt.Errorf("server with ID %s does not exist: %w", id, domain.ErrServerNotFound))
	}

	// Update fields only if they are provided (non-empty)
	if name != "" {
		// Check if new name already exists (but not for current server)
		srv, err := uc.serverRepo.GetByName(ctx, name)
		if err != nil {
			return tracing.TraceWithErr(span, fmt.Errorf("serverRepo.GetByName: %w", err))
		}

		if srv != nil {
			return tracing.TraceWithErr(span, fmt.Errorf("server with name %s already exists: %w", name, domain.ErrNameExists))
		}
	}

	// Update server
	err = uc.serverRepo.Update(ctx, id, name, ipv4)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("serverRepo.Update: %w", err))
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
		return tracing.TraceWithErr(span, fmt.Errorf("serverRepo.ExistsWithID: %w", err))
	}
	if !exist {
		return tracing.TraceWithErr(span, fmt.Errorf("server with ID %s does not exist: %w", serverID, domain.ErrServerNotFound))
	}

	// Delete the server
	err = uc.serverRepo.Delete(ctx, serverID)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("serverRepo.Delete: %w", err))
	}

	err = uc.publisher.PublishServerDelete(ctx, []string{serverID})
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("publisher.PublishServerDelete: %w", err))
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

	start := time.Now()
	var servers []domain.Server

	// Skip header row and process data rows
	for i, row := range rows[1:] {
		if len(row) < 3 {
			continue
		}

		serverName := row[0]
		serverIPv4 := row[1]
		portStr := row[2]

		if serverName == "" || serverIPv4 == "" || portStr == "" {
			result.FailureCount++
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("row:%d - missing required fields", i))
			continue
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			result.FailureCount++
			result.FailureServers = append(result.FailureServers, fmt.Sprintf("row:%d name:%s - invalid port", i, serverName))
			continue
		}

		server := domain.Server{
			Name:      serverName,
			IPv4:      serverIPv4,
			Port:      port,
			Status:    domain.StatusOffline,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		servers = append(servers, server)
	}

	// Bulk insert servers
	err = uc.serverRepo.CreateBatch(ctx, servers)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("serverRepo.CreateBatch: %w", err))
	}
	elapsed := time.Since(start)
	uc.log.Info("Imported servers from Excel ", "count: ", len(servers), "elapsed: ", elapsed)

	result.SuccessCount = len(servers)
	for _, srv := range servers {
		result.SuccessServers = append(result.SuccessServers, srv.Name)
	}

	// Publish event
	err = uc.publisher.PublishServerCreate(ctx, servers)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("publisher.PublishServerCreate: %w", err))
	}

	return result, nil
}

func (uc *serverUsecase) ExportServersToExcel(ctx context.Context, name, status, ipv4 string, from, to int, sort, order string) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.ExportServersToExcel")
	defer span.Finish()

	// Get servers from repository
	servers, _, err := uc.serverRepo.List(ctx, name, status, ipv4, from, to, sort, order)
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("serverRepo.List: %w", err))
	}

	// Create Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	// Set headers
	headers := []string{"ID", "Name", "Port", "Status", "Created At", "Updated At", "IPv4"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cell, header)
	}

	// Add server data
	for i, server := range *servers {
		row := i + 2 // Start from row 2 (after headers)
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), server.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), server.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), server.Port)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), server.Status)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), server.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), server.UpdatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", row), server.IPv4)
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to write Excel file to buffer: %w", err))
	}

	return buffer.Bytes(), nil
}

func (uc *serverUsecase) UpdateServerStatus(ctx context.Context, serverID string, status string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.UpdateServerStatus")
	defer span.Finish()

	// Check if server exists
	exist, err := uc.serverRepo.ExistsWithID(ctx, serverID)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("serverRepo.ExistsWithID: %w", err))
	}
	if !exist {
		return tracing.TraceWithErr(span, fmt.Errorf("server with ID %s does not exist: %w", serverID, domain.ErrServerNotFound))
	}

	// Validate status
	if !domain.IsStatusValid(status) {
		return tracing.TraceWithErr(span, fmt.Errorf("invalid server status"))
	}

	err = uc.serverRepo.UpdateStatus(ctx, serverID, status)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to update server status: %w", err))
	}

	return nil
}

func (uc *serverUsecase) BulkUpdateServerStatus(ctx context.Context, updates []domain.ServerStatusUpdate) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverUsecase.BulkUpdateServerStatus")
	defer span.Finish()

	if len(updates) == 0 {
		return nil
	}

	// Validate all statuses
	for _, update := range updates {
		if !domain.IsStatusValid(update.Status) {
			return tracing.TraceWithErr(span, fmt.Errorf("invalid server status for server %s: %s", update.ServerID, update.Status))
		}
	}

	// Perform bulk update
	err := uc.serverRepo.BulkUpdateStatus(ctx, updates)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to bulk update server status: %w", err))
	}

	return nil
}
