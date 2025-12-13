package benchmarks

import (
	"fmt"
	"sync"

	"github.com/nadmax/dbcompare/internal/config"
	"github.com/nadmax/dbcompare/internal/database"
	"github.com/nadmax/dbcompare/internal/generator"
	internalmodels "github.com/nadmax/dbcompare/internal/models"
	"github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type SurrealDBBenchmark struct {
	BaseBenchmark
	db  *database.SurrealDB
	gen *generator.Generator
}

type SurrealRecord struct {
	ID          *models.RecordID `json:"id,omitempty"`
	Name        string           `json:"name"`
	Email       string           `json:"email"`
	Age         int              `json:"age"`
	Balance     float64          `json:"balance"`
	CreatedAt   string           `json:"created_at"`
	Description string           `json:"description"`
	IsActive    bool             `json:"is_active"`
}

func NewSurrealDBBenchmark(db *database.SurrealDB, cfg *config.Config) *SurrealDBBenchmark {
	return &SurrealDBBenchmark{
		BaseBenchmark: BaseBenchmark{
			name:   "SurrealDB",
			config: cfg,
		},
		db:  db,
		gen: generator.NewDefault(),
	}
}

func (s *SurrealDBBenchmark) Setup() error {
	fmt.Println("Setting up SurrealDB schema...")
	return s.db.CreateSchema()
}

func (s *SurrealDBBenchmark) Teardown() error {
	return s.db.Close()
}

func (s *SurrealDBBenchmark) Run() ([]internalmodels.BenchmarkResult, error) {
	results := make([]internalmodels.BenchmarkResult, 0)

	if result, err := s.bulkInsert(); err != nil {
		fmt.Printf("⚠ Bulk Insert failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := s.sequentialRead(); err != nil {
		fmt.Printf("⚠ Sequential Read failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := s.randomRead(); err != nil {
		fmt.Printf("⚠ Random Read failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := s.updateOperations(); err != nil {
		fmt.Printf("⚠ Update failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := s.concurrentReads(); err != nil {
		fmt.Printf("⚠ Concurrent Reads failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	if result, err := s.concurrentWrites(); err != nil {
		fmt.Printf("⚠ Concurrent Writes failed: %v\n", err)
	} else {
		results = append(results, *result)
	}

	return results, nil
}

func (s *SurrealDBBenchmark) bulkInsert() (*internalmodels.BenchmarkResult, error) {
	result := internalmodels.NewBenchmarkResult("Bulk Insert", s.name, s.config.Benchmark.RecordCount)

	errorCount := 0
	batchSize := s.config.Benchmark.BatchSize
	ctx := s.db.Context()

	for i := 0; i < s.config.Benchmark.RecordCount; i += batchSize {
		end := i + batchSize
		if end > s.config.Benchmark.RecordCount {
			end = s.config.Benchmark.RecordCount
		}

		for j := i; j < end; j++ {
			record := s.gen.GenerateRecord(j + 1)

			surrealRec := SurrealRecord{
				Name:        record.Name,
				Email:       record.Email,
				Age:         record.Age,
				Balance:     record.Balance,
				CreatedAt:   record.CreatedAt.Format("2006-01-02T15:04:05Z"),
				Description: record.Description,
				IsActive:    record.IsActive,
			}

			_, err := surrealdb.Create[SurrealRecord](ctx, s.db.DB(), models.Table("test_records"), surrealRec)
			if err != nil {
				errorCount++
			}
		}

		s.logProgress("Bulk Insert", end, s.config.Benchmark.RecordCount)
	}

	result.Complete(errorCount)
	s.logComplete("Bulk Insert", result)
	return result, nil
}

func (s *SurrealDBBenchmark) sequentialRead() (*internalmodels.BenchmarkResult, error) {
	result := internalmodels.NewBenchmarkResult("Sequential Read", s.name, s.config.Benchmark.RecordCount)
	ctx := s.db.Context()
	records, err := surrealdb.Select[[]SurrealRecord](ctx, s.db.DB(), models.Table("test_records"))
	errorCount := 0

	if err != nil {
		errorCount = 1
	} else {
		_ = len(*records)
	}

	result.Complete(errorCount)
	s.logComplete("Sequential Read", result)
	return result, nil
}

func (s *SurrealDBBenchmark) randomRead() (*internalmodels.BenchmarkResult, error) {
	count := s.config.Benchmark.RandomReads
	result := internalmodels.NewBenchmarkResult("Random Read", s.name, count)
	ctx := s.db.Context()

	errorCount := 0

	allRecords, err := surrealdb.Select[[]SurrealRecord](ctx, s.db.DB(), models.Table("test_records"))
	if err != nil || len(*allRecords) == 0 {
		return result, fmt.Errorf("failed to get records for random read: %w", err)
	}

	for i := 0; i < count && i < len(*allRecords); i++ {
		idx := s.gen.GenerateRandomID(len(*allRecords)) - 1
		if idx >= 0 && idx < len(*allRecords) && (*allRecords)[idx].ID != nil {
			_, err := surrealdb.Select[SurrealRecord](ctx, s.db.DB(), *(*allRecords)[idx].ID)
			if err != nil {
				errorCount++
			}
		}
		s.logProgress("Random Read", i+1, count)
	}

	result.Complete(errorCount)
	s.logComplete("Random Read", result)
	return result, nil
}

func (s *SurrealDBBenchmark) updateOperations() (*internalmodels.BenchmarkResult, error) {
	count := s.config.Benchmark.Updates
	result := internalmodels.NewBenchmarkResult("Update Operations", s.name, count)
	ctx := s.db.Context()

	errorCount := 0

	allRecords, err := surrealdb.Select[[]SurrealRecord](ctx, s.db.DB(), models.Table("test_records"))
	if err != nil || len(*allRecords) == 0 {
		return result, fmt.Errorf("failed to get records for update: %w", err)
	}

	for i := 0; i < count && i < len(*allRecords); i++ {
		idx := s.gen.GenerateRandomID(len(*allRecords)) - 1
		if idx >= 0 && idx < len(*allRecords) && (*allRecords)[idx].ID != nil {
			newBalance := s.gen.GenerateUpdateValue("balance").(float64)

			updateData := map[string]any{
				"balance": newBalance,
			}

			_, err := surrealdb.Update[SurrealRecord](ctx, s.db.DB(), *(*allRecords)[idx].ID, updateData)
			if err != nil {
				errorCount++
			}
		}
		s.logProgress("Update", i+1, count)
	}

	result.Complete(errorCount)
	s.logComplete("Update Operations", result)
	return result, nil
}

func (s *SurrealDBBenchmark) concurrentReads() (*internalmodels.BenchmarkResult, error) {
	goroutines := s.config.Benchmark.ConcurrentGoroutines
	readsPerGoroutine := 100 // Reduced for SurrealDB
	totalReads := goroutines * readsPerGoroutine

	result := internalmodels.NewBenchmarkResult("Concurrent Reads", s.name, totalReads)
	ctx := s.db.Context()

	// Get all records first
	allRecords, err := surrealdb.Select[[]SurrealRecord](ctx, s.db.DB(), models.Table("test_records"))
	if err != nil || len(*allRecords) == 0 {
		result.Complete(1)
		return result, fmt.Errorf("failed to get records for concurrent read: %w", err)
	}

	var wg sync.WaitGroup
	errorChan := make(chan error, goroutines)
	errorCount := 0

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < readsPerGoroutine && j < len(*allRecords); j++ {
				idx := s.gen.GenerateRandomID(len(*allRecords)) - 1
				if idx >= 0 && idx < len(*allRecords) && (*allRecords)[idx].ID != nil {
					_, readErr := surrealdb.Select[SurrealRecord](ctx, s.db.DB(), *(*allRecords)[idx].ID)
					if readErr != nil {
						errorChan <- readErr
					}
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
	s.logComplete("Concurrent Reads", result)
	return result, nil
}

func (s *SurrealDBBenchmark) concurrentWrites() (*internalmodels.BenchmarkResult, error) {
	goroutines := s.config.Benchmark.ConcurrentGoroutines
	writesPerGoroutine := 50 // Reduced for SurrealDB
	totalWrites := goroutines * writesPerGoroutine

	result := internalmodels.NewBenchmarkResult("Concurrent Writes", s.name, totalWrites)
	ctx := s.db.Context()

	var wg sync.WaitGroup
	errorChan := make(chan error, goroutines)
	errorCount := 0

	for i := range goroutines {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := range writesPerGoroutine {
				record := s.gen.GenerateRecord(200000 + routineID*writesPerGoroutine + j)

				surrealRec := SurrealRecord{
					Name:        record.Name,
					Email:       record.Email,
					Age:         record.Age,
					Balance:     record.Balance,
					CreatedAt:   record.CreatedAt.Format("2006-01-02T15:04:05Z"),
					Description: record.Description,
					IsActive:    record.IsActive,
				}

				_, writeErr := surrealdb.Create[SurrealRecord](ctx, s.db.DB(), models.Table("test_records"), surrealRec)
				if writeErr != nil {
					errorChan <- writeErr
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
	s.logComplete("Concurrent Writes", result)
	return result, nil
}
