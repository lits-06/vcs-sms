package domain

import "time"

type ServerStatus string

const (
	StatusOnline  ServerStatus = "ON"
	StatusOffline ServerStatus = "OFF"
)

type ServerState struct {
	ID        string       `json:"id"`
	ServerID  string       `json:"server_id"`
	Status    ServerStatus `json:"status"` // "online" or "offline"
	Timestamp time.Time    `json:"timestamp"`
}
type Server struct {
	ID     string       `json:"server_id"`
	Port   int          `json:"port"`
	Status ServerStatus `json:"status"`
}

type HealthCheckResult struct {
	ServerID     string        `json:"server_id"`
	Status       string        `json:"status"` // "online", "offline"
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
	CheckedAt    time.Time     `json:"checked_at"`
}
