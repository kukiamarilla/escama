package commands

import (
	"context"
	"time"

	"escama/domain"
	"escama/domain/events"

	"github.com/google/uuid"
)

type CreateExpenseCommand struct {
	ID          *string
	Name        string
	CategoryID  string
	Amount      float64
	Description *string
	Date        time.Time
}

type CreateExpenseHandler struct {
	Save    func(ctx context.Context, expense *domain.Expense) error
	Publish func(ctx context.Context, events []events.DomainEvent) error
}

func (h *CreateExpenseHandler) Handle(ctx context.Context, cmd CreateExpenseCommand) error {
	if cmd.ID == nil {
		id := uuid.New().String()
		cmd.ID = &id
	}
	expense := domain.NewExpense(*cmd.ID, cmd.CategoryID, cmd.Amount, cmd.Description, cmd.Date)

	if err := h.Save(ctx, expense); err != nil {
		return err
	}

	if err := h.Publish(ctx, expense.UncommittedEvents()); err != nil {
		return err
	}

	return nil
}
