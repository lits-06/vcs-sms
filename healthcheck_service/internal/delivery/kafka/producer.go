package kafka

import (
	"context"
	"encoding/json"
	"sync"
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

	batch []kafka.Message
	mu    sync.Mutex

	flushInterval time.Duration
	maxBatchSize  int
	done          chan struct{}
}

func NewProducer(log logger.Logger, cfg *config.Config) *Producer {
	return &Producer{
		log:           log,
		cfg:           cfg,
		batch:         make([]kafka.Message, 0, 100),
		flushInterval: 5 * time.Second,
		maxBatchSize:  100,
		done:          make(chan struct{}),
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
	p.writer = p.getNewKafkaWriter(updateTopic)

	go p.batchLoop()
}

func (p *Producer) Close() {
	close(p.done)
	p.writer.Close()
}

func (p *Producer) PublishStateChange(ctx context.Context, server *domain.Server) error {
	data, err := json.Marshal(server)
	if err != nil {
		p.log.Errorf("json.Marshal: %v", err)
		return err
	}

	msg := kafka.Message{
		Key:   []byte(server.ServerID),
		Value: data,
	}

	p.mu.Lock()
	p.batch = append(p.batch, msg)
	flush := len(p.batch) >= p.maxBatchSize
	p.mu.Unlock()

	if flush {
		p.flush(ctx)
	}

	return nil
}

func (p *Producer) batchLoop() {
	ticker := time.NewTicker(p.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.flush(context.Background())
		case <-p.done:
			p.flush(context.Background())
			return
		}
	}
}

func (p *Producer) flush(ctx context.Context) error {
	p.mu.Lock()
	batch := p.batch
	if len(batch) == 0 {
		p.mu.Unlock()
		return nil
	}
	p.batch = make([]kafka.Message, 0, 100)
	p.mu.Unlock()

	start := time.Now()
	err := p.writer.WriteMessages(ctx, batch...)
	elapsed := time.Since(start)

	if err != nil {
		p.log.Errorf("Failed to state batch: %v", err)
	} else {
		var totalBytes int32 = 0
		for _, msg := range batch {
			totalBytes += utils.TotalSize(&msg)
		}

		p.log.Infof("Produced state batch: %d messages, %d bytes, took %s", len(batch), totalBytes, elapsed)
	}

	return err
}
