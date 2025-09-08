package domain

import "time"

// ServerStateRecord represents a server state change event
type ServerStateRecord struct {
	ID          string    `json:"id"`
	ServerID    string    `json:"server_id"`
	Status      string    `json:"status"` // "online" or "offline"
	Timestamp   time.Time `json:"timestamp"`
	Duration    int64     `json:"duration,omitempty"` // Duration in seconds (for offline events)
	ProcessedAt time.Time `json:"processed_at"`
}

// UptimeRecord represents uptime calculation for a specific time period
type UptimeRecord struct {
	ServerID     string    `json:"server_id"`
	Date         string    `json:"date"` // Format: YYYY-MM-DD
	UptimeSecond int64     `json:"uptime_seconds"`
	TotalSeconds int64     `json:"total_seconds"`
	Percentage   float64   `json:"percentage"`
	LastUpdated  time.Time `json:"last_updated"`
}

// ServerSnapshot represents current server state
type ServerSnapshot struct {
	ServerID      string    `json:"server_id"`
	CurrentStatus string    `json:"current_status"`
	LastOnline    time.Time `json:"last_online"`
	LastOffline   time.Time `json:"last_offline,omitempty"`
	TotalUptime   int64     `json:"total_uptime_seconds"`
	LastUpdated   time.Time `json:"last_updated"`
}