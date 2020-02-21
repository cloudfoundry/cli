package ccv3

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
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
// NOTE: This only returns the first page of results. We are intentionally not using the paginate helper to fetch all
// pages here because we only needed the first page for the `cf events` output. If we need to, we can refactor this
// later to fetch all pages and make `cf events` only filter down to the first page.
func (client *Client) GetEvents(query ...Query) ([]Event, Warnings, error) {
	var responseBody struct {
		Resources []Event `json:"resources"`
	}

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetEventsRequest,
		ResponseBody: &responseBody,
		Query:        query,
	})

	return responseBody.Resources, warnings, err
}
