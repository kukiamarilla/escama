package repositories

import (
	"context"

	"escama/domain"
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
