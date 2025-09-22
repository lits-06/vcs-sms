package postgres

import (
	"fmt"
	"time"

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

func NewPostgresDB(cfg *Config) (*gorm.DB, error) {
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
		return nil, fmt.Errorf("gorm.Open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("db.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxConn)
	sqlDB.SetMaxIdleConns(minConns)
	sqlDB.SetConnMaxLifetime(maxConnLifetime)
	sqlDB.SetConnMaxIdleTime(maxConnIdleTime)

	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("sqlDB.Ping: %w", err)
	}

	return db, nil
}