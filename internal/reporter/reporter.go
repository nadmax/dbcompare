package reporter

import "github.com/nadmax/dbcompare/internal/models"

type Reporter interface {
	Name() string
	Generate(suite *models.BenchmarkSuite) error
}
