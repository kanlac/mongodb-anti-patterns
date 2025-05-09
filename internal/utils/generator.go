package utils

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"mongo-bench/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GenerateRandomEvent generates a random event
func GenerateRandomEvent() models.Event {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Use current time
	currentTime := time.Now()

	// Randomly select event type
	eventType := models.EventTypes[r.Intn(len(models.EventTypes))]

	// Randomly select description
	description := models.Descriptions[r.Intn(len(models.Descriptions))]

	// Randomly select severity
	severity := models.SeverityLevels[r.Intn(len(models.SeverityLevels))]

	// Randomly select source system
	sourceSystem := models.SourceSystems[r.Intn(len(models.SourceSystems))]

	// Randomly select source IP
	sourceIP := models.ServerIPs[r.Intn(len(models.ServerIPs))]

	// Randomly select affected components (1-3)
	numComponents := 1 + r.Intn(3)
	components := make([]string, 0, numComponents)
	componentsCopy := make([]string, len(models.ComponentsPool))
	copy(componentsCopy, models.ComponentsPool)

	for i := 0; i < numComponents && i < len(componentsCopy); i++ {
		idx := r.Intn(len(componentsCopy))
		components = append(components, componentsCopy[idx])
		// Avoid duplicates
		componentsCopy = append(componentsCopy[:idx], componentsCopy[idx+1:]...)
	}

	// Randomly select status
	status := models.StatusOptions[r.Intn(len(models.StatusOptions))]

	// Build event
	event := models.Event{
		Timestamp:          currentTime,
		EventType:          eventType,
		Description:        description,
		Severity:           severity,
		SourceSystem:       sourceSystem,
		SourceIP:           sourceIP,
		AffectedComponents: components,
		Recommendation:     fmt.Sprintf("Recommended action for %s issue", strings.ToLower(severity.Label)),
		Status:             status,
		Tags:               []string{strings.Split(eventType, " ")[0], "Automated"},
		Metadata:           make(map[string]interface{}),
	}

	// Add some random values to metadata
	event.Metadata["eventId"] = primitive.NewObjectID().Hex()

	if severity.Level >= 3 {
		event.Metadata["priorityFollow"] = true
	}

	// If status is "Resolved", add resolution time and notes
	if status == "Resolved" {
		resolvedTime := currentTime.Add(-time.Duration(1+r.Intn(24)) * time.Hour)
		event.ResolvedAt = &resolvedTime
		event.ResolutionNotes = fmt.Sprintf("Issue resolved by applying standard procedure #%d", r.Intn(100)+1)
		event.AssignedTo = models.TeamOptions[r.Intn(len(models.TeamOptions))]
	} else if status == "In Progress" {
		event.AssignedTo = models.TeamOptions[r.Intn(len(models.TeamOptions))]
	}

	// Add specific metadata based on event type
	switch {
	case strings.Contains(eventType, "System"):
		event.Metadata["cpuUsage"] = 50.0 + r.Float64()*50.0
		event.Metadata["memoryUsage"] = 40.0 + r.Float64()*50.0
	case strings.Contains(eventType, "Security"):
		event.Metadata["attemptCount"] = 5 + r.Intn(20)
		event.Metadata["ipBlocked"] = r.Intn(2) == 1
	case strings.Contains(eventType, "Database"):
		event.Metadata["queryTime"] = 100 + r.Intn(9900)
		event.Metadata["connectionCount"] = 10 + r.Intn(90)
	case strings.Contains(eventType, "API"):
		event.Metadata["statusCode"] = []int{408, 500, 502, 503}[r.Intn(4)]
		event.Metadata["responseTime"] = 1000 + r.Intn(9000)
	}

	return event
}

// InsertEvent inserts a single event into MongoDB
func InsertEvent(ctx context.Context, collection *mongo.Collection, event models.Event, wg *sync.WaitGroup) {
	defer wg.Done()

	result, err := collection.InsertOne(ctx, event)
	if err != nil {
		log.Printf("Failed to insert event: %v", err)
		return
	}

	log.Printf("Event inserted, ID: %v", result.InsertedID)
	log.Printf("Event type: %s, Severity: %s (%d), Status: %s",
		event.EventType, event.Severity.Label, event.Severity.Level, event.Status)
}

// GenerateAndInsertEvents generates and inserts multiple events
func GenerateAndInsertEvents(ctx context.Context, collection *mongo.Collection) int {
	numEvents := 40000

	log.Printf("Concurrently generating %d events...", numEvents)

	var wg sync.WaitGroup

	// Concurrently generate and insert events
	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		event := GenerateRandomEvent()

		// Use goroutine to concurrently insert events
		go InsertEvent(ctx, collection, event, &wg)
	}

	// Wait for all event insertions to complete
	wg.Wait()
	log.Printf("%d events successfully inserted.", numEvents)

	return numEvents
}
