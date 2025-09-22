package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/server_service/config"
	"github.com/lits-06/vcs-sms/server_service/internal/domain"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

type Producer struct {
	log    logger.Logger
	cfg    *config.Config
	createWriter *kafka.Writer
	deleteWriter *kafka.Writer
}

func NewProducer(log logger.Logger, cfg *config.Config) *Producer {
	return &Producer{
		log: log,
		cfg: cfg,
	}
}

func (p *Producer) getNewKafkaWriter(topic string) *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(p.cfg.Kafka.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: writerRequiredAcks,
		WriteTimeout: writerWriteTimeout,
		ReadTimeout:  writerReadTimeout,
		MaxAttempts:  writerMaxAttempts,
		Logger:       kafka.LoggerFunc(p.log.Debugf),
		ErrorLogger:  kafka.LoggerFunc(p.log.Errorf),
		Compression:  compress.Snappy,
		Async:        true,
		BatchSize:    100,
		BatchTimeout: 100 * time.Millisecond,
	}

	return w
}

func (p *Producer) Run() {
	p.createWriter = p.getNewKafkaWriter(createTopic)
	p.deleteWriter = p.getNewKafkaWriter(deleteTopic)
}

func (p *Producer) Close() {
	p.createWriter.Close()
	p.deleteWriter.Close()	
}

func (p *Producer) PublishServerCreate(ctx context.Context, server *domain.Server) error {
	data, err := json.Marshal(server)
	if err != nil {
		p.log.Error("json.Marshal: %v", err)
		return err
	}

	msg := kafka.Message{
		Key:   []byte(server.ID),
		Value: data,
	}

	return p.createWriter.WriteMessages(ctx, msg)
}

func (p *Producer) PublishServerDelete(ctx context.Context, serverID string) error {
	msg := kafka.Message{
		Key:   []byte(serverID),
		Value: []byte(serverID),
	}

	return p.deleteWriter.WriteMessages(ctx, msg)
}
