package events

import "time"

type IncomeUpdated struct {
	IncomeID    string    `json:"income_id"`
	CategoryID  string    `json:"category_id"`
	Amount      float64   `json:"amount"`
	Description *string   `json:"description,omitempty"`
	Date        time.Time `json:"date"`
	Occurred    time.Time `json:"occurred"`
}

func (e IncomeUpdated) EventType() string {
	return "IncomeUpdated"
}

func (e IncomeUpdated) OccurredAt() time.Time {
	return e.Occurred
}

func NewIncomeUpdated(incomeID, categoryID string, amount float64, description *string, date time.Time) IncomeUpdated {
	return IncomeUpdated{
		IncomeID:    incomeID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
		Occurred:    time.Now(),
	}
}
