package run

import (
	"context"
	"fmt"
	"log"
	"strings"

	"mongo-bench/internal/database"
	"mongo-bench/internal/utils"

	"github.com/spf13/cobra"
)

var (
	// Command parameters
	mongoURI      string
	mongoUsername string
	mongoPassword string
	mongoDatabase string
	testsName     []string
)

// NewRunCmd creates a run benchmark command
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute MongoDB benchmarks",
		Long:  "Execute a series of MongoDB query benchmark tests and report execution time and memory usage for each test",
		Run:   runBenchmarkCmd,
	}

	// Add parameters
	cmd.Flags().StringVar(&mongoURI, "uri", "mongodb://localhost:27017", "MongoDB connection URI")
	cmd.Flags().StringVar(&mongoUsername, "username", "admin", "MongoDB username")
	cmd.Flags().StringVar(&mongoPassword, "password", "password", "MongoDB password")
	cmd.Flags().StringVar(&mongoDatabase, "database", "eventstore", "MongoDB database name")
	cmd.Flags().StringSliceVar(&testsName, "test", []string{}, "Specify test name to run")

	return cmd
}

// Execute benchmark command
func runBenchmarkCmd(cmd *cobra.Command, args []string) {
	// Create context
	ctx := context.Background()

	// Configure MongoDB connection
	config := database.MongoConfig{
		URI:      mongoURI,
		Username: mongoUsername,
		Password: mongoPassword,
		Database: mongoDatabase,
	}

	// Connect to MongoDB
	client, err := database.ConnectMongoDB(ctx, config)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	// Get collection
	eventsCollection := database.GetEventsCollection(client, config.Database)

	// Create query context
	queryContext := &utils.QueryContext{
		Ctx:        ctx,
		Collection: eventsCollection,
	}

	// Get all test functions
	testPairs := utils.GetQueryTestPairs()

	// If specific tests are specified, only run these tests
	var selectedTests []utils.QueryTestPair
	if len(testsName) > 0 {
		testMap := make(map[string]bool)
		for _, name := range testsName {
			testMap[name] = true
		}

		for _, pair := range testPairs {
			if testMap[pair.Name] {
				selectedTests = append(selectedTests, pair)
			}
		}

		if len(selectedTests) == 0 {
			log.Fatalf("No matching tests found for the specified test names")
		}
		testPairs = selectedTests
	}

	// Run all tests
	fmt.Println("\nRunning query benchmark tests...")
	fmt.Println(strings.Repeat("=", 50))

	// Record all test results
	var results []utils.ProfileResult

	// Execute tests one by one
	for _, pair := range testPairs {
		fmt.Printf("\nRunning test: %s\n", pair.Name)
		fmt.Println(strings.Repeat("-", 40))

		// Execute test and analyze performance
		result, err := utils.ProfileFunc(pair.Name, func() error {
			return pair.TestFunc(queryContext)
		})
		if err != nil {
			log.Printf("Test failed: %v", err)
			continue
		}

		// Save result
		results = append(results, result)
		fmt.Println(strings.Repeat("-", 40))
	}

	// Print all test results summary
	fmt.Println("\nTest Results Summary:")
	fmt.Println(strings.Repeat("=", 50))
	for _, result := range results {
		fmt.Println(result.String())
		fmt.Println(strings.Repeat("-", 30))
	}
}
