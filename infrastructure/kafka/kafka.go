package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/usecases/server"
	"github.com/segmentio/kafka-go"
)

type Client struct {
	writer *kafka.Writer
	reader *kafka.Reader
}

type ServerStatusMessage struct {
	ServerID  string    `json:"server_id"`
	Status    string    `json:"status"`
	Port      int       `json:"port"`
	Host      string    `json:"host"`
	Timestamp time.Time `json:"timestamp"`
}

func NewKafkaClient(cfg *config.Config) (*Client, error) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(fmt.Sprintf("localhost:%d", 9092)), // From config
		Topic:    "server-status",
		Balancer: &kafka.LeastBytes{},
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{fmt.Sprintf("localhost:%d", 9092)},
		Topic:   "server-status",
		GroupID: "server-monitor-group",
	})

	return &Client{
		writer: writer,
		reader: reader,
	}, nil
}

// PublishServerStatus publishes server status to Kafka
func (c *Client) PublishServerStatus(ctx context.Context, msg server.StatusRecord) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal server status message: %w", err)
	}

	err = c.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(msg.ServerID),
		Value: data,
		Time:  time.Now(),
	})

	if err != nil {
		return fmt.Errorf("failed to publish server status: %w", err)
	}

	return nil
}

// ConsumeServerStatus consumes server status messages
func (c *Client) ConsumeServerStatus(ctx context.Context, handler func(*server.StatusRecord) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				return fmt.Errorf("failed to read message: %w", err)
			}

			var statusMsg server.StatusRecord
			if err := json.Unmarshal(msg.Value, &statusMsg); err != nil {
				continue // Skip invalid messages
			}

			if err := handler(&statusMsg); err != nil {
				// Log error but continue processing
				continue
			}
		}
	}
}

// Close closes the Kafka client
func (c *Client) Close() error {
	if err := c.writer.Close(); err != nil {
		return err
	}
	return c.reader.Close()
}
