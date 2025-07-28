package events

import "time"

type ExpenseDeleted struct {
	ExpenseID string    `json:"expense_id"`
	Occurred  time.Time `json:"occurred"`
}

func (e ExpenseDeleted) EventType() string {
	return "ExpenseDeleted"
}

func (e ExpenseDeleted) OccurredAt() time.Time {
	return e.Occurred
}

func NewExpenseDeleted(expenseID string) ExpenseDeleted {
	return ExpenseDeleted{
		ExpenseID: expenseID,
		Occurred:  time.Now(),
	}
}
