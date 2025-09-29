package dto

import "time"

type UptimeRequest struct {
	Email     string    `json:"email" validate:"required,email" example:"admin@example.com"`
	StartDate time.Time `json:"start_date" validate:"required" example:"2025-01-01T00:00:00Z"`
	EndDate   time.Time `json:"end_date" validate:"required" example:"2025-01-31T23:59:59Z"`
}