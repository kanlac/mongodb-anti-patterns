package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Collection names
const (
	EventsCollectionName = "events"
)

// Index names
const (
	TimestampIndex        = "timestamp_idx"
	EventTypeIndex        = "eventType_idx"
	SeverityLevelIndex    = "severityLevel_idx"
	StatusIndex           = "status_idx"
	CompoundSeverityIndex = "compound_severity_idx"
)

// MongoConfig represents MongoDB connection configuration
type MongoConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

// ConnectMongoDB establishes a connection to MongoDB
func ConnectMongoDB(ctx context.Context, config MongoConfig) (*mongo.Client, error) {
	// Create client options
	clientOptions := options.Client().ApplyURI(config.URI)

	// Add credentials if provided
	if config.Username != "" && config.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: config.Username,
			Password: config.Password,
		})
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Verify the connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return client, nil
}

// GetEventsCollection returns the events collection
func GetEventsCollection(client *mongo.Client, database string) *mongo.Collection {
	return client.Database(database).Collection(EventsCollectionName)
}

// CreateEventIndexes creates indexes for the events collection
func CreateEventIndexes(ctx context.Context, collection *mongo.Collection) ([]string, error) {
	// Create indexes for better query performance
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "timestamp", Value: -1}},
			Options: options.Index().SetName(TimestampIndex),
		},
		{
			Keys:    bson.D{{Key: "eventType", Value: 1}},
			Options: options.Index().SetName(EventTypeIndex),
		},
		{
			Keys:    bson.D{{Key: "severity.level", Value: -1}},
			Options: options.Index().SetName(SeverityLevelIndex),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName(StatusIndex),
		},
		{
			Keys: bson.D{
				{Key: "severity.level", Value: -1},
				{Key: "timestamp", Value: -1},
			},
			Options: options.Index().SetName(CompoundSeverityIndex),
		},
	}

	// Create indexes with a timeout
	indexCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Create the indexes
	result, err := collection.Indexes().CreateMany(indexCtx, indexes)
	if err != nil {
		log.Printf("Failed to create indexes: %v", err)
		return nil, err
	}

	return result, nil
}

// PingMongoDB checks if MongoDB is accessible
func PingMongoDB(ctx context.Context, client *mongo.Client) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.Ping(pingCtx, readpref.Primary())
}
