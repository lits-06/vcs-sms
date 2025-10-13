package repository

import (
	"bytes"
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

func NewESRepository(client *elasticsearch.Client, cfg *config.Config) domain.Repository {
	return &esRepository{
		client:      client,
		snapshotIdx: cfg.Elasticsearch.SnapshotIndex,
		recordIdx:   cfg.Elasticsearch.RecordIndex,
	}
}

func (r *esRepository) GetAllServersSnapshot(ctx context.Context) ([]domain.Server, error) {
	var servers []domain.Server

	query := `{
		"size": 10000,
		"query": {
			"match_all": {}
		}
	}`

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.snapshotIdx),
		r.client.Search.WithBody(strings.NewReader(query)),
		r.client.Search.WithTrackTotalHits(true),
		r.client.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("res.IsError.Decode: %s", err)
		} else {
			return nil, fmt.Errorf("res.IsError [%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	var rBody map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&rBody); err != nil {
		return nil, err
	}

	hits := rBody["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		sourceBytes, err := json.Marshal(source)
		if err != nil {
			return nil, err
		}

		var server domain.Server
		if err := json.Unmarshal(sourceBytes, &server); err != nil {
			return nil, err
		}

		servers = append(servers, server)
	}

	return servers, nil
}

func (r *esRepository) SaveServerSnapshot(ctx context.Context, server *domain.Server) error {
	sbytes, err := json.Marshal(server)
	if err != nil {
		return err
	}

	res, err := r.client.Index(
		r.snapshotIdx,
		strings.NewReader(string(sbytes)),
		r.client.Index.WithContext(ctx),
		r.client.Index.WithDocumentID(server.ServerID),
		r.client.Index.WithRefresh("false"), // Changed to false for better performance
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

func (r *esRepository) IndexServerState(ctx context.Context, server *domain.Server) error {
	sbytes, err := json.Marshal(server)
	if err != nil {
		return err
	}

	res, err := r.client.Index(
		r.recordIdx,
		strings.NewReader(string(sbytes)),
		r.client.Index.WithContext(ctx),
		r.client.Index.WithRefresh("false"), // Changed to false for better performance
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

func (r *esRepository) BulkIndexServerStates(ctx context.Context, servers []domain.Server) error {
	if len(servers) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, server := range servers {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": r.recordIdx,
			},
		}
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		buf.Write(metaBytes)
		buf.WriteByte('\n')

		docBytes, err := json.Marshal(server)
		if err != nil {
			return err
		}
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}

	res, err := r.client.Bulk(
		bytes.NewReader(buf.Bytes()),
		r.client.Bulk.WithContext(ctx),
		r.client.Bulk.WithRefresh("false"),
	)

	if err != nil {
		return fmt.Errorf("client.Bulk: %w", err)
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

func (r *esRepository) BulkSaveServerSnapshots(ctx context.Context, servers []domain.Server) error {
	if len(servers) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, server := range servers {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": r.snapshotIdx,
				"_id":    server.ServerID,
			},
		}
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		buf.Write(metaBytes)
		buf.WriteByte('\n')

		docBytes, err := json.Marshal(server)
		if err != nil {
			return err
		}
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}

	res, err := r.client.Bulk(
		bytes.NewReader(buf.Bytes()),
		r.client.Bulk.WithContext(ctx),
		r.client.Bulk.WithRefresh("false"),
	)

	if err != nil {
		return fmt.Errorf("client.Bulk: %w", err)
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

func (r *esRepository) DeleteServerSnapshot(ctx context.Context, serverID string) error {
	res, err := r.client.Delete(
		r.snapshotIdx,
		serverID,
		r.client.Delete.WithContext(ctx),
		r.client.Delete.WithRefresh("true"),
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
