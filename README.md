[CI](https://github.com/nadmax/dbcompare/actions/workflows/ci.yml/badge.svg)

# DBCompare

A comprehensive Go benchmarking suite comparing PostgreSQL and SurrealDB performance across various operations.

## Quick Start

### Option 1: Docker (Recommended - Everything Containerized)

```sh
# 1. Clone the repository
git clone https://github.com/nadmax/dbcompare.git
cd dbcompare

# 2. Create Docker config
cp configs/config.docker.yml configs/config.yml

# 3. Run complete comparison suite
make docker-run
# Results will be in ./results/
```

### Option 2: Local Development with Docker Databases

```sh
# 1. Clone and setup
git clone https://github.com/nadmax/dbcompare.git
cd dbcompare
go mod download

# 2. Start databases in Docker
make docker-up-db

# 3. Configure
cp configs/config.example.yml configs/config.yml
# Edit configs/config.yml with your database connections

# 4. Run comparisons locally
make run

# Or run specific database
make run-postgres
make run-surrealdb
```

### Option 3: Fully Local (No Docker)

```sh
# 1. Install PostgreSQL, and SurrealDB locally
# 2. Configure connection strings in configs/config.yml
# 3. Run: make run
```

## Configuration

Example `configs/config.yml`:

```yaml
databases:
  postgres:
    enabled: true
    host: localhost
    port: 5432
    user: postgres
    password: postgres
    database: postgres
    sslmode: disable
    max_connections: 25

  surrealdb:
    enabled: true
    url: "ws://localhost:8000/rpc"
    namespace: test
    database: test
    user: root
    password: root

comparison:
  record_count: 100000
  batch_size: 1000
  random_reads: 10000
  updates: 10000
  transactions: 1000
  concurrent_goroutines: 10

output:
  format: ["console", "csv", "json"]
  directory: "./results"
  filename_prefix: "dbcompare"
```

## Benchmarks

The suite runs the following benchmarks:

1. **Bulk Insert** - Large batch insertions
2. **Sequential Read** - Full table scans
3. **Random Read** - Point lookups by ID
4. **Indexed Read** - Queries using indexed columns
5. **Update Operations** - Single record updates
6. **Bulk Update** - Batch update operations
7. **Complex Queries** - Aggregations, JOINs, GROUP BY
8. **Concurrent Writes** - Multi-threaded inserts
9. **Concurrent Reads** - Multi-threaded queries
10. **Transaction Performance** - ACID operations
11. **Full-Text Search** - Text search capabilities (if supported)
12. **JSON Operations** - JSON field queries (if supported)

## Results

Results are saved in the `results/` directory in multiple formats:

- **Console**: Real-time output with progress
- **CSV**: `results/dbcompare_YYYYMMDD_HHMMSS.csv`
- **JSON**: `results/dbcompare_YYYYMMDD_HHMMSS.json`

Example output:

## Development

### Building

```sh
make build
```

### Clean

```sh
make clean
```

## Docker Support

The project includes full Docker support with multiple options:

### Run Everything in Docker

```sh
# Build and run complete comparison suite
make docker-run
# Stop everything
make docker-down
```

### Docker Commands Reference

```sh
make docker-build          # Build Docker image
make docker-up             # Start all services including app
make docker-down           # Stop all containers
make docker-clean          # Remove containers, volumes, and images
make docker-logs           # View all logs
```

## Requirements

- Go 1.25.3
- PostgreSQL 18
- SurrealDB v2

## License

MIT License - see LICENSE file for details
