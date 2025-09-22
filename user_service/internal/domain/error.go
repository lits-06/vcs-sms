package domain

import "errors"

var (
	ErrUserExists      = errors.New("user with this email already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidScope    = errors.New("invalid scope")
	ErrNoValidScopesFound = errors.New("no valid scopes found")
)
