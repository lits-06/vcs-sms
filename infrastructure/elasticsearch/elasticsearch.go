package elasticsearch

// import (
// 	"fmt"

// 	"github.com/elastic/go-elasticsearch/v8"
// 	"github.com/lits-06/vcs-sms/internal/config"
// )

// // Client wraps elasticsearch client
// type Client struct {
// 	*elasticsearch.Client
// 	Index string
// }

// // NewElasticsearchClient creates a new Elasticsearch client
// func NewElasticsearchClient(cfg config.ElasticsearchConfig) (*Client, error) {
// 	esConfig := elasticsearch.Config{
// 		Addresses: []string{
// 			fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
// 		},
// 	}

// 	if cfg.Username != "" && cfg.Password != "" {
// 		esConfig.Username = cfg.Username
// 		esConfig.Password = cfg.Password
// 	}

// 	client, err := elasticsearch.NewClient(esConfig)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
// 	}

// 	// Test connection
// 	_, err = client.Info()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
// 	}

// 	return &Client{
// 		Client: client,
// 		Index:  cfg.Index,
// 	}, nil
// }
