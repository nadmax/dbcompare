package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/nadmax/dbcompare/internal/config"
)

type PostgresDB struct {
	db     *sql.DB
	config *config.PostgresConfig
}

func NewPostgresDB(cfg *config.PostgresConfig) (*PostgresDB, error) {
	connStr := cfg.ConnectionString()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxConnections / 5)

	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to ping database: %w (also failed to close: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{
		db:     db,
		config: cfg,
	}, nil
}

func (p *PostgresDB) DB() *sql.DB {
	return p.db
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

func (p *PostgresDB) CreateSchema() error {
	queries := []string{
		`DROP TABLE IF EXISTS benchmark_records CASCADE`,
		`CREATE TABLE benchmark_records (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL,
			age INTEGER NOT NULL,
			balance DECIMAL(10,2) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			description TEXT,
			is_active BOOLEAN NOT NULL DEFAULT true
		)`,
		`CREATE INDEX idx_email ON benchmark_records(email)`,
		`CREATE INDEX idx_age ON benchmark_records(age)`,
		`CREATE INDEX idx_balance ON benchmark_records(balance)`,
		`CREATE INDEX idx_created_at ON benchmark_records(created_at)`,
		`CREATE INDEX idx_active ON benchmark_records(is_active)`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

func (p *PostgresDB) TruncateTable() error {
	_, err := p.db.Exec("TRUNCATE TABLE benchmark_records RESTART IDENTITY CASCADE")
	return err
}

func (p *PostgresDB) GetStats() (map[string]any, error) {
	stats := make(map[string]any)

	var count int64
	err := p.db.QueryRow("SELECT COUNT(*) FROM benchmark_records").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["row_count"] = count

	var tableSize string
	err = p.db.QueryRow(`
		SELECT pg_size_pretty(pg_total_relation_size('benchmark_records'))
	`).Scan(&tableSize)
	if err != nil {
		return nil, err
	}
	stats["table_size"] = tableSize

	rows, err := p.db.Query(`
		SELECT indexname, pg_size_pretty(pg_relation_size(indexrelid))
		FROM pg_stat_user_indexes
		WHERE tablename = 'benchmark_records'
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", err)
		}
	}()

	indexes := make(map[string]string)
	for rows.Next() {
		var name, size string
		if err := rows.Scan(&name, &size); err != nil {
			return nil, err
		}
		indexes[name] = size
	}
	stats["indexes"] = indexes

	return stats, nil
}
