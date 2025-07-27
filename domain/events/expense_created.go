package events

import "time"

type ExpenseCreated struct {
	ExpenseID   string    `json:"expense_id"`
	CategoryID  string    `json:"category_id"`
	Amount      float64   `json:"amount"`
	Description *string   `json:"description,omitempty"`
	Date        time.Time `json:"date"`
	Occurred    time.Time `json:"occurred"`
}

func (e ExpenseCreated) EventType() string {
	return "ExpenseCreated"
}

func (e ExpenseCreated) OccurredAt() time.Time {
	return e.Occurred
}
