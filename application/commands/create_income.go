package commands

import (
	"context"
	"time"

	"escama/domain"
	"escama/domain/events"

	"github.com/google/uuid"
)

type CreateIncomeCommand struct {
	ID          *string
	CategoryID  string
	Amount      float64
	Description *string
	Date        time.Time
}

type CreateIncomeHandler struct {
	Save    func(ctx context.Context, income *domain.Income) error
	Publish func(ctx context.Context, events []events.DomainEvent) error
}

func (h *CreateIncomeHandler) Handle(ctx context.Context, cmd CreateIncomeCommand) error {
	if cmd.ID == nil {
		id := uuid.New().String()
		cmd.ID = &id
	}
	income := domain.NewIncome(*cmd.ID, cmd.CategoryID, cmd.Amount, cmd.Description, cmd.Date)

	if err := h.Save(ctx, income); err != nil {
		return err
	}

	if err := h.Publish(ctx, income.UncommittedEvents()); err != nil {
		return err
	}

	return nil
}
