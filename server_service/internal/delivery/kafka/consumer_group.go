package kafka

import (
	"context"
	"sync"

	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/lits-06/vcs-sms/server_service/internal/domain"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

type ConsumerGroup struct {
	Brokers  []string
	log      logger.Logger
	serverUC domain.UseCase
}

func NewConsumerGroup(brokers []string, log logger.Logger, serverUC domain.UseCase) *ConsumerGroup {
	return &ConsumerGroup{
		Brokers:  brokers,
		log:      log,
		serverUC: serverUC,
	}
}

func (cg *ConsumerGroup) getNewKafkaReader(kafkaURL []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:                kafkaURL,
		Topic:                  topic,
		GroupID:                groupID,
		MinBytes:               minBytes,
		MaxBytes:               maxBytes,
		QueueCapacity:          queueCapacity,
		HeartbeatInterval:      heartbeatInterval,
		CommitInterval:         commitInterval,
		PartitionWatchInterval: partitionWatchInterval,
		Logger:                 kafka.LoggerFunc(cg.log.Debugf),
		ErrorLogger:            kafka.LoggerFunc(cg.log.Errorf),
		MaxAttempts:            maxAttempts,
		Dialer: &kafka.Dialer{
			Timeout: dialTimeout,
		},
	})
}

func (cg *ConsumerGroup) getNewKafkaWriter(topic string) *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(cg.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: writerRequiredAcks,
		MaxAttempts:  writerMaxAttempts,
		Logger:       kafka.LoggerFunc(cg.log.Debugf),
		ErrorLogger:  kafka.LoggerFunc(cg.log.Errorf),
		Compression:  compress.Snappy,
		ReadTimeout:  writerReadTimeout,
		WriteTimeout: writerWriteTimeout,
	}
	return w
}

func (cg *ConsumerGroup) consumeUpdateServerStatus(
	ctx context.Context,
	cancel context.CancelFunc,
	groupID string,
	topic string,
	workerNum int,
) {
	r := cg.getNewKafkaReader(cg.Brokers, topic, groupID)
	defer cancel()
	defer func() {
		if err := r.Close(); err != nil {
			cg.log.Error("r.Close: %v", err)
		}
	}()

	w := cg.getNewKafkaWriter(deadLetterQueueTopic)
	defer func() {
		if err := w.Close(); err != nil {
			cg.log.Error("w.Close: %v", err)
			cancel()
		}
	}()

	cg.log.Infof("Starting consumer group: %v", r.Config().GroupID)
	wg := &sync.WaitGroup{}
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go cg.updateWorker(ctx, cancel, wg, r, w, i)
	}
	wg.Wait()
}

func (cg *ConsumerGroup) RunConsumers(ctx context.Context, cancel context.CancelFunc) {
	go cg.consumeUpdateServerStatus(ctx, cancel, updateGroupID, updateTopic, updateWorkerCount)
}
