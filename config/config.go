package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	SMTP          SMTPConfig          `mapstructure:"smtp"`
	Logging       LoggingConfig       `mapstructure:"log"`
	Monitoring    MonitoringConfig    `mapstructure:"monitoring"`
	App           AppConfig           `mapstructure:"app"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host" validate:"required"`
	Port            int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	Name            string        `mapstructure:"name" validate:"required"`
	User            string        `mapstructure:"user" validate:"required"`
	Password        string        `mapstructure:"password" validate:"required"`
	SSLMode         string        `mapstructure:"ssl_mode" validate:"required,oneof=disable require verify-ca verify-full"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"min=1"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"min=1"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host" validate:"required"`
	Port         int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db" validate:"min=0,max=15"`
	PoolSize     int    `mapstructure:"pool_size" validate:"min=1"`
	MinIdleConns int    `mapstructure:"min_idle_conns" validate:"min=0"`
}

type ElasticsearchConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Index    string `mapstructure:"index" validate:"required"`
}

type JWTConfig struct {
	Secret string        `mapstructure:"secret" validate:"required,min=32"`
	Expiry time.Duration `mapstructure:"expiry" validate:"required"`
}

type SMTPConfig struct {
	Host       string `mapstructure:"host" validate:"required"`
	Port       int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Username   string `mapstructure:"username" validate:"required"`
	Password   string `mapstructure:"password" validate:"required"`
	From       string `mapstructure:"from" validate:"required,email"`
	AdminEmail string `mapstructure:"admin_email" validate:"required,email"`
}

type LoggingConfig struct {
	Level       string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	File        string `mapstructure:"file" validate:"required"`
	MaxSize     int    `mapstructure:"max_size" validate:"min=1"`
	MaxBackups  int    `mapstructure:"max_backups" validate:"min=0"`
	MaxAge      int    `mapstructure:"max_age" validate:"min=1"`
	Compression bool   `mapstructure:"compression"`
}

type MonitoringConfig struct {
	Interval time.Duration `mapstructure:"interval" validate:"required"`
}

type AppConfig struct {
	Environment string `mapstructure:"environment" validate:"required,oneof=development staging production"`
	Name        string `mapstructure:"name" validate:"required"`
	Version     string `mapstructure:"version" validate:"required"`
}

var cfg *Config

// Load loads configuration from various sources
func Load() (*Config, error) {
	// Set up Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../config") // For when running from cmd directory
	viper.AddConfigPath("../")       // For config.yaml in root directory

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		log.Println("No config file found, using environment variables and defaults")
	}

	// Unmarshal config
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	cfg = config
	return config, nil
}

// Get returns the global config instance
func Get() *Config {
	if cfg == nil {
		panic("config not loaded. Call Load() first")
	}
	return cfg
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetElasticsearchURL returns the Elasticsearch URL
func (c *ElasticsearchConfig) GetElasticsearchURL() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}

// GetSMTPAddr returns the SMTP address
func (c *SMTPConfig) GetSMTPAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsProduction returns true if running in production environment
func (c *AppConfig) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *AppConfig) IsDevelopment() bool {
	return c.Environment == "development"
}
