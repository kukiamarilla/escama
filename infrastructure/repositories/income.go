package repositories

import (
	"context"
	"fmt"
	"time"

	"escama/domain"
	"escama/domain/events"
	"escama/infrastructure/eventstore"
)

// IncomeRepository maneja la persistencia de agregados Income vía Event Store
type IncomeRepository struct {
	eventStore eventstore.EventStore
}

func NewIncomeRepository(eventStore eventstore.EventStore) *IncomeRepository {
	return &IncomeRepository{
		eventStore: eventStore,
	}
}

// Save persiste los eventos uncommitted del agregado Income
func (r *IncomeRepository) Save(ctx context.Context, income *domain.Income) error {
	uncommittedEvents := income.UncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	if err := r.eventStore.Store(ctx, income.ID, "Income", uncommittedEvents); err != nil {
		return err
	}

	income.ClearUncommittedEvents()
	return nil
}

// GetByID reconstruye un agregado Income desde sus eventos
func (r *IncomeRepository) GetByID(ctx context.Context, id string) (*domain.Income, error) {
	storedEvents, err := r.eventStore.Load(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load events for income %s: %w", id, err)
	}

	if len(storedEvents) == 0 {
		return nil, nil // No existe
	}

	// Reconstruir el agregado desde los eventos
	var income *domain.Income
	for _, storedEvent := range storedEvents {
		switch storedEvent.EventType {
		case "IncomeCreated":
			// Este es el evento de creación, crear el agregado
			if income == nil {
				income = &domain.Income{
					ID: id,
				}
			}

			// Aplicar el evento al agregado
			if err := r.applyIncomeCreated(income, storedEvent); err != nil {
				return nil, fmt.Errorf("failed to apply IncomeCreated event: %w", err)
			}

		case "IncomeUpdated":
			if income == nil {
				return nil, fmt.Errorf("received IncomeUpdated event before IncomeCreated for income %s", id)
			}

			if err := r.applyIncomeUpdated(income, storedEvent); err != nil {
				return nil, fmt.Errorf("failed to apply IncomeUpdated event: %w", err)
			}

		case "IncomeDeleted":
			// Marcar como eliminado, pero mantener el agregado para propósitos de auditoría
			// En una implementación más compleja podrías tener un flag IsDeleted
		}
	}

	if income != nil {
		income.ClearUncommittedEvents() // Los eventos ya están persistidos
	}

	return income, nil
}

func (r *IncomeRepository) applyIncomeCreated(income *domain.Income, storedEvent events.StoredEvent) error {
	categoryID := r.getStringFromPayload(storedEvent.Payload, "CategoryID", "category_id")
	amount := r.getFloat64FromPayload(storedEvent.Payload, "Amount", "amount")
	description := r.getStringPtrFromPayload(storedEvent.Payload, "Description", "description")
	date := r.getTimeFromPayload(storedEvent.Payload, "Date", "date")

	if date.IsZero() {
		date = storedEvent.OccurredAt
	}

	income.CategoryID = categoryID
	income.Amount = amount
	income.Description = description
	income.Date = date

	return nil
}

func (r *IncomeRepository) applyIncomeUpdated(income *domain.Income, storedEvent events.StoredEvent) error {
	categoryID := r.getStringFromPayload(storedEvent.Payload, "CategoryID", "category_id")
	amount := r.getFloat64FromPayload(storedEvent.Payload, "Amount", "amount")
	description := r.getStringPtrFromPayload(storedEvent.Payload, "Description", "description")
	date := r.getTimeFromPayload(storedEvent.Payload, "Date", "date")

	income.CategoryID = categoryID
	income.Amount = amount
	income.Description = description
	income.Date = date

	return nil
}

// Helper functions
func (r *IncomeRepository) getStringFromPayload(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			if strVal, ok := val.(string); ok {
				return strVal
			}
		}
	}
	return ""
}

func (r *IncomeRepository) getFloat64FromPayload(payload map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			switch v := val.(type) {
			case float64:
				return v
			case float32:
				return float64(v)
			case int:
				return float64(v)
			case int64:
				return float64(v)
			}
		}
	}
	return 0
}

func (r *IncomeRepository) getStringPtrFromPayload(payload map[string]interface{}, keys ...string) *string {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			if strVal, ok := val.(string); ok && strVal != "" {
				return &strVal
			}
		}
	}
	return nil
}

func (r *IncomeRepository) getTimeFromPayload(payload map[string]interface{}, keys ...string) time.Time {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			switch v := val.(type) {
			case string:
				// RFC3339 format
				if parsedDate, err := time.Parse(time.RFC3339, v); err == nil {
					return parsedDate
				}
				// Otros formatos comunes
				formats := []string{
					"2006-01-02T15:04:05Z07:00",
					"2006-01-02T15:04:05.000Z",
					"2006-01-02T15:04:05",
					"2006-01-02 15:04:05",
					"2006-01-02",
				}
				for _, format := range formats {
					if parsedDate, err := time.Parse(format, v); err == nil {
						return parsedDate
					}
				}
			case time.Time:
				return v
			}
		}
	}
	return time.Time{}
}
