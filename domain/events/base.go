package events

import (
	"time"
)

type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

type StoredEvent struct {
	ID            string                 `bson:"_id"`
	AggregateID   string                 `bson:"aggregate_id"`
	AggregateType string                 `bson:"aggregate_type"`
	EventType     string                 `bson:"event_type"`
	Payload       map[string]interface{} `bson:"payload"`
	OccurredAt    time.Time              `bson:"occurred_at"`
}
