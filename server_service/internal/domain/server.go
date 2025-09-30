package domain

import "time"

type Server struct {
	ID        string    `json:"server_id" db:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey" example:"550e8400-e29b-41d4-a716-446655440000"` // Server ID
	Name      string    `json:"name" db:"name" gorm:"column:name;uniqueIndex" validate:"required" example:"web-server-01"`                              // Server name
	Port      int       `json:"port" gorm:"port" validate:"omitempty,min=1024,max=65535" example:"80"`                                                  // Port number
	Status    string    `json:"status" db:"status" gorm:"column:status" validate:"oneof=ON OFF" example:"ON" enums:"ON,OFF"`                            // Server status
	CreatedAt time.Time `json:"created_at" db:"created_at" gorm:"column:created_at" example:"2023-01-01T00:00:00Z"`                                     // Creation timestamp
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" gorm:"column:updated_at;autoUpdateTime" example:"2023-01-01T00:00:00Z"`                      // Last update timestamp
	IPv4      string    `json:"ipv4" db:"ipv4" gorm:"column:ipv4" validate:"omitempty,ipv4" example:"192.168.1.100"`                                    // IPv4 address
}

func (Server) TableName() string {
	return "servers"
}

const (
	StatusOnline  string = "ON"
	StatusOffline string = "OFF"
)

func IsStatusValid(status string) bool {
	switch status {
	case StatusOnline, StatusOffline:
		return true
	}
	return false
}
