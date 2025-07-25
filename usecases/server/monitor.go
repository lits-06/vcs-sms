package server

import (
	"context"
	"sync"
	"time"

	"github.com/lits-06/vcs-sms/entity"
	"github.com/lits-06/vcs-sms/pkg/logger"
)

const (
	ServerListCacheKey = "server"
	WorkerCount        = 10
	ServerChannelSize  = 10000
)

type MonitorService struct {
	serverRepo    Repository
	cacheRepo     CacheRepository
	recordRepo    RecordRepository
	kafkaProducer Producer
	provider      Provider
	ticker        *time.Ticker
	logger        logger.Logger

	serverChan     chan *entity.Server
	stopChan       chan bool
	workerWg       sync.WaitGroup // WaitGroup cho workers
	monitorWg      sync.WaitGroup // WaitGroup cho monitoring
	wg             sync.WaitGroup
	workersStarted bool
}

func NewMonitorService(
	serverRepo ServerRepository,
	cacheRepo CacheRepository,
	recordRepo RecordRepository,
	kafkaProducer Producer,
	provider Provider,
	logger logger.Logger,
	interval time.Duration,
) *MonitorService {
	return &MonitorService{
		serverRepo:    serverRepo,
		cacheRepo:     cacheRepo,
		recordRepo:    recordRepo,
		kafkaProducer: kafkaProducer,
		provider:      provider,
		ticker:        time.NewTicker(interval), // Example interval
		logger:        logger,
		stopChan:      make(chan bool),
	}
}

func (s *MonitorService) Start() {
	s.logger.Info("Starting monitoring service")

	// Initial check
	s.checkAllServers()

	go func() {
		defer s.ticker.Stop()

		for {
			select {
			case <-s.ticker.C:
				s.checkAllServers()
			case <-s.stopChan:
				s.logger.Info("Monitoring service stopped")
				return
			}
		}
	}()
}

// startWorkerPool khởi tạo worker pool một lần duy nhất
func (s *MonitorService) startWorkerPool(ctx context.Context) {
	if s.workersStarted {
		return
	}

	for i := 0; i < WorkerCount; i++ {
		s.workerWg.Add(1)
		go s.persistentWorker(ctx, i)
	}

	s.workersStarted = true
}

// persistentWorker - worker chạy liên tục, không bị tạo lại
func (s *MonitorService) persistentWorker(ctx context.Context, workerID int) {
	defer s.workerWg.Done()

	for {
		select {
		case server, ok := <-s.serverChan:
			if !ok {
				return
			}

			if err := s.checkSingleServer(ctx, server); err != nil {
				s.logger.Error("Failed to check server",
					"server_id", server.ID,
					"server_name", server.Name,
					"error", err,
				)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (s *MonitorService) Stop() {
	s.logger.Info("Stopping monitoring service")
	close(s.stopChan)
	s.wg.Wait()
}

func (s *MonitorService) checkAllServers() error {
	servers, err := s.getAllServers()
	if err != nil {
		s.logger.Error("Failed to get server list", "error", err)
		return nil, err
	}

	if len(*servers) == 0 {
		return nil
	}

}

func (s *MonitorService) getAllServers() (*[]entity.Server, error) {
	cacheServers, err := s.cacheRepo.GetServerList(context.Background())
	if err != nil && cacheServers != nil {
		return cacheServers, nil
	}

	servers, err := s.serverRepo.List(context.Background(), ServerFilter{}, ServerSort{}, ServerPagination{})
	if err != nil {
		return nil, err
	}

	if err := s.cacheRepo.SetServerList(context.Background(), servers); err != nil {
		s.logger.Warn("Failed to set server list in cache", "error", err)
	}

	return servers, nil
}
