package models

import "time"

type EventFields struct {
	BasicFields
	Timestamp   time.Time
	Description string
}
