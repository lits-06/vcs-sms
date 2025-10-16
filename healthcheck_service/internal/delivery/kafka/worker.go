package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/lits-06/vcs-sms/healthcheck_service/internal/domain"
	"github.com/lits-06/vcs-sms/pkg/utils"
	"github.com/segmentio/kafka-go"
)

const (
	retryAttempts = 1
	retryDelay    = 1 * time.Second
	batchSize     = 2000
	batchTimeout  = 1 * time.Second
)

func (cg *ConsumerGroup) createWorker(
	ctx context.Context,
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
	r *kafka.Reader,
	w *kafka.Writer,
	workerID int,
) {
	defer wg.Done()
	defer cancel()

	var (
		batch      []domain.Server
		messages   []kafka.Message
		batchTimer = time.NewTicker(batchTimeout)
	)

	defer batchTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			// Flush remaining batch before exiting
			if len(batch) > 0 {
				cg.processBatchIndex(ctx, w, r, batch, messages, workerID)
			}
			return

		case <-batchTimer.C:
			// Timeout reached, process whatever we have
			if len(batch) > 0 {
				cg.processBatchIndex(ctx, w, r, batch, messages, workerID)
				batch = batch[:0]
				messages = messages[:0]
			}

		default:
			// Try to fetch a message with a short timeout
			fetchCtx, fetchCancel := context.WithTimeout(ctx, 100*time.Millisecond)
			m, err := r.FetchMessage(fetchCtx)
			fetchCancel()

			if err != nil {
				cg.log.Warnf("r.FetchMessage: %v", err)
				continue
			}

			var server domain.Server
			if err := json.Unmarshal(m.Value, &server); err != nil {
				cg.log.Errorf("json.Unmarshal: %v", err)
				continue
			}

			if server.Timestamp.IsZero() {
				server.Timestamp = time.Now()
			}

			batch = append(batch, server)
			messages = append(messages, m)

			// Process batch if it reaches the size limit
			if len(batch) >= batchSize {
				cg.processBatchIndex(ctx, w, r, batch, messages, workerID)
				batch = batch[:0]
				messages = messages[:0]
			}
		}
	}
}

func (cg *ConsumerGroup) processBatchIndex(
	ctx context.Context,
	w *kafka.Writer,
	r *kafka.Reader,
	batch []domain.Server,
	messages []kafka.Message,
	workerID int,
) {
	if len(batch) == 0 {
		return
	}

	start := time.Now()

	var totalBytes int32 = 0
	for _, msg := range messages {
		totalBytes += utils.TotalSize(&msg)
	}

	cg.log.Infof("Worker %d: Processing batch - messages: %d, bytes: %d, topic: %s, group: %s",
		workerID, len(batch), totalBytes, r.Config().Topic, r.Config().GroupID)

	if err := retry.Do(func() error {
		return cg.healthCheckUC.BulkIndexServerStates(ctx, batch)
	},
		retry.Attempts(retryAttempts),
		retry.Delay(retryDelay),
		retry.Context(ctx),
	); err != nil {
		cg.log.Warnf("Worker %d: Failed to process batch: %v", workerID, err)
		return
	}

	// Commit all messages in batch
	if err := r.CommitMessages(ctx, messages...); err != nil {
		cg.log.Warnf("Worker %d: Failed to commit batch: %v", workerID, err)
		return
	}

	elapsed := time.Since(start)

	partition := make(map[int]int)
	for _, msg := range messages {
		partition[msg.Partition]++
	}

	var partitionInfo string
	for p, count := range partition {
		if partitionInfo != "" {
			partitionInfo += ", "
		}
		partitionInfo += fmt.Sprintf("partition%d:%d", p, count)
	}

	cg.log.Infof("Worker %d: Successfully processed batch - messages: %d, bytes: %d, duration: %s, topic: %s, group: %s [%s]",
		workerID, len(batch), totalBytes, elapsed, r.Config().Topic, r.Config().GroupID, partitionInfo)
}

func (cg *ConsumerGroup) deleteWorker(
	ctx context.Context,
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
	r *kafka.Reader,
	w *kafka.Writer,
	workerID int,
) {
	defer wg.Done()
	defer cancel()

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			cg.log.Warnf("r.FetchMessage: %v", err)
			continue
		}

		cg.log.Infof(
			"WORKER: %v, message at topic/partition/offset %v/%v/%v: %s = %s\n",
			workerID,
			m.Topic,
			m.Partition,
			m.Offset,
			string(m.Key),
			string(m.Value),
		)

		serverID := string(m.Value)

		if err := retry.Do(func() error {
			err := cg.healthCheckUC.DeleteServerSnapshot(ctx, serverID)
			if err != nil {
				return err
			}
			cg.log.Infof("Deleted server snapshot: %v", serverID)
			return nil
		},
			retry.Attempts(retryAttempts),
			retry.Delay(retryDelay),
			retry.Context(ctx),
		); err != nil {
			if err := cg.publishErrorMessage(ctx, w, m, err); err != nil {
				cg.log.Warnf("cg.publishErrorMessage: %v", err)
				continue
			}
			cg.log.Warnf("cg.healthCheckUC.DeleteServerSnapshot: %v", err)
			continue
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			cg.log.Warnf("r.CommitMessages: %v", err)
			continue
		}
	}
}

func (cg *ConsumerGroup) updateWorker(
	ctx context.Context,
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
	r *kafka.Reader,
	w *kafka.Writer,
	workerID int,
) {
	defer wg.Done()
	defer cancel()

	var (
		batch      []domain.Server
		messages   []kafka.Message
		batchTimer = time.NewTicker(batchTimeout)
	)

	defer batchTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			// Flush remaining batch before exiting
			if len(batch) > 0 {
				cg.processBatchIndex(ctx, w, r, batch, messages, workerID)
			}
			return

		case <-batchTimer.C:
			// Timeout reached, process whatever we have
			if len(batch) > 0 {
				cg.processBatchIndex(ctx, w, r, batch, messages, workerID)
				batch = batch[:0]
				messages = messages[:0]
			}

		default:
			// Try to fetch a message with a short timeout
			fetchCtx, fetchCancel := context.WithTimeout(ctx, 100*time.Millisecond)
			m, err := r.FetchMessage(fetchCtx)
			fetchCancel()

			if err != nil {
				cg.log.Warnf("r.FetchMessage: %v", err)
				continue
			}

			var server domain.Server
			if err := json.Unmarshal(m.Value, &server); err != nil {
				cg.log.Errorf("json.Unmarshal: %v", err)
				continue
			}

			if server.Timestamp.IsZero() {
				server.Timestamp = time.Now()
			}

			batch = append(batch, server)
			messages = append(messages, m)

			// Process batch if it reaches the size limit
			if len(batch) >= batchSize {
				cg.processBatchIndex(ctx, w, r, batch, messages, workerID)
				batch = batch[:0]
				messages = messages[:0]
			}
		}
	}
}
