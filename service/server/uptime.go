package server

import (
	"context"

	"github.com/lits-06/vcs-sms/pkg/logger"
)

type uptimeService struct {
	recordRepo  RecordRepository
	mailService MailService
	logger      logger.Logger
}

func NewUptimeService(recordRepo RecordRepository, mailService MailService, logger logger.Logger) UptimeService {
	return &uptimeService{
		recordRepo:  recordRepo,
		mailService: mailService,
		logger:      logger.With("service", "uptime"),
	}
}

func (s *uptimeService) ReportStats(ctx context.Context, req *UptimeRequest) error {
	stats, err := s.recordRepo.GetUptimeStats(ctx, req)
	if err != nil {
		s.logger.Error("failed to get uptime stats", "error", err)
		return err
	}

	if err := s.mailService.SendUptimeReport(stats); err != nil {
		s.logger.Error("failed to send uptime report", "error", err)
		return err
	}

	return nil
}
