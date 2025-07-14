package entities

import "time"

type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username" validate:"required,min=3,max=50"`
	Email     string    `json:"email" db:"email" validate:"required,email"`
	Password  string    `json:"-" db:"password" validate:"required,min=6"`
	Roles     []string  `json:"roles" db:"roles" validate:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
