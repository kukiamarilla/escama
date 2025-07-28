package domain

import (
	"time"

	"escama/domain/events"
)

type Expense struct {
	ID          string
	CategoryID  string
	Amount      float64
	Description *string
	Date        time.Time

	uncommitted []events.DomainEvent
}

func NewExpense(id, categoryID string, amount float64, description *string, date time.Time) *Expense {
	exp := &Expense{
		ID:          id,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
	}

	event := events.ExpenseCreated{
		ExpenseID:   id,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
		Occurred:    time.Now().UTC(),
	}
	exp.uncommitted = append(exp.uncommitted, event)

	return exp
}

func (e *Expense) UncommittedEvents() []events.DomainEvent {
	return e.uncommitted
}

func (e *Expense) ClearUncommittedEvents() {
	e.uncommitted = nil
}

func (e *Expense) Update(categoryID string, amount float64, description *string, date time.Time) {
	e.CategoryID = categoryID
	e.Amount = amount
	e.Description = description
	e.Date = date

	event := events.NewExpenseUpdated(e.ID, categoryID, amount, description, date)
	e.uncommitted = append(e.uncommitted, event)
}

func (e *Expense) Delete() {
	event := events.NewExpenseDeleted(e.ID)
	e.uncommitted = append(e.uncommitted, event)
}
