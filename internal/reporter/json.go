package reporter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nadmax/dbcompare/internal/models"
)

type JSONReporter struct {
	filename string
}

func NewJSONReporter(filename string) *JSONReporter {
	return &JSONReporter{
		filename: filename,
	}
}

func (j *JSONReporter) Name() string {
	return "JSON"
}

func (j *JSONReporter) Generate(suite *models.BenchmarkSuite) error {
	file, err := os.Create(j.filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Warning: failed to close JSON file: %v\n", err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(suite); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	fmt.Printf("✓ JSON report saved to: %s\n", j.filename)
	return nil
}
