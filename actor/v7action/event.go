package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type Event ccv3.Event

func (actor Actor) GetRecentEventsByApplicationNameAndSpace(appName string, spaceGUID string) ([]Event, Warnings, error) {
	var allWarnings Warnings

	app, appWarnings, appErr := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, appWarnings...)
	if appErr != nil {
		return nil, allWarnings, appErr
	}

	ccEvents, warnings, err := actor.CloudControllerClient.GetEvents(
		ccv3.Query{Key: ccv3.TargetGUIDFilter, Values: []string{app.GUID}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{"-created_at"}},
	)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return nil, allWarnings, err
	}

	var events []Event
	for _, ccEvent := range ccEvents {
		events = append(events, Event(ccEvent))
	}

	return events, allWarnings, nil
}
