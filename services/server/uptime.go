package server

import (
	"context"

	"github.com/lits-06/vcs-sms/pkg/logger"
)

type UptimeService struct {
	recordRepo RecordRepository
	logger     logger.Logger
}

func NewUptimeService(recordRepo RecordRepository, logger logger.Logger) *UptimeService {
	return &UptimeService{
		recordRepo: recordRepo,
		logger:     logger.With("service", "uptime"),
	}
}

func (s *UptimeService) GetUptimeStats(ctx context.Context, req *UptimeRequest) error {
	stats, err := s.recordRepo.GetUptimeStats(ctx, req)
	if err != nil {
		s.logger.Error("failed to get uptime stats", "error", err)
		return err
	}
	return nil
}
