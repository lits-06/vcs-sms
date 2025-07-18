package entity

import (
	"time"
)

// Server represents a server entity
type Server struct {
	ID        string       `json:"id" db:"id" gorm:"primaryKey;column:id" validate:"required"`
	Name      string       `json:"name" db:"name" gorm:"column:name;uniqueIndex" validate:"required"`
	Host      string       `json:"host" gorm:"-" validate:"omitempty"`
	Port      int          `json:"port" gorm:"-" validate:"omitempty,min=1024,max=65535"`
	Status    ServerStatus `json:"status" db:"status" gorm:"column:status" validate:"omitempty,oneof=ON OFF"`
	CreatedAt time.Time    `json:"created_at" db:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at" gorm:"column:updated_at,autoUpdateTime"`
	IPv4      string       `json:"ipv4" db:"ipv4" gorm:"column:ipv4" validate:"omitempty,ipv4"`
}

func (Server) TableName() string {
	return "servers"
}

// ServerStatus represents server status constants
type ServerStatus string

const (
	StatusOnline  ServerStatus = "ON"
	StatusOffline ServerStatus = "OFF"
)
