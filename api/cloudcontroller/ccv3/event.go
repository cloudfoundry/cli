package ccv3

import (
	"time"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/internal"
)

type Event struct {
	GUID      string
	CreatedAt time.Time
	Type      string
	ActorName string
	Data      map[string]interface{}
}

func (e *Event) UnmarshalJSON(data []byte) error {
	var ccEvent struct {
		GUID      string    `json:"guid"`
		CreatedAt time.Time `json:"created_at"`
		Type      string    `json:"type"`
		Actor     struct {
			Name string `json:"name"`
		} `json:"actor"`
		Data map[string]interface{} `json:"data"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccEvent)
	if err != nil {
		return err
	}

	e.GUID = ccEvent.GUID
	e.CreatedAt = ccEvent.CreatedAt
	e.Type = ccEvent.Type
	e.ActorName = ccEvent.Actor.Name
	e.Data = ccEvent.Data

	return nil
}

// GetEvents uses the /v3/audit_events endpoint to retrieve a list of audit events.
func (client *Client) GetEvents(query ...Query) ([]Event, Warnings, error) {
	var events []Event

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetEventsRequest,
		Query:        query,
		ResponseBody: Event{},
		AppendToList: func(item interface{}) error {
			events = append(events, item.(Event))
			return nil
		},
	})

	return events, warnings, err
}
