package postgres

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"strings"
// 	"time"

// 	"github.com/lib/pq"
// 	"github.com/lits-06/vcs-sms/internal/domain"
// )

// type serverRepository struct {
// 	db *sql.DB
// }

// // NewServerRepository creates a new server repository
// func NewServerRepository(db *sql.DB) domain.ServerRepository {
// 	return &serverRepository{db: db}
// }

// // Create creates a new server
// func (r *serverRepository) Create(ctx context.Context, server *domain.Server) error {
// 	query := `
// 		INSERT INTO servers (name, host, port, status, description, tags, created_at, updated_at, last_checked)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
// 		RETURNING id`

// 	now := time.Now()
// 	server.CreatedAt = now
// 	server.UpdatedAt = now
// 	server.LastChecked = now
// 	server.Status = string(domain.StatusUnknown)

// 	err := r.db.QueryRowContext(ctx, query,
// 		server.Name,
// 		server.Host,
// 		server.Port,
// 		server.Status,
// 		server.Description,
// 		pq.Array(server.Tags),
// 		server.CreatedAt,
// 		server.UpdatedAt,
// 		server.LastChecked,
// 	).Scan(&server.ID)

// 	if err != nil {
// 		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
// 			return domain.ErrServerExists
// 		}
// 		return fmt.Errorf("failed to create server: %w", err)
// 	}

// 	return nil
// }

// // GetByID retrieves a server by ID
// func (r *serverRepository) GetByID(ctx context.Context, id int64) (*domain.Server, error) {
// 	query := `
// 		SELECT id, name, host, port, status, description, tags, created_at, updated_at, last_checked
// 		FROM servers WHERE id = $1`

// 	server := &domain.Server{}
// 	err := r.db.QueryRowContext(ctx, query, id).Scan(
// 		&server.ID,
// 		&server.Name,
// 		&server.Host,
// 		&server.Port,
// 		&server.Status,
// 		&server.Description,
// 		pq.Array(&server.Tags),
// 		&server.CreatedAt,
// 		&server.UpdatedAt,
// 		&server.LastChecked,
// 	)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, domain.ErrServerNotFound
// 		}
// 		return nil, fmt.Errorf("failed to get server: %w", err)
// 	}

// 	return server, nil
// }

// // Update updates a server
// func (r *serverRepository) Update(ctx context.Context, id int64, server *domain.Server) error {
// 	query := `
// 		UPDATE servers
// 		SET name = $2, host = $3, port = $4, description = $5, tags = $6, updated_at = $7
// 		WHERE id = $1`

// 	server.UpdatedAt = time.Now()

// 	result, err := r.db.ExecContext(ctx, query,
// 		id,
// 		server.Name,
// 		server.Host,
// 		server.Port,
// 		server.Description,
// 		pq.Array(server.Tags),
// 		server.UpdatedAt,
// 	)

// 	if err != nil {
// 		return fmt.Errorf("failed to update server: %w", err)
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return fmt.Errorf("failed to get rows affected: %w", err)
// 	}

// 	if rowsAffected == 0 {
// 		return domain.ErrServerNotFound
// 	}

// 	return nil
// }

// // Delete deletes a server
// func (r *serverRepository) Delete(ctx context.Context, id int64) error {
// 	query := `DELETE FROM servers WHERE id = $1`

// 	result, err := r.db.ExecContext(ctx, query, id)
// 	if err != nil {
// 		return fmt.Errorf("failed to delete server: %w", err)
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return fmt.Errorf("failed to get rows affected: %w", err)
// 	}

// 	if rowsAffected == 0 {
// 		return domain.ErrServerNotFound
// 	}

// 	return nil
// }

// // List retrieves servers with filtering, sorting, and pagination
// func (r *serverRepository) List(ctx context.Context, req domain.ServerListRequest) (*domain.ServerListResponse, error) {
// 	// Build WHERE clause
// 	whereClause, args := r.buildWhereClause(req.Filter)

// 	// Build ORDER BY clause
// 	orderClause := r.buildOrderClause(req.Sort)

// 	// Calculate offset
// 	offset := (req.Pagination.Page - 1) * req.Pagination.Limit

// 	// Query for total count
// 	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM servers %s", whereClause)
// 	var total int64
// 	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to count servers: %w", err)
// 	}

// 	// Query for servers
// 	query := fmt.Sprintf(`
// 		SELECT id, name, host, port, status, description, tags, created_at, updated_at, last_checked
// 		FROM servers %s %s LIMIT $%d OFFSET $%d`,
// 		whereClause, orderClause, len(args)+1, len(args)+2)

// 	args = append(args, req.Pagination.Limit, offset)

// 	rows, err := r.db.QueryContext(ctx, query, args...)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query servers: %w", err)
// 	}
// 	defer rows.Close()

// 	var servers []domain.Server
// 	for rows.Next() {
// 		var server domain.Server
// 		err := rows.Scan(
// 			&server.ID,
// 			&server.Name,
// 			&server.Host,
// 			&server.Port,
// 			&server.Status,
// 			&server.Description,
// 			pq.Array(&server.Tags),
// 			&server.CreatedAt,
// 			&server.UpdatedAt,
// 			&server.LastChecked,
// 		)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to scan server: %w", err)
// 		}
// 		servers = append(servers, server)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, fmt.Errorf("failed to iterate rows: %w", err)
// 	}

// 	totalPages := int((total + int64(req.Pagination.Limit) - 1) / int64(req.Pagination.Limit))

// 	return &domain.ServerListResponse{
// 		Servers:    servers,
// 		Total:      total,
// 		Page:       req.Pagination.Page,
// 		Limit:      req.Pagination.Limit,
// 		TotalPages: totalPages,
// 	}, nil
// }

// // GetAll retrieves all servers
// func (r *serverRepository) GetAll(ctx context.Context) ([]domain.Server, error) {
// 	query := `
// 		SELECT id, name, host, port, status, description, tags, created_at, updated_at, last_checked
// 		FROM servers ORDER BY id`

// 	rows, err := r.db.QueryContext(ctx, query)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query servers: %w", err)
// 	}
// 	defer rows.Close()

// 	var servers []domain.Server
// 	for rows.Next() {
// 		var server domain.Server
// 		err := rows.Scan(
// 			&server.ID,
// 			&server.Name,
// 			&server.Host,
// 			&server.Port,
// 			&server.Status,
// 			&server.Description,
// 			pq.Array(&server.Tags),
// 			&server.CreatedAt,
// 			&server.UpdatedAt,
// 			&server.LastChecked,
// 		)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to scan server: %w", err)
// 		}
// 		servers = append(servers, server)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, fmt.Errorf("failed to iterate rows: %w", err)
// 	}

// 	return servers, nil
// }

// // BulkCreate creates multiple servers
// func (r *serverRepository) BulkCreate(ctx context.Context, servers []domain.Server) error {
// 	if len(servers) == 0 {
// 		return nil
// 	}

// 	tx, err := r.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback()

// 	stmt, err := tx.PrepareContext(ctx, `
// 		INSERT INTO servers (name, host, port, status, description, tags, created_at, updated_at, last_checked)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`)
// 	if err != nil {
// 		return fmt.Errorf("failed to prepare statement: %w", err)
// 	}
// 	defer stmt.Close()

// 	now := time.Now()
// 	for _, server := range servers {
// 		_, err := stmt.ExecContext(ctx,
// 			server.Name,
// 			server.Host,
// 			server.Port,
// 			string(domain.StatusUnknown),
// 			server.Description,
// 			pq.Array(server.Tags),
// 			now,
// 			now,
// 			now,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("failed to insert server: %w", err)
// 		}
// 	}

// 	if err := tx.Commit(); err != nil {
// 		return fmt.Errorf("failed to commit transaction: %w", err)
// 	}

// 	return nil
// }

// // GetByHostPort retrieves a server by host and port
// func (r *serverRepository) GetByHostPort(ctx context.Context, host string, port int) (*domain.Server, error) {
// 	query := `
// 		SELECT id, name, host, port, status, description, tags, created_at, updated_at, last_checked
// 		FROM servers WHERE host = $1 AND port = $2`

// 	server := &domain.Server{}
// 	err := r.db.QueryRowContext(ctx, query, host, port).Scan(
// 		&server.ID,
// 		&server.Name,
// 		&server.Host,
// 		&server.Port,
// 		&server.Status,
// 		&server.Description,
// 		pq.Array(&server.Tags),
// 		&server.CreatedAt,
// 		&server.UpdatedAt,
// 		&server.LastChecked,
// 	)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, domain.ErrServerNotFound
// 		}
// 		return nil, fmt.Errorf("failed to get server: %w", err)
// 	}

// 	return server, nil
// }

// // buildWhereClause builds WHERE clause for filtering
// func (r *serverRepository) buildWhereClause(filter domain.ServerFilter) (string, []interface{}) {
// 	var conditions []string
// 	var args []interface{}
// 	argIndex := 1

// 	if filter.Name != "" {
// 		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
// 		args = append(args, "%"+filter.Name+"%")
// 		argIndex++
// 	}

// 	if filter.Host != "" {
// 		conditions = append(conditions, fmt.Sprintf("host ILIKE $%d", argIndex))
// 		args = append(args, "%"+filter.Host+"%")
// 		argIndex++
// 	}

// 	if filter.Status != "" {
// 		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
// 		args = append(args, string(filter.Status))
// 		argIndex++
// 	}

// 	if len(filter.Tags) > 0 {
// 		conditions = append(conditions, fmt.Sprintf("tags && $%d", argIndex))
// 		args = append(args, pq.Array(filter.Tags))
// 		argIndex++
// 	}

// 	if len(conditions) == 0 {
// 		return "", args
// 	}

// 	return "WHERE " + strings.Join(conditions, " AND "), args
// }

// // buildOrderClause builds ORDER BY clause for sorting
// func (r *serverRepository) buildOrderClause(sort domain.ServerSort) string {
// 	if sort.Field == "" {
// 		return "ORDER BY id ASC"
// 	}

// 	order := "ASC"
// 	if sort.Order == "desc" {
// 		order = "DESC"
// 	}

// 	return fmt.Sprintf("ORDER BY %s %s", sort.Field, order)
// }
