# MongoDB Benchmark Tool

This is a Go-based event generation and query benchmark tool for randomly writing event data to MongoDB databases and providing performance benchmarks. The program randomly generates 0-3 event records every minute and concurrently inserts them into the database, while offering various query functionalities for performance testing.

## Features

- Periodic (every minute) generation of random events, with customizable intervals
- Random generation of different types, severity levels, and statuses of events
- Events contain rich metadata, reflecting real-world scenarios
- Uses MongoDB native time types for timestamps, facilitating date operations and queries
- Structured severity design for easy sorting and filtering
- Concurrent writing to improve performance
- Built-in various query tests with performance data analysis (execution time and memory usage)
- Includes anti-pattern vs. optimized solution comparison tests, demonstrating MongoDB query optimization techniques
- Uses compound indexes to optimize query performance

## Project Structure

The project adopts a modular design with the following structure:

```
├── cmd
│   ├── generate     # Event generation subcommand
│   └── run          # Benchmark test subcommand
├── internal
│   ├── database     # Database connection management
│   ├── models       # Data model definitions
│   └── utils        # Utility functions and query implementations
├── main.go          # Main program entry
├── go.mod           # Go module definition
└── go.sum           # Dependency verification
```

## Event Structure

Each event contains the following fields:

- **timestamp**: Time of the event occurrence (native MongoDB time type)
- **eventType**: Event type (such as "System Warning", "Security Incident", etc.)
- **description**: Event description
- **severity**: Event severity level, including numerical level, label, and color
- **sourceSystem**: Source system
- **sourceIP**: Source IP address
- **affectedComponents**: List of affected components
- **recommendation**: Handling recommendation
- **status**: Event status
- **assignedTo**: Team assigned to (if applicable)
- **resolvedAt**: Resolution time (if applicable, native MongoDB time type)
- **resolutionNotes**: Resolution notes (if applicable)
- **tags**: List of tags
- **metadata**: Additional metadata related to the event type

## Prerequisites

- Go 1.21 or higher
- Running MongoDB instance (default at localhost:27017)

## Installation and Building

1. Clone the repository:

```bash
git clone https://github.com/yourusername/mongo-bench.git
cd mongo-bench
```

2. Fetch dependencies:

```bash
go mod tidy
```

3. Build the program:

```bash
go build -o mongo-bench
```

## Usage

The program provides two main subcommands, which can be executed directly using `go run` or by building and running the binary:

### 1. Generate Events

```bash
# Using go run
go run main.go generate [flags]

# Or using the built binary
./mongo-bench generate [flags]
```

Optional parameters:
- `--uri`: MongoDB connection URI (default: "mongodb://localhost:27017")
- `--username`: MongoDB username (default: "admin")
- `--password`: MongoDB password (default: "password")
- `--database`: MongoDB database name (default: "eventstore")
- `--interval`: Event generation interval in seconds (default: 60)

### 2. Run Benchmarks

```bash
# Using go run
go run main.go run [flags]

# Or using the built binary
./mongo-bench run [flags]
```

Optional parameters:
- `--uri`: MongoDB connection URI (default: "mongodb://localhost:27017")
- `--username`: MongoDB username (default: "admin")
- `--password`: MongoDB password (default: "password")
- `--database`: MongoDB database name (default: "eventstore")
- `--test`: Specify the test name to run, leave empty to run all tests

Available tests include:
- Query recent events
- Count by severity
- Query high severity events
- Query events within a time range
- Query with index hint
- Various anti-pattern and optimization solution comparison tests

## Example Usage

1. Generate events with default configuration:

```bash
go run main.go generate
```

2. Custom generation interval:

```bash
go run main.go generate --interval=30
```

3. Run all query tests:

```bash
go run main.go run
```

4. Run a specific query test:

```bash
go run main.go run --test="Query High Severity Events"
```

## MongoDB Query Optimization

This project demonstrates three common MongoDB query optimization techniques:

### 1. Projection Optimization

Reduce network transfer and memory usage by retrieving only necessary fields.

- **Anti-Pattern**: Retrieve entire documents but use only a few fields
- **Optimization**: Use projection to retrieve only necessary fields

### 2. Aggregation Pipeline Optimization

Improve aggregation efficiency by optimizing the pipeline order.

- **Anti-Pattern**: Aggregate first, then filter
- **Optimization**: Filter first, then aggregate, reducing the number of documents to process

### 3. Index Optimization

Create indexes for commonly queried and sorted fields to accelerate query operations.

- **Anti-Pattern**: Sort on unindexed fields
- **Optimization**: Create indexes for sort fields and use index hints ($hint)

> **Note**: In performance comparison tests, index creation is a one-time operation that is not included in query execution time. This is because in real applications, indexes are typically created in advance, not during each query.

### Reliability Improvements for Performance Measurements

To ensure the stability and reliability of test results, this tool implements the following optimizations:

1. **Test Warm-up**: Performs warm-up queries before formal measurement to reduce the cache effect of the first query
2. **Repeated Execution**: Executes each test multiple times and takes the average value to improve statistical significance
3. **Memory Calculation Optimization**: Improves memory usage statistics algorithms to avoid unstable results caused by extremely small values
4. **Detailed Logging**: Provides more detailed logs of the testing process and results for analyzing performance bottlenecks
5. **Cache Flushing**: Actively flushes MongoDB cache before and between each test phase, including:
   - Clearing the query plan cache (`planCacheClear` command)
   - Executing multiple random queries with different patterns to flush the data cache
   - Forcing garbage collection to release memory resources
   These measures greatly reduce the interference of MongoDB's built-in caching mechanisms on test results

## License

MIT 