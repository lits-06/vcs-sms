package kafka

// Config kafka config
type Config struct {
	Brokers []string `mapstructure:"brokers"`
	GroupID string   `mapstructure:"groupID"`
}

// TopicConfig kafka topic config
type TopicConfig struct {
	TopicName         string `mapstructure:"topicName"`
	Partitions        int    `mapstructure:"partitions"`
	ReplicationFactor int    `mapstructure:"replicationFactor"`
}
