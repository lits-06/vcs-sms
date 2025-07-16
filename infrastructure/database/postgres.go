package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/entity"
	"github.com/lits-06/vcs-sms/usecases/server"
)

type PostgresDB struct {
	*sql.DB
}

// NewPostgresConnection tạo kết nối PostgreSQL với connection pooling
func NewPostgresConnection() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DB_HOST, config.DB_PORT, config.DB_USER, config.DB_PASSWORD, config.DB_NAME, config.DB_SSL_MODE,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Cấu hình connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test kết nối
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

type serverRepository struct {
	db *sql.DB
}

// NewServerRepository tạo instance mới của server repository
func NewServerRepository(db *sql.DB) server.Repository {
	return &serverRepository{
		db: db,
	}
}

func (r *serverRepository) Create(ctx context.Context, srv *entity.Server) error {
	query := `
        INSERT INTO servers (id, name, status, created_at, updated_at, last_checked, ipv4)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	_, err := r.db.ExecContext(ctx, query,
		srv.ID,
		srv.Name,
		string(srv.Status),
		srv.CreatedAt,
		srv.UpdatedAt,
		srv.LastChecked,
		srv.IPv4,
	)

	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	return nil
}

func (r *serverRepository) GetByID(ctx context.Context, id string) (*entity.Server, error) {
	query := `
        SELECT id, name, status, created_at, updated_at, last_checked, ipv4
        FROM servers WHERE id = $1
    `

	var srv entity.Server
	var status string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&srv.ID,
		&srv.Name,
		&status,
		&srv.CreatedAt,
		&srv.UpdatedAt,
		&srv.LastChecked,
		&srv.IPv4,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Không tìm thấy
		}
		return nil, fmt.Errorf("failed to get server by ID: %w", err)
	}

	srv.Status = entity.ServerStatus(status)
	return &srv, nil
}

func (r *serverRepository) GetByName(ctx context.Context, name string) (*entity.Server, error) {
	query := `
        SELECT id, name, status, created_at, updated_at, last_checked, ipv4
        FROM servers WHERE name = $1
    `

	var srv entity.Server
	var status string

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&srv.ID,
		&srv.Name,
		&status,
		&srv.CreatedAt,
		&srv.UpdatedAt,
		&srv.LastChecked,
		&srv.IPv4,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get server by name: %w", err)
	}

	srv.Status = entity.ServerStatus(status)
	return &srv, nil
}

func (r *serverRepository) List(ctx context.Context, filter server.ServerFilter, sort server.ServerSort, pagination server.ServerPagination) ([]entity.Server, int, error) {
	var args []interface{}

	// Xây dựng WHERE clause
	whereClause := r.buildWhereClause(filter, &args)

	// Đếm tổng số records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM servers %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count servers: %w", err)
	}

	// Xây dựng ORDER BY clause
	orderClause := r.buildOrderClause(sort)

	// Xây dựng LIMIT và OFFSET
	limitClause := r.buildLimitClause(pagination, &args)

	// Query chính
	query := fmt.Sprintf(`
        SELECT id, name, status, created_at, updated_at, last_checked, ipv4
        FROM servers %s %s %s
    `, whereClause, orderClause, limitClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query servers: %w", err)
	}
	defer rows.Close()

	var servers []entity.Server
	for rows.Next() {
		var srv entity.Server
		var status string

		err := rows.Scan(
			&srv.ID,
			&srv.Name,
			&status,
			&srv.CreatedAt,
			&srv.UpdatedAt,
			&srv.LastChecked,
			&srv.IPv4,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan server: %w", err)
		}

		srv.Status = entity.ServerStatus(status)
		servers = append(servers, srv)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return servers, total, nil
}

func (r *serverRepository) buildWhereClause(filter server.ServerFilter, args *[]interface{}) string {
	var conditions []string
	argIndex := len(*args) + 1

	if filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		*args = append(*args, "%"+filter.Name+"%")
		argIndex++
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		*args = append(*args, string(filter.Status))
		argIndex++
	}

	if filter.IPv4 != "" {
		conditions = append(conditions, fmt.Sprintf("ipv4 = $%d", argIndex))
		*args = append(*args, filter.IPv4)
		argIndex++
	}

	if len(conditions) == 0 {
		return ""
	}

	return "WHERE " + strings.Join(conditions, " AND ")
}

func (r *serverRepository) buildOrderClause(sort server.ServerSort) string {
	field := "created_at" // Default sort field
	if sort.Sort != "" {
		field = sort.Sort
	}

	order := "ASC" // Default order
	if sort.Order == server.SortDesc {
		order = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", field, order)
}

func (r *serverRepository) buildLimitClause(pagination server.ServerPagination, args *[]interface{}) string {
	if pagination.To < pagination.From {
		return ""
	}

	argIndex := len(*args) + 1

	if pagination.From > 0 {
		if pagination.To == 0 {
			// Chỉ có FROM - không giới hạn số lượng
			*args = append(*args, pagination.From)
			return fmt.Sprintf("OFFSET $%d", argIndex)
		} else {
			// Cả FROM và TO - giới hạn số lượng
			*args = append(*args, pagination.To-pagination.From, pagination.From)
			return fmt.Sprintf("LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		}
	}

	if pagination.To > 0 {
		// Chỉ có TO - giới hạn từ đầu
		*args = append(*args, pagination.To)
		return fmt.Sprintf("LIMIT $%d", argIndex)
	}

	return ""
}

func (r *serverRepository) Update(ctx context.Context, srv *entity.Server) error {
	query := `
        UPDATE servers 
        SET name = $2, status = $3, updated_at = $4, ipv4 = $5
        WHERE id = $1
    `

	result, err := r.db.ExecContext(ctx, query,
		srv.ID,
		srv.Name,
		string(srv.Status),
		time.Now(),
		srv.IPv4,
	)

	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("server with ID %s not found", srv.ID)
	}

	return nil
}

func (r *serverRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM servers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("server with ID %s not found", id)
	}

	return nil
}

func (r *serverRepository) ExistsWithID(ctx context.Context, id string) (bool, error) {
	query := `SELECT 1 FROM servers WHERE id = $1 LIMIT 1`

	var exists int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if server exists by ID: %w", err)
	}

	return true, nil
}

func (r *serverRepository) ExistsWithName(ctx context.Context, name string) (bool, error) {
	query := `SELECT 1 FROM servers WHERE name = $1 LIMIT 1`

	var exists int
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if server exists by name: %w", err)
	}

	return true, nil
}
