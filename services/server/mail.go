package server

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"gopkg.in/gomail.v2"
)

type mailService struct {
	config *config.SMTPConfig
	logger logger.Logger
}

func NewMailService(cfg *config.SMTPConfig, log logger.Logger) *mailService {
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
	m.SetHeader("To", s.config.AdminEmail)
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
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Server Uptime Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f4f4f4; padding: 20px; border-radius: 5px; }
        .stats { background-color: #e8f5e8; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .warning { background-color: #fff3cd; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .error { background-color: #f8d7da; padding: 15px; margin: 20px 0; border-radius: 5px; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .online { color: #28a745; font-weight: bold; }
        .offline { color: #dc3545; font-weight: bold; }
        .good { background-color: #d4edda; }
        .warning-row { background-color: #fff3cd; }
        .bad { background-color: #f8d7da; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📊 Server Uptime Report</h1>
        <p><strong>Period:</strong> {{.Stats.StartDate.Format "2006-01-02"}} to {{.Stats.EndDate.Format "2006-01-02"}}</p>
    </div>

    <div class="stats">
        <h2>📈 Overall Statistics</h2>
        <ul>
            <li><strong>Total Servers:</strong> {{.Stats.TotalServers}}</li>
            <li><strong>Online Servers:</strong> <span class="online">{{.Stats.OnlineServers}}</span></li>
            <li><strong>Offline Servers:</strong> <span class="offline">{{.Stats.OfflineServers}}</span></li>
            <li><strong>Average Uptime:</strong> {{printf "%.2f" .Stats.UptimePercentage}}%</li>
        </ul>
    </div>
</body>
</html>
`
	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, stats); err != nil {
		return "", err
	}

	return buf.String(), nil
}
