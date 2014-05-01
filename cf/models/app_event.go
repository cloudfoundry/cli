package models

import "time"

type EventFields struct {
	Guid        string
	Name        string
	Timestamp   time.Time
	Description string
	ActorName   string
}
