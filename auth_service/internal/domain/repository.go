package domain

import (
	"context"
	"time"
)

type CacheRepository interface {
	SetRefreshToken(ctx context.Context, token, email string, ttl time.Duration) error
	DeleteRefreshToken(ctx context.Context, token string) error
	GetEmailByRefreshToken(ctx context.Context, token string) (string, error)
	AddBlacklist(ctx context.Context, token string, exp int64) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}
