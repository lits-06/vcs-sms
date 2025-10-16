package usecase

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/report_service/config"
	"github.com/lits-06/vcs-sms/report_service/internal/domain"
	"github.com/opentracing/opentracing-go"
	"github.com/xuri/excelize/v2"
	"gopkg.in/gomail.v2"
)

type reportUseCase struct {
	repo domain.Repository
	cfg  *config.Config
	log  logger.Logger
}

func NewReportUseCase(repo domain.Repository, cfg *config.Config, log logger.Logger) domain.UseCase {
	return &reportUseCase{repo: repo, cfg: cfg, log: log}
}

func (uc *reportUseCase) ReportStats(ctx context.Context, email string, startDate, endDate time.Time) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "reportUseCase.ReportStats")
	defer span.Finish()

	start := time.Now()
	stats, err := uc.repo.GetUptimeStats(ctx, startDate, endDate)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to get uptime stats: %w", err))
	}
	elapsed := time.Since(start)
	uc.log.Infof("Fetched uptime stats for %d servers in %s", stats.TotalServers, elapsed)

	return uc.sendUptimeReport(ctx, email, stats)
}

func (uc *reportUseCase) sendUptimeReport(ctx context.Context, email string, stats *domain.UptimeStats) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "reportUseCase.sendUptimeReport")
	defer span.Finish()

	start := time.Now()
	htmlBody, err := uc.generateUptimeReportHTML(ctx, stats)
	elapsed := time.Since(start)
	uc.log.Infof("Generated uptime report HTML in %s", elapsed)

	if err != nil {
		return tracing.TraceWithErr(span, err)
	}

	emails := []string{}
	if email != "" {
		emails = append(emails, email)
	}
	emails = append(emails, uc.cfg.Smtp.To...)

	m := gomail.NewMessage()
	m.SetHeader("From", uc.cfg.Smtp.From)
	m.SetHeader("To", emails...)
	m.SetHeader("Subject", fmt.Sprintf("Server Uptime Report - %s to %s",
		stats.StartDate.Format(time.DateTime),
		stats.EndDate.Format(time.DateTime)))
	m.SetBody("text/html", htmlBody)

	if stats.TotalServers > 0 {
		start := time.Now()
		excelBytes, err := uc.generateUptimeExcel(ctx, stats.ServerDetails)
		elapsed := time.Since(start)
		uc.log.Infof("Generated uptime report Excel in %s", elapsed)
		if err != nil {
			return tracing.TraceWithErr(span, fmt.Errorf("failed to generate excel report: %w", err))
		}

		m.Attach("uptime_report.xlsx", gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(excelBytes)
			return err
		}))
	}

	start = time.Now()
	d := gomail.NewDialer(uc.cfg.Smtp.Host, uc.cfg.Smtp.Port, uc.cfg.Smtp.Username, uc.cfg.Smtp.Password)
	if err := d.DialAndSend(m); err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to send email: %w", err))
	}
	elapsed = time.Since(start)
	uc.log.Infof("Sent uptime report email to %v in %s", emails, elapsed)

	return nil
}

func (uc *reportUseCase) generateUptimeReportHTML(ctx context.Context, stats *domain.UptimeStats) (string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "reportUseCase.generateUptimeReportHTML")
	defer span.Finish()

	rootPath, err := os.Getwd()
	if err != nil {
		return "", tracing.TraceWithErr(span, fmt.Errorf("failed to find project root: %w", err))
	}

	templatePath := filepath.Join(rootPath, "report_service", "internal", "templates", "uptime_report.html")

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", tracing.TraceWithErr(span, fmt.Errorf("failed to parse template: %w", err))
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, stats); err != nil {
		return "", tracing.TraceWithErr(span, fmt.Errorf("failed to execute template: %w", err))
	}

	return buf.String(), nil
}

func (uc *reportUseCase) generateUptimeExcel(ctx context.Context, details []domain.ServerUptimeDetail) ([]byte, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "reportUseCase.generateUptimeExcel")
	defer span.Finish()

	f := excelize.NewFile()
	sheet := "UptimeReport"
	index, _ := f.NewSheet(sheet)

	f.DeleteSheet("Sheet1")

	// Header
	headers := []string{"Server ID", "Uptime Percentage"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Data
	for r, d := range details {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r+2), d.ServerID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r+2), d.UptimePercentage)
	}

	f.SetActiveSheet(index)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
