package services

// import (
// 	"context"
// 	"fmt"
// 	"net"
// 	"sync"
// 	"time"

// 	"go.uber.org/zap"

// 	"VCS-Checkpoint1/internal/domain"
// 	"VCS-Checkpoint1/internal/repository"
// )

// type MonitoringService struct {
// 	serverRepo    repository.ServerRepository
// 	uptimeService *UptimeService
// 	logger        *zap.Logger
// 	stopChan      chan bool
// 	wg            sync.WaitGroup
// }

// func NewMonitoringService(
// 	serverRepo repository.ServerRepository,
// 	uptimeService *UptimeService,
// 	logger *zap.Logger,
// ) *MonitoringService {
// 	return &MonitoringService{
// 		serverRepo:    serverRepo,
// 		uptimeService: uptimeService,
// 		logger:        logger,
// 		stopChan:      make(chan bool),
// 	}
// }

// func (s *MonitoringService) Start(ctx context.Context, interval time.Duration) {
// 	s.logger.Info("Starting monitoring service", zap.Duration("interval", interval))

// 	ticker := time.NewTicker(interval)
// 	defer ticker.Stop()

// 	// Initial check
// 	s.checkAllServers(ctx)

// 	for {
// 		select {
// 		case <-ticker.C:
// 			s.checkAllServers(ctx)
// 		case <-s.stopChan:
// 			s.logger.Info("Monitoring service stopped")
// 			return
// 		case <-ctx.Done():
// 			s.logger.Info("Monitoring service context cancelled")
// 			return
// 		}
// 	}
// }

// func (s *MonitoringService) Stop() {
// 	s.logger.Info("Stopping monitoring service")
// 	close(s.stopChan)
// 	s.wg.Wait()
// }

// func (s *MonitoringService) checkAllServers(ctx context.Context) {
// 	servers, err := s.serverRepo.GetAll(ctx)
// 	if err != nil {
// 		s.logger.Error("Failed to get servers for monitoring", zap.Error(err))
// 		return
// 	}

// 	s.logger.Info("Checking server status", zap.Int("server_count", len(servers)))

// 	// Use worker pool to check servers concurrently
// 	const workerCount = 10
// 	serverChan := make(chan *domain.Server, len(servers))

// 	// Start workers
// 	for i := 0; i < workerCount; i++ {
// 		s.wg.Add(1)
// 		go s.worker(ctx, serverChan)
// 	}

// 	// Send servers to workers
// 	for _, server := range servers {
// 		serverChan <- server
// 	}
// 	close(serverChan)

// 	s.wg.Wait()
// }

// func (s *MonitoringService) worker(ctx context.Context, serverChan <-chan *domain.Server) {
// 	defer s.wg.Done()

// 	for server := range serverChan {
// 		s.checkServer(ctx, server)
// 	}
// }

// func (s *MonitoringService) checkServer(ctx context.Context, server *domain.Server) {
// 	isOnline := s.pingServer(server.Host, server.Port)

// 	newStatus := "offline"
// 	if isOnline {
// 		newStatus = "online"
// 	}

// 	// Update status if changed
// 	if server.Status != newStatus {
// 		server.Status = newStatus
// 		server.LastChecked = time.Now()

// 		if err := s.serverRepo.Update(ctx, server); err != nil {
// 			s.logger.Error("Failed to update server status",
// 				zap.Error(err),
// 				zap.Int64("server_id", server.ID),
// 				zap.String("new_status", newStatus),
// 			)
// 		}
// 	}

// 	// Record uptime data
// 	if err := s.uptimeService.RecordServerStatus(ctx, server.ID, newStatus); err != nil {
// 		s.logger.Error("Failed to record server status",
// 			zap.Error(err),
// 			zap.Int64("server_id", server.ID),
// 			zap.String("status", newStatus),
// 		)
// 	}

// 	s.logger.Debug("Server check completed",
// 		zap.Int64("server_id", server.ID),
// 		zap.String("host", server.Host),
// 		zap.Int("port", server.Port),
// 		zap.String("status", newStatus),
// 	)
// }

// func (s *MonitoringService) pingServer(host string, port int) bool {
// 	timeout := 5 * time.Second
// 	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
// 	if err != nil {
// 		return false
// 	}
// 	defer conn.Close()
// 	return true
// }
