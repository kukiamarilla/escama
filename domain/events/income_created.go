package events

import "time"

type IncomeCreated struct {
	IncomeID    string
	CategoryID  string
	Amount      float64
	Description *string
	Date        time.Time
	Occurred    time.Time
}

func (e IncomeCreated) EventType() string {
	return "IncomeCreated"
}

func (e IncomeCreated) OccurredAt() time.Time {
	return e.Occurred
}
