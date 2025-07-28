package events

import "time"

type IncomeDeleted struct {
	IncomeID string    `json:"income_id"`
	Occurred time.Time `json:"occurred"`
}

func (e IncomeDeleted) EventType() string {
	return "IncomeDeleted"
}

func (e IncomeDeleted) OccurredAt() time.Time {
	return e.Occurred
}

func NewIncomeDeleted(incomeID string) IncomeDeleted {
	return IncomeDeleted{
		IncomeID: incomeID,
		Occurred: time.Now(),
	}
}
