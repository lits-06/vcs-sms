package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/lits-06/vcs-sms/healthcheck_service/config"
	"github.com/lits-06/vcs-sms/healthcheck_service/internal/domain"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/utils"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

type Producer struct {
	log    logger.Logger
	cfg    *config.Config
	writer *kafka.Writer
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
		BatchSize:    10000,
		BatchTimeout: 500 * time.Millisecond,
	}

	return w
}

func (p *Producer) Run() {
	p.writer = p.getNewKafkaWriter(updateTopic)
}

func (p *Producer) Close() {
	p.writer.Close()
}

func (p *Producer) PublishStateChange(ctx context.Context, server []domain.Server) error {
	var msgs []kafka.Message
	for _, server := range server {
		data, err := json.Marshal(server)
		if err != nil {
			p.log.Errorf("Failed to marshal server data: %v", err)
			continue
		}

		msg := kafka.Message{
			Key:   []byte(server.ServerID),
			Value: data,
		}
		msgs = append(msgs, msg)
	}

	start := time.Now()
	err := p.writer.WriteMessages(ctx, msgs...)
	elapsed := time.Since(start)

	if err != nil {
		p.log.Errorf("Failed to state batch: %v", err)
	} else {
		var totalBytes int32 = 0
		for _, msg := range msgs {
			totalBytes += utils.TotalSize(&msg)
		}

		p.log.Infof("Produced state batch: %d messages, %d bytes, took %s", len(msgs), totalBytes, elapsed)
	}

	return nil
}
