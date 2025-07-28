package server

import (
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
