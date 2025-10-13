package kafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/lits-06/vcs-sms/server_service/internal/domain"
	"github.com/segmentio/kafka-go"
)

const (
	retryAttempts = 1
	retryDelay    = 1 * time.Second
	batchSize     = 1000
	batchTimeout  = 100 * time.Millisecond
)

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
		batch      []domain.ServerStatusUpdate
		messages   []kafka.Message
		batchTimer = time.NewTicker(batchTimeout)
	)

	defer batchTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				cg.processBatch(ctx, r, batch, messages, workerID)
			}
			return

		case <-batchTimer.C:
			if len(batch) > 0 {
				cg.processBatch(ctx, r, batch, messages, workerID)
				batch = batch[:0]
				messages = messages[:0]
			}

		default:
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

			batch = append(batch, domain.ServerStatusUpdate{
				ServerID: server.ID,
				Status:   server.Status,
			})
			messages = append(messages, m)

			if len(batch) >= batchSize {
				cg.processBatch(ctx, r, batch, messages, workerID)
				batch = batch[:0]
				messages = messages[:0]
			}
		}
	}
}

func (cg *ConsumerGroup) processBatch(
	ctx context.Context,
	r *kafka.Reader,
	batch []domain.ServerStatusUpdate,
	messages []kafka.Message,
	workerID int,
) {
	if len(batch) == 0 {
		return
	}

	start := time.Now()

	if err := retry.Do(func() error {
		return cg.serverUC.BulkUpdateServerStatus(ctx, batch)
	},
		retry.Attempts(retryAttempts),
		retry.Delay(retryDelay),
		retry.Context(ctx),
	); err != nil {
		cg.log.Warnf("Failed to process batch: %v", err)
		return
	}

	// Commit all messages in batch
	if err := r.CommitMessages(ctx, messages...); err != nil {
		cg.log.Warnf("Failed to commit batch: %v", err)
		return
	}

	cg.log.Infof("Worker %d: successfully processed batch of %d server updates in %s",
		workerID, len(batch), time.Since(start))
}
