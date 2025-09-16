package config

import (
	"flag"
	"time"

	"github.com/lits-06/vcs-sms/pkg/kafka"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/probes"
	"github.com/lits-06/vcs-sms/pkg/tracing"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Healthcheck microservice config path")
}

type Config struct {
	ServiceName string         `mapstructure:"serviceName"`
	Logger      *logger.Config `mapstructure:"logger"`

	Kafka  *kafka.Config   `mapstructure:"kafka"`
	Probes probes.Config   `mapstructure:"probes"`
	Jaeger *tracing.Config `mapstructure:"jaeger"`

	Elasticsearch Elasticsearch `mapstructure:"elasticsearch"`
	Redis         Redis         `mapstructure:"redis"`
	Duration      time.Duration `mapstructure:"duration"`
}

type Elasticsearch struct {
	Address       string `mapstructure:"address"`
	SnapshotIndex string `mapstructure:"snapshotIndex"`
	RecordIndex   string `mapstructure:"recordIndex"`
}

type Redis struct {
	Address     string `mapstructure:"address"`
	Password    string `mapstructure:"password"`
	DB          int    `mapstructure:"db"`
	SnapshotKey string `mapstructure:"snapshotKey"`
}
