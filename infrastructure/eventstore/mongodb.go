package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"escama/domain/events"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoEventStore implementación de MongoDB del EventStore
type MongoEventStore struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func NewMongoEventStore() (*MongoEventStore, error) {
	connectionString := os.Getenv("MONGODB_CONNECTION_STRING")
	if connectionString == "" {
		return nil, fmt.Errorf("MONGODB_CONNECTION_STRING environment variable is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database("escama")
	collection := database.Collection("events")

	fmt.Println("✅ Connected to MongoDB successfully")

	return &MongoEventStore{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

func (s *MongoEventStore) Store(ctx context.Context, aggregateID, aggregateType string, domainEvents []events.DomainEvent) error {
	var docs []interface{}

	for _, event := range domainEvents {
		payload, err := s.serializeEvent(event)
		if err != nil {
			return fmt.Errorf("failed to serialize event: %w", err)
		}

		storedEvent := events.StoredEvent{
			ID:            fmt.Sprintf("%s-%d", aggregateID, time.Now().UnixNano()),
			AggregateID:   aggregateID,
			AggregateType: aggregateType,
			EventType:     event.EventType(),
			Payload:       payload,
			OccurredAt:    event.OccurredAt(),
		}

		docs = append(docs, storedEvent)
	}

	if len(docs) == 0 {
		return nil
	}

	_, err := s.collection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to insert events: %w", err)
	}

	return nil
}

func (s *MongoEventStore) Load(ctx context.Context, aggregateID string) ([]events.StoredEvent, error) {
	filter := bson.M{"aggregate_id": aggregateID}

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"occurred_at": 1})

	cursor, err := s.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer cursor.Close(ctx)

	var storedEvents []events.StoredEvent
	if err := cursor.All(ctx, &storedEvents); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return storedEvents, nil
}

func (s *MongoEventStore) GetAllEvents(ctx context.Context, startDate, endDate *time.Time) ([]events.StoredEvent, error) {
	filter := bson.M{}

	// Agregar filtros de fecha si se proporcionan
	// Filtramos por la fecha del movimiento en el payload, no por occurred_at
	if startDate != nil || endDate != nil {
		dateFilter := bson.M{}
		if startDate != nil {
			// Incluir toda la fecha de inicio (desde las 00:00:00)
			dateFilter["$gte"] = startDate.Format("2006-01-02")
		}
		if endDate != nil {
			// Incluir toda la fecha de fin (hasta las 23:59:59)
			endOfDay := endDate.AddDate(0, 0, 1).Format("2006-01-02")
			dateFilter["$lt"] = endOfDay
		}

		// Crear un filtro OR para buscar en diferentes campos de fecha del payload
		orFilter := bson.A{
			bson.M{"payload.Date": dateFilter},
			bson.M{"payload.date": dateFilter},
		}

		// Si también hay fechas en formato completo, incluirlas
		if startDate != nil && endDate != nil {
			startDateTime := *startDate
			endDateTime := endDate.AddDate(0, 0, 1).Add(-time.Nanosecond)

			orFilter = append(orFilter,
				bson.M{"payload.Date": bson.M{"$gte": startDateTime, "$lte": endDateTime}},
				bson.M{"payload.date": bson.M{"$gte": startDateTime, "$lte": endDateTime}},
			)
		}

		filter["$or"] = orFilter
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"occurred_at": -1})

	cursor, err := s.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query all events: %w", err)
	}
	defer cursor.Close(ctx)

	var storedEvents []events.StoredEvent
	if err := cursor.All(ctx, &storedEvents); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return storedEvents, nil
}

func (s *MongoEventStore) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.client.Disconnect(ctx)
}

func (s *MongoEventStore) serializeEvent(event events.DomainEvent) (map[string]interface{}, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}
