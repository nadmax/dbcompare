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

type OracleBenchmark struct {
	BaseBenchmark
	db  *database.OracleDB
	gen *generator.Generator
}

func NewOracleBenchmark(db *database.OracleDB, cfg *config.Config) *OracleBenchmark {
	return &OracleBenchmark{
		BaseBenchmark: BaseBenchmark{
			name:   "Oracle",
			config: cfg,
		},
		db:  db,
		gen: generator.NewDefault(),
	}
}

func (o *OracleBenchmark) Setup() error {
	fmt.Println("Setting up Oracle schema...")
	return o.db.CreateSchema()
}

func (o *OracleBenchmark) Teardown() error {
	return o.db.Close()
}

func (o *OracleBenchmark) Run() ([]models.BenchmarkResult, error) {
	results := make([]models.BenchmarkResult, 0)

	if result, err := o.bulkInsert(); err != nil {
		fmt.Printf("⚠ Bulk Insert failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := o.sequentialRead(); err != nil {
		fmt.Printf("⚠ Sequential Read failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := o.randomRead(); err != nil {
		fmt.Printf("⚠ Random Read failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := o.indexedQuery(); err != nil {
		fmt.Printf("⚠ Indexed Query failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := o.updateOperations(); err != nil {
		fmt.Printf("⚠ Update failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := o.complexQuery(); err != nil {
		fmt.Printf("⚠ Complex Query failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := o.concurrentReads(); err != nil {
		fmt.Printf("⚠ Concurrent Reads failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := o.concurrentWrites(); err != nil {
		fmt.Printf("⚠ Concurrent Writes failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := o.transactionPerformance(); err != nil {
		fmt.Printf("⚠ Transaction Performance failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	return results, nil
}

func (o *OracleBenchmark) bulkInsert() (*models.BenchmarkResult, error) {
	result := models.NewBenchmarkResult("Bulk Insert", o.name, o.config.Benchmark.RecordCount)

	ctx := context.Background()
	tx, err := o.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO benchmark_records (name, email, age, balance, created_at, description, is_active)
		VALUES (:1, :2, :3, :4, :5, :6, :7)
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	errorCount := 0
	batchSize := o.config.Benchmark.BatchSize

	for i := 0; i < o.config.Benchmark.RecordCount; i++ {
		record := o.gen.GenerateRecord(i + 1)
		isActiveInt := 0
		if record.IsActive {
			isActiveInt = 1
		}

		_, err := stmt.Exec(
			record.Name,
			record.Email,
			record.Age,
			record.Balance,
			record.CreatedAt,
			record.Description,
			isActiveInt,
		)
		if err != nil {
			errorCount++
		}

		o.logProgress("Bulk Insert", i+1, o.config.Benchmark.RecordCount)

		if (i+1)%batchSize == 0 {
			if err := tx.Commit(); err != nil {
				return nil, err
			}
			tx, err = o.db.DB().BeginTx(ctx, nil)
			if err != nil {
				return nil, err
			}
			stmt, err = tx.Prepare(`
				INSERT INTO benchmark_records (name, email, age, balance, created_at, description, is_active)
				VALUES (:1, :2, :3, :4, :5, :6, :7)
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
	o.logComplete("Bulk Insert", result)
	return result, nil
}

func (o *OracleBenchmark) sequentialRead() (*models.BenchmarkResult, error) {
	result := models.NewBenchmarkResult("Sequential Read", o.name, o.config.Benchmark.RecordCount)

	rows, err := o.db.DB().Query(fmt.Sprintf("SELECT * FROM benchmark_records WHERE ROWNUM <= %d", o.config.Benchmark.RecordCount))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	count := 0
	errorCount := 0
	for rows.Next() {
		var r models.TestRecord
		var isActiveInt int
		err := rows.Scan(&r.ID, &r.Name, &r.Email, &r.Age, &r.Balance, &r.CreatedAt, &r.Description, &isActiveInt)
		if err != nil {
			errorCount++
		}
		r.IsActive = isActiveInt == 1
		count++
		o.logProgress("Sequential Read", count, o.config.Benchmark.RecordCount)
	}

	result.Complete(errorCount)
	o.logComplete("Sequential Read", result)
	return result, nil
}

func (o *OracleBenchmark) randomRead() (*models.BenchmarkResult, error) {
	count := o.config.Benchmark.RandomReads
	result := models.NewBenchmarkResult("Random Read", o.name, count)

	errorCount := 0
	for i := 0; i < count; i++ {
		id := o.gen.GenerateRandomID(o.config.Benchmark.RecordCount)
		var r models.TestRecord
		var isActiveInt int
		err := o.db.DB().QueryRow("SELECT * FROM benchmark_records WHERE id = :1", id).
			Scan(&r.ID, &r.Name, &r.Email, &r.Age, &r.Balance, &r.CreatedAt, &r.Description, &isActiveInt)
		if err != nil && err != sql.ErrNoRows {
			errorCount++
		}
		r.IsActive = isActiveInt == 1
		o.logProgress("Random Read", i+1, count)
	}

	result.Complete(errorCount)
	o.logComplete("Random Read", result)
	return result, nil
}

func (o *OracleBenchmark) indexedQuery() (*models.BenchmarkResult, error) {
	result := models.NewBenchmarkResult("Indexed Query", o.name, 1000)

	errorCount := 0
	for i := 0; i < 1000; i++ {
		age := 20 + (i % 50)
		rows, err := o.db.DB().Query("SELECT * FROM benchmark_records WHERE age = :1 AND ROWNUM <= 10", age)
		if err != nil {
			errorCount++
		} else {
			rows.Close()
		}
		o.logProgress("Indexed Query", i+1, 1000)
	}

	result.Complete(errorCount)
	o.logComplete("Indexed Query", result)
	return result, nil
}

func (o *OracleBenchmark) updateOperations() (*models.BenchmarkResult, error) {
	count := o.config.Benchmark.Updates
	result := models.NewBenchmarkResult("Update Operations", o.name, count)

	errorCount := 0
	for i := 0; i < count; i++ {
		id := o.gen.GenerateRandomID(o.config.Benchmark.RecordCount)
		newBalance := o.gen.GenerateUpdateValue("balance").(float64)
		_, err := o.db.DB().Exec("UPDATE benchmark_records SET balance = :1 WHERE id = :2", newBalance, id)
		if err != nil {
			errorCount++
		}
		o.logProgress("Update", i+1, count)
	}

	result.Complete(errorCount)
	o.logComplete("Update Operations", result)
	return result, nil
}

func (o *OracleBenchmark) complexQuery() (*models.BenchmarkResult, error) {
	result := models.NewBenchmarkResult("Complex Query", o.name, 1)

	rows, err := o.db.DB().Query(`
		SELECT 
			age,
			COUNT(*) as user_count,
			AVG(balance) as avg_balance,
			MAX(balance) as max_balance,
			MIN(balance) as min_balance
		FROM benchmark_records
		WHERE is_active = 1 AND age > 25
		GROUP BY age
		HAVING COUNT(*) > 5
		ORDER BY AVG(balance) DESC
		FETCH FIRST 50 ROWS ONLY
	`)

	errorCount := 0
	if err != nil {
		errorCount = 1
	} else {
		defer rows.Close()
		for rows.Next() {
			var age, count int
			var avg, max, min float64
			if err := rows.Scan(&age, &count, &avg, &max, &min); err != nil {
				errorCount++
			}
		}
	}

	result.Complete(errorCount)
	o.logComplete("Complex Query", result)
	return result, nil
}

func (o *OracleBenchmark) concurrentReads() (*models.BenchmarkResult, error) {
	goroutines := o.config.Benchmark.ConcurrentGoroutines
	readsPerGoroutine := 1000
	totalReads := goroutines * readsPerGoroutine

	result := models.NewBenchmarkResult("Concurrent Reads", o.name, totalReads)

	var wg sync.WaitGroup
	errorChan := make(chan error, goroutines)
	errorCount := 0

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < readsPerGoroutine; j++ {
				id := o.gen.GenerateRandomID(o.config.Benchmark.RecordCount)
				var r models.TestRecord
				var isActiveInt int
				err := o.db.DB().QueryRow("SELECT * FROM benchmark_records WHERE id = :1", id).
					Scan(&r.ID, &r.Name, &r.Email, &r.Age, &r.Balance, &r.CreatedAt, &r.Description, &isActiveInt)
				if err != nil && err != sql.ErrNoRows {
					errorChan <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errorChan)

	for range errorChan {
		errorCount++
	}

	result.Complete(errorCount)
	o.logComplete("Concurrent Reads", result)
	return result, nil
}

func (o *OracleBenchmark) concurrentWrites() (*models.BenchmarkResult, error) {
	goroutines := o.config.Benchmark.ConcurrentGoroutines
	writesPerGoroutine := 100
	totalWrites := goroutines * writesPerGoroutine

	result := models.NewBenchmarkResult("Concurrent Writes", o.name, totalWrites)

	var wg sync.WaitGroup
	errorChan := make(chan error, goroutines)
	errorCount := 0

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				record := o.gen.GenerateRecord(200000 + routineID*writesPerGoroutine + j)
				isActiveInt := 0
				if record.IsActive {
					isActiveInt = 1
				}
				_, err := o.db.DB().Exec(`
					INSERT INTO benchmark_records (name, email, age, balance, created_at, description, is_active)
					VALUES (:1, :2, :3, :4, :5, :6, :7)
				`, record.Name, record.Email, record.Age, record.Balance, record.CreatedAt, record.Description, isActiveInt)
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
	o.logComplete("Concurrent Writes", result)
	return result, nil
}

func (o *OracleBenchmark) transactionPerformance() (*models.BenchmarkResult, error) {
	count := o.config.Benchmark.Transactions
	result := models.NewBenchmarkResult("Transaction Performance", o.name, count)

	errorCount := 0
	for i := 0; i < count; i++ {
		tx, err := o.db.DB().Begin()
		if err != nil {
			errorCount++
			continue
		}

		id1 := o.gen.GenerateRandomID(o.config.Benchmark.RecordCount)
		id2 := o.gen.GenerateRandomID(o.config.Benchmark.RecordCount)

		_, err = tx.Exec("UPDATE benchmark_records SET balance = balance - 10 WHERE id = :1", id1)
		if err != nil {
			tx.Rollback()
			errorCount++
			continue
		}

		_, err = tx.Exec("UPDATE benchmark_records SET balance = balance + 10 WHERE id = :1", id2)
		if err != nil {
			tx.Rollback()
			errorCount++
			continue
		}

		if err := tx.Commit(); err != nil {
			errorCount++
		}

		o.logProgress("Transactions", i+1, count)
	}

	result.Complete(errorCount)
	o.logComplete("Transaction Performance", result)
	return result, nil
}
