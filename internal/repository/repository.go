package repository

import (
	"database/sql"

	"github.com/lits-06/vcs-sms/internal/domain"
	"github.com/lits-06/vcs-sms/internal/infrastructure/elasticsearch"
	"github.com/lits-06/vcs-sms/internal/infrastructure/redis"
	"github.com/lits-06/vcs-sms/internal/repository/cache"
	"github.com/lits-06/vcs-sms/internal/repository/postgres"
)

// Repositories holds all repository implementations
type Repositories struct {
	Server domain.ServerRepository
	Uptime domain.UptimeRepository
	User   domain.UserRepository
	Cache  domain.CacheRepository
}

// NewRepositories creates a new repositories instance
func NewRepositories(db *sql.DB, redisClient *redis.Client, esClient *elasticsearch.Client) *Repositories {
	return &Repositories{
		Server: postgres.NewServerRepository(db),
		Uptime: postgres.NewUptimeRepository(db, esClient),
		User:   postgres.NewUserRepository(db),
		Cache:  cache.NewCacheRepository(redisClient),
	}
}
