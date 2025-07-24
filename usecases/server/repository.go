package server

import (
	"context"
	"time"

	"github.com/lits-06/vcs-sms/entity"
)

// Repository defines the interface for server data operations
type Repository interface {
	// CRUD operations
	Create(ctx context.Context, server *entity.Server) error
	GetByID(ctx context.Context, id string) (*entity.Server, error)
	GetByName(ctx context.Context, name string) (*entity.Server, error)
	Update(ctx context.Context, server *entity.Server) error
	Delete(ctx context.Context, id string) error

	// Query operations
	List(ctx context.Context, filter ServerFilter, sort ServerSort, pagination ServerPagination) (*[]entity.Server, int, error)

	// Validation operations
	ExistsWithID(ctx context.Context, id string) (bool, error)
	ExistsWithName(ctx context.Context, name string) (bool, error)
}

// StatusRecord represents status data stored in Elasticsearch
type StatusRecord struct {
	ServerID  string    `json:"server_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// UptimeStats represents uptime statistics
type UptimeStats struct {
	TotalServers     int     `json:"total_servers"`
	OnlineServers    int     `json:"online_servers"`
	OfflineServers   int     `json:"offline_servers"`
	UptimePercentage float64 `json:"uptime_percentage"`
}

type Record interface {
	Create(ctx context.Context, record *StatusRecord) error
	GetByID(ctx context.Context, id string) (*StatusRecord, error)
	Update(ctx context.Context, record *StatusRecord) error
	Delete(ctx context.Context, id string) error
}
