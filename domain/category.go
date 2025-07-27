package domain

import (
	"time"

	"escama/domain/events"
)

type Category struct {
	ID   string
	Name string

	uncommitted []events.DomainEvent
}

func NewCategory(id, name string) *Category {
	c := &Category{
		ID:   id,
		Name: name,
	}

	event := events.CategoryCreated{
		CategoryID: id,
		Name:       name,
		Occurred:   time.Now().UTC(),
	}
	c.uncommitted = append(c.uncommitted, event)

	return c
}

func (c *Category) UncommittedEvents() []events.DomainEvent {
	return c.uncommitted
}

func (c *Category) ClearUncommittedEvents() {
	c.uncommitted = nil
}
