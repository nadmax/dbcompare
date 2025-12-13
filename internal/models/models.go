package models

import "time"

type TestRecord struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Age         int       `json:"age"`
	Balance     float64   `json:"balance"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
}

type BenchmarkResult struct {
	Operation    string         `json:"operation"`
	Database     string         `json:"database"`
	Duration     time.Duration  `json:"duration"`
	RecordsCount int            `json:"records_count"`
	Throughput   float64        `json:"throughput"`
	ErrorCount   int            `json:"error_count"`
	ErrorRate    float64        `json:"error_rate"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type BenchmarkSuite struct {
	Results   []BenchmarkResult `json:"results"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Duration  time.Duration     `json:"duration"`
	Config    map[string]any    `json:"config"`
}

func NewBenchmarkResult(operation, database string, recordsCount int) *BenchmarkResult {
	return &BenchmarkResult{
		Operation:    operation,
		Database:     database,
		RecordsCount: recordsCount,
		StartTime:    time.Now(),
		Metadata:     make(map[string]any),
	}
}

func (r *BenchmarkResult) Complete(errorCount int) {
	r.EndTime = time.Now()
	r.Duration = r.EndTime.Sub(r.StartTime)
	r.ErrorCount = errorCount

	if r.RecordsCount > 0 {
		r.ErrorRate = float64(errorCount) / float64(r.RecordsCount)
		r.Throughput = float64(r.RecordsCount) / r.Duration.Seconds()
	}
}

func (r *BenchmarkResult) SetMetadata(key string, value any) {
	r.Metadata[key] = value
}
