package repositories

import (
	"context"
	"fmt"
	"time"

	"escama/domain"
	"escama/domain/events"
	"escama/infrastructure/eventstore"
)

// ExpenseRepository maneja la persistencia de agregados Expense vía Event Store
type ExpenseRepository struct {
	eventStore eventstore.EventStore
}

func NewExpenseRepository(eventStore eventstore.EventStore) *ExpenseRepository {
	return &ExpenseRepository{
		eventStore: eventStore,
	}
}

// Save persiste los eventos uncommitted del agregado Expense
func (r *ExpenseRepository) Save(ctx context.Context, expense *domain.Expense) error {
	uncommittedEvents := expense.UncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	if err := r.eventStore.Store(ctx, expense.ID, "Expense", uncommittedEvents); err != nil {
		return err
	}

	expense.ClearUncommittedEvents()
	return nil
}

// GetByID reconstruye un agregado Expense desde sus eventos
func (r *ExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	storedEvents, err := r.eventStore.Load(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load events for expense %s: %w", id, err)
	}

	if len(storedEvents) == 0 {
		return nil, nil // No existe
	}

	// Reconstruir el agregado desde los eventos
	var expense *domain.Expense
	for _, storedEvent := range storedEvents {
		switch storedEvent.EventType {
		case "ExpenseCreated":
			// Este es el evento de creación, crear el agregado
			if expense == nil {
				expense = &domain.Expense{
					ID: id,
				}
			}

			// Aplicar el evento al agregado
			if err := r.applyExpenseCreated(expense, storedEvent); err != nil {
				return nil, fmt.Errorf("failed to apply ExpenseCreated event: %w", err)
			}

		case "ExpenseUpdated":
			if expense == nil {
				return nil, fmt.Errorf("received ExpenseUpdated event before ExpenseCreated for expense %s", id)
			}

			if err := r.applyExpenseUpdated(expense, storedEvent); err != nil {
				return nil, fmt.Errorf("failed to apply ExpenseUpdated event: %w", err)
			}

		case "ExpenseDeleted":
			// Marcar como eliminado, pero mantener el agregado para propósitos de auditoría
			// En una implementación más compleja podrías tener un flag IsDeleted
		}
	}

	if expense != nil {
		expense.ClearUncommittedEvents() // Los eventos ya están persistidos
	}

	return expense, nil
}

func (r *ExpenseRepository) applyExpenseCreated(expense *domain.Expense, storedEvent events.StoredEvent) error {
	categoryID := r.getStringFromPayload(storedEvent.Payload, "CategoryID", "category_id")
	amount := r.getFloat64FromPayload(storedEvent.Payload, "Amount", "amount")
	description := r.getStringPtrFromPayload(storedEvent.Payload, "Description", "description")
	date := r.getTimeFromPayload(storedEvent.Payload, "Date", "date")

	if date.IsZero() {
		date = storedEvent.OccurredAt
	}

	expense.CategoryID = categoryID
	expense.Amount = amount
	expense.Description = description
	expense.Date = date

	return nil
}

func (r *ExpenseRepository) applyExpenseUpdated(expense *domain.Expense, storedEvent events.StoredEvent) error {
	categoryID := r.getStringFromPayload(storedEvent.Payload, "CategoryID", "category_id")
	amount := r.getFloat64FromPayload(storedEvent.Payload, "Amount", "amount")
	description := r.getStringPtrFromPayload(storedEvent.Payload, "Description", "description")
	date := r.getTimeFromPayload(storedEvent.Payload, "Date", "date")

	expense.CategoryID = categoryID
	expense.Amount = amount
	expense.Description = description
	expense.Date = date

	return nil
}

// Helper functions
func (r *ExpenseRepository) getStringFromPayload(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			if strVal, ok := val.(string); ok {
				return strVal
			}
		}
	}
	return ""
}

func (r *ExpenseRepository) getFloat64FromPayload(payload map[string]interface{}, keys ...string) float64 {
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

func (r *ExpenseRepository) getStringPtrFromPayload(payload map[string]interface{}, keys ...string) *string {
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			if strVal, ok := val.(string); ok && strVal != "" {
				return &strVal
			}
		}
	}
	return nil
}

func (r *ExpenseRepository) getTimeFromPayload(payload map[string]interface{}, keys ...string) time.Time {
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
