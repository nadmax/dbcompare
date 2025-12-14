package benchmarks

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/nadmax/dbcompare/internal/config"
	"github.com/nadmax/dbcompare/internal/database"
	"github.com/nadmax/dbcompare/internal/generator"
	"github.com/nadmax/dbcompare/internal/models"
)

type PostgresBenchmark struct {
	BaseBenchmark
	db  *database.PostgresDB
	gen *generator.Generator
}

func NewPostgresBenchmark(db *database.PostgresDB, cfg *config.Config) *PostgresBenchmark {
	return &PostgresBenchmark{
		BaseBenchmark: BaseBenchmark{
			name:   "PostgreSQL",
			config: cfg,
		},
		db:  db,
		gen: generator.NewDefault(),
	}
}

func (p *PostgresBenchmark) Setup() error {
	fmt.Println("Setting up PostgreSQL schema...")
	return p.db.CreateSchema()
}

func (p *PostgresBenchmark) Teardown() error {
	return p.db.Close()
}

func (p *PostgresBenchmark) Run() ([]models.BenchmarkResult, error) {
	results := make([]models.BenchmarkResult, 0)

	if result, err := p.bulkInsert(); err != nil {
		fmt.Printf("⚠ Bulk Insert failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := p.sequentialRead(); err != nil {
		fmt.Printf("⚠ Sequential Read failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := p.randomRead(); err != nil {
		fmt.Printf("⚠ Random Read failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := p.indexedQuery(); err != nil {
		fmt.Printf("⚠ Indexed Query failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := p.updateOperations(); err != nil {
		fmt.Printf("⚠ Update failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := p.complexQuery(); err != nil {
		fmt.Printf("⚠ Complex Query failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := p.concurrentReads(); err != nil {
		fmt.Printf("⚠ Concurrent Reads failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := p.concurrentWrites(); err != nil {
		fmt.Printf("⚠ Concurrent Writes failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := p.transactionPerformance(); err != nil {
		fmt.Printf("⚠ Transaction Performance failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	return results, nil
}

func (p *PostgresBenchmark) bulkInsert() (*models.BenchmarkResult, error) {
	result := models.NewBenchmarkResult("Bulk Insert", p.name, p.config.Benchmark.RecordCount)

	ctx := context.Background()
	tx, err := p.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			fmt.Printf("Warning: failed to rollback transition: %v\n", err)
		}
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO benchmark_records (name, email, age, balance, created_at, description, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			fmt.Printf("Warning: failed to close statement: %v\n", err)
		}
	}()

	errorCount := 0
	batchSize := p.config.Benchmark.BatchSize

	for i := 0; i < p.config.Benchmark.RecordCount; i++ {
		record := p.gen.GenerateRecord(i + 1)
		_, err := stmt.Exec(
			record.Name,
			record.Email,
			record.Age,
			record.Balance,
			record.CreatedAt,
			record.Description,
			record.IsActive,
		)
		if err != nil {
			errorCount++
		}

		p.logProgress("Bulk Insert", i+1, p.config.Benchmark.RecordCount)

		if (i+1)%batchSize == 0 {
			if err := tx.Commit(); err != nil {
				return nil, err
			}
			tx, err = p.db.DB().BeginTx(ctx, nil)
			if err != nil {
				return nil, err
			}
			stmt, err = tx.Prepare(`
				INSERT INTO benchmark_records (name, email, age, balance, created_at, description, is_active)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	result.Complete(errorCount)
	p.logComplete("Bulk Insert", result)
	return result, nil
}

func (p *PostgresBenchmark) sequentialRead() (*models.BenchmarkResult, error) {
	result := models.NewBenchmarkResult("Sequential Read", p.name, p.config.Benchmark.RecordCount)

	rows, err := p.db.DB().Query("SELECT * FROM benchmark_records LIMIT $1", p.config.Benchmark.RecordCount)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", err)
		}
	}()

	count := 0
	errorCount := 0
	for rows.Next() {
		var r models.TestRecord
		err := rows.Scan(&r.ID, &r.Name, &r.Email, &r.Age, &r.Balance, &r.CreatedAt, &r.Description, &r.IsActive)
		if err != nil {
			errorCount++
		}
		count++
		p.logProgress("Sequential Read", count, p.config.Benchmark.RecordCount)
	}

	result.Complete(errorCount)
	p.logComplete("Sequential Read", result)
	return result, nil
}

func (p *PostgresBenchmark) randomRead() (*models.BenchmarkResult, error) {
	count := p.config.Benchmark.RandomReads
	result := models.NewBenchmarkResult("Random Read", p.name, count)

	errorCount := 0
	for i := range count {
		id := p.gen.GenerateRandomID(p.config.Benchmark.RecordCount)
		var r models.TestRecord
		err := p.db.DB().QueryRow("SELECT * FROM benchmark_records WHERE id = $1", id).
			Scan(&r.ID, &r.Name, &r.Email, &r.Age, &r.Balance, &r.CreatedAt, &r.Description, &r.IsActive)
		if err != nil && err != sql.ErrNoRows {
			errorCount++
		}
		p.logProgress("Random Read", i+1, count)
	}

	result.Complete(errorCount)
	p.logComplete("Random Read", result)
	return result, nil
}

func (p *PostgresBenchmark) indexedQuery() (*models.BenchmarkResult, error) {
	result := models.NewBenchmarkResult("Indexed Query", p.name, 1000)

	errorCount := 0
	for i := range 1000 {
		age := 20 + (i % 50)
		rows, err := p.db.DB().Query("SELECT * FROM benchmark_records WHERE age = $1 LIMIT 10", age)
		if err != nil {
			errorCount++
		} else {
			if err := rows.Close(); err != nil {
				fmt.Printf("Warning: failed to close rows: %v", err)
			}
		}
		p.logProgress("Indexed Query", i+1, 1000)
	}

	result.Complete(errorCount)
	p.logComplete("Indexed Query", result)
	return result, nil
}

func (p *PostgresBenchmark) updateOperations() (*models.BenchmarkResult, error) {
	count := p.config.Benchmark.Updates
	result := models.NewBenchmarkResult("Update Operations", p.name, count)

	errorCount := 0
	for i := range count {
		id := p.gen.GenerateRandomID(p.config.Benchmark.RecordCount)
		newBalance := p.gen.GenerateUpdateValue("balance").(float64)
		_, err := p.db.DB().Exec("UPDATE benchmark_records SET balance = $1 WHERE id = $2", newBalance, id)
		if err != nil {
			errorCount++
		}
		p.logProgress("Update", i+1, count)
	}

	result.Complete(errorCount)
	p.logComplete("Update Operations", result)
	return result, nil
}

func (p *PostgresBenchmark) complexQuery() (*models.BenchmarkResult, error) {
	result := models.NewBenchmarkResult("Complex Query", p.name, 1)

	rows, err := p.db.DB().Query(`
		SELECT 
			age,
			COUNT(*) as user_count,
			AVG(balance) as avg_balance,
			MAX(balance) as max_balance,
			MIN(balance) as min_balance
		FROM benchmark_records
		WHERE is_active = true AND age > 25
		GROUP BY age
		HAVING COUNT(*) > 5
		ORDER BY avg_balance DESC
		LIMIT 50
	`)

	errorCount := 0
	if err != nil {
		errorCount = 1
	} else {
		defer func() {
			if err := rows.Close(); err != nil {
				fmt.Printf("Warning: failed to close rows: %v", err)
			}
		}()
		for rows.Next() {
			var age, count int
			var avg, max, min float64
			if err := rows.Scan(&age, &count, &avg, &max, &min); err != nil {
				errorCount++
			}
		}
	}

	result.Complete(errorCount)
	p.logComplete("Complex Query", result)
	return result, nil
}

func (p *PostgresBenchmark) concurrentReads() (*models.BenchmarkResult, error) {
	goroutines := p.config.Benchmark.ConcurrentGoroutines
	readsPerGoroutine := 1000
	totalReads := goroutines * readsPerGoroutine

	result := models.NewBenchmarkResult("Concurrent Reads", p.name, totalReads)

	var wg sync.WaitGroup
	errorChan := make(chan error, goroutines)
	errorCount := 0

	for i := range goroutines {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for range readsPerGoroutine {
				id := p.gen.GenerateRandomID(p.config.Benchmark.RecordCount)
				var r models.TestRecord
				err := p.db.DB().QueryRow("SELECT * FROM benchmark_records WHERE id = $1", id).
					Scan(&r.ID, &r.Name, &r.Email, &r.Age, &r.Balance, &r.CreatedAt, &r.Description, &r.IsActive)
				if err != nil && err != sql.ErrNoRows {
					errorChan <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	for range errorChan {
		errorCount++
	}

	result.Complete(errorCount)
	p.logComplete("Concurrent Reads", result)
	return result, nil
}

func (p *PostgresBenchmark) concurrentWrites() (*models.BenchmarkResult, error) {
	goroutines := p.config.Benchmark.ConcurrentGoroutines
	writesPerGoroutine := 100
	totalWrites := goroutines * writesPerGoroutine

	result := models.NewBenchmarkResult("Concurrent Writes", p.name, totalWrites)

	var wg sync.WaitGroup
	errorChan := make(chan error, goroutines)
	errorCount := 0

	for i := range goroutines {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := range writesPerGoroutine {
				record := p.gen.GenerateRecord(200000 + routineID*writesPerGoroutine + j)
				_, err := p.db.DB().Exec(`
					INSERT INTO benchmark_records (name, email, age, balance, created_at, description, is_active)
					VALUES ($1, $2, $3, $4, $5, $6, $7)
				`, record.Name, record.Email, record.Age, record.Balance, record.CreatedAt, record.Description, record.IsActive)
				if err != nil {
					errorChan <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	for range errorChan {
		errorCount++
	}

	result.Complete(errorCount)
	p.logComplete("Concurrent Writes", result)
	return result, nil
}

func (p *PostgresBenchmark) transactionPerformance() (*models.BenchmarkResult, error) {
	count := p.config.Benchmark.Transactions
	result := models.NewBenchmarkResult("Transaction Performance", p.name, count)

	errorCount := 0
	for i := range count {
		tx, err := p.db.DB().Begin()
		if err != nil {
			errorCount++
			continue
		}

		id1 := p.gen.GenerateRandomID(p.config.Benchmark.RecordCount)
		id2 := p.gen.GenerateRandomID(p.config.Benchmark.RecordCount)

		_, err = tx.Exec("UPDATE benchmark_records SET balance = balance - 10 WHERE id = $1", id1)
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				fmt.Printf("Warning: failed to rollback transition: %v\n", txErr)
			}
			errorCount++
			continue
		}

		_, err = tx.Exec("UPDATE benchmark_records SET balance = balance + 10 WHERE id = $1", id2)
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				fmt.Printf("Warning: failed to rollback transition: %v\n", txErr)
			}
			errorCount++
			continue
		}

		if err := tx.Commit(); err != nil {
			errorCount++
		}

		p.logProgress("Transactions", i+1, count)
	}

	result.Complete(errorCount)
	p.logComplete("Transaction Performance", result)
	return result, nil
}
