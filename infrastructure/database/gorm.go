package database

import (
	"context"
	"fmt"
	"time"

	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/entity"
	"github.com/lits-06/vcs-sms/usecases/server"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&entity.Server{}); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	return nil
}

type GormDB struct {
	*gorm.DB
}

func NewGormConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Open connection
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

type gormServerRepository struct {
	db *gorm.DB
}

// NewServerRepository creates a new GORM server repository
func NewServerRepository(db *gorm.DB) server.Repository {
	return &gormServerRepository{
		db: db,
	}
}

func (r *gormServerRepository) Create(ctx context.Context, srv *entity.Server) error {
	if err := r.db.WithContext(ctx).Create(srv).Error; err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	return nil
}

func (r *gormServerRepository) GetByID(ctx context.Context, id string) (*entity.Server, error) {
	var srv entity.Server
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&srv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get server by ID: %w", err)
	}
	return &srv, nil
}

func (r *gormServerRepository) GetByName(ctx context.Context, name string) (*entity.Server, error) {
	var srv entity.Server
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&srv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get server by name: %w", err)
	}
	return &srv, nil
}

func (r *gormServerRepository) Update(ctx context.Context, srv *entity.Server) error {
	data := make(map[string]interface{})
	if srv.Name != "" {
		data["name"] = srv.Name
	}
	if srv.Status != "" {
		data["status"] = srv.Status
	}
	if srv.IPv4 != "" {
		data["ipv4"] = srv.IPv4
	}

	result := r.db.WithContext(ctx).Model(srv).Where("id = ?", srv.ID).Updates(data)

	if result.Error != nil {
		return fmt.Errorf("failed to update server: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("server with ID %s not found", srv.ID)
	}

	return nil
}

func (r *gormServerRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Server{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete server: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("server with ID %s not found", id)
	}

	return nil
}

func (r *gormServerRepository) List(ctx context.Context, filter server.ServerFilter, sort server.ServerSort, pagination server.ServerPagination) (*[]entity.Server, int, error) {
	var servers []entity.Server
	var total int64

	// Build base query with filters
	query := r.db.WithContext(ctx).Model(&entity.Server{})
	query = r.applyFilters(query, filter)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count servers: %w", err)
	}

	// Apply sorting
	query = r.applySorting(query, sort)

	// Apply pagination
	query = r.applyPagination(query, pagination)

	// Execute query
	if err := query.Find(&servers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list servers: %w", err)
	}

	return &servers, int(total), nil
}

func (r *gormServerRepository) applyFilters(query *gorm.DB, filter server.ServerFilter) *gorm.DB {
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if filter.Status != "" {
		query = query.Where("status = ?", string(filter.Status))
	}

	if filter.IPv4 != "" {
		query = query.Where("ipv4 = ?", filter.IPv4)
	}

	return query
}

func (r *gormServerRepository) applySorting(query *gorm.DB, sort server.ServerSort) *gorm.DB {
	field := "created_at" // Default sort field
	if sort.Sort != "" {
		field = sort.Sort
	}

	order := "ASC" // Default order
	if sort.Order == server.SortDesc {
		order = "DESC"
	}

	return query.Order(fmt.Sprintf("%s %s", field, order))
}

func (r *gormServerRepository) applyPagination(query *gorm.DB, pagination server.ServerPagination) *gorm.DB {
	if pagination.To < pagination.From {
		return query
	}

	if pagination.From > 0 {
		query = query.Offset(pagination.From)
	}

	if pagination.To > 0 {
		limit := pagination.To
		if pagination.From > 0 {
			limit = pagination.To - pagination.From
		}
		query = query.Limit(limit)
	}

	return query
}

func (r *gormServerRepository) ExistsWithID(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Server{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if server exists by ID: %w", err)
	}
	return count > 0, nil
}

func (r *gormServerRepository) ExistsWithName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Server{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if server exists by name: %w", err)
	}
	return count > 0, nil
}
