package entities

import (
	"time"
)

// Server represents a server entity
type Server struct {
	ID          string       `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Status      ServerStatus `json:"status" db:"status"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	LastChecked time.Time    `json:"last_checked" db:"last_checked"`
	IPv4        string       `json:"ipv4" db:"ipv4" validate:"ipv4"`
}

// ServerStatus represents server status constants
type ServerStatus string

const (
	StatusOnline  ServerStatus = "online"
	StatusOffline ServerStatus = "offline"
	StatusUnknown ServerStatus = "unknown"
)

// ServerFilter represents filtering criteria for servers
type ServerFilter struct {
	Name   string       `json:"name,omitempty"`
	Status ServerStatus `json:"status,omitempty"`
	IPv4   string       `json:"ipv4,omitempty"`
}

// SortOrder represents sorting direction
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// ServerSort represents sorting criteria
type ServerSort struct {
	Field string    `json:"field,omitempty"` // name, status, created_at, updated_at, last_checked
	Order SortOrder `json:"order,omitempty"` // asc, desc
}

// Pagination represents pagination parameters
type ServerPagination struct {
	From  int `json:"from,omitempty"`  // offset
	To    int `json:"to,omitempty"`    // limit (or you can use Size instead)
	Size  int `json:"size,omitempty"`  // page size
	Page  int `json:"page,omitempty"`  // page number (alternative to from/to)
	Total int `json:"total,omitempty"` // total count (for response)
}
