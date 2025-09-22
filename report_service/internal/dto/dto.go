package dto

import "time"

type UptimeRequest struct {
	Email     string    `json:"email" validate:"required,email"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}