package utils

import (
	"fmt"
	"runtime"
	"time"
)

// ProfileResult represents the result of a profiled function execution
type ProfileResult struct {
	Name          string
	ExecutionTime time.Duration
	MemoryUsage   uint64
}

// String returns a formatted string representation of ProfileResult
func (r ProfileResult) String() string {
	return fmt.Sprintf("Profile [%s]:\n- Execution time: %v\n- Memory usage: %.2f MB",
		r.Name,
		r.ExecutionTime,
		float64(r.MemoryUsage)/(1024*1024))
}

// TimerFunc is the type of the function used to measure execution time and memory usage
type TimerFunc func() error

// ProfileFunc executes and profiles a function, returning timing and memory usage metrics
func ProfileFunc(name string, fn func() error) (ProfileResult, error) {
	var memStats runtime.MemStats

	// Get baseline memory stats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	baselineAlloc := memStats.TotalAlloc

	// Record start time
	startTime := time.Now()

	// Execute the function
	err := fn()

	// Record execution time
	execTime := time.Since(startTime)

	// Get final memory stats
	runtime.ReadMemStats(&memStats)
	finalAlloc := memStats.TotalAlloc

	// Calculate memory allocated during execution
	allocatedMem := finalAlloc - baselineAlloc

	// Create result
	result := ProfileResult{
		Name:          name,
		ExecutionTime: execTime,
		MemoryUsage:   allocatedMem,
	}

	// Return result and any error from the function
	return result, err
}
