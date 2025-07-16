package cache

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"time"

// 	"github.com/lits-06/vcs-sms/internal/domain"
// 	"github.com/lits-06/vcs-sms/internal/infrastructure/redis"
// )

// type cacheRepository struct {
// 	client *redis.Client
// }

// // NewCacheRepository creates a new cache repository
// func NewCacheRepository(client *redis.Client) domain.CacheRepository {
// 	return &cacheRepository{client: client}
// }

// // Set stores a value in cache
// func (r *cacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
// 	data, err := json.Marshal(value)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal value: %w", err)
// 	}

// 	err = r.client.Set(ctx, key, data, expiration).Err()
// 	if err != nil {
// 		return fmt.Errorf("failed to set cache: %w", err)
// 	}

// 	return nil
// }

// // Get retrieves a value from cache
// func (r *cacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
// 	data, err := r.client.Get(ctx, key).Result()
// 	if err != nil {
// 		return fmt.Errorf("failed to get cache: %w", err)
// 	}

// 	err = json.Unmarshal([]byte(data), dest)
// 	if err != nil {
// 		return fmt.Errorf("failed to unmarshal value: %w", err)
// 	}

// 	return nil
// }

// // Delete removes a value from cache
// func (r *cacheRepository) Delete(ctx context.Context, key string) error {
// 	err := r.client.Del(ctx, key).Err()
// 	if err != nil {
// 		return fmt.Errorf("failed to delete cache: %w", err)
// 	}

// 	return nil
// }

// // DeletePattern removes all keys matching a pattern
// func (r *cacheRepository) DeletePattern(ctx context.Context, pattern string) error {
// 	keys, err := r.client.Keys(ctx, pattern).Result()
// 	if err != nil {
// 		return fmt.Errorf("failed to get keys: %w", err)
// 	}

// 	if len(keys) > 0 {
// 		err = r.client.Del(ctx, keys...).Err()
// 		if err != nil {
// 			return fmt.Errorf("failed to delete keys: %w", err)
// 		}
// 	}

// 	return nil
// }
