package eventbus

import (
	"context"
	"log"

	"escama/domain/events"
	"escama/infrastructure/projections"
)

// ProjectionSubscriber suscriptor que actualiza las proyecciones cuando ocurren eventos
type ProjectionSubscriber struct {
	projectionStore *projections.ProjectionStore
}

func NewProjectionSubscriber(projectionStore *projections.ProjectionStore) *ProjectionSubscriber {
	return &ProjectionSubscriber{
		projectionStore: projectionStore,
	}
}

// Handle procesa eventos y actualiza las proyecciones
func (s *ProjectionSubscriber) Handle(ctx context.Context, domainEvents []events.DomainEvent) error {
	for _, domainEvent := range domainEvents {
		// Convertir el evento de dominio a evento almacenado para procesar
		storedEvent := events.StoredEvent{
			EventType:  domainEvent.EventType(),
			Payload:    s.eventToPayload(domainEvent),
			OccurredAt: domainEvent.OccurredAt(),
		}

		if err := s.projectionStore.ProcessEvent(ctx, storedEvent); err != nil {
			log.Printf("Error processing event for projections: %v", err)
			// No retornamos error para no fallar el comando, solo loggeamos
		}
	}

	return nil
}

// eventToPayload convierte un evento de dominio a payload map
func (s *ProjectionSubscriber) eventToPayload(event events.DomainEvent) map[string]interface{} {
	payload := make(map[string]interface{})

	switch e := event.(type) {
	case events.CategoryCreated:
		payload["CategoryID"] = e.CategoryID
		payload["Name"] = e.Name

	case events.ExpenseCreated:
		payload["ExpenseID"] = e.ExpenseID
		payload["CategoryID"] = e.CategoryID
		payload["Amount"] = e.Amount
		payload["Description"] = e.Description
		payload["Date"] = e.Date

	case events.IncomeCreated:
		payload["IncomeID"] = e.IncomeID
		payload["CategoryID"] = e.CategoryID
		payload["Amount"] = e.Amount
		payload["Description"] = e.Description
		payload["Date"] = e.Date

	case events.ExpenseUpdated:
		payload["ExpenseID"] = e.ExpenseID
		payload["CategoryID"] = e.CategoryID
		payload["Amount"] = e.Amount
		payload["Description"] = e.Description
		payload["Date"] = e.Date

	case events.IncomeUpdated:
		payload["IncomeID"] = e.IncomeID
		payload["CategoryID"] = e.CategoryID
		payload["Amount"] = e.Amount
		payload["Description"] = e.Description
		payload["Date"] = e.Date

	case events.ExpenseDeleted:
		payload["ExpenseID"] = e.ExpenseID

	case events.IncomeDeleted:
		payload["IncomeID"] = e.IncomeID

	default:
		log.Printf("Unknown event type for projection: %T", event)
	}

	return payload
}
