package postgres

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	DBName   string `yaml:"dbName"`
	SSLMode  bool   `yaml:"sslMode"`
	Password string `yaml:"password"`
}

const (
	maxConn           = 50
	minConns          = 10
	maxConnIdleTime   = 1 * time.Minute
	maxConnLifetime   = 3 * time.Minute
)

func NewPostgresConn(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%t",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.Port,
		cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default})
	if err != nil {
		return nil, errors.Wrap(err, "gorm.Open")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "db.DB")
	}

	sqlDB.SetMaxOpenConns(maxConn)
	sqlDB.SetMaxIdleConns(minConns)
	sqlDB.SetConnMaxLifetime(maxConnLifetime)
	sqlDB.SetConnMaxIdleTime(maxConnIdleTime)

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	// if err := sqlDB.PingContext(ctx); err != nil {
	// 	return nil, fmt.Errorf("failed to ping database: %w", err)
	// }

	// if err := db.AutoMigrate(&entity.Server{}); err != nil {
	// 	return nil, fmt.Errorf("failed to auto migrate: %w", err)
	// }

	return db, nil
}