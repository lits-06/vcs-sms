package domain

import "time"

type ServerStatus string

const (
	StatusOnline  ServerStatus = "ON"
	StatusOffline ServerStatus = "OFF"
)

type Server struct {
	ServerID  string       `json:"server_id"`
	Port      int          `json:"port"`
	Status    ServerStatus `json:"status"` // "online" or "offline"
	Timestamp time.Time    `json:"timestamp"`
}
