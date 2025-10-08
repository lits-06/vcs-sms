package kafka

import (
	"context"
	"encoding/json"
	"sync"
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

	createBatch []kafka.Message
	deleteBatch []kafka.Message
	createMutex sync.Mutex
	deleteMutex sync.Mutex

	flushInterval time.Duration
	maxBatchSize  int
	done          chan struct{}
}

func NewProducer(log logger.Logger, cfg *config.Config) *Producer {
	return &Producer{
		log:           log,
		cfg:           cfg,
		createBatch:   make([]kafka.Message, 0, 100),
		deleteBatch:   make([]kafka.Message, 0, 100),
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
	p.createWriter = p.getNewKafkaWriter(createTopic)
	p.deleteWriter = p.getNewKafkaWriter(deleteTopic)

	go p.batchLoop()
}

func (p *Producer) Close() {
	close(p.done)
	p.createWriter.Close()
	p.deleteWriter.Close()
}

func (p *Producer) PublishServerCreate(ctx context.Context, server *domain.Server) error {
	data, err := json.Marshal(server)
	if err != nil {
		p.log.Errorf("json.Marshal: %v", err)
		return err
	}

	msg := kafka.Message{
		Key:   []byte(server.ID),
		Value: data,
	}

	p.createMutex.Lock()
	p.createBatch = append(p.createBatch, msg)
	flush := len(p.createBatch) >= p.maxBatchSize
	p.createMutex.Unlock()

	if flush {
		go p.flushCreate(ctx)
	}

	return nil
}

func (p *Producer) PublishServerDelete(ctx context.Context, serverID string) error {
	msg := kafka.Message{
		Key:   []byte(serverID),
		Value: []byte(serverID),
	}

	p.deleteMutex.Lock()
	p.deleteBatch = append(p.deleteBatch, msg)
	flush := len(p.deleteBatch) >= p.maxBatchSize
	p.deleteMutex.Unlock()

	if flush {
		go p.flushDelete(ctx)
	}

	return nil
}

func (p *Producer) batchLoop() {
	ticker := time.NewTicker(p.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.flushCreate(context.Background())
			p.flushDelete(context.Background())
		case <-p.done:
			p.flushCreate(context.Background())
			p.flushDelete(context.Background())
			return
		}
	}
}

func (p *Producer) flushCreate(ctx context.Context) error {
	p.createMutex.Lock()
	batch := p.createBatch
	if len(batch) == 0 {
		p.createMutex.Unlock()
		return nil
	}
	p.createBatch = make([]kafka.Message, 0, p.maxBatchSize)
	p.createMutex.Unlock()

	start := time.Now()
	err := p.createWriter.WriteMessages(ctx, batch...)
	elapsed := time.Since(start)
	if err != nil {
		p.log.Errorf("Failed to write create batch: %v", err)
	} else {
		var totalBytes int32 = 0
		for _, msg := range batch {
			totalBytes += utils.TotalSize(&msg)
		}
		p.log.Infof("Produced create batch: %d messages, %d bytes, took %s", len(batch), totalBytes, elapsed)
	}

	return err
}

func (p *Producer) flushDelete(ctx context.Context) error {
	p.deleteMutex.Lock()
	batch := p.deleteBatch
	if len(batch) == 0 {
		p.deleteMutex.Unlock()
		return nil
	}
	p.deleteBatch = make([]kafka.Message, 0, p.maxBatchSize)
	p.deleteMutex.Unlock()

	start := time.Now()
	err := p.deleteWriter.WriteMessages(ctx, batch...)
	elapsed := time.Since(start)
	if err != nil {
		p.log.Errorf("Failed to write delete batch: %v", err)
	} else {
		var totalBytes int32 = 0
		for _, msg := range batch {
			totalBytes += utils.TotalSize(&msg)
		}
		p.log.Infof("Produced delete batch: %d messages, %d bytes, took %s", len(batch), totalBytes, elapsed)
	}

	return err
}
