package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lits-06/vcs-sms/entity"
	"github.com/lits-06/vcs-sms/pkg/logger"
)

const (
	ServerListCacheKey = "monitor:servers:active"
	WorkerCount        = 10
	ServerChannelSize  = 10000
	ServerCacheTTL     = 5 * time.Minute
	ServerListTTL      = 2 * time.Minute
)

type MonitorService struct {
	serverRepo    Repository
	cacheRepo     CacheRepository
	recordRepo    RecordRepository
	kafkaProducer Producer
	provider      Provider
	ticker        *time.Ticker
	logger        logger.Logger

	// Channel và worker pool tối ưu
	serverChan     chan *entity.Server // Persistent channel
	stopChan       chan bool
	workerWg       sync.WaitGroup // WaitGroup cho workers
	monitorWg      sync.WaitGroup // WaitGroup cho monitoring
	workersStarted bool
}

func NewMonitorService(
	serverRepo Repository,
	cacheRepo CacheRepository,
	recordRepo RecordRepository,
	kafkaProducer Producer,
	provider Provider,
	logger logger.Logger,
	interval time.Duration,
) *MonitorService {
	return &MonitorService{
		serverRepo:     serverRepo,
		cacheRepo:      cacheRepo,
		recordRepo:     recordRepo,
		kafkaProducer:  kafkaProducer,
		provider:       provider,
		ticker:         time.NewTicker(interval),
		logger:         logger,
		serverChan:     make(chan *entity.Server, ServerChannelSize), // Tạo 1 lần duy nhất
		stopChan:       make(chan bool),
		workersStarted: false,
	}
}

func (s *MonitorService) Start(ctx context.Context) {
	s.logger.Info("Starting monitoring service")

	// Start worker pool một lần duy nhất
	s.startWorkerPool(ctx)

	// Initial check
	if err := s.checkAllServers(ctx); err != nil {
		s.logger.Error("Failed initial server check", "error", err)
	}

	// Start monitoring goroutine
	s.monitorWg.Add(1)
	go s.monitoringLoop(ctx)
}

// startWorkerPool khởi tạo worker pool một lần duy nhất
func (s *MonitorService) startWorkerPool(ctx context.Context) {
	if s.workersStarted {
		return
	}

	s.logger.Info("Starting worker pool", "worker_count", WorkerCount)

	for i := 0; i < WorkerCount; i++ {
		s.workerWg.Add(1)
		go s.persistentWorker(ctx, i)
	}

	s.workersStarted = true
}

// persistentWorker - worker chạy liên tục, không bị tạo lại
func (s *MonitorService) persistentWorker(ctx context.Context, workerID int) {
	defer s.workerWg.Done()

	workerLogger := s.logger.With("worker_id", workerID)
	workerLogger.Debug("Worker started")

	for {
		select {
		case server, ok := <-s.serverChan:
			if !ok {
				workerLogger.Debug("Server channel closed, worker stopping")
				return
			}

			if err := s.checkSingleServer(ctx, server); err != nil {
				workerLogger.Error("Failed to check server",
					"server_id", server.ID,
					"server_name", server.Name,
					"error", err,
				)
			}

		case <-ctx.Done():
			workerLogger.Debug("Context cancelled, worker stopping")
			return
		}
	}
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

	// Signal monitoring loop to stop
	close(s.stopChan)

	// Wait for monitoring loop to finish
	s.monitorWg.Wait()

	// Close server channel to signal workers to stop
	close(s.serverChan)

	// Wait for all workers to finish
	s.workerWg.Wait()

	s.logger.Info("All workers and monitoring stopped")
}

// checkAllServers tối ưu - không tạo slice mới mỗi lần
func (s *MonitorService) checkAllServers(ctx context.Context) error {
	// 1. Lấy danh sách servers từ cache/database
	servers, err := s.getServersWithCaching(ctx)
	if err != nil {
		return fmt.Errorf("failed to get servers: %w", err)
	}

	if servers == nil || len(*servers) == 0 {
		s.logger.Debug("No servers found for monitoring")
		return nil
	}

	serverCount := len(*servers)
	s.logger.Debug("Starting server health checks", "server_count", serverCount)

	// 2. Đẩy servers vào channel - không cần tạo worker mới
	for i := range *servers {
		server := &(*servers)[i] // Lấy pointer để tránh copy

		select {
		case s.serverChan <- server:
			// Successfully sent to channel
		case <-ctx.Done():
			s.logger.Info("Context cancelled while sending servers to channel")
			return ctx.Err()
		default:
			// Channel is full, log warning
			s.logger.Warn("Server channel is full, skipping server",
				"server_id", server.ID,
				"channel_size", len(s.serverChan),
			)
		}
	}

	s.logger.Debug("Completed sending servers to workers", "server_count", serverCount)
	return nil
}

// getServersWithCaching tối ưu cache strategy
func (s *MonitorService) getServersWithCaching(ctx context.Context) (*[]entity.Server, error) {
	// 1. Thử lấy từ cache trước
	if s.cacheRepo != nil {
		servers, err := s.cacheRepo.GetServerList(ctx, ServerListCacheKey)
		if err == nil && servers != nil && len(*servers) > 0 {
			s.logger.Debug("Loaded servers from cache", "count", len(*servers))
			return servers, nil
		}
		s.logger.Debug("Cache miss for server list", "error", err)
	}

	// 2. Fallback: lấy từ database với GetAllActive
	servers, err := s.serverRepo.GetAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active servers from database: %w", err)
	}

	// 3. Cache lại kết quả
	if s.cacheRepo != nil && servers != nil && len(*servers) > 0 {
		if err := s.cacheRepo.SetServerList(ctx, ServerListCacheKey, servers, ServerListTTL); err != nil {
			s.logger.Warn("Failed to cache server list", "error", err)
		} else {
			s.logger.Debug("Cached server list", "count", len(*servers))
		}
	}

	return servers, nil
}

// checkSingleServer kiểm tra status của một server
func (s *MonitorService) checkSingleServer(ctx context.Context, server *entity.Server) error {
	startTime := time.Now()

	// 1. Check status through Provider
	actualStatus, err := s.provider.GetServerStatus(ctx, server.ID)
	if err != nil {
		s.logger.Warn("Failed to get server status from provider",
			"server_id", server.ID,
			"error", err,
		)
		actualStatus = entity.StatusOffline // Default to offline if check fails
	}

	responseTime := time.Since(startTime).Milliseconds()

	// 2. Tạo status record
	record := &StatusRecord{
		ServerID:  server.ID,
		Status:    string(actualStatus),
		Timestamp: time.Now(),
	}

	// 3. Update database nếu status thay đổi
	if server.Status != actualStatus {
		s.logger.Info("Server status changed",
			"server_id", server.ID,
			"old_status", string(server.Status),
			"new_status", string(actualStatus),
		)

		// Update database
		updatedServer := *server // Copy server
		updatedServer.Status = actualStatus
		updatedServer.UpdatedAt = time.Now()

		if err := s.serverRepo.Update(ctx, &updatedServer); err != nil {
			s.logger.Error("Failed to update server status in database",
				"server_id", server.ID,
				"error", err,
			)
		} else {
			// Update the original server object
			server.Status = actualStatus
			server.UpdatedAt = updatedServer.UpdatedAt
		}

		// Invalidate cache khi có thay đổi
		if s.cacheRepo != nil {
			s.cacheRepo.DeleteServer(ctx, server.ID)
			s.cacheRepo.DeleteServerList(ctx, ServerListCacheKey)
		}
	}

	// 4. Send to Kafka (async)
	if s.kafkaProducer != nil {
		go func() {
			if err := s.kafkaProducer.SendServerStatus(context.Background(), record); err != nil {
				s.logger.Error("Failed to send status to Kafka",
					"server_id", server.ID,
					"error", err,
				)
			}
		}()
	}

	// 5. Store in Elasticsearch (async)
	if s.recordRepo != nil {
		go func() {
			if err := s.recordRepo.Create(context.Background(), record); err != nil {
				s.logger.Error("Failed to store record in Elasticsearch",
					"server_id", server.ID,
					"error", err,
				)
			}
		}()
	}

	s.logger.Debug("Server check completed",
		"server_id", server.ID,
		"status", string(actualStatus),
		"response_time_ms", responseTime,
	)

	return nil
}

// GetServerHealthStats trả về thống kê hiện tại của monitoring service
func (s *MonitorService) GetServerHealthStats() map[string]interface{} {
	return map[string]interface{}{
		"workers_active":    s.workersStarted,
		"worker_count":      WorkerCount,
		"channel_capacity":  ServerChannelSize,
		"channel_length":    len(s.serverChan),
		"channel_available": ServerChannelSize - len(s.serverChan),
	}
}
