package commands

import (
	"context"
	"fmt"

	"escama/domain/events"
	"escama/infrastructure/repositories"
)

type DeleteExpenseCommand struct {
	ID string
}

type DeleteExpenseHandler struct {
	Repository *repositories.ExpenseRepository
	Publish    func(ctx context.Context, events []events.DomainEvent) error
}

func (h *DeleteExpenseHandler) Handle(ctx context.Context, cmd DeleteExpenseCommand) error {
	// Cargar el gasto existente
	expense, err := h.Repository.GetByID(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to load expense: %w", err)
	}

	if expense == nil {
		return fmt.Errorf("expense not found: %s", cmd.ID)
	}

	// Eliminar el gasto
	expense.Delete()

	// Guardar cambios
	if err := h.Repository.Save(ctx, expense); err != nil {
		return fmt.Errorf("failed to save expense: %w", err)
	}

	// Publicar eventos
	if err := h.Publish(ctx, expense.UncommittedEvents()); err != nil {
		return fmt.Errorf("failed to publish events: %w", err)
	}

	return nil
}
