package usecase

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/pkg/utils"
	"github.com/lits-06/vcs-sms/report_service/config"
	"github.com/lits-06/vcs-sms/report_service/internal/domain"
	"gopkg.in/gomail.v2"
)

type reportUseCase struct {
	repo domain.Repository
	cfg  *config.Config
}

func NewReportUseCase(repo domain.Repository, cfg *config.Config) domain.UseCase {
	return &reportUseCase{repo: repo, cfg: cfg}
}

func (uc *reportUseCase) ReportStats(ctx context.Context, req *domain.UptimeRequest) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "reportUseCase.ReportStats")
	defer span.Finish()

	stats, err := uc.repo.GetUptimeStats(ctx, req)
	if err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to get uptime stats: %w", err))
	}

	return uc.sendUptimeReport(ctx, stats)
}

func (uc *reportUseCase) sendUptimeReport(ctx context.Context, stats *domain.UptimeStats) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "reportUseCase.sendUptimeReport")
	defer span.Finish()

	htmlBody, err := uc.generateUptimeReportHTML(ctx, stats)

	if err != nil {
		return tracing.TraceWithErr(span, err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", uc.cfg.Smtp.From)
	m.SetHeader("To", uc.cfg.Smtp.To...)
	m.SetHeader("Subject", fmt.Sprintf("Server Uptime Report - %s to %s",
		stats.StartDate.Format(time.DateTime),
		stats.EndDate.Format(time.DateTime)))
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(uc.cfg.Smtp.Host, uc.cfg.Smtp.Port, uc.cfg.Smtp.Username, uc.cfg.Smtp.Password)
	if err := d.DialAndSend(m); err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to send email: %w", err))
	}

	return nil
}

func (uc *reportUseCase) generateUptimeReportHTML(ctx context.Context, stats *domain.UptimeStats) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "reportUseCase.generateUptimeReportHTML")
	defer span.Finish()

	projectRoot, err := utils.FindProjectRoot()
	if err != nil {
		return "", tracing.TraceWithErr(span, fmt.Errorf("failed to find project root: %w", err))
	}

	templatePath := filepath.Join(projectRoot, "report_service", "internal", "templates", "uptime_report.html")

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
