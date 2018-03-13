package ccv2

import (
	"encoding/json"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

type Event struct {
	// GUID is the unique event identifier.
	GUID string

	// Type is the type of the event.
	Type constant.EventType

	// The GUID of the actor.
	Actor string

	// The actor type.
	ActorType string

	// The name of the actor.
	ActorName string

	// The GUID of the actee.
	Actee string

	// The actee type.
	ActeeType string

	// The name of the actee.
	ActeeName string

	// The event creation time.
	Timestamp time.Time

	// Metadata about the event
	Metadata map[string]interface{}
}

func (event *Event) UnmarshalJSON(data []byte) error {
	var ccEvent struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Type      string                 `json:"type,omitempty"`
			Actor     string                 `json:"actor,omitempty"`
			ActorType string                 `json:"actor_type,omitempty"`
			ActorName string                 `json:"actor_name,omitempty"`
			Actee     string                 `json:"actee,omitempty"`
			ActeeType string                 `json:"actee_type,omitempty"`
			ActeeName string                 `json:"actee_name,omitempty"`
			Timestamp *time.Time             `json:"timestamp"`
			Metadata  map[string]interface{} `json:"metadata"`
		} `json:"entity"`
	}

	if err := json.Unmarshal(data, &ccEvent); err != nil {
		return err
	}

	event.GUID = ccEvent.Metadata.GUID
	event.Type = constant.EventType(ccEvent.Entity.Type)
	event.Actor = ccEvent.Entity.Actor
	event.ActorType = ccEvent.Entity.ActorType
	event.ActorName = ccEvent.Entity.ActorName
	event.Actee = ccEvent.Entity.Actee
	event.ActeeType = ccEvent.Entity.ActeeType
	event.ActeeName = ccEvent.Entity.ActeeName
	if ccEvent.Entity.Timestamp != nil {
		event.Timestamp = *ccEvent.Entity.Timestamp
	}
	event.Metadata = ccEvent.Entity.Metadata

	return nil
}

// GetEvents returns back a list of Events based off of the provided queries.
func (client *Client) GetEvents(filters ...Filter) ([]Event, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetEventsRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullEventsList []Event
	warnings, err := client.paginate(request, Event{}, func(item interface{}) error {
		if event, ok := item.(Event); ok {
			fullEventsList = append(fullEventsList, event)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Event{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullEventsList, warnings, err
}
