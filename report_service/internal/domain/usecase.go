package domain

import "context"

type UseCase interface {
	ReportStats(ctx context.Context, req *UptimeRequest) error
}
