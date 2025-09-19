package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lits-06/vcs-sms/pkg/constants"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/redis"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Auth microservice config path")
}

type Config struct {
	ServiceName string          `mapstructure:"serviceName"`
	Port        string          `mapstructure:"port"`
	Logger      *logger.Config  `mapstructure:"logger"`
	Redis       Redis           `mapstructure:"redis"`
	Grpc        Grpc            `mapstructure:"grpc"`
	Jaeger      *tracing.Config `mapstructure:"jaeger"`
	JWT         JWT             `mapstructure:"jwt"`
}

type Grpc struct {
	UserServicePort string `mapstructure:"userServicePort"`
}

type Redis struct {
	*redis.Config
	RefreshKey   string `mapstructure:"refreshKey"`
	BlacklistKey string `mapstructure:"blacklistKey"`
}

type JWT struct {
	AccessSecretKey  string        `mapstructure:"accessSecretKey"`
	RefreshSecretKey string        `mapstructure:"refreshSecretKey"`
	AccessTTL        time.Duration `mapstructure:"accessTTL"`
	RefreshTTL       time.Duration `mapstructure:"refreshTTL"`
}

func InitConfig() (*Config, error) {
	if configPath == "" {
		configPathFromEnv := os.Getenv(constants.ConfigPath)
		if configPathFromEnv != "" {
			configPath = configPathFromEnv
		} else {
			getwd, err := os.Getwd()
			if err != nil {
				return nil, errors.Wrap(err, "os.Getwd")
			}
			configPath = fmt.Sprintf("%s/api_gateway_service/config/config.yaml", getwd)
		}
	}

	cfg := &Config{}

	viper.SetConfigType(constants.Yaml)
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viper.ReadInConfig: %w", err)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("viper.Unmarshal: %w", err)
	}

	port := os.Getenv(constants.HttpPort)
	if port != "" {
		cfg.Port = port
	}

	jaegerAddr := os.Getenv(constants.JaegerHostPort)
	if jaegerAddr != "" {
		cfg.Jaeger.HostPort = jaegerAddr
	}

	return cfg, nil
}
