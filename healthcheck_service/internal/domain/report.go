package domain

import "time"

type ServerStateRecord struct {
	ID        string    `json:"id"`
	ServerID  string    `json:"server_id"`
	Status    string    `json:"status"` // "online" or "offline"
	Timestamp time.Time `json:"timestamp"`
}

// ServerSnapshot represents current server state
type ServerSnapshot struct {
	ServerID      string    `json:"server_id"`
	CurrentStatus string    `json:"current_status"`
	LastOnline    time.Time `json:"last_online"`
	LastOffline   time.Time `json:"last_offline,omitempty"`
	LastUpdated   time.Time `json:"last_updated"`
}
