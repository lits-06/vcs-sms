package domain

import (
	"context"
	"time"
)

type Repository interface {
	GetUptimeStats(ctx context.Context, startDate, endDate time.Time) (*UptimeStats, error)
}
