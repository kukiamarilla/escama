package events

import "time"

type CategoryCreated struct {
	CategoryID string    `json:"category_id"`
	Name       string    `json:"name"`
	Occurred   time.Time `json:"occurred"`
}

func (e CategoryCreated) EventType() string {
	return "CategoryCreated"
}

func (e CategoryCreated) OccurredAt() time.Time {
	return e.Occurred
}
