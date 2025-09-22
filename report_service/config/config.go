package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lits-06/vcs-sms/pkg/constants"
	"github.com/lits-06/vcs-sms/pkg/jwt"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Report microservice config path")
}

type Config struct {
	ServiceName  string              `mapstructure:"serviceName"`
	Port             string              `mapstructure:"port"`
	Duration         time.Duration       `mapstructure:"duration"`
	Logger           *logger.Config      `mapstructure:"logger"`
	Elasticsearch	Elasticsearch       `mapstructure:"elasticsearch"`
	Jaeger           *tracing.Config     `mapstructure:"jaeger"`
	Smtp			Smtp `mapstructure:"smtp"`
	JWT             *jwt.Config      `mapstructure:"jwt"`
}

type Elasticsearch struct {
	Address       string `mapstructure:"address"`
	SnapshotIndex string `mapstructure:"snapshotIndex"`
	RecordIndex   string `mapstructure:"recordIndex"`
}

type Smtp struct {
	Host       string   `mapstructure:"host" validate:"required"`
	Port       int      `mapstructure:"port" validate:"required,min=1,max=65535"`
	Username   string   `mapstructure:"username" validate:"required"`
	Password   string   `mapstructure:"password" validate:"required"`
	From       string   `mapstructure:"from" validate:"required,email"`
	To         []string `mapstructure:"to" validate:"required,dive,email"`
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
			configPath = fmt.Sprintf("%s/report_service/config/config.yaml", getwd)
		}
	}

	cfg := &Config{}

	viper.SetConfigType(constants.Yaml)
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "viper.ReadInConfig")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, errors.Wrap(err, "viper.Unmarshal")
	}
	return cfg, nil
}