package repositories

import (
	"context"

	"escama/domain"
	"escama/infrastructure/eventstore"
)

// IncomeRepository maneja la persistencia de agregados Income vía Event Store
type IncomeRepository struct {
	eventStore eventstore.EventStore
}

func NewIncomeRepository(eventStore eventstore.EventStore) *IncomeRepository {
	return &IncomeRepository{
		eventStore: eventStore,
	}
}

// Save persiste los eventos uncommitted del agregado Income
func (r *IncomeRepository) Save(ctx context.Context, income *domain.Income) error {
	uncommittedEvents := income.UncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	if err := r.eventStore.Store(ctx, income.ID, "Income", uncommittedEvents); err != nil {
		return err
	}

	income.ClearUncommittedEvents()
	return nil
}
