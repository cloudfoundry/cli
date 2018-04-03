package task

import (
	"reflect"
	"time"

	boshuifmt "github.com/cloudfoundry/bosh-cli/ui/fmt"
)

const (
	EventTypeDeprecation = "deprecation"
	EventTypeWarning     = "warning"

	EventStateStarted    = "started"
	EventStateFinished   = "finished"
	EventStateFailed     = "failed"
	EventStateInProgress = "in_progress"
)

type Event struct {
	TaskID   int
	UnixTime int64 `json:"time"` // e.g 1451020321

	Type    string // e.g. "deprecation"
	Message string

	State string   // e.g. "started"
	Stage string   // e.g. "Preparing deployment"
	Task  string   // e.g. "Binding deployment"
	Tags  []string // e.g. ["api"]

	Total    int // e.g. 0
	Index    int // e.g. 0
	Progress int // e.g. 0

	Data  EventData
	Error *EventError

	StartEvent *Event
}

type EventData struct {
	Error string // e.g. "'api2/2' is not running after update"
}

type EventError struct {
	Code    int    // e.g. 100
	Message string // e.g. "Bosh::Director::Lock::TimeoutError"
}

func (e Event) IsSame(other Event) bool {
	return e.IsSameGroup(other) && e.Task == other.Task
}

func (e Event) IsSameGroup(other Event) bool {
	return e.IsSameTaskID(other) && len(e.Stage) != 0 && e.Stage == other.Stage && reflect.DeepEqual(e.Tags, other.Tags)
}

func (e Event) IsSameTaskID(other Event) bool {
	return e.TaskID == other.TaskID
}

func (e Event) Time() time.Time {
	return time.Unix(e.UnixTime, 0).UTC()
}

func (e Event) TimeAsStr() string {
	return e.Time().Format(boshuifmt.TimeFullFmt)
}

func (e Event) TimeAsHoursStr() string {
	return e.Time().Format(boshuifmt.TimeHoursFmt)
}

func (e Event) DurationAsStr(later Event) string {
	return boshuifmt.Duration(later.Time().Sub(e.Time()))
}

func (e Event) DurationSinceStartAsStr() string {
	if e.StartEvent != nil {
		return e.StartEvent.DurationAsStr(e)
	}
	return ""
}

func (e Event) IsWorthKeeping() bool {
	return e.State != EventStateInProgress
}
