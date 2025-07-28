package eventbus

import (
	"context"
	"fmt"

	"escama/domain/events"
)

// EventPublisher define el contrato para publicar eventos de dominio
type EventPublisher interface {
	Publish(ctx context.Context, events []events.DomainEvent) error
}

// InMemoryEventPublisher implementación simple que loggea los eventos y actualiza proyecciones
type InMemoryEventPublisher struct {
	projectionSubscriber *ProjectionSubscriber
}

func NewInMemoryEventPublisher() *InMemoryEventPublisher {
	return &InMemoryEventPublisher{}
}

func (p *InMemoryEventPublisher) SetProjectionSubscriber(subscriber *ProjectionSubscriber) {
	p.projectionSubscriber = subscriber
}

func (p *InMemoryEventPublisher) Publish(ctx context.Context, domainEvents []events.DomainEvent) error {
	for _, event := range domainEvents {
		fmt.Printf("📢 Published event: %s at %s\n", event.EventType(), event.OccurredAt().Format("15:04:05"))

		// Debug: mostrar el evento completo
		fmt.Printf("   Event details: %+v\n", event)

		// Aquí podrías integrar con sistemas externos:
		// - Message queues (RabbitMQ, Apache Kafka)
		// - Event streaming platforms
		// - Webhooks
		// - Notificaciones push

		if err := p.handleEvent(event); err != nil {
			return fmt.Errorf("failed to handle event %s: %w", event.EventType(), err)
		}
	}

	// Actualizar proyecciones si hay un suscriptor configurado
	if p.projectionSubscriber != nil {
		if err := p.projectionSubscriber.Handle(ctx, domainEvents); err != nil {
			fmt.Printf("⚠️  Error updating projections: %v\n", err)
			// No devolvemos el error para no fallar el comando principal
		}
	}

	return nil
}

func (p *InMemoryEventPublisher) handleEvent(event events.DomainEvent) error {
	// Aquí puedes agregar lógica específica para cada tipo de evento
	switch event.EventType() {
	case "CategoryCreated":
		fmt.Printf("   ✓ New category created successfully!\n")
	case "ExpenseCreated":
		fmt.Printf("   ✓ New expense recorded successfully!\n")
	case "IncomeCreated":
		fmt.Printf("   ✓ New income recorded successfully!\n")
	}
	return nil
}
