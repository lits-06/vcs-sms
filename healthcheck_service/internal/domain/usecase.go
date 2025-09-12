package domain

import "context"

type UseCase interface {
	StartHealthCheckScheduler(ctx context.Context) error
	StopHealthCheckScheduler(ctx context.Context) error
	CheckServerHealth(ctx context.Context, serverID string) (*HealthCheckResult, error)
	GetServerStatus(ctx context.Context, serverID string) (*ServerSnapshot, error)
}
