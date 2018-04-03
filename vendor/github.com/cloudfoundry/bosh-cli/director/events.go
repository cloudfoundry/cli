package director

import (
	"fmt"
	"net/url"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type EventImpl struct {
	client Client

	id             string
	parentID       string
	timestamp      time.Time
	user           string
	action         string
	objectType     string
	objectName     string
	taskID         string
	deploymentName string
	instance       string
	context        map[string]interface{}
	error          string
}

type EventResp struct {
	ID             string                 `json:"id"`
	Timestamp      int64                  `json:"timestamp"`
	User           string                 `json:"user"`
	Action         string                 `json:"action"`
	ObjectType     string                 `json:"object_type"`
	ObjectName     string                 `json:"object_name"`
	TaskID         string                 `json:"task"`
	DeploymentName string                 `json:"deployment"`
	Instance       string                 `json:"instance"`
	ParentID       string                 `json:"parent_id,omitempty"`
	Context        map[string]interface{} `json:"context"`
	Error          string                 `json:"error"`
}

func (e EventImpl) ID() string                      { return e.id }
func (e EventImpl) ParentID() string                { return e.parentID }
func (e EventImpl) Timestamp() time.Time            { return e.timestamp }
func (e EventImpl) User() string                    { return e.user }
func (e EventImpl) Action() string                  { return e.action }
func (e EventImpl) ObjectType() string              { return e.objectType }
func (e EventImpl) ObjectName() string              { return e.objectName }
func (e EventImpl) TaskID() string                  { return e.taskID }
func (e EventImpl) DeploymentName() string          { return e.deploymentName }
func (e EventImpl) Instance() string                { return e.instance }
func (e EventImpl) Context() map[string]interface{} { return e.context }
func (e EventImpl) Error() string                   { return e.error }

func NewEventFromResp(client Client, r EventResp) EventImpl {
	return EventImpl{
		client: client,

		id:             r.ID,
		parentID:       r.ParentID,
		timestamp:      time.Unix(r.Timestamp, 0).UTC(),
		user:           r.User,
		action:         r.Action,
		objectType:     r.ObjectType,
		objectName:     r.ObjectName,
		taskID:         r.TaskID,
		deploymentName: r.DeploymentName,
		instance:       r.Instance,
		context:        r.Context,
		error:          r.Error,
	}
}

func (d DirectorImpl) Events(opts EventsFilter) ([]Event, error) {
	events := []Event{}

	eventResps, err := d.client.Events(opts)
	if err != nil {
		return events, err
	}

	for _, r := range eventResps {
		events = append(events, NewEventFromResp(d.client, r))
	}

	return events, nil
}

func (c Client) Events(opts EventsFilter) ([]EventResp, error) {
	var events []EventResp

	u, err := url.Parse("/events")

	if err != nil {
		panic("Parse err is non-nil.")
	}

	q := u.Query()
	if len(opts.BeforeID) > 0 {
		q.Set("before_id", opts.BeforeID)
	}
	if len(opts.Before) > 0 {
		q.Set("before_time", opts.Before)
	}
	if len(opts.After) > 0 {
		q.Set("after_time", opts.After)
	}
	if len(opts.Deployment) > 0 {
		q.Set("deployment", opts.Deployment)
	}
	if len(opts.Task) > 0 {
		q.Set("task", opts.Task)
	}
	if len(opts.Instance) > 0 {
		q.Set("instance", opts.Instance)
	}
	if len(opts.User) > 0 {
		q.Set("user", opts.User)
	}
	if len(opts.Action) > 0 {
		q.Set("action", opts.Action)
	}
	if len(opts.ObjectType) > 0 {
		q.Set("object_type", opts.ObjectType)
	}
	if len(opts.ObjectName) > 0 {
		q.Set("object_name", opts.ObjectName)
	}

	u.RawQuery = q.Encode()

	path := u.String()

	err = c.clientRequest.Get(path, &events)
	if err != nil {
		return events, bosherr.WrapErrorf(err, "Finding events")
	}

	return events, nil
}

func (d DirectorImpl) Event(id string) (Event, error) {
	eventResp, err := d.client.Event(id)
	if err != nil {
		return EventImpl{}, err
	}

	return NewEventFromResp(d.client, eventResp), nil
}

func (c Client) Event(id string) (EventResp, error) {
	var event EventResp

	err := c.clientRequest.Get(fmt.Sprintf("/events/%s", id), &event)
	if err != nil {
		return event, bosherr.WrapErrorf(err, "Finding event '%s'", id)
	}

	return event, nil
}
