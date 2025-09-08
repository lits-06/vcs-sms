package domain

import (
	"context"
	"time"
)

type Repository interface {
	// GetUptimeStats calculates uptime statistics for all servers in a given time range
	GetUptimeStats(ctx context.Context, req *UptimeRequest) (*UptimeStats, error)
	
	// GetServerUptimeStats calculates uptime statistics for specific servers
	GetServerUptimeStats(ctx context.Context, serverIDs []string, startDate, endDate time.Time) ([]ServerUptimeDetail, error)
	
	// GetServerCurrentStatus gets current status of servers
	GetServerCurrentStatus(ctx context.Context, serverIDs []string) (map[string]ServerSnapshot, error)
	
	// GetServerStateEvents gets state change events for servers in a time range
	GetServerStateEvents(ctx context.Context, serverIDs []string, startDate, endDate time.Time) ([]ServerStateRecord, error)
}
