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
	ServerID  string        `json:"server_id"`
	Status    string        `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Interval  time.Duration `json:"interval"` // in seconds
}

type UptimeRequest struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
	Email     string    `json:"email" validate:"omitempty,email"`
}

// UptimeStats represents uptime statistics
type UptimeStats struct {
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	TotalServers     int       `json:"total_servers"`
	OnlineServers    int       `json:"online_servers"`
	OfflineServers   int       `json:"offline_servers"`
	UptimePercentage float64   `json:"uptime_percentage"`
}

type RecordRepository interface {
	// Create(ctx context.Context, record *StatusRecord) error
	CreateBatch(ctx context.Context, records []*StatusRecord) error
	GetUptimeStats(ctx context.Context, req *UptimeRequest) (*UptimeStats, error)
}

type CacheRepository interface {
	GetServerList(ctx context.Context) (*[]entity.Server, error)
	SetServerList(ctx context.Context, servers *[]entity.Server) error
	GetServer(ctx context.Context, serverID string) (*entity.Server, error)
	SetServer(ctx context.Context, serverID string, server *entity.Server) error
	DeleteServer(ctx context.Context, serverID string) error
	DeleteServerList(ctx context.Context) error
}

type Provider interface {
	CreateServer(ctx context.Context, server *entity.Server) error
	DeleteServer(ctx context.Context, serverID string) error
	UpdateServer(ctx context.Context, server *entity.Server) error
	StartServer(ctx context.Context, serverID string) error
	StopServer(ctx context.Context, serverID string) error
	GetServerStatus(ctx context.Context, serverID string) (entity.ServerStatus, error)
}

type MailService interface {
	SendUptimeReport(ctx context.Context, email string, stats *UptimeStats) error
}
