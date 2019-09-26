package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Event struct {
	GUID      string
	CreatedAt string
	Type      string
	ActorName string
}

func (e Event) MarshalJson() ([]byte, error) {
	return nil, nil
}

func (e *Event) UnmarshalJSON(data []byte) error {
	var ccEvent struct {
		GUID      string `json:"guid"`
		CreatedAt string `json:"created_at"`
		Type      string `json:"type"`
		Actor     struct {
			Name string `json:"name"`
		} `json:"actor"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccEvent)
	if err != nil {
		return err
	}

	e.GUID = ccEvent.GUID
	e.CreatedAt = ccEvent.CreatedAt
	e.Type = ccEvent.Type
	e.ActorName = ccEvent.Actor.Name

	return nil
}

func (client *Client) GetEvents(query ...Query) ([]Event, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetEventsRequest,
		Query:       query,
	})

	if err != nil {
		return nil, nil, err // untested
	}

	var eventResponse struct {
		Resources []Event `json:"resources"`
	}

	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &eventResponse,
	}
	err = client.connection.Make(request, &response)

	return eventResponse.Resources, response.Warnings, err
}
