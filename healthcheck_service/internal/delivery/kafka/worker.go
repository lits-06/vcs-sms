package kafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/lits-06/vcs-sms/healthcheck_service/internal/domain"
	"github.com/segmentio/kafka-go"
)

const (
	retryAttempts = 1
	retryDelay    = 1 * time.Second
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

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			cg.log.Error("r.FetchMessage: %v", err)
			return
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

		var server domain.Server
		if err := json.Unmarshal(m.Value, &server); err != nil {
			cg.log.Error("json.Unmarshal: %v", err)
			continue
		}

		if server.Timestamp.IsZero() {
			server.Timestamp = time.Now()
		}

		if err := retry.Do(func() error {
			err := cg.healthCheckUC.IndexServerState(ctx, &server)
			if err != nil {
				return err
			}
			cg.log.Infof("Indexed server state, created: %v", server)
			return nil
		},
			retry.Attempts(retryAttempts),
			retry.Delay(retryDelay),
			retry.Context(ctx),
		); err != nil {
			if err := cg.publishErrorMessage(ctx, w, m, err); err != nil {
				cg.log.Error("cg.publishErrorMessage: %v", err)
				continue
			}
			cg.log.Error("cg.healthCheckUC.IndexServerState: %v", err)
			continue
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			cg.log.Error("r.CommitMessages: %v", err)
			continue
		}
	}
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
			cg.log.Error("r.FetchMessage: %v", err)
			return
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

		var serverID string
		if err := json.Unmarshal(m.Value, &serverID); err != nil {
			cg.log.Error("json.Unmarshal: %v", err)
			continue
		}

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
				cg.log.Error("cg.publishErrorMessage: %v", err)
				continue
			}
			cg.log.Error("cg.healthCheckUC.DeleteServerSnapshot: %v", err)
			continue
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			cg.log.Error("r.CommitMessages: %v", err)
			continue
		}
	}
}
