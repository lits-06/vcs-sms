package domain

import "context"

type Repository interface {
	GetUptimeStats(ctx context.Context, req *UptimeRequest) (*UptimeStats, error)
}
