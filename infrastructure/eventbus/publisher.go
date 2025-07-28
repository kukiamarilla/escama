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

// InMemoryEventPublisher implementación simple que loggea los eventos
type InMemoryEventPublisher struct{}

func NewInMemoryEventPublisher() *InMemoryEventPublisher {
	return &InMemoryEventPublisher{}
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
