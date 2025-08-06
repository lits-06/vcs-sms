package server

import (
	"context"
	"sync"
	"time"

	"github.com/lits-06/vcs-sms/entity"
	"github.com/lits-06/vcs-sms/pkg/logger"
)

const (
	WorkerCount       = 10
	ServerChannelSize = 10000
)

type MonitorService struct {
	serverRepo Repository
	cacheRepo  CacheRepository
	recordRepo RecordRepository
	provider   Provider
	logger     logger.Logger

	serverChan chan *entity.Server

	ticker   *time.Ticker
	interval time.Duration

	buffer []*StatusRecord // Buffer để lưu server tạm thời

	stopChan  chan bool
	monitorWg sync.WaitGroup // WaitGroup cho monitoring
}

func NewMonitorService(
	serverRepo Repository,
	cacheRepo CacheRepository,
	recordRepo RecordRepository,
	provider Provider,
	logger logger.Logger,
	interval time.Duration,
) *MonitorService {
	return &MonitorService{
		serverRepo: serverRepo,
		cacheRepo:  cacheRepo,
		recordRepo: recordRepo,
		provider:   provider,
		logger:     logger.With("service", "monitor"),

		serverChan: make(chan *entity.Server, ServerChannelSize),
		stopChan:   make(chan bool),
		buffer:     make([]*StatusRecord, 0, ServerChannelSize), // Buffer size can be adjusted

		ticker:   time.NewTicker(interval), // Example interval
		interval: time.Duration(interval.Seconds()),
	}
}

func (s *MonitorService) Start(ctx context.Context) {
	s.logger.Info("Starting monitoring service")

	s.monitorWg.Add(1)
	go s.monitoringLoop(ctx)
}

// monitoringLoop - vòng lặp monitoring chính
func (s *MonitorService) monitoringLoop(ctx context.Context) {
	defer s.monitorWg.Done()
	defer s.ticker.Stop()

	for {
		select {
		case <-s.ticker.C:
			if err := s.checkAllServers(ctx); err != nil {
				s.logger.Error("Failed to check servers", "error", err)
			}

			s.recordRepo.CreateBatch(ctx, s.buffer)
			s.buffer = s.buffer[:0] // Clear buffer after processing

		case <-s.stopChan:
			s.logger.Info("Monitoring service stopped")
			return

		case <-ctx.Done():
			s.logger.Info("Monitoring service context cancelled")
			return
		}
	}
}

func (s *MonitorService) Stop() {
	s.logger.Info("Stopping monitoring service")
	close(s.stopChan)
	s.monitorWg.Wait()
	close(s.serverChan)
	s.logger.Info("Monitoring service stopped")
}

func (s *MonitorService) checkAllServers(ctx context.Context) error {
	servers, err := s.getAllServers(ctx)
	if err != nil {
		s.logger.Error("Failed to get server list", "error", err)
		return err
	}

	if len(*servers) == 0 {
		return nil
	}

	for _, server := range *servers {
		server := &server

		actualStatus, err := s.provider.GetServerStatus(ctx, server.ID)
		if err != nil {
			s.logger.Warn("Failed to get server status from provider",
				"server_id", server.ID,
				"error", err,
			)
			actualStatus = entity.StatusOffline // Default to offline if check fails
		}

		// Tính interval dựa trên status
		var recordInterval time.Duration
		if actualStatus == entity.StatusOnline {
			recordInterval = s.interval // Sử dụng interval của ticker khi server ON
		} else {
			recordInterval = 0 // Interval = 0 khi server OFF
		}

		// 2. Tạo status record
		record := &StatusRecord{
			ServerID:  server.ID,
			Status:    string(actualStatus),
			Timestamp: time.Now(),
			Interval:  recordInterval,
		}

		// 3. Update database nếu status thay đổi
		if server.Status != actualStatus {
			// Update database
			updatedServer := *server // Copy server
			updatedServer.Status = actualStatus

			if err := s.serverRepo.Update(ctx, &updatedServer); err != nil {
				s.logger.Error("Failed to update server status in database",
					"server_id", server.ID,
					"error", err,
				)
			}

			// Update cache
			if err := s.cacheRepo.SetServer(ctx, server.ID, &updatedServer); err != nil {
				s.logger.Warn("Failed to update server status in cache",
					"server_id", server.ID,
					"error", err,
				)
			}

			if err := s.provider.UpdateServer(ctx, &updatedServer); err != nil {
				s.logger.Warn("Failed to update server status in provider",
					"server_id", server.ID,
					"error", err,
				)
			}
		}

		s.buffer = append(s.buffer, record)
	}

	return nil
}

func (s *MonitorService) getAllServers(ctx context.Context) (*[]entity.Server, error) {
	cacheServers, err := s.cacheRepo.GetServerList(ctx)
	if err == nil && cacheServers != nil && len(*cacheServers) > 0 {
		return cacheServers, nil
	}

	servers, _, err := s.serverRepo.List(ctx, ServerFilter{}, ServerSort{}, ServerPagination{})
	if err != nil {
		return nil, err
	}

	if servers == nil || len(*servers) > 0 {
		if err := s.cacheRepo.SetServerList(ctx, servers); err != nil {
			s.logger.Warn("Failed to set server list in cache", "error", err)
		}
	}

	return servers, nil
}
