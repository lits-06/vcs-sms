package kafka

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/lits-06/vcs-sms/healthcheck_service/config"
	"github.com/lits-06/vcs-sms/healthcheck_service/internal/domain"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

type ConsumerGroup struct {
	Brokers       []string
	log           logger.Logger
	cfg           *config.Config
	healthCheckUC domain.UseCase
}

func NewConsumerGroup(brokers []string, log logger.Logger, cfg *config.Config, healthCheckUC domain.UseCase) *ConsumerGroup {
	return &ConsumerGroup{
		Brokers:       brokers,
		log:           log,
		cfg:           cfg,
		healthCheckUC: healthCheckUC,
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

func (cg *ConsumerGroup) consumerCreateServer(
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
			cg.log.Errorf("r.Close: %v", err)
		}
	}()

	w := cg.getNewKafkaWriter(deadLetterQueueTopic)
	defer func() {
		if err := w.Close(); err != nil {
			cg.log.Errorf("w.Close: %v", err)
			cancel()
		}
	}()

	cg.log.Infof("Starting consumer group: %v", r.Config().GroupID)
	wg := &sync.WaitGroup{}
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go cg.createWorker(ctx, cancel, wg, r, w, i)
	}
	wg.Wait()
}

func (cg *ConsumerGroup) consumeDeleteServer(
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
			cg.log.Errorf("r.Close: %v", err)
		}
	}()

	w := cg.getNewKafkaWriter(deadLetterQueueTopic)
	defer func() {
		if err := w.Close(); err != nil {
			cg.log.Errorf("w.Close: %v", err)
			cancel()
		}
	}()

	cg.log.Infof("Starting consumer group: %v", r.Config().GroupID)
	wg := &sync.WaitGroup{}
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go cg.deleteWorker(ctx, cancel, wg, r, w, i)
	}
	wg.Wait()
}

func (cg *ConsumerGroup) consumeUpdateServerState(
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
			cg.log.Errorf("r.Close: %v", err)
		}
	}()

	w := cg.getNewKafkaWriter(deadLetterQueueTopic)
	defer func() {
		if err := w.Close(); err != nil {
			cg.log.Errorf("w.Close: %v", err)
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

func (cg *ConsumerGroup) publishErrorMessage(ctx context.Context, w *kafka.Writer, m kafka.Message, err error) error {
	errMsg := &domain.ErrorMessage{
		Offset:    m.Offset,
		Error:     err.Error(),
		Time:      m.Time.UTC(),
		Partition: m.Partition,
		Topic:     m.Topic,
	}

	cg.log.Debugf("Publishing error message: %v", errMsg)

	errMsgBytes, err := json.Marshal(errMsg)
	if err != nil {
		cg.log.Errorf("json.Marshal: %v", err)
		return err
	}

	return w.WriteMessages(ctx, kafka.Message{
		Value: errMsgBytes,
	})
}

func (cg *ConsumerGroup) RunConsumers(ctx context.Context, cancel context.CancelFunc) {
	go cg.consumerCreateServer(ctx, cancel, createGroupID, createTopic, createWorkerCount)
	go cg.consumeDeleteServer(ctx, cancel, deleteGroupID, deleteTopic, deleteWorkerCount)
	go cg.consumeUpdateServerState(ctx, cancel, updateGroupID, updateTopic, updateWorkerCount)
}
