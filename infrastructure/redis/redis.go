package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/entity"
	"github.com/lits-06/vcs-sms/services/server"
	"github.com/redis/go-redis/v9"
)

const (
	// Cache keys
	ServerKey        = "server:%s"         // server:{id}
	ServerProcessKey = "server_process:%s" // server_process:{id}
)

func NewRedisClient(cfg *config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return rdb, nil
}

type redisRepository struct {
	client *redis.Client
}

func NewCacheRepository(client *redis.Client) server.CacheRepository {
	return &redisRepository{
		client: client,
	}
}

// GetServerList lấy danh sách server từ cache
func (r *redisRepository) GetServerList(ctx context.Context) (*[]entity.Server, error) {
	var allKeys []string
	var cursor uint64

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, "server:*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan server keys: %w", err)
		}

		allKeys = append(allKeys, keys...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	if len(allKeys) == 0 {
		return nil, nil
	}

	values, err := r.client.MGet(ctx, allKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get server values: %w", err)
	}

	servers := make([]entity.Server, 0, len(values))

	for i, value := range values {
		if value == nil {
			// Key đã expire hoặc bị xóa trong lúc query
			continue
		}

		valueStr, ok := value.(string)
		if !ok {
			continue
		}

		var srv entity.Server
		if err := json.Unmarshal([]byte(valueStr), &srv); err != nil {
			// Log warning nhưng không fail toàn bộ operation
			fmt.Printf("Failed to unmarshal server from key %s: %v\n", allKeys[i], err)
			continue
		}

		servers = append(servers, srv)
	}

	if len(servers) == 0 {
		return nil, nil
	}

	return &servers, nil
}

// SetServerList lưu danh sách server vào cache
func (r *redisRepository) SetServerList(ctx context.Context, servers *[]entity.Server) error {
	if servers == nil || len(*servers) == 0 {
		return nil
	}

	// Sử dụng Pipeline để batch operations
	pipe := r.client.Pipeline()

	for _, srv := range *servers {
		data, err := json.Marshal(srv)
		if err != nil {
			fmt.Printf("Failed to marshal server %s: %v\n", srv.ID, err)
			continue
		}

		key := fmt.Sprintf(ServerKey, srv.ID)
		pipe.Set(ctx, key, data, 0)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute pipeline for SetServerList: %w", err)
	}

	return nil
}

// GetServer lấy thông tin server từ cache theo ID
func (r *redisRepository) GetServer(ctx context.Context, serverID string) (*entity.Server, error) {
	if serverID == "" {
		return nil, fmt.Errorf("server ID cannot be empty")
	}

	key := fmt.Sprintf(ServerKey, serverID)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get server from cache: %w", err)
	}

	var srv entity.Server
	if err := json.Unmarshal([]byte(data), &srv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal server: %w", err)
	}

	return &srv, nil
}

// SetServer lưu thông tin server vào cache
func (r *redisRepository) SetServer(ctx context.Context, serverID string, srv *entity.Server) error {
	if serverID == "" {
		return fmt.Errorf("server ID cannot be empty")
	}
	if srv == nil {
		return fmt.Errorf("server cannot be nil")
	}

	data, err := json.Marshal(srv)
	if err != nil {
		return fmt.Errorf("failed to marshal server: %w", err)
	}

	key := fmt.Sprintf(ServerKey, serverID)
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to set server in cache: %w", err)
	}

	return nil
}

// DeleteServer xóa server khỏi cache
func (r *redisRepository) DeleteServer(ctx context.Context, serverID string) error {
	if serverID == "" {
		return fmt.Errorf("server ID cannot be empty")
	}

	key := fmt.Sprintf(ServerKey, serverID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete server from cache: %w", err)
	}

	return nil
}

// DeleteServerList xóa danh sách server khỏi cache
func (r *redisRepository) DeleteServerList(ctx context.Context) error {
	var cursor uint64

	for {
		// Scan keys với batch size nhỏ
		keys, nextCursor, err := r.client.Scan(ctx, cursor, "server:*", 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan server keys: %w", err)
		}

		if len(keys) > 0 {
			// Sử dụng Pipeline để batch delete
			pipe := r.client.Pipeline()
			for _, key := range keys {
				pipe.Del(ctx, key)
			}

			// Execute pipeline
			if _, err := pipe.Exec(ctx); err != nil {
				return fmt.Errorf("failed to execute delete pipeline: %w", err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
