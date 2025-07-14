package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"VCS-Checkpoint1/internal/domain"
	"VCS-Checkpoint1/internal/repository"
)

type ExcelService struct {
	serverRepo repository.ServerRepository
	logger     *zap.Logger
}

func NewExcelService(serverRepo repository.ServerRepository, logger *zap.Logger) *ExcelService {
	return &ExcelService{
		serverRepo: serverRepo,
		logger:     logger,
	}
}

func (s *ExcelService) ImportServers(ctx context.Context, file io.Reader, filename string) (*domain.ImportResult, error) {
	s.logger.Info("Starting server import", zap.String("filename", filename))

	// Read file content
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Open Excel file
	f, err := excelize.OpenReader(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel file must have at least 2 rows (header + data)")
	}

	// Skip header row
	var servers []*domain.Server
	var errors []string
	successCount := 0

	for i, row := range rows[1:] {
		if len(row) < 4 {
			errors = append(errors, fmt.Sprintf("Row %d: insufficient columns", i+2))
			continue
		}

		name := row[0]
		host := row[1]
		portStr := row[2]
		description := ""
		if len(row) > 3 {
			description = row[3]
		}

		// Validate port
		port, err := strconv.Atoi(portStr)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: invalid port '%s'", i+2, portStr))
			continue
		}

		server := &domain.Server{
			Name:        name,
			Host:        host,
			Port:        port,
			Status:      "offline", // Default status
			Description: description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Create server
		if err := s.serverRepo.Create(ctx, server); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: failed to create server '%s': %v", i+2, name, err))
			continue
		}

		servers = append(servers, server)
		successCount++
	}

	result := &domain.ImportResult{
		TotalRows:    len(rows) - 1,
		SuccessCount: successCount,
		ErrorCount:   len(errors),
		Errors:       errors,
	}

	s.logger.Info("Server import completed",
		zap.Int("total_rows", result.TotalRows),
		zap.Int("success_count", result.SuccessCount),
		zap.Int("error_count", result.ErrorCount),
	)

	return result, nil
}

func (s *ExcelService) ExportServers(ctx context.Context) ([]byte, string, error) {
	s.logger.Info("Starting server export")

	servers, err := s.serverRepo.GetAll(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get servers: %w", err)
	}

	// Create new Excel file
	f := excelize.NewFile()
	sheetName := "Servers"

	// Create sheet
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create sheet: %w", err)
	}

	// Set headers
	headers := []string{"ID", "Name", "Host", "Port", "Status", "Description", "Created At", "Updated At", "Last Checked"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
	}

	// Set data
	for i, server := range servers {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), server.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), server.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), server.Host)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), server.Port)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), server.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), server.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), server.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), server.UpdatedAt.Format("2006-01-02 15:04:05"))

		lastChecked := ""
		if !server.LastChecked.IsZero() {
			lastChecked = server.LastChecked.Format("2006-01-02 15:04:05")
		}
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), lastChecked)
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Generate filename
	filename := fmt.Sprintf("servers_%s.xlsx", time.Now().Format("20060102_150405"))

	// Save to buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", fmt.Errorf("failed to write Excel file: %w", err)
	}

	s.logger.Info("Server export completed",
		zap.Int("server_count", len(servers)),
		zap.String("filename", filename),
	)

	return buf.Bytes(), filename, nil
}
