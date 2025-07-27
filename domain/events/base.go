package domain

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}
