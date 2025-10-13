package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/pkg/utils"
	"github.com/lits-06/vcs-sms/server_service/config"
	"github.com/lits-06/vcs-sms/server_service/internal/domain"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

type Producer struct {
	log          logger.Logger
	cfg          *config.Config
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
		BatchSize:    10000,
		BatchTimeout: 500 * time.Millisecond,
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

func (p *Producer) PublishServerCreate(ctx context.Context, servers []domain.Server) error {
	var msgs []kafka.Message
	for _, srv := range servers {
		value, err := json.Marshal(srv)
		if err != nil {
			p.log.Errorf("json.Marshal: %v", err)
			continue
		}
		msgs = append(msgs, kafka.Message{
			Key:   []byte(srv.ID),
			Value: value,
		})
	}

	start := time.Now()
	err := p.createWriter.WriteMessages(ctx, msgs...)
	elapsed := time.Since(start)
	if err != nil {
		p.log.Errorf("Failed to write create batch: %v", err)
	} else {
		var totalBytes int32 = 0
		for _, msg := range msgs {
			totalBytes += utils.TotalSize(&msg)
		}
		p.log.Infof("Produced create batch: %d messages, %d bytes, took %s", len(msgs), totalBytes, elapsed)
	}

	return nil
}

func (p *Producer) PublishServerDelete(ctx context.Context, serverID []string) error {
	var msgs []kafka.Message
	for _, id := range serverID {
		msgs = append(msgs, kafka.Message{
			Key:   []byte(id),
			Value: []byte(id),
		})
	}

	start := time.Now()
	err := p.deleteWriter.WriteMessages(ctx, msgs...)
	elapsed := time.Since(start)
	if err != nil {
		p.log.Errorf("Failed to write delete batch: %v", err)
	} else {
		var totalBytes int32 = 0
		for _, msg := range msgs {
			totalBytes += utils.TotalSize(&msg)
		}
		p.log.Infof("Produced create batch: %d messages, %d bytes, took %s", len(msgs), totalBytes, elapsed)
	}

	return nil
}
