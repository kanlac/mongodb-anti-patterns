# MongoDB Benchmark Tool

This is a Go-based event generation and query benchmark tool for randomly writing event data to MongoDB databases and providing performance benchmarks. 

## Features

- Connect to your own MongoDB instance
- Built-in various query tests with performance data analysis (execution time and memory usage)
- Add your own tests inside the queries.go file!

## Example Usage

1. Generate events:

```bash
go build .
./mongo-bench generate
```

2. Run all query tests:

```bash
./mongo-bench run
```

3. Run a specific query test:

```bash
./mongo-bench run --test="Query High Severity Events"
```

## Example Result

```
Running all query benchmark tests...
==================================================

Running test: FindAllFieldsAntiPattern
----------------------------------------
Running anti-pattern: Querying all fields when only a few are needed
Found 0 high severity events
----------------------------------------

Running test: FindWithProjectionOptimized
----------------------------------------
Running optimized solution: Using projection to return only needed fields
Found 0 high severity events (projected fields)
----------------------------------------

Running test: AggregateBeforeFilterAntiPattern
----------------------------------------
Running anti-pattern: Performing aggregations before filtering
2025/05/11 11:54:48 Test failed: (BSONObjectTooLarge) BSON size limit hit while building Message. Size: 98217662 (0x5DAAEBE); maxSize: 16793600(16MB)

Running test: FilterBeforeAggregateOptimized
----------------------------------------
Running optimized solution: Filtering data before aggregation
Found 1 severity groups with optimized aggregation
----------------------------------------

Test Results Summary:
==================================================
Profile [FindAllFieldsAntiPattern]:
- Execution time: 3.516625ms
- Memory usage: 0.01 MB
------------------------------
Profile [FindWithProjectionOptimized]:
- Execution time: 1.674125ms
- Memory usage: 0.01 MB
------------------------------
Profile [FilterBeforeAggregateOptimized]:
- Execution time: 903.701417ms
- Memory usage: 0.02 MB
------------------------------
```