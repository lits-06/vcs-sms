package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
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
	Logging       LoggingConfig       `mapstructure:"logging"`
	Monitoring    MonitoringConfig    `mapstructure:"monitoring"`
	App           AppConfig           `mapstructure:"app"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"server.port" validate:"required,min=1,max=65535"`
	ReadTimeout  time.Duration `mapstructure:"server.read_timeout"`
	WriteTimeout time.Duration `mapstructure:"server.write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"server.idle_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"postgre.host" validate:"required"`
	Port            int           `mapstructure:"postgre.port" validate:"required,min=1,max=65535"`
	Name            string        `mapstructure:"postgre.name" validate:"required"`
	User            string        `mapstructure:"postgre.user" validate:"required"`
	Password        string        `mapstructure:"postgre.password" validate:"required"`
	SSLMode         string        `mapstructure:"postgre.ssl_mode" validate:"required,oneof=disable require verify-ca verify-full"`
	MaxOpenConns    int           `mapstructure:"postgre.max_open_conns" validate:"min=1"`
	MaxIdleConns    int           `mapstructure:"postgre.max_idle_conns" validate:"min=1"`
	ConnMaxLifetime time.Duration `mapstructure:"postgre.conn_max_lifetime"`
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
	Level       string `mapstructure:"log.level" validate:"required,oneof=debug info warn error"`
	File        string `mapstructure:"log.file" validate:"required"`
	MaxSize     int    `mapstructure:"log.max_size" validate:"min=1"`
	MaxBackups  int    `mapstructure:"log.max_backups" validate:"min=0"`
	MaxAge      int    `mapstructure:"log.max_age" validate:"min=1"`
	Compression bool   `mapstructure:"log.compression"`
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
	// Load .env file if it exists
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using environment variables and config files")
	}

	// Set up Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	// viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")
	// viper.AddConfigPath("./config")

	// Enable automatic env vars
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// // Set defaults
	// setDefaults()

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

	// Validate configuration
	// if err := validateConfig(config); err != nil {
	// 	return nil, fmt.Errorf("config validation failed: %w", err)
	// }

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

// // setDefaults sets default values for configuration
// func setDefaults() {
// 	// Server defaults
// 	viper.SetDefault("server.port", 8080)
// 	viper.SetDefault("server.read_timeout", "30s")
// 	viper.SetDefault("server.write_timeout", "30s")
// 	viper.SetDefault("server.idle_timeout", "120s")

// 	// Database defaults
// 	viper.SetDefault("database.host", "localhost")
// 	viper.SetDefault("database.port", 5432)
// 	viper.SetDefault("database.ssl_mode", "disable")
// 	viper.SetDefault("database.max_open_conns", 25)
// 	viper.SetDefault("database.max_idle_conns", 25)
// 	viper.SetDefault("database.conn_max_lifetime", "5m")

// 	// Redis defaults
// 	viper.SetDefault("redis.host", "localhost")
// 	viper.SetDefault("redis.port", 6379)
// 	viper.SetDefault("redis.db", 0)
// 	viper.SetDefault("redis.pool_size", 100)
// 	viper.SetDefault("redis.min_idle_conns", 10)

// 	// Elasticsearch defaults
// 	viper.SetDefault("elasticsearch.host", "localhost")
// 	viper.SetDefault("elasticsearch.port", 9200)
// 	viper.SetDefault("elasticsearch.index", "server-logs")

// 	// JWT defaults
// 	viper.SetDefault("jwt.expiry", "24h")

// 	// SMTP defaults
// 	viper.SetDefault("smtp.port", 587)

// 	// Logging defaults
// 	viper.SetDefault("logging.level", "info")
// 	viper.SetDefault("logging.file", "app.log")
// 	viper.SetDefault("logging.max_size", 1)
// 	viper.SetDefault("logging.max_backups", 10)
// 	viper.SetDefault("logging.max_age", 30)
// 	viper.SetDefault("logging.compression", true)

// 	// Monitoring defaults
// 	viper.SetDefault("monitoring.interval", "60s")

// 	// App defaults
// 	viper.SetDefault("app.environment", "development")
// 	viper.SetDefault("app.name", "Server Management System")
// 	viper.SetDefault("app.version", "1.0.0")
// }

// validateConfig validates the loaded configuration
// func validateConfig(config *Config) error {
// 	// Add custom validation logic here
// 	if config.Database.Host == "" {
// 		return fmt.Errorf("database host is required")
// 	}

// 	if config.JWT.Secret == "" {
// 		return fmt.Errorf("JWT secret is required")
// 	}

// 	if len(config.JWT.Secret) < 32 {
// 		return fmt.Errorf("JWT secret must be at least 32 characters long")
// 	}

// 	return nil
// }

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
