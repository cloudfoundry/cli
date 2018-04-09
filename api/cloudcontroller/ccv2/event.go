package ccv2

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Event represents a Cloud Controller Event
type Event struct {
	// GUID is the unique event identifier.
	GUID string

	// Type is the type of event.
	Type constant.EventType

	// ActorGUID is the GUID of the actor initiating an event.
	ActorGUID string

	// ActorType is the type of actor initiating an event.
	ActorType string

	// ActorName is the name of the actor initiating an event.
	ActorName string

	// ActeeGUID is the GUID of the cc object affected by an event.
	ActeeGUID string

	// ActeeType is the type of the cc object affected by an event.
	ActeeType string

	// ActeeName is the name of the cc object affected by an event.
	ActeeName string

	// Timestamp is the event creation time.
	Timestamp time.Time

	// Metadata contains additional information about the event.
	Metadata map[string]interface{}
}

// UnmarshalJSON helps unmarshal a Cloud Controller Event response.
func (event *Event) UnmarshalJSON(data []byte) error {
	var ccEvent struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Type      string                 `json:"type,omitempty"`
			ActorGUID string                 `json:"actor,omitempty"`
			ActorType string                 `json:"actor_type,omitempty"`
			ActorName string                 `json:"actor_name,omitempty"`
			ActeeGUID string                 `json:"actee,omitempty"`
			ActeeType string                 `json:"actee_type,omitempty"`
			ActeeName string                 `json:"actee_name,omitempty"`
			Timestamp *time.Time             `json:"timestamp"`
			Metadata  map[string]interface{} `json:"metadata"`
		} `json:"entity"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccEvent)
	if err != nil {
		return err
	}

	event.GUID = ccEvent.Metadata.GUID
	event.Type = constant.EventType(ccEvent.Entity.Type)
	event.ActorGUID = ccEvent.Entity.ActorGUID
	event.ActorType = ccEvent.Entity.ActorType
	event.ActorName = ccEvent.Entity.ActorName
	event.ActeeGUID = ccEvent.Entity.ActeeGUID
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
