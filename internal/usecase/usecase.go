package usecase

import (
	"github.com/lits-06/vcs-sms/internal/config"
	"github.com/lits-06/vcs-sms/internal/domain"
	"github.com/lits-06/vcs-sms/internal/infrastructure/logger"
	"github.com/lits-06/vcs-sms/internal/repository"
	"github.com/lits-06/vcs-sms/internal/usecase/auth"
	"github.com/lits-06/vcs-sms/internal/usecase/monitor"
	"github.com/lits-06/vcs-sms/internal/usecase/report"
	"github.com/lits-06/vcs-sms/internal/usecase/server"
)

// UseCases holds all use case implementations
type UseCases struct {
	Server        domain.ServerUseCase
	ServerMonitor domain.MonitorUseCase
	Auth          domain.AuthUseCase
	ReportService domain.ReportUseCase
}

// NewUseCases creates a new use cases instance
func NewUseCases(repos *repository.Repositories, logger *logger.Logger, cfg *config.Config) *UseCases {
	return &UseCases{
		Server:        server.NewServerUseCase(repos, logger),
		ServerMonitor: monitor.NewMonitorUseCase(repos, logger, cfg),
		Auth:          auth.NewAuthUseCase(repos, logger, cfg),
		ReportService: report.NewReportUseCase(repos, logger, cfg),
	}
}
