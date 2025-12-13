package database

import (
	"context"
	"fmt"

	"github.com/nadmax/dbcompare/internal/config"
	"github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type SurrealDB struct {
	db     *surrealdb.DB
	config *config.SurrealDBConfig
	ctx    context.Context
}

func NewSurrealDB(cfg *config.SurrealDBConfig) (*SurrealDB, error) {
	ctx := context.Background()
	db, err := surrealdb.FromEndpointURLString(ctx, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if err = db.Use(ctx, cfg.Namespace, cfg.Database); err != nil {
		return nil, fmt.Errorf("failed to use namespace/database: %w", err)
	}

	authData := &surrealdb.Auth{
		Username: cfg.User,
		Password: cfg.Password,
	}
	_, err = db.SignIn(ctx, authData)
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}

	return &SurrealDB{
		db:     db,
		config: cfg,
		ctx:    ctx,
	}, nil
}

func (s *SurrealDB) DB() *surrealdb.DB {
	return s.db
}

func (s *SurrealDB) Context() context.Context {
	return s.ctx
}

func (s *SurrealDB) Close() error {
	return s.db.Close(s.ctx)
}

func (s *SurrealDB) CreateSchema() error {
	_, err := surrealdb.Delete[[]map[string]any](s.ctx, s.db, models.Table("test_records"))
	if err != nil {
		fmt.Printf("Note: Could not delete test_records (might not exist): %v\n", err)
	}

	return nil
}

func (s *SurrealDB) TruncateTable() error {
	_, err := surrealdb.Delete[[]map[string]any](s.ctx, s.db, models.Table("test_records"))
	return err
}

func (s *SurrealDB) GetStats() (map[string]any, error) {
	stats := make(map[string]any)

	records, err := surrealdb.Select[[]map[string]any](s.ctx, s.db, models.Table("test_records"))
	if err != nil {
		return nil, err
	}

	stats["row_count"] = len(*records)
	stats["namespace"] = s.config.Namespace
	stats["database"] = s.config.Database

	return stats, nil
}
