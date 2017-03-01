package v3action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// IsolationSegment represents a V3 actor IsolationSegment.
type IsolationSegment ccv3.IsolationSegment

// CreateIsolationSegment returns a new Isolation Segment
func (actor Actor) CreateIsolationSegment(name string) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.CreateIsolationSegment(name)
	if _, ok := err.(cloudcontroller.UnprocessableEntityError); ok {
		return Warnings(warnings), nil
	}
	return Warnings(warnings), err
}
