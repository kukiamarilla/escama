package events

import "time"

type ExpenseUpdated struct {
	ExpenseID   string    `json:"expense_id"`
	CategoryID  string    `json:"category_id"`
	Amount      float64   `json:"amount"`
	Description *string   `json:"description,omitempty"`
	Date        time.Time `json:"date"`
	Occurred    time.Time `json:"occurred"`
}

func (e ExpenseUpdated) EventType() string {
	return "ExpenseUpdated"
}

func (e ExpenseUpdated) OccurredAt() time.Time {
	return e.Occurred
}

func NewExpenseUpdated(expenseID, categoryID string, amount float64, description *string, date time.Time) ExpenseUpdated {
	return ExpenseUpdated{
		ExpenseID:   expenseID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
		Occurred:    time.Now(),
	}
}
