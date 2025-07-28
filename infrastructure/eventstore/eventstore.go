package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"escama/domain/events"
)

// EventStore define el contrato para persistir eventos
type EventStore interface {
	Store(ctx context.Context, aggregateID, aggregateType string, domainEvents []events.DomainEvent) error
	Load(ctx context.Context, aggregateID string) ([]events.StoredEvent, error)
	GetAllEvents(ctx context.Context, startDate, endDate *time.Time) ([]events.StoredEvent, error)
}

// InMemoryEventStore implementación en memoria del EventStore
type InMemoryEventStore struct {
	events    map[string][]events.StoredEvent
	allEvents []events.StoredEvent // Para queries globales
}

func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		events:    make(map[string][]events.StoredEvent),
		allEvents: make([]events.StoredEvent, 0),
	}
}

func (s *InMemoryEventStore) Store(ctx context.Context, aggregateID, aggregateType string, domainEvents []events.DomainEvent) error {
	for _, event := range domainEvents {
		payload, err := s.serializeEvent(event)
		if err != nil {
			return fmt.Errorf("failed to serialize event: %w", err)
		}

		storedEvent := events.StoredEvent{
			ID:            fmt.Sprintf("%s-%d", aggregateID, len(s.events[aggregateID])),
			AggregateID:   aggregateID,
			AggregateType: aggregateType,
			EventType:     event.EventType(),
			Payload:       payload,
			OccurredAt:    event.OccurredAt(),
		}

		s.events[aggregateID] = append(s.events[aggregateID], storedEvent)
		s.allEvents = append(s.allEvents, storedEvent) // Mantener lista global
	}

	return nil
}

func (s *InMemoryEventStore) Load(ctx context.Context, aggregateID string) ([]events.StoredEvent, error) {
	storedEvents, exists := s.events[aggregateID]
	if !exists {
		return []events.StoredEvent{}, nil
	}
	return storedEvents, nil
}

func (s *InMemoryEventStore) GetAllEvents(ctx context.Context, startDate, endDate *time.Time) ([]events.StoredEvent, error) {
	var filteredEvents []events.StoredEvent

	for _, event := range s.allEvents {
		// Filtrar por fechas si se proporcionan
		if startDate != nil && event.OccurredAt.Before(*startDate) {
			continue
		}
		if endDate != nil && event.OccurredAt.After(*endDate) {
			continue
		}
		filteredEvents = append(filteredEvents, event)
	}

	return filteredEvents, nil
}

func (s *InMemoryEventStore) serializeEvent(event events.DomainEvent) (map[string]interface{}, error) {
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
