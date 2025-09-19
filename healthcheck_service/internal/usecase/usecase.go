package usecase

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/lits-06/vcs-sms/healthcheck_service/config"
	"github.com/lits-06/vcs-sms/healthcheck_service/internal/domain"
	"github.com/lits-06/vcs-sms/pkg/logger"
)

type healthCheckUseCase struct {
	repo      domain.Repository
	cacheRepo domain.CacheRepository
	publisher domain.EventPublisher

	log        logger.Logger
	ticker     *time.Ticker
	workerPool chan struct{}
}

func NewHealthCheckUseCase(cfg *config.Config, log logger.Logger, repo domain.Repository, cacheRepo domain.CacheRepository, publisher domain.EventPublisher) domain.UseCase {
	return &healthCheckUseCase{
		repo:      repo,
		cacheRepo: cacheRepo,
		publisher: publisher,
		log:       log,
		ticker:    time.NewTicker(cfg.Duration),
	}
}

func (h *healthCheckUseCase) StartHealthCheckScheduler(ctx context.Context) {
	for {
		select {
		case <-h.ticker.C:
			servers, err := h.getAllServersSnapshot(ctx)
			if err != nil {
				h.log.Error("getAllServer: %v", err)
				continue
			}

			// Check health of all servers concurrently
			h.CheckServersHealth(ctx, servers)

		case <-ctx.Done():
			h.log.Info("Health check scheduler stopped")
			return
		}
	}
}

func (h *healthCheckUseCase) StopHealthCheckScheduler(ctx context.Context) {
	h.ticker.Stop()
}

func (h *healthCheckUseCase) CheckServersHealth(ctx context.Context, servers *[]domain.Server) {
	var wg sync.WaitGroup

	for i, server := range *servers {
		wg.Add(1)
		go func(i int, server *domain.Server) {
			defer wg.Done()
			snapshot, err := h.checkServerHealth(server)
			if err != nil {
				h.log.Error("checkServerHealth: %v", err)
				return
			}

			if snapshot.Status != server.Status {
				state := domain.Server{
					ServerID:  server.ServerID,
					Port:      server.Port,
					Status:    snapshot.Status,
					Timestamp: time.Now(),
				}
				err = h.publisher.PublishStateChange(ctx, &state)
				if err != nil {
					h.log.Error("PublishStateChange: %v", err)
					return
				}
			}
		}(i, &server)
	}

	wg.Wait()
	h.log.Info("Health check completed for all servers")
}

func (h *healthCheckUseCase) getAllServersSnapshot(ctx context.Context) (*[]domain.Server, error) {
	servers, err := h.cacheRepo.GetAllServersSnapshot(ctx)
	if err == nil && len(*servers) > 0 {
		h.log.Debug("Retrieved %d servers from cache", len(*servers))
		return servers, nil
	}

	if err != nil {
		h.log.Debug("Cache miss for servers: %v", err)
	}

	servers, err = h.repo.GetAllServersSnapshot(ctx)
	if err != nil {
		h.log.Error("repo.GetAllServersSnapshot: %v", err)
		return nil, fmt.Errorf("repo.GetAllServersSnapshot: %w", err)
	}

	if len(*servers) > 0 {
		if err = h.cacheRepo.SetAllServersSnapshot(ctx, servers); err != nil {
			h.log.Warn("cacheRepo.SetAllServersSnapshot: %v", err)
		}
	}

	h.log.Debug("Retrieved %d servers from DB", len(*servers))
	return servers, nil
}

func (h *healthCheckUseCase) checkServerHealth(server *domain.Server) (*domain.Server, error) {
	addr := fmt.Sprintf("localhost:%d/health", server.Port)
	timeout := 1 * time.Second
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	responseTime := time.Since(start)
	if err != nil {
		h.log.Debug("Failed to ping id:%s port:%d (took %v): %v", server.ServerID, server.Port, responseTime, err)
		return &domain.Server{
			ServerID: server.ServerID,
			Port:     server.Port,
			Status:   domain.StatusOffline,
		}, err
	}

	defer conn.Close()

	return &domain.Server{
		ServerID: server.ServerID,
		Port:     server.Port,
		Status:   domain.StatusOnline,
	}, nil
}

func (h *healthCheckUseCase) IndexServerState(ctx context.Context, server *domain.Server) error {
	err := h.repo.IndexServerState(ctx, server)
	if err != nil {
		h.log.Error("h.repo.IndexServerState: %v", err)
		return err
	}

	err = h.repo.SaveServerSnapshot(ctx, server)
	if err != nil {
		h.log.Error("h.repo.SaveServerSnapshot: %v", err)
		return err
	}

	err = h.cacheRepo.SaveServerSnapshot(ctx, server)
	if err != nil {
		h.log.Error("h.cacheRepo.SaveServerSnapshot: %v", err)
		return err
	}

	return nil
}
