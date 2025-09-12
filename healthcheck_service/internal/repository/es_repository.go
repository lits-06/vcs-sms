package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/lits-06/vcs-sms/healthcheck_service/config"
	"github.com/lits-06/vcs-sms/healthcheck_service/internal/domain"
)

type esRepository struct {
	client      *elasticsearch.Client
	snapshotIdx string
	recordIdx   string
}

func NewESRepository(client *elasticsearch.Client, cfg *config.Config) domain.StateRepository {
	return &esRepository{
		client:      client,
		snapshotIdx: cfg.Elasticsearch.SnapshotIndex,
		recordIdx:   cfg.Elasticsearch.RecordIndex,
	}
}

func (r *esRepository) SaveServerSnapshot(ctx context.Context, snapshot *domain.ServerSnapshot) error {
	ssbytes, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	res, err := r.client.Index(
		r.snapshotIdx,
		strings.NewReader(string(ssbytes)),
		r.client.Index.WithContext(ctx),
		r.client.Index.WithDocumentID(snapshot.ServerID),
		r.client.Index.WithRefresh("true"),
	)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return fmt.Errorf("res.IsError.Decode: %s", err)
		} else {
			return fmt.Errorf("res.IsError [%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	return nil
}
