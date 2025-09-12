package repository

import (
	"context"
	"encoding/json"

	"github.com/lits-06/vcs-sms/healthcheck_service/config"
	"github.com/lits-06/vcs-sms/healthcheck_service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type redisRepository struct {
	client      *redis.Client
	snapshotKey string
}

func NewCacheRepository(client *redis.Client, cfg *config.Config) domain.CacheRepository {
	return &redisRepository{
		client:      client,
		snapshotKey: cfg.Redis.SnapshotKey,
	}
}

func (r *redisRepository) GetServerState(ctx context.Context, serverID string) (*domain.ServerSnapshot, error) {
	val, err := r.client.Get(ctx, r.snapshotKey+serverID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var snapshot domain.ServerSnapshot
	if err := json.Unmarshal([]byte(val), &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}
