package reporter

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/nadmax/dbcompare/internal/models"
)

type CSVReporter struct {
	filename string
}

func NewCSVReporter(filename string) *CSVReporter {
	return &CSVReporter{
		filename: filename,
	}
}

func (c *CSVReporter) Name() string {
	return "CSV"
}

func (c *CSVReporter) Generate(suite *models.BenchmarkSuite) error {
	file, err := os.Create(c.filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Database",
		"Operation",
		"Duration (ms)",
		"Records Count",
		"Throughput (ops/s)",
		"Error Count",
		"Error Rate (%)",
		"Start Time",
		"End Time",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, result := range suite.Results {
		row := []string{
			result.Database,
			result.Operation,
			fmt.Sprintf("%.2f", float64(result.Duration.Milliseconds())),
			fmt.Sprintf("%d", result.RecordsCount),
			fmt.Sprintf("%.2f", result.Throughput),
			fmt.Sprintf("%d", result.ErrorCount),
			fmt.Sprintf("%.4f", result.ErrorRate*100),
			result.StartTime.Format("2006-01-02 15:04:05"),
			result.EndTime.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	fmt.Printf("✓ CSV report saved to: %s\n", c.filename)
	return nil
}
