package kafka

// Config kafka config
type Config struct {
	Brokers []string `mapstructure:"brokers"`
}
