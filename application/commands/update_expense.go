package commands

import (
	"context"
	"fmt"
	"time"

	"escama/domain/events"
	"escama/infrastructure/repositories"
)

type UpdateExpenseCommand struct {
	ID          string
	CategoryID  string
	Amount      float64
	Description *string
	Date        time.Time
}

type UpdateExpenseHandler struct {
	Repository *repositories.ExpenseRepository
	Publish    func(ctx context.Context, events []events.DomainEvent) error
}

func (h *UpdateExpenseHandler) Handle(ctx context.Context, cmd UpdateExpenseCommand) error {
	// Cargar el gasto existente
	expense, err := h.Repository.GetByID(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to load expense: %w", err)
	}

	if expense == nil {
		return fmt.Errorf("expense not found: %s", cmd.ID)
	}

	// Actualizar el gasto
	expense.Update(cmd.CategoryID, cmd.Amount, cmd.Description, cmd.Date)

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
