package domain

import (
	"context"
	"time"
)

type UseCase interface {
	ReportStats(ctx context.Context, email string, startDate, endDate time.Time) error
}
