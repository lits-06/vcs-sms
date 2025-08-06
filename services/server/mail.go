package server

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/utils"
	"gopkg.in/gomail.v2"
)

type mailService struct {
	config *config.SMTPConfig
	logger logger.Logger
}

func NewMailService(cfg *config.SMTPConfig, log logger.Logger) MailService {
	return &mailService{
		config: cfg,
		logger: log,
	}
}

func (s *mailService) SendUptimeReport(stats *UptimeStats) error {
	htmlBody, err := s.generateUptimeReportHTML(stats)

	if err != nil {
		s.logger.Error("failed to generate uptime report HTML", "error", err)
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", s.config.AdminEmail...)
	m.SetHeader("Subject", fmt.Sprintf("Server Uptime Report - %s to %s",
		stats.StartDate.Format("2006-01-02"),
		stats.EndDate.Format("2006-01-02")))
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)
	if err := d.DialAndSend(m); err != nil {
		s.logger.Error("failed to send uptime report email", "error", err)
		return err
	}

	return nil
}

func (s *mailService) generateUptimeReportHTML(stats *UptimeStats) (string, error) {
	projectRoot, err := utils.FindProjectRoot()
	if err != nil {
		s.logger.Error("failed to find project root", "error", err)
		return "", err
	}

	templatePath := filepath.Join(projectRoot, "templates", "uptime_report.html")

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		s.logger.Error("failed to parse uptime report template", "error", err)
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, stats); err != nil {
		return "", err
	}

	return buf.String(), nil
}
