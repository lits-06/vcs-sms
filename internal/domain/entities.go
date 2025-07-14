package domain

import (
	"context"
	"time"
)

// Server represents a server entity
// type Server struct {
// 	ID          int64     `json:"id" db:"id"`
// 	Name        string    `json:"name" db:"name" validate:"required,min=1,max=255"`
// 	Host        string    `json:"host" db:"host" validate:"required,hostname_or_ip"`
// 	Port        int       `json:"port" db:"port" validate:"required,min=1,max=65535"`
// 	Status      string    `json:"status" db:"status"`
// 	Description string    `json:"description" db:"description"`
// 	Tags        []string  `json:"tags" db:"tags"`
// 	CreatedAt   time.Time `json:"created_at" db:"created_at"`
// 	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
// 	LastChecked time.Time `json:"last_checked" db:"last_checked"`
// }

// ServerStatus represents server status constants
// type ServerStatus string

// const (
// 	StatusOnline  ServerStatus = "online"
// 	StatusOffline ServerStatus = "offline"
// 	StatusUnknown ServerStatus = "unknown"
// )

// UptimeRecord represents uptime tracking data
type UptimeRecord struct {
	ID        int64     `json:"id" db:"id"`
	ServerID  int64     `json:"server_id" db:"server_id"`
	Status    string    `json:"status" db:"status"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// ServerFilter represents filtering options for server queries
type ServerFilter struct {
	Name   string       `json:"name,omitempty"`
	Host   string       `json:"host,omitempty"`
	Status ServerStatus `json:"status,omitempty"`
	Tags   []string     `json:"tags,omitempty"`
}

// ServerSort represents sorting options
type ServerSort struct {
	Field string `json:"field" validate:"oneof=id name host port status created_at updated_at last_checked"`
	Order string `json:"order" validate:"oneof=asc desc"`
}

// Pagination represents pagination parameters
type Pagination struct {
	Page  int `json:"page" validate:"min=1"`
	Limit int `json:"limit" validate:"min=1,max=100"`
}

// ServerListRequest represents request for listing servers
type ServerListRequest struct {
	Filter     ServerFilter `json:"filter"`
	Sort       ServerSort   `json:"sort"`
	Pagination Pagination   `json:"pagination"`
}

// ServerListResponse represents response for listing servers
type ServerListResponse struct {
	Servers    []Server `json:"servers"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	TotalPages int      `json:"total_pages"`
}

// CreateServerRequest represents request for creating a server
type CreateServerRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=255"`
	Host        string   `json:"host" validate:"required,hostname_or_ip"`
	Port        int      `json:"port" validate:"required,min=1,max=65535"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// UpdateServerRequest represents request for updating a server
type UpdateServerRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Host        *string  `json:"host,omitempty" validate:"omitempty,hostname_or_ip"`
	Port        *int     `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	Description *string  `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ImportServerRequest represents request for importing servers
type ImportServerRequest struct {
	Servers []CreateServerRequest `json:"servers" validate:"required,dive"`
}

// DailyReport represents daily server status report
type DailyReport struct {
	Date                time.Time            `json:"date"`
	TotalServers        int64                `json:"total_servers"`
	OnlineServers       int64                `json:"online_servers"`
	OfflineServers      int64                `json:"offline_servers"`
	AverageUptimeRatio  float64              `json:"average_uptime_ratio"`
	ServerUptimeDetails []ServerUptimeDetail `json:"server_uptime_details"`
}

// ServerUptimeDetail represents uptime details for a specific server
type ServerUptimeDetail struct {
	ServerID    int64   `json:"server_id"`
	ServerName  string  `json:"server_name"`
	UptimeRatio float64 `json:"uptime_ratio"`
	TotalChecks int64   `json:"total_checks"`
	OnlineTime  int64   `json:"online_time"` // in minutes
}

// User represents a user entity
type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Roles     []string  `json:"roles" db:"roles"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

// Repository interfaces

// ServerRepository defines methods for server data access
type ServerRepository interface {
	Create(ctx context.Context, server *Server) error
	GetByID(ctx context.Context, id int64) (*Server, error)
	Update(ctx context.Context, id int64, server *Server) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, req ServerListRequest) (*ServerListResponse, error)
	GetAll(ctx context.Context) ([]Server, error)
	BulkCreate(ctx context.Context, servers []Server) error
	GetByHostPort(ctx context.Context, host string, port int) (*Server, error)
}

// UptimeRepository defines methods for uptime data access
type UptimeRepository interface {
	RecordUptime(ctx context.Context, record *UptimeRecord) error
	GetUptimeStats(ctx context.Context, serverID int64, from, to time.Time) (*ServerUptimeDetail, error)
	GetDailyReport(ctx context.Context, date time.Time) (*DailyReport, error)
	BulkRecordUptime(ctx context.Context, records []UptimeRecord) error
}

// UserRepository defines methods for user data access
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id int64, user *User) error
	Delete(ctx context.Context, id int64) error
}

// CacheRepository defines methods for caching
type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
}

// Use case interfaces

// ServerUseCase defines methods for server business logic
type ServerUseCase interface {
	Create(ctx context.Context, req CreateServerRequest) (*Server, error)
	GetByID(ctx context.Context, id int64) (*Server, error)
	Update(ctx context.Context, id int64, req UpdateServerRequest) (*Server, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, req ServerListRequest) (*ServerListResponse, error)
	Import(ctx context.Context, req ImportServerRequest) error
	Export(ctx context.Context, filter ServerFilter) ([]byte, error)
}

// MonitorUseCase defines methods for server monitoring
type MonitorUseCase interface {
	CheckServerStatus(ctx context.Context, server *Server) error
	StartMonitoring(ctx context.Context) error
	StopMonitoring() error
}

// AuthUseCase defines methods for authentication
type AuthUseCase interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (*User, error)
	HasPermission(ctx context.Context, userID int64, scope string) (bool, error)
}

// ReportUseCase defines methods for reporting
type ReportUseCase interface {
	GenerateDailyReport(ctx context.Context, date time.Time) (*DailyReport, error)
	SendDailyReport(ctx context.Context, date time.Time) error
	StartDailyReporting(ctx context.Context) error
	StopDailyReporting() error
}
