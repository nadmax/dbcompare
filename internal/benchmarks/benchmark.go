package benchmarks

import (
	"fmt"
	"log"
	"time"

	"github.com/nadmax/dbcompare/internal/config"
	"github.com/nadmax/dbcompare/internal/database"
	"github.com/nadmax/dbcompare/internal/models"
)

type Benchmark interface {
	Name() string
	Setup() error
	Run() ([]models.BenchmarkResult, error)
	Teardown() error
}

type Runner struct {
	config     *config.Config
	benchmarks map[string]Benchmark
}

func NewRunner(cfg *config.Config) *Runner {
	runner := &Runner{
		config:     cfg,
		benchmarks: make(map[string]Benchmark),
	}
	if cfg.Databases.Postgres.Enabled {
		pgDB, err := database.NewPostgresDB(&cfg.Databases.Postgres)
		if err != nil {
			log.Printf("Warning: Failed to initialize PostgreSQL: %v", err)
		} else {
			runner.benchmarks["postgres"] = NewPostgresBenchmark(pgDB, cfg)
		}
	}

	if cfg.Databases.Oracle.Enabled {
		oracleDB, err := database.NewOracleDB(&cfg.Databases.Oracle)
		if err != nil {
			log.Printf("Warning: Failed to initialize Oracle: %v", err)
		} else {
			runner.benchmarks["oracle"] = NewOracleBenchmark(oracleDB, cfg)
		}
	}

	if cfg.Databases.SurrealDB.Enabled {
		surrealDB, err := database.NewSurrealDB(&cfg.Databases.SurrealDB)
		if err != nil {
			log.Printf("Warning: Failed to initialize SurrealDB: %v", err)
		} else {
			runner.benchmarks["surrealdb"] = NewSurrealDBBenchmark(surrealDB, cfg)
		}
	}

	return runner
}

func (r *Runner) Run(filter string) (*models.BenchmarkSuite, error) {
	suite := &models.BenchmarkSuite{
		Results: make([]models.BenchmarkResult, 0),
		Config:  make(map[string]interface{}),
	}
	suite.StartTime = time.Now()

	for name, bench := range r.benchmarks {
		if filter != "" && filter != name {
			continue
		}

		fmt.Printf("\n=== Running %s Benchmarks ===\n", bench.Name())

		if err := bench.Setup(); err != nil {
			log.Printf("Setup failed for %s: %v", name, err)
			continue
		}

		results, err := bench.Run()
		if err != nil {
			log.Printf("Benchmark failed for %s: %v", name, err)
		} else {
			suite.Results = append(suite.Results, results...)
		}

		if err := bench.Teardown(); err != nil {
			log.Printf("Teardown warning for %s: %v", name, err)
		}
	}

	suite.EndTime = time.Now()
	suite.Duration = suite.EndTime.Sub(suite.StartTime)

	return suite, nil
}

type BaseBenchmark struct {
	name   string
	config *config.Config
}

func (b *BaseBenchmark) Name() string {
	return b.name
}

func (b *BaseBenchmark) logProgress(operation string, current, total int) {
	if total > 0 && current%10000 == 0 {
		percent := float64(current) / float64(total) * 100
		fmt.Printf("\r%s: %.1f%% (%d/%d)", operation, percent, current, total)
	}
}

func (b *BaseBenchmark) logComplete(operation string, result *models.BenchmarkResult) {
	fmt.Printf("\r%s: ✓ Duration: %v, Throughput: %.0f ops/s, Errors: %d (%.2f%%)\n",
		operation,
		result.Duration,
		result.Throughput,
		result.ErrorCount,
		result.ErrorRate*100)
}
