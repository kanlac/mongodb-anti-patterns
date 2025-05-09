package main

import (
	"log"
	"os"

	"mongo-bench/cmd/generate"
	"mongo-bench/cmd/run"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mongo-bench",
		Short: "MongoDB Benchmark Tool",
		Long: `MongoDB Benchmark Tool is a tool for generating simulated event data and testing MongoDB query performance.
The tool provides two main functions:
1. Generate random event data and write to MongoDB
2. Execute a series of query benchmark tests and analyze performance`,
	}

	rootCmd.AddCommand(
		generate.NewGenerateCmd(),
		run.NewRunCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Printf("Command execution failed: %v", err)
		os.Exit(1)
	}
}
