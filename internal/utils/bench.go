package utils

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OptimizationPair represents a pair of anti-pattern and its corresponding optimized solution
type OptimizationPair struct {
	AntiPatternName string
	OptimizedName   string
	AntiPatternFunc QueryTestFunc
	OptimizedFunc   QueryTestFunc
	Description     string
}

// GetOptimizationPairs returns all optimization comparison pairs
func GetOptimizationPairs() []OptimizationPair {
	return []OptimizationPair{
		{
			AntiPatternName: "Anti-pattern 1 - Full Document Retrieval",
			OptimizedName:   "Optimization 1 - Using Projection",
			AntiPatternFunc: FindAllFieldsAntiPattern,
			OptimizedFunc:   FindWithProjectionOptimized,
			Description:     "Only retrieve needed fields with projection to reduce network transfer and memory usage",
		},
		{
			AntiPatternName: "Anti-pattern 2 - Aggregate Without Filtering",
			OptimizedName:   "Optimization 2 - Filter Before Aggregation",
			AntiPatternFunc: AggregateBeforeFilterAntiPattern,
			OptimizedFunc:   FilterBeforeAggregateOptimized,
			Description:     "Filter data before aggregation to reduce the number of documents to process",
		},
		{
			AntiPatternName: "Anti-pattern 3 - Sorting Without Index",
			OptimizedName:   "Optimization 3 - Using Index for Sorting",
			AntiPatternFunc: SortWithoutIndexAntiPattern,
			OptimizedFunc:   SortWithIndexOptimized,
			Description:     "Create indexes for sorting fields to speed up sorting operations",
		},
	}
}

// RunOptimizationComparison runs optimization comparison tests
func RunOptimizationComparison(qc *QueryContext) error {
	pairs := GetOptimizationPairs()

	fmt.Println("\n=================== MongoDB Query Optimization Comparison ===================")

	// Test repetitions to increase reliability
	const testRepetitions = 3

	for i, pair := range pairs {
		fmt.Printf("\nOptimization Group %d: %s\n", i+1, pair.Description)
		fmt.Println(strings.Repeat("-", 60))

		// For the third optimization group, pre-create the index, not counted in performance measurement
		if i == 2 { // Third group is index optimization
			fmt.Println("Pre-creating index for fair comparison (this operation time is not included in performance measurement)...")
			// Create index
			indexModel := mongo.IndexModel{
				Keys: bson.D{
					{Key: "sourceSystem", Value: 1},
					{Key: "status", Value: 1},
				},
				Options: options.Index().SetName("sourceSystem_status_idx"),
			}

			_, err := qc.Collection.Indexes().CreateOne(qc.Ctx, indexModel)
			if err != nil {
				fmt.Printf("Note: Failed to create index: %v, continuing test but results may not be accurate\n", err)
			} else {
				fmt.Println("Index created successfully, continuing with performance comparison...")
			}

			// Warm-up query to reduce cache effects
			fmt.Println("Running warm-up queries to reduce cache effects...")
			_ = pair.AntiPatternFunc(qc)
			_ = pair.OptimizedFunc(qc)
			fmt.Println("Warm-up complete, starting formal testing...")
		}

		// Run anti-pattern test multiple times, take average
		fmt.Printf("\nExecuting anti-pattern: %s (repeated %d times for average)\n", pair.AntiPatternName, testRepetitions)
		var antiPatternTotalTime time.Duration
		var antiPatternTotalMem uint64

		for r := 0; r < testRepetitions; r++ {
			result, err := ProfileFunc(fmt.Sprintf("%s (run %d/%d)", pair.AntiPatternName, r+1, testRepetitions), func() error {
				return pair.AntiPatternFunc(qc)
			})

			if err != nil {
				fmt.Printf("Anti-pattern test execution failed: %v\n", err)
				continue
			}

			antiPatternTotalTime += result.ExecutionTime
			antiPatternTotalMem += result.MemoryUsage
		}

		// Calculate average
		antiPatternResult := ProfileResult{
			Name:          pair.AntiPatternName,
			ExecutionTime: antiPatternTotalTime / testRepetitions,
			MemoryUsage:   antiPatternTotalMem / testRepetitions,
		}

		// Run optimized test multiple times, take average
		fmt.Printf("\nExecuting optimized solution: %s (repeated %d times for average)\n", pair.OptimizedName, testRepetitions)
		var optimizedTotalTime time.Duration
		var optimizedTotalMem uint64

		for r := 0; r < testRepetitions; r++ {
			result, err := ProfileFunc(fmt.Sprintf("%s (run %d/%d)", pair.OptimizedName, r+1, testRepetitions), func() error {
				return pair.OptimizedFunc(qc)
			})

			if err != nil {
				fmt.Printf("Optimized test execution failed: %v\n", err)
				continue
			}

			optimizedTotalTime += result.ExecutionTime
			optimizedTotalMem += result.MemoryUsage
		}

		// Calculate average
		optimizedResult := ProfileResult{
			Name:          pair.OptimizedName,
			ExecutionTime: optimizedTotalTime / testRepetitions,
			MemoryUsage:   optimizedTotalMem / testRepetitions,
		}

		// Calculate performance improvement
		timeDiff := antiPatternResult.ExecutionTime - optimizedResult.ExecutionTime

		// Prevent division by zero
		var timeImprovement float64
		if antiPatternResult.ExecutionTime > 0 {
			timeImprovement = float64(timeDiff) / float64(antiPatternResult.ExecutionTime) * 100
		} else {
			timeImprovement = 0
		}

		memDiff := int64(antiPatternResult.MemoryUsage) - int64(optimizedResult.MemoryUsage)

		// Prevent division by zero and abnormal percentages due to very small values
		var memImprovement float64
		if antiPatternResult.MemoryUsage > 1024 { // Ensure at least 1KB baseline
			memImprovement = float64(memDiff) / float64(antiPatternResult.MemoryUsage) * 100
		} else {
			memImprovement = 0
		}

		// For the third group of tests, if configured to use the same function, expect results close to zero
		if i == 2 {
			if pair.AntiPatternName == pair.OptimizedName {
				fmt.Println("Note: The third group of tests is currently configured to use the same function, expected performance difference should be close to zero")
			} else {
				fmt.Println("Note: The third group of tests is currently configured to use different functions")
			}
		}

		// Print performance comparison
		fmt.Println(strings.Repeat("-", 60))
		fmt.Println("Performance Comparison:")
		fmt.Printf("Execution Time: %.2f%% improvement (Anti-pattern: %v, Optimized: %v, Difference: %v)\n",
			timeImprovement,
			antiPatternResult.ExecutionTime,
			optimizedResult.ExecutionTime,
			timeDiff)

		fmt.Printf("Memory Usage: %.2f%% improvement (Anti-pattern: %.2f MB, Optimized: %.2f MB, Difference: %.2f MB)\n",
			memImprovement,
			float64(antiPatternResult.MemoryUsage)/(1024*1024),
			float64(optimizedResult.MemoryUsage)/(1024*1024),
			float64(memDiff)/(1024*1024))

		fmt.Println(strings.Repeat("=", 60))
	}

	return nil
}
