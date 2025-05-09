package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SeverityLevel represents the severity of an event
type SeverityLevel struct {
	Level int    `bson:"level"`
	Label string `bson:"label"`
	Color string `bson:"color"`
}

// Event represents an event record
type Event struct {
	ID                 primitive.ObjectID     `bson:"_id,omitempty"`
	Timestamp          time.Time              `bson:"timestamp"`
	EventType          string                 `bson:"eventType"`
	Description        string                 `bson:"description"`
	Severity           SeverityLevel          `bson:"severity"`
	SourceSystem       string                 `bson:"sourceSystem"`
	SourceIP           string                 `bson:"sourceIp,omitempty"`
	AffectedComponents []string               `bson:"affectedComponents"`
	Recommendation     string                 `bson:"recommendation"`
	Status             string                 `bson:"status"`
	AssignedTo         string                 `bson:"assignedTo,omitempty"`
	ResolvedAt         *time.Time             `bson:"resolvedAt,omitempty"`
	ResolutionNotes    string                 `bson:"resolutionNotes,omitempty"`
	Tags               []string               `bson:"tags"`
	Metadata           map[string]interface{} `bson:"metadata"`
}

// Constants and sample data
var (
	EventTypes = []string{
		"System Warning",
		"Security Incident",
		"Database Exception",
		"API Error",
		"Network Issue",
		"Application Error",
		"Hardware Failure",
		"System Maintenance",
	}

	Descriptions = []string{
		"CPU usage exceeds 90%",
		"Multiple failed login attempts detected",
		"Database connection pool exhausted",
		"Third-party API consistently returning timeout errors",
		"Network latency increased significantly",
		"Application crashed unexpectedly",
		"Disk I/O performance degraded",
		"Memory usage approaching limit",
		"Scheduled system update required",
		"Suspicious network traffic detected",
	}

	SourceSystems = []string{
		"Server A",
		"Authentication Service",
		"Main Database",
		"Payment Service",
		"Network Gateway",
		"User Management System",
		"API Gateway",
		"Monitoring System",
	}

	ServerIPs = []string{
		"192.168.1.100",
		"192.168.1.101",
		"192.168.1.150",
		"192.168.2.50",
		"10.0.0.15",
		"10.0.0.16",
	}

	ComponentsPool = []string{
		"Web Server",
		"Database",
		"User Authentication System",
		"Order System",
		"User Management",
		"Checkout Process",
		"Payment Processing",
		"Reporting System",
		"Notification Service",
		"Admin Panel",
	}

	StatusOptions = []string{
		"Unhandled",
		"In Progress",
		"Resolved",
		"Scheduled",
	}

	TeamOptions = []string{
		"Security Team",
		"Database Team",
		"Network Team",
		"DevOps Team",
		"Application Team",
		"Operations Team",
	}

	SeverityLevels = []SeverityLevel{
		{Level: 0, Label: "Information", Color: "blue"},
		{Level: 1, Label: "Low", Color: "green"},
		{Level: 2, Label: "Medium", Color: "yellow"},
		{Level: 3, Label: "High", Color: "orange"},
		{Level: 4, Label: "Critical", Color: "red"},
	}
)
