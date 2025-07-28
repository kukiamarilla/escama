package repositories

import (
	"context"

	"escama/domain"
	"escama/infrastructure/eventstore"
)

// CategoryRepository maneja la persistencia de agregados Category vía Event Store
type CategoryRepository struct {
	eventStore eventstore.EventStore
}

func NewCategoryRepository(eventStore eventstore.EventStore) *CategoryRepository {
	return &CategoryRepository{
		eventStore: eventStore,
	}
}

// Save persiste los eventos uncommitted del agregado Category
func (r *CategoryRepository) Save(ctx context.Context, category *domain.Category) error {
	uncommittedEvents := category.UncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	if err := r.eventStore.Store(ctx, category.ID, "Category", uncommittedEvents); err != nil {
		return err
	}

	category.ClearUncommittedEvents()
	return nil
}
