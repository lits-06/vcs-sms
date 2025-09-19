package repository

import (
	"context"
	"time"

	"github.com/lits-06/vcs-sms/auth_service/config"
	"github.com/lits-06/vcs-sms/auth_service/internal/domain"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
)

type redisRepository struct {
	client       *redis.Client
	refreshKey   string
	blacklistKey string
}

func NewCacheRepository(client *redis.Client, cfg *config.Config) domain.CacheRepository {
	return &redisRepository{
		client:       client,
		refreshKey:   cfg.Redis.RefreshKey,
		blacklistKey: cfg.Redis.BlacklistKey,
	}
}

func (r *redisRepository) SetRefreshToken(ctx context.Context, email, refreshToken string, expiration time.Duration) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "redisRepository.SetRefreshToken")
	defer span.Finish()

	return r.client.Set(ctx, r.refreshKey+refreshToken, email, expiration).Err()
}

func (r *redisRepository) GetEmailByRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "redisRepository.GetEmailByRefreshToken")
	defer span.Finish()

	email, err := r.client.Get(ctx, r.refreshKey+refreshToken).Result()
	if err != nil {
		return "", err
	}
	return email, nil
}

func (r *redisRepository) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "redisRepository.DeleteRefreshToken")
	defer span.Finish()

	return r.client.Del(ctx, r.refreshKey+refreshToken).Err()
}

func (r *redisRepository) AddBlacklist(ctx context.Context, token string, expiration int64) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "redisRepository.AddBlacklist")
	defer span.Finish()

	return r.client.Set(ctx, r.blacklistKey+token, "blacklisted", time.Duration(expiration)*time.Second).Err()
}

func (r *redisRepository) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "redisRepository.IsBlacklisted")
	defer span.Finish()

	result, err := r.client.Get(ctx, r.blacklistKey+token).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return result == "blacklisted", nil
}
