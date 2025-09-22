package domain

import "errors"

var (
	ErrServerExists       = errors.New("server already exists")
	ErrNameExists         = errors.New("server name already exists")
	ErrServerNotFound     = errors.New("server not found")
)
