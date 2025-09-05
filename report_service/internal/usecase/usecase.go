package usecase

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"github.com/lits-06/vcs-sms/pkg/utils"
	"github.com/lits-06/vcs-sms/report_service/internal/domain"
	"gopkg.in/gomail.v2"
)

type reportUseCase struct {
	repo domain.Repository
}

func NewReportUseCase(repo domain.Repository) domain.UseCase {
	return &reportUseCase{repo: repo}
}

func (uc *reportUseCase) ReportStats(ctx context.Context, req *domain.UptimeRequest) error {
	stats, err := uc.repo.GetUptimeStats(ctx, req)
	if err != nil {
		return err
	}

	return uc.sendUptimeReport(stats)
}

func (uc *reportUseCase) sendUptimeReport(stats *domain.UptimeStats) error {
	htmlBody, err := uc.generateUptimeReportHTML(stats)

	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", uc.config.From)
	m.SetHeader("To", uc.config.AdminEmail...)
	m.SetHeader("Subject", fmt.Sprintf("Server Uptime Report - %s to %s",
		stats.StartDate.Format(time.DateTime),
		stats.EndDate.Format(time.DateTime)))
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(uc.config.Host, uc.config.Port, uc.config.Username, uc.config.Password)
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func (uc *reportUseCase) generateUptimeReportHTML(stats *domain.UptimeStats) (string, error) {
	projectRoot, err := utils.FindProjectRoot()
	if err != nil {
		return "", err
	}

	templatePath := filepath.Join(projectRoot, "report_service", "internal", "templates", "uptime_report.html")

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, stats); err != nil {
		return "", err
	}

	return buf.String(), nil
}
