package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"time"

	"go.uber.org/zap"

	"VCS-Checkpoint1/internal/domain"
	"VCS-Checkpoint1/internal/repository"
)

type ReportService struct {
	serverRepo    repository.ServerRepository
	uptimeService *UptimeService
	logger        *zap.Logger
	smtpConfig    domain.SMTPConfig
}

func NewReportService(
	serverRepo repository.ServerRepository,
	uptimeService *UptimeService,
	logger *zap.Logger,
	smtpConfig domain.SMTPConfig,
) *ReportService {
	return &ReportService{
		serverRepo:    serverRepo,
		uptimeService: uptimeService,
		logger:        logger,
		smtpConfig:    smtpConfig,
	}
}

func (s *ReportService) StartDailyReports(ctx context.Context) {
	s.logger.Info("Starting daily report service")

	// Calculate time until next midnight
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	timeUntilMidnight := nextMidnight.Sub(now)

	// Wait until midnight, then send reports daily
	timer := time.NewTimer(timeUntilMidnight)
	defer timer.Stop()

	select {
	case <-timer.C:
		// Send first report at midnight
		s.sendDailyReport(ctx)

		// Then send daily
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.sendDailyReport(ctx)
			case <-ctx.Done():
				s.logger.Info("Daily report service stopped")
				return
			}
		}
	case <-ctx.Done():
		s.logger.Info("Daily report service stopped before first report")
		return
	}
}

func (s *ReportService) sendDailyReport(ctx context.Context) {
	s.logger.Info("Generating daily report")

	yesterday := time.Now().AddDate(0, 0, -1)
	startOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	report, err := s.generateDailyReport(ctx, startOfDay, endOfDay)
	if err != nil {
		s.logger.Error("Failed to generate daily report", zap.Error(err))
		return
	}

	if err := s.sendEmailReport(report); err != nil {
		s.logger.Error("Failed to send daily report email", zap.Error(err))
		return
	}

	s.logger.Info("Daily report sent successfully")
}

func (s *ReportService) generateDailyReport(ctx context.Context, from, to time.Time) (*domain.DailyReport, error) {
	servers, err := s.serverRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers: %w", err)
	}

	totalServers := len(servers)
	onlineServers := 0
	offlineServers := 0
	var totalUptime float64

	for _, server := range servers {
		if server.Status == "online" {
			onlineServers++
		} else {
			offlineServers++
		}

		// Calculate uptime for this server
		stats, err := s.uptimeService.CalculateUptime(ctx, server.ID, from, to)
		if err != nil {
			s.logger.Warn("Failed to calculate uptime for server",
				zap.Error(err),
				zap.Int64("server_id", server.ID),
			)
			continue
		}
		totalUptime += stats.UptimePercentage
	}

	averageUptime := float64(0)
	if totalServers > 0 {
		averageUptime = totalUptime / float64(totalServers)
	}

	return &domain.DailyReport{
		Date:           from,
		TotalServers:   totalServers,
		OnlineServers:  onlineServers,
		OfflineServers: offlineServers,
		AverageUptime:  averageUptime,
		GeneratedAt:    time.Now(),
	}, nil
}

func (s *ReportService) sendEmailReport(report *domain.DailyReport) error {
	subject := fmt.Sprintf("Daily Server Report - %s", report.Date.Format("2006-01-02"))

	emailTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>Daily Server Report</title>
</head>
<body>
    <h2>Daily Server Report - {{.Date.Format "2006-01-02"}}</h2>
    
    <h3>Summary</h3>
    <ul>
        <li><strong>Total Servers:</strong> {{.TotalServers}}</li>
        <li><strong>Online Servers:</strong> {{.OnlineServers}}</li>
        <li><strong>Offline Servers:</strong> {{.OfflineServers}}</li>
        <li><strong>Average Uptime:</strong> {{printf "%.2f" .AverageUptime}}%</li>
    </ul>
    
    <p><em>Report generated at: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</em></p>
</body>
</html>
`

	tmpl, err := template.New("email").Parse(emailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, report); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	return s.sendEmail(s.smtpConfig.AdminEmail, subject, body.String())
}

func (s *ReportService) sendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.smtpConfig.Username, s.smtpConfig.Password, s.smtpConfig.Host)

	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.smtpConfig.Host, s.smtpConfig.Port)
	return smtp.SendMail(addr, auth, s.smtpConfig.From, []string{to}, []byte(msg))
}
