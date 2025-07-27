package commands

import (
	"context"

	"escama/domain"
	"escama/domain/events"

	"github.com/google/uuid"
)

type CreateCategoryCommand struct {
	ID   *string
	Name string
}

type CreateCategoryHandler struct {
	Save    func(ctx context.Context, category *domain.Category) error
	Publish func(ctx context.Context, events []events.DomainEvent) error
}

func (h *CreateCategoryHandler) Handle(ctx context.Context, cmd CreateCategoryCommand) error {
	if cmd.ID == nil {
		id := uuid.New().String()
		cmd.ID = &id
	}
	category := domain.NewCategory(*cmd.ID, cmd.Name)

	if err := h.Save(ctx, category); err != nil {
		return err
	}

	if err := h.Publish(ctx, category.UncommittedEvents()); err != nil {
		return err
	}

	return nil
}
