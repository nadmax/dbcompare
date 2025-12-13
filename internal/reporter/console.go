package reporter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nadmax/dbcompare/internal/models"
)

type ConsoleReporter struct{}

func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

func (c *ConsoleReporter) Name() string {
	return "Console"
}

func (c *ConsoleReporter) Generate(suite *models.BenchmarkSuite) error {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("BENCHMARK RESULTS SUMMARY")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("\nTotal Duration: %v\n", suite.Duration)
	fmt.Printf("Start Time: %s\n", suite.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("End Time: %s\n\n", suite.EndTime.Format("2006-01-02 15:04:05"))

	dbResults := make(map[string][]models.BenchmarkResult)
	for _, result := range suite.Results {
		dbResults[result.Database] = append(dbResults[result.Database], result)
	}

	for db, results := range dbResults {
		c.printDatabaseResults(db, results)
	}

	c.printComparisonTable(suite.Results)

	c.printPerformanceSummary(suite.Results)

	fmt.Println(strings.Repeat("=", 100))

	return nil
}

func (c *ConsoleReporter) printDatabaseResults(database string, results []models.BenchmarkResult) {
	fmt.Printf("\n┌─ %s %s\n", database, strings.Repeat("─", 90-len(database)))
	fmt.Printf("│\n")
	fmt.Printf("│ %-30s %12s %15s %12s %10s\n", "Operation", "Duration", "Throughput", "Records", "Errors")
	fmt.Printf("│ %s\n", strings.Repeat("─", 95))

	for _, result := range results {
		errorPercent := result.ErrorRate * 100
		errorIndicator := "✓"
		if errorPercent > 0 {
			errorIndicator = "⚠"
		}

		fmt.Printf("│ %-30s %12v %12.0f/s %12d %s %.2f%%\n",
			result.Operation,
			result.Duration,
			result.Throughput,
			result.RecordsCount,
			errorIndicator,
			errorPercent,
		)
	}
	fmt.Printf("└%s\n", strings.Repeat("─", 97))
}

func (c *ConsoleReporter) printComparisonTable(results []models.BenchmarkResult) {
	opResults := make(map[string][]models.BenchmarkResult)
	for _, result := range results {
		opResults[result.Operation] = append(opResults[result.Operation], result)
	}

	fmt.Println("\n┌─ OPERATION COMPARISON")
	fmt.Println("│")

	for operation, opRes := range opResults {
		if len(opRes) <= 1 {
			continue
		}

		fmt.Printf("│ %s\n", operation)
		fmt.Printf("│ %s\n", strings.Repeat("─", 95))

		sort.Slice(opRes, func(i, j int) bool {
			return opRes[i].Throughput > opRes[j].Throughput
		})

		fastest := opRes[0].Throughput
		for i, result := range opRes {
			rank := fmt.Sprintf("#%d", i+1)
			percentDiff := ((result.Throughput - fastest) / fastest) * 100

			indicator := "🥇"
			if i == 1 {
				indicator = "🥈"
			} else if i == 2 {
				indicator = "🥉"
			} else if i > 2 {
				indicator = "  "
			}

			fmt.Printf("│   %s %-3s %-15s %12.0f/s (%+6.1f%%) %12v\n",
				indicator,
				rank,
				result.Database,
				result.Throughput,
				percentDiff,
				result.Duration,
			)
		}
		fmt.Printf("│\n")
	}
	fmt.Printf("└%s\n", strings.Repeat("─", 97))
}

func (c *ConsoleReporter) printPerformanceSummary(results []models.BenchmarkResult) {
	dbScores := make(map[string]int)

	opResults := make(map[string][]models.BenchmarkResult)
	for _, result := range results {
		opResults[result.Operation] = append(opResults[result.Operation], result)
	}

	for _, opRes := range opResults {
		if len(opRes) <= 1 {
			continue
		}

		sort.Slice(opRes, func(i, j int) bool {
			return opRes[i].Throughput > opRes[j].Throughput
		})

		if len(opRes) > 0 {
			dbScores[opRes[0].Database] += 3
		}
		if len(opRes) > 1 {
			dbScores[opRes[1].Database] += 2
		}
		if len(opRes) > 2 {
			dbScores[opRes[2].Database] += 1
		}
	}

	type scoreEntry struct {
		database string
		score    int
	}
	scores := make([]scoreEntry, 0, len(dbScores))
	for db, score := range dbScores {
		scores = append(scores, scoreEntry{db, score})
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	fmt.Println("\n┌─ OVERALL PERFORMANCE RANKING")
	fmt.Println("│")
	for i, entry := range scores {
		medal := "🏆"
		switch i {
		case 1:
			medal = "🥈"
		case 2:
			medal = "🥉"
		}
		fmt.Printf("│ %s #%d %-15s Score: %d\n", medal, i+1, entry.database, entry.score)
	}
	fmt.Printf("└%s\n", strings.Repeat("─", 97))
}
