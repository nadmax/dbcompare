package database

import (
	"database/sql"
	"fmt"

	_ "github.com/godror/godror"
	"github.com/nadmax/dbcompare/internal/config"
)

type OracleDB struct {
	db     *sql.DB
	config *config.OracleConfig
}

func NewOracleDB(cfg *config.OracleConfig) (*OracleDB, error) {
	db, err := sql.Open("godror", cfg.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxConnections / 5)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &OracleDB{
		db:     db,
		config: cfg,
	}, nil
}

func (o *OracleDB) DB() *sql.DB {
	return o.db
}

func (o *OracleDB) Close() error {
	return o.db.Close()
}

func (o *OracleDB) CreateSchema() error {
	_, _ = o.db.Exec(`
		BEGIN
			EXECUTE IMMEDIATE 'DROP TABLE benchmark_records CASCADE CONSTRAINTS';
		EXCEPTION
			WHEN OTHERS THEN
				IF SQLCODE != -942 THEN
					RAISE;
				END IF;
		END;
	`)

	_, _ = o.db.Exec(`
		BEGIN
			EXECUTE IMMEDIATE 'DROP SEQUENCE benchmark_records_seq';
		EXCEPTION
			WHEN OTHERS THEN
				IF SQLCODE != -2289 THEN
					RAISE;
				END IF;
		END;
	`)

	queries := []string{
		`CREATE SEQUENCE benchmark_records_seq START WITH 1 INCREMENT BY 1`,
		`CREATE TABLE benchmark_records (
			id NUMBER DEFAULT benchmark_records_seq.NEXTVAL PRIMARY KEY,
			name VARCHAR2(100) NOT NULL,
			email VARCHAR2(100) NOT NULL,
			age NUMBER NOT NULL,
			balance NUMBER(10,2) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			description CLOB,
			is_active NUMBER(1) NOT NULL
		)`,
		`CREATE INDEX idx_email ON benchmark_records(email)`,
		`CREATE INDEX idx_age ON benchmark_records(age)`,
		`CREATE INDEX idx_balance ON benchmark_records(balance)`,
		`CREATE INDEX idx_created_at ON benchmark_records(created_at)`,
		`CREATE INDEX idx_active ON benchmark_records(is_active)`,
	}

	for _, query := range queries {
		if _, err := o.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

func (o *OracleDB) TruncateTable() error {
	_, err := o.db.Exec("TRUNCATE TABLE benchmark_records")
	if err != nil {
		return err
	}

	// Reset sequence
	_, err = o.db.Exec("ALTER SEQUENCE benchmark_records_seq RESTART START WITH 1")
	return err
}

func (o *OracleDB) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var count int64
	err := o.db.QueryRow("SELECT COUNT(*) FROM benchmark_records").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["row_count"] = count

	var tableSize float64
	err = o.db.QueryRow(`
		SELECT ROUND(SUM(bytes)/1024/1024, 2)
		FROM user_segments
		WHERE segment_name = 'BENCHMARK_RECORDS'
	`).Scan(&tableSize)
	if err != nil {
		return nil, err
	}
	stats["table_size_mb"] = tableSize

	rows, err := o.db.Query(`
		SELECT segment_name, ROUND(bytes/1024/1024, 2) as size_mb
		FROM user_segments
		WHERE segment_type = 'INDEX'
		AND segment_name LIKE 'IDX_%'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]float64)
	for rows.Next() {
		var name string
		var size float64
		if err := rows.Scan(&name, &size); err != nil {
			return nil, err
		}
		indexes[name] = size
	}
	stats["indexes_mb"] = indexes

	return stats, nil
}
