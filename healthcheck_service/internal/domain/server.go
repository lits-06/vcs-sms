package domain

import "time"

type Server struct {
	ID     string       `json:"server_id" db:"id" gorm:"primaryKey;column:id"`
	Name   string       `json:"name" db:"name" gorm:"column:name;uniqueIndex" validate:"required"`
	Port   int          `json:"-" gorm:"-" validate:"omitempty,min=1024,max=65535"`
	Status ServerStatus `json:"status" db:"status" gorm:"column:status" validate:"omitempty,oneof=ON OFF"`
	IPv4   string       `json:"ipv4" db:"ipv4" gorm:"column:ipv4" validate:"omitempty,ipv4"`
}

type ServerStatus string

const (
	StatusOnline  ServerStatus = "ON"
	StatusOffline ServerStatus = "OFF"
)

type ServerState struct {
	ID        string    `json:"id"`
	ServerID  string    `json:"server_id"`
	Status    string    `json:"status"` // "online" or "offline"
	Timestamp time.Time `json:"timestamp"`
}

// ServerSnapshot represents current server state
type ServerSnapshot struct {
	ServerID      string `json:"server_id"`
	CurrentStatus string `json:"current_status"`
}

type HealthCheckResult struct {
	ServerID     string        `json:"server_id"`
	Status       string        `json:"status"` // "online", "offline"
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
	CheckedAt    time.Time     `json:"checked_at"`
}
