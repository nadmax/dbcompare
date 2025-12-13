package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Databases DatabasesConfig `yaml:"databases"`
	Benchmark BenchmarkConfig `yaml:"benchmark"`
	Output    OutputConfig    `yaml:"output"`
}

type DatabasesConfig struct {
	Postgres  PostgresConfig  `yaml:"postgres"`
	Oracle    OracleConfig    `yaml:"oracle"`
	SurrealDB SurrealDBConfig `yaml:"surrealdb"`
}

type PostgresConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	Database       string `yaml:"database"`
	SSLMode        string `yaml:"sslmode"`
	MaxConnections int    `yaml:"max_connections"`
}

type OracleConfig struct {
	Enabled          bool   `yaml:"enabled"`
	ConnectionString string `yaml:"connection_string"`
	MaxConnections   int    `yaml:"max_connections"`
}

type SurrealDBConfig struct {
	Enabled   bool   `yaml:"enabled"`
	URL       string `yaml:"url"`
	Namespace string `yaml:"namespace"`
	Database  string `yaml:"database"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
}

type BenchmarkConfig struct {
	RecordCount          int `yaml:"record_count"`
	BatchSize            int `yaml:"batch_size"`
	RandomReads          int `yaml:"random_reads"`
	Updates              int `yaml:"updates"`
	Transactions         int `yaml:"transactions"`
	ConcurrentGoroutines int `yaml:"concurrent_goroutines"`
}

type OutputConfig struct {
	Format         []string `yaml:"format"`
	Directory      string   `yaml:"directory"`
	FilenamePrefix string   `yaml:"filename_prefix"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.Benchmark.RecordCount == 0 {
		cfg.Benchmark.RecordCount = 100000
	}
	if cfg.Benchmark.BatchSize == 0 {
		cfg.Benchmark.BatchSize = 1000
	}
	if cfg.Output.Directory == "" {
		cfg.Output.Directory = "./results"
	}
	if cfg.Output.FilenamePrefix == "" {
		cfg.Output.FilenamePrefix = "dbcompare"
	}

	return &cfg, nil
}

func (c *PostgresConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}
