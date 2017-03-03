package v3action

import (
	"fmt"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// IsolationSegment represents a V3 actor IsolationSegment.
type IsolationSegment ccv3.IsolationSegment

// IsolationSegmentNotFoundError represents the error that occurs when the
// isolation segment is not found.
type IsolationSegmentNotFoundError struct {
	Name string
}

func (e IsolationSegmentNotFoundError) Error() string {
	return fmt.Sprintf("Isolation Segment '%s' not found.", e.Name)
}

// IsolationSegmentAlreadyExistsError gets returned when an isolation segment
// already exists.
type IsolationSegmentAlreadyExistsError struct {
	Name string
}

func (e IsolationSegmentAlreadyExistsError) Error() string {
	return fmt.Sprintf("Isolation Segment '%s' already exists.", e.Name)
}

// CreateIsolationSegmentByName creates a given isolation segment.
func (actor Actor) CreateIsolationSegmentByName(name string) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.CreateIsolationSegment(name)
	if _, ok := err.(cloudcontroller.UnprocessableEntityError); ok {
		return Warnings(warnings), IsolationSegmentAlreadyExistsError{Name: name}
	}
	return Warnings(warnings), err
}

// DeleteIsolationSegmentByName deletes the given isolation segment.
func (actor Actor) DeleteIsolationSegmentByName(name string) (Warnings, error) {
	isolationSegment, warnings, err := actor.GetIsolationSegmentByName(name)
	allWarnings := append(Warnings{}, warnings...)
	if err != nil {
		return allWarnings, err
	}

	apiWarnings, err := actor.CloudControllerClient.DeleteIsolationSegment(isolationSegment.GUID)
	return append(allWarnings, apiWarnings...), err
}

// GetIsolationSegmentByName returns the requested isolation segment.
func (actor Actor) GetIsolationSegmentByName(name string) (IsolationSegment, Warnings, error) {
	isolationSegments, warnings, err := actor.CloudControllerClient.GetIsolationSegments(url.Values{ccv3.NameFilter: []string{name}})
	if err != nil {
		return IsolationSegment{}, Warnings(warnings), err
	}

	if len(isolationSegments) == 0 {
		return IsolationSegment{}, Warnings(warnings), IsolationSegmentNotFoundError{Name: name}
	}

	return IsolationSegment(isolationSegments[0]), Warnings(warnings), nil
}
