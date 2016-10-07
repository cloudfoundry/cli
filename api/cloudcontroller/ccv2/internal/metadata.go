package internal

import "time"

type Metadata struct {
	GUID      string    `json:"guid"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update time for a given object, it can be null
	UpdatedAt *time.Time `json:"updated_at"`
}
