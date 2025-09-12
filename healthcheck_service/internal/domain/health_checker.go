package domain

import "context"

type HealthChecker interface {
	CheckServer(ctx context.Context, server *Server) (*HealthCheckResult, error)
}
