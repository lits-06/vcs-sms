package entity

import (
	"time"
)

// Server represents a server entity
type Server struct {
	ID          string       `json:"id" db:"id" validate:"required"`
	Name        string       `json:"name" db:"name" validate:"required"`
	Status      ServerStatus `json:"status" db:"status" validate:"omitempty,oneof=ON OFF UNKNOWN"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	LastChecked time.Time    `json:"last_checked" db:"last_checked"`
	IPv4        string       `json:"ipv4" db:"ipv4" validate:"omitempty,ipv4"`
}

// ServerStatus represents server status constants
type ServerStatus string

const (
	StatusOnline  ServerStatus = "ON"
	StatusOffline ServerStatus = "OFF"
	StatusUnknown ServerStatus = "UNKNOWN"
)
