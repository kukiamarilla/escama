package commands

import (
	"context"
	"fmt"
	"time"

	"escama/domain/events"
	"escama/infrastructure/repositories"
)

type UpdateIncomeCommand struct {
	ID          string
	CategoryID  string
	Amount      float64
	Description *string
	Date        time.Time
}

type UpdateIncomeHandler struct {
	Repository *repositories.IncomeRepository
	Publish    func(ctx context.Context, events []events.DomainEvent) error
}

func (h *UpdateIncomeHandler) Handle(ctx context.Context, cmd UpdateIncomeCommand) error {
	// Cargar el ingreso existente
	income, err := h.Repository.GetByID(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to load income: %w", err)
	}

	if income == nil {
		return fmt.Errorf("income not found: %s", cmd.ID)
	}

	// Actualizar el ingreso
	income.Update(cmd.CategoryID, cmd.Amount, cmd.Description, cmd.Date)

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
