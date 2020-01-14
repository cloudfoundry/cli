package v7action

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/util/generic"
)

type Event struct {
	GUID        string
	Time        time.Time
	Type        string
	ActorName   string
	Description string
}

func (actor Actor) GetRecentEventsByApplicationNameAndSpace(appName string, spaceGUID string) ([]Event, Warnings, error) {
	var allWarnings Warnings

	app, appWarnings, appErr := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, appWarnings...)
	if appErr != nil {
		return nil, allWarnings, appErr
	}

	ccEvents, warnings, err := actor.CloudControllerClient.GetEvents(
		ccv3.Query{Key: ccv3.TargetGUIDFilter, Values: []string{app.GUID}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
	)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return nil, allWarnings, err
	}

	var events []Event
	for _, ccEvent := range ccEvents {
		events = append(events, Event{
			GUID:        ccEvent.GUID,
			Time:        ccEvent.CreatedAt,
			Type:        ccEvent.Type,
			ActorName:   ccEvent.ActorName,
			Description: generateDescription(ccEvent.Data),
		})

	}

	return events, allWarnings, nil
}

var knownMetadataKeys = []string{
	"index",
	"reason",
	"cell_id",
	"instance",
	"exit_description",
	"exit_status",
	"recursive",
	"disk_quota",
	"instances",
	"memory",
	"state",
	"command",
	"environment_json",
}

func generateDescription(data map[string]interface{}) string {
	mappedData := generic.NewMap(data)
	if mappedData.Has("request") {
		mappedData = generic.NewMap(mappedData.Get("request"))
	}
	return formatDescription(mappedData, knownMetadataKeys)
}

func formatDescription(metadata generic.Map, keys []string) string {
	parts := []string{}
	for _, key := range keys {
		value := metadata.Get(key)
		if value != nil {
			parts = append(parts, fmt.Sprintf("%s: %s", key, formatDescriptionPart(value)))
		}
	}
	return strings.Join(parts, ", ")
}

func formatDescriptionPart(val interface{}) string {
	switch val := val.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, byte('f'), -1, 64)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%s", val)
	}
}
