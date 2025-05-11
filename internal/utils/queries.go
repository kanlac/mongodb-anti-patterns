package utils

import (
	"context"
	"fmt"
	"time"

	"mongo-bench/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// QueryContext holds the context and collection for queries
type QueryContext struct {
	Ctx        context.Context
	Collection *mongo.Collection
}

// QueryTestFunc defines a function type for query tests
type QueryTestFunc func(ctx *QueryContext) error

// QueryTestPair holds a test name and its function
type QueryTestPair struct {
	Name     string
	TestFunc QueryTestFunc
}

// GetQueryTestPairs returns all query test pairs
func GetQueryTestPairs() []QueryTestPair {
	return []QueryTestPair{
		{"FindAllFieldsAntiPattern", FindAllFieldsAntiPattern},
		{"FindWithProjectionOptimized", FindWithProjectionOptimized},
		{"AggregateBeforeFilterAntiPattern", AggregateBeforeFilterAntiPattern},
		{"FilterBeforeAggregateOptimized", FilterBeforeAggregateOptimized},
		{"FindRecentEvents", FindRecentEvents},
		{"FindHighSeverityEvents", FindHighSeverityEvents},
		{"AggregateEventsBySeverity", AggregateEventsBySeverity},
		{"FindEventsWithProjection", FindEventsWithProjection},
		{"FindEventsByTimeRange", FindEventsByTimeRange},
		{"ComplexAggregation", ComplexAggregation},
		{"FindEventsWithSorting", FindEventsWithSorting},
	}
}

// Optimization anti-pattern and optimized solution pairs

// Pair 1: Projection optimization
// Anti-pattern: Query entire documents when only a few fields are needed
func FindAllFieldsAntiPattern(ctx *QueryContext) error {
	fmt.Println("Running anti-pattern: Querying all fields when only a few are needed")

	// Find recent high severity events, but return all fields
	filter := bson.M{
		"severity.level": bson.M{"$gte": 3},
		"timestamp": bson.M{
			"$gte": time.Now().Add(-24 * time.Hour),
		},
	}

	cursor, err := ctx.Collection.Find(ctx.Ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx.Ctx)

	var events []models.Event
	if err = cursor.All(ctx.Ctx, &events); err != nil {
		return err
	}

	fmt.Printf("Found %d high severity events\n", len(events))
	return nil
}

// Optimized solution: Use projection to return only needed fields
func FindWithProjectionOptimized(ctx *QueryContext) error {
	fmt.Println("Running optimized solution: Using projection to return only needed fields")

	// Find recent high severity events, but return only necessary fields
	filter := bson.M{
		"severity.level": bson.M{"$gte": 3},
		"timestamp": bson.M{
			"$gte": time.Now().Add(-24 * time.Hour),
		},
	}

	projection := bson.M{
		"eventType":    1,
		"severity":     1,
		"timestamp":    1,
		"sourceSystem": 1,
		"_id":          0,
	}

	opts := options.Find().SetProjection(projection)
	cursor, err := ctx.Collection.Find(ctx.Ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx.Ctx)

	var events []bson.M
	if err = cursor.All(ctx.Ctx, &events); err != nil {
		return err
	}

	fmt.Printf("Found %d high severity events (projected fields)\n", len(events))
	return nil
}

// Pair 2: Aggregation pipeline optimization
// Anti-pattern: Do complex aggregations before filtering data
func AggregateBeforeFilterAntiPattern(ctx *QueryContext) error {
	fmt.Println("Running anti-pattern: Performing aggregations before filtering")

	pipeline := mongo.Pipeline{
		{{"$group", bson.M{
			"_id": "$eventType",
			// "_id":    "$severity.level",
			"count":  bson.M{"$sum": 1},
			"events": bson.M{"$push": "$$ROOT"},
		}}},
		{{"$match", bson.M{
			"_id": bson.M{"$eq": "System Warning"},
		}}},
	}

	cursor, err := ctx.Collection.Aggregate(ctx.Ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx.Ctx)

	var results []bson.M
	if err = cursor.All(ctx.Ctx, &results); err != nil {
		return err
	}

	fmt.Printf("Found %d severity groups after aggregation\n", len(results))
	return nil
}

// Optimized solution: Filter data before performing aggregation
func FilterBeforeAggregateOptimized(ctx *QueryContext) error {
	fmt.Println("Running optimized solution: Filtering data before aggregation")

	pipeline := mongo.Pipeline{
		{{"$match", bson.M{
			"eventType": bson.M{"$eq": "System Warning"},
		}}},
		{{"$group", bson.M{
			"_id":          "$eventType",
			"count":        bson.M{"$sum": 1},
			"avgTimestamp": bson.M{"$avg": bson.M{"$toLong": "$timestamp"}},
		}}},
	}

	cursor, err := ctx.Collection.Aggregate(ctx.Ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx.Ctx)

	var results []bson.M
	if err = cursor.All(ctx.Ctx, &results); err != nil {
		return err
	}

	fmt.Printf("Found %d severity groups with optimized aggregation\n", len(results))
	return nil
}

// Individual benchmark query functions

// FindRecentEvents finds the most recent events
func FindRecentEvents(ctx *QueryContext) error {
	fmt.Println("Finding most recent events")

	// Find most recent events
	opts := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(10)

	cursor, err := ctx.Collection.Find(ctx.Ctx, bson.M{}, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx.Ctx)

	var events []models.Event
	if err = cursor.All(ctx.Ctx, &events); err != nil {
		return err
	}

	fmt.Printf("Found %d recent events\n", len(events))

	// Display event timestamps
	for i, e := range events {
		fmt.Printf("  %d. %s - %s (Severity: %d)\n",
			i+1, e.Timestamp.Format(time.RFC3339), e.EventType, e.Severity.Level)
	}

	return nil
}

// FindHighSeverityEvents finds high severity events
func FindHighSeverityEvents(ctx *QueryContext) error {
	fmt.Println("Finding high severity events")

	// Find high severity events
	filter := bson.M{
		"severity.level": bson.M{"$gte": 3},
	}

	opts := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(10)

	cursor, err := ctx.Collection.Find(ctx.Ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx.Ctx)

	var events []models.Event
	if err = cursor.All(ctx.Ctx, &events); err != nil {
		return err
	}

	fmt.Printf("Found %d high severity events\n", len(events))

	// Display high severity events
	for i, e := range events {
		fmt.Printf("  %d. [Level %d] %s - %s\n",
			i+1, e.Severity.Level, e.EventType, e.Description)
	}

	return nil
}

// AggregateEventsBySeverity aggregates events by severity level
func AggregateEventsBySeverity(ctx *QueryContext) error {
	fmt.Println("Aggregating events by severity level")

	// Aggregate by severity level
	pipeline := mongo.Pipeline{
		{{"$group", bson.M{
			"_id":   "$severity.level",
			"count": bson.M{"$sum": 1},
			"label": bson.M{"$first": "$severity.label"},
		}}},
		{{"$sort", bson.M{"_id": 1}}},
	}

	cursor, err := ctx.Collection.Aggregate(ctx.Ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx.Ctx)

	var results []bson.M
	if err = cursor.All(ctx.Ctx, &results); err != nil {
		return err
	}

	fmt.Printf("Found %d severity groups\n", len(results))

	// Display severity distribution
	for _, r := range results {
		fmt.Printf("  Level %v (%v): %v events\n",
			r["_id"], r["label"], r["count"])
	}

	return nil
}

// FindEventsWithProjection finds events with projection
func FindEventsWithProjection(ctx *QueryContext) error {
	fmt.Println("Finding events with field projection")

	// Query with projection
	filter := bson.M{
		"eventType": bson.M{"$regex": ".*Database.*", "$options": "i"},
	}

	projection := bson.M{
		"eventType":    1,
		"description":  1,
		"severity":     1,
		"sourceSystem": 1,
		"timestamp":    1,
		"_id":          0,
	}

	opts := options.Find().
		SetProjection(projection).
		SetLimit(5)

	cursor, err := ctx.Collection.Find(ctx.Ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx.Ctx)

	var events []bson.M
	if err = cursor.All(ctx.Ctx, &events); err != nil {
		return err
	}

	fmt.Printf("Found %d database-related events\n", len(events))

	// Display projected events
	for i, e := range events {
		fmt.Printf("  %d. %v - %v\n",
			i+1, e["eventType"], e["description"])
	}

	return nil
}

// FindEventsByTimeRange finds events within a time range
func FindEventsByTimeRange(ctx *QueryContext) error {
	fmt.Println("Finding events within a time range")

	// Define a time range (last 24 hours)
	end := time.Now()
	start := end.Add(-24 * time.Hour)

	// Query for events in time range
	filter := bson.M{
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	opts := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(10)

	cursor, err := ctx.Collection.Find(ctx.Ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx.Ctx)

	var events []models.Event
	if err = cursor.All(ctx.Ctx, &events); err != nil {
		return err
	}

	fmt.Printf("Found %d events in the last 24 hours\n", len(events))

	// Display time range events
	for i, e := range events {
		fmt.Printf("  %d. %s - %s\n",
			i+1, e.Timestamp.Format(time.RFC3339), e.EventType)
	}

	return nil
}

// ComplexAggregation performs a complex aggregation
func ComplexAggregation(ctx *QueryContext) error {
	fmt.Println("Performing complex aggregation")

	// Complex aggregate query with multiple stages
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{
			"severity.level": bson.M{"$gte": 2},
		}}},
		{{"$group", bson.M{
			"_id":         "$eventType",
			"count":       bson.M{"$sum": 1},
			"avgSeverity": bson.M{"$avg": "$severity.level"},
			"systems":     bson.M{"$addToSet": "$sourceSystem"},
		}}},
		{{"$sort", bson.M{"avgSeverity": -1}}},
		{{"$limit", 5}},
	}

	cursor, err := ctx.Collection.Aggregate(ctx.Ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx.Ctx)

	var results []bson.M
	if err = cursor.All(ctx.Ctx, &results); err != nil {
		return err
	}

	fmt.Printf("Complex aggregation produced %d result groups\n", len(results))

	// Display complex aggregation results
	for i, r := range results {
		fmt.Printf("  %d. Event Type: %v\n", i+1, r["_id"])
		fmt.Printf("     Count: %v, Avg Severity: %.2f\n",
			r["count"], r["avgSeverity"])
		fmt.Printf("     Affected Systems: %v\n", r["systems"])
	}

	return nil
}

// FindEventsWithSorting finds events with sorting
func FindEventsWithSorting(ctx *QueryContext) error {
	fmt.Println("Finding events with sorting options")

	// Query with sort
	filter := bson.M{
		"sourceSystem": bson.M{"$in": []string{
			"Database",
			"Main Database",
			"Authentication Service",
		}},
	}

	opts := options.Find().
		SetSort(bson.M{"severity.level": -1, "timestamp": -1}).
		SetLimit(10)

	cursor, err := ctx.Collection.Find(ctx.Ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx.Ctx)

	var events []models.Event
	if err = cursor.All(ctx.Ctx, &events); err != nil {
		return err
	}

	fmt.Printf("Found %d database/auth service events\n", len(events))

	// Display sorted events
	for i, e := range events {
		fmt.Printf("  %d. [Level %d] %s - %s (%s)\n",
			i+1, e.Severity.Level, e.SourceSystem, e.EventType,
			e.Timestamp.Format(time.RFC3339))
	}

	return nil
}
