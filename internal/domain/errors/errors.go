package errors

import "errors"

var (
	// Server errors
	ErrServerNotFound    = errors.New("server not found")
	ErrServerExists      = errors.New("server already exists")
	ErrInvalidServerData = errors.New("invalid server data")
	ErrServerUnreachable = errors.New("server unreachable")
	ErrServerNameExists  = errors.New("server name already exists")
	ErrServerIDExists    = errors.New("server ID already exists")

	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")

	// General errors
	ErrInvalidInput       = errors.New("invalid input")
	ErrInternalServer     = errors.New("internal server error")
	ErrDatabaseConnection = errors.New("database connection error")
	ErrCacheConnection    = errors.New("cache connection error")
)
