package generate

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"mongo-bench/internal/database"
	"mongo-bench/internal/models"
	"mongo-bench/internal/utils"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// Command parameters
	mongoURI      string
	mongoUsername string
	mongoPassword string
	mongoDatabase string
	duration      int
	concurrency   int
	interval      int
)

// NewGenerateCmd creates a generate command
func NewGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate random event data",
		Long:  "Generate random event data and insert into MongoDB at specified intervals",
		Run:   generateCmd,
	}

	// Add parameters
	cmd.Flags().StringVar(&mongoURI, "uri", "mongodb://localhost:27017", "MongoDB connection URI")
	cmd.Flags().StringVar(&mongoUsername, "username", "admin", "MongoDB username")
	cmd.Flags().StringVar(&mongoPassword, "password", "password", "MongoDB password")
	cmd.Flags().StringVar(&mongoDatabase, "database", "eventstore", "MongoDB database name")
	cmd.Flags().IntVar(&duration, "duration", 0, "How long to run in minutes (0 for infinite)")
	cmd.Flags().IntVar(&concurrency, "concurrency", 5, "Number of concurrent insertion operations")
	cmd.Flags().IntVar(&interval, "interval", 60, "Interval between batch generations in seconds")

	return cmd
}

// Execute generate command
func generateCmd(cmd *cobra.Command, args []string) {
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

	// Get events collection
	eventsCollection := database.GetEventsCollection(client, config.Database)

	// Log startup information
	fmt.Printf("Connected to MongoDB: %s/%s\n", config.URI, config.Database)
	fmt.Printf("Starting event generator with %d second interval\n", interval)
	fmt.Println("Press Ctrl+C to stop")

	// Generate an event immediately on startup
	generateAndInsertEvents(ctx, eventsCollection, concurrency)

	// Calculate end time if duration is set
	var endTime time.Time
	if duration > 0 {
		endTime = time.Now().Add(time.Duration(duration) * time.Minute)
		fmt.Printf("Event generator will run for %d minutes (until %s)\n",
			duration, endTime.Format("15:04:05"))
	}

	// Start ticker for periodic event generation
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// Main loop
	for {
		select {
		case <-ticker.C:
			// Check if we should stop based on duration
			if duration > 0 && time.Now().After(endTime) {
				fmt.Println("Duration completed, stopping event generator")
				return
			}

			// Generate and insert events concurrently
			generateAndInsertEvents(ctx, eventsCollection, concurrency)
		}
	}
}

// Generate and insert events concurrently
func generateAndInsertEvents(ctx context.Context, collection *mongo.Collection, concurrency int) {
	eventCount := rand.Intn(4) // 0-3 events
	if eventCount == 0 {
		fmt.Println("No events generated in this interval")
		return
	}

	fmt.Printf("Generating %d events...\n", eventCount)

	// Create a wait group to manage concurrency
	var wg sync.WaitGroup
	wg.Add(eventCount)

	// Generate and insert events
	for i := 0; i < eventCount; i++ {
		// Create a new event
		e := utils.GenerateRandomEvent()

		// Insert event concurrently
		go func(evt models.Event) {
			defer wg.Done()
			_, err := collection.InsertOne(ctx, evt)
			if err != nil {
				log.Printf("Failed to insert event: %v", err)
			}
		}(e)

		// Log event details
		isResolved := e.Status == "Resolved"
		fmt.Printf("  Event generated: Type=%s, Severity=%d, Resolved=%t\n",
			e.EventType, e.Severity.Level, isResolved)
	}

	// Wait for all insertions to complete
	wg.Wait()
	fmt.Println("All events successfully inserted")
}
