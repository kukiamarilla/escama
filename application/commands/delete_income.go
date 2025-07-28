package commands

import (
	"context"
	"fmt"

	"escama/domain/events"
	"escama/infrastructure/repositories"
)

type DeleteIncomeCommand struct {
	ID string
}

type DeleteIncomeHandler struct {
	Repository *repositories.IncomeRepository
	Publish    func(ctx context.Context, events []events.DomainEvent) error
}

func (h *DeleteIncomeHandler) Handle(ctx context.Context, cmd DeleteIncomeCommand) error {
	// Cargar el ingreso existente
	income, err := h.Repository.GetByID(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to load income: %w", err)
	}

	if income == nil {
		return fmt.Errorf("income not found: %s", cmd.ID)
	}

	// Eliminar el ingreso
	income.Delete()

	// Guardar cambios
	if err := h.Repository.Save(ctx, income); err != nil {
		return fmt.Errorf("failed to save income: %w", err)
	}

	// Publicar eventos
	if err := h.Publish(ctx, income.UncommittedEvents()); err != nil {
		return fmt.Errorf("failed to publish events: %w", err)
	}

	return nil
}
