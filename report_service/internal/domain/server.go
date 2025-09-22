package domain

import "time"

// ServerStateRecord represents a server state change event
type Server struct {
	ServerID    string    `json:"server_id"`
	Port       	int       `json:"port"`
	Status      string    `json:"status"` // "online" or "offline"
	Timestamp   time.Time `json:"timestamp"`
}