package domain

import (
	"time"

	"escama/domain/events"
)

type Income struct {
	ID          string
	CategoryID  string
	Amount      float64
	Description *string
	Date        time.Time

	uncommitted []events.DomainEvent
}

func NewIncome(id, categoryID string, amount float64, description *string, date time.Time) *Income {
	inc := &Income{
		ID:          id,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
	}

	event := events.IncomeCreated{
		IncomeID:    id,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
		Occurred:    time.Now().UTC(),
	}
	inc.uncommitted = append(inc.uncommitted, event)

	return inc
}

func (i *Income) UncommittedEvents() []events.DomainEvent {
	return i.uncommitted
}

func (i *Income) ClearUncommittedEvents() {
	i.uncommitted = nil
}

func (i *Income) Update(categoryID string, amount float64, description *string, date time.Time) {
	i.CategoryID = categoryID
	i.Amount = amount
	i.Description = description
	i.Date = date

	event := events.NewIncomeUpdated(i.ID, categoryID, amount, description, date)
	i.uncommitted = append(i.uncommitted, event)
}

func (i *Income) Delete() {
	event := events.NewIncomeDeleted(i.ID)
	i.uncommitted = append(i.uncommitted, event)
}
