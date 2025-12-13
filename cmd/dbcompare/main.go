package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nadmax/dbcompare/internal/benchmarks"
	"github.com/nadmax/dbcompare/internal/config"
	"github.com/nadmax/dbcompare/internal/reporter"
)

func main() {
	configPath := flag.String("config", "configs/config.yml", "Path to configuration file")
	dbFilter := flag.String("db", "", "Run only specific database (postgres, oracle, surrealdb)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║          DBCompare: Benchmark Suite                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	reporters := createReporters(cfg)
	runner := benchmarks.NewRunner(cfg)
	results, err := runner.Run(*dbFilter)
	if err != nil {
		log.Fatalf("Benchmark execution failed: %v", err)
	}

	for _, rep := range reporters {
		if err := rep.Generate(results); err != nil {
			log.Printf("Failed to generate %s report: %v", rep.Name(), err)
		}
	}

	fmt.Println("\n✓ Benchmarks completed successfully")
}

func createReporters(cfg *config.Config) []reporter.Reporter {
	reporters := make([]reporter.Reporter, 0)
	for _, format := range cfg.Output.Format {
		switch format {
		case "console":
			reporters = append(reporters, reporter.NewConsoleReporter())
		case "csv":
			filename := fmt.Sprintf("%s/%s_%s.csv",
				cfg.Output.Directory,
				cfg.Output.FilenamePrefix,
				time.Now().Format("20060102_150405"))
			reporters = append(reporters, reporter.NewCSVReporter(filename))
		case "json":
			filename := fmt.Sprintf("%s/%s_%s.json",
				cfg.Output.Directory,
				cfg.Output.FilenamePrefix,
				time.Now().Format("20060102_150405"))
			reporters = append(reporters, reporter.NewJSONReporter(filename))
		}
	}

	if err := os.MkdirAll(cfg.Output.Directory, 0755); err != nil {
		log.Printf("Warning: Could not create results directory: %v", err)
	}

	return reporters
}
