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

func (r *redisRepository) GetAllServersSnapshot(ctx context.Context) (*[]domain.Server, error) {
	var (
		cursor  uint64
		servers []domain.Server
	)

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, r.snapshotKey+"*", 1000).Result()
		if err != nil {
			return nil, err
		}

		if len(keys) > 0 {
			values, err := r.client.MGet(ctx, keys...).Result()
			if err != nil {
				return nil, err
			}

			for _, val := range values {
				if val == nil {
					continue
				}
				var server domain.Server
				if err := json.Unmarshal([]byte(val.(string)), &server); err == nil {
					servers = append(servers, server)
				}
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return &servers, nil
}

func (r *redisRepository) SetAllServersSnapshot(ctx context.Context, servers *[]domain.Server) error {
	pipe := r.client.Pipeline()
	for _, server := range *servers {
		sbytes, err := json.Marshal(server)
		if err != nil {
			return err
		}
		pipe.Set(ctx, r.snapshotKey+":"+server.ServerID, sbytes, 0)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *redisRepository) GetServerSnapshot(ctx context.Context, serverID string) (*domain.Server, error) {
	val, err := r.client.Get(ctx, r.snapshotKey+":"+serverID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var server domain.Server
	if err := json.Unmarshal([]byte(val), &server); err != nil {
		return nil, err
	}

	return &server, nil
}

func (r *redisRepository) SaveServerSnapshot(ctx context.Context, server *domain.Server) error {
	sbytes, err := json.Marshal(server)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, r.snapshotKey+":"+server.ServerID, sbytes, 0).Err()
}

func (r *redisRepository) DeleteServerSnapshot(ctx context.Context, serverID string) error {
	err := r.client.Del(ctx, r.snapshotKey+":"+serverID).Err()
	if err != nil && err != redis.Nil {
		return nil
	}
	return err
}
