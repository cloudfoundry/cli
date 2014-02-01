package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"generic"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const APP_EVENT_TIMESTAMP_FORMAT = "2006-01-02T15:04:05-07:00"

type EventResourceOldV2 struct {
	Resource
	Entity struct {
		Timestamp       time.Time
		ExitDescription string `json:"exit_description"`
		ExitStatus      int    `json:"exit_status"`
		InstanceIndex   int    `json:"instance_index"`
	}
}

func (resource EventResourceOldV2) ToFields() interface{} {
	event := cf.EventFields{}
	description := fmt.Sprintf("instance: %d, reason: %s, exit_status: %s", resource.Entity.InstanceIndex, resource.Entity.ExitDescription, strconv.Itoa(resource.Entity.ExitStatus))
	event.Guid = resource.Metadata.Guid
	event.Name = "app crashed"
	event.Timestamp = resource.Entity.Timestamp
	event.Description = description
	return event
}

type EventResourceNewV2 struct {
	Resource
	Entity struct {
		Timestamp time.Time
		Type      string
		Metadata  map[string]interface{}
	}
}

var KNOWN_METADATA_KEYS = []string{
	"index",
	"reason",
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

func (resource EventResourceNewV2) ToFields() interface{} {
	event := cf.EventFields{}
	event.Guid = resource.Metadata.Guid
	event.Name = resource.Entity.Type
	event.Timestamp = resource.Entity.Timestamp

	metadata := generic.NewMap(resource.Entity.Metadata)
	if metadata.Has("request") {
		metadata = generic.NewMap(metadata.Get("request"))
	}
	event.Description = formatDescription(metadata, KNOWN_METADATA_KEYS)

	return event
}

func formatDescription(metadata generic.Map, keys []string) string {
	parts := []string{}
	for _, key := range keys {
		value := metadata.Get(key)
		if value != nil {
			parts = append(parts, fmt.Sprintf("%s: %s", key, String(value)))
		}
	}
	return strings.Join(parts, ", ")
}

func String(val interface{}) string {
	switch val := val.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, byte('f'), -1, 64)
	case bool:
		if val {
			return "true"
		} else {
			return "false"
		}
	default:
		return fmt.Sprintf("%s", val)
	}
}

type ListEventsCallback func(events []cf.EventFields) (fetchNext bool)

type AppEventsRepository interface {
	ListEvents(appGuid string, cb ListEventsCallback) net.ApiResponse
}

type CloudControllerAppEventsRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerAppEventsRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerAppEventsRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppEventsRepository) ListEvents(appGuid string, cb ListEventsCallback) (apiResponse net.ApiResponse) {
	apiResponse = repo.newV2ListEvents(appGuid, cb)
	if apiResponse.IsNotFound() {
		apiResponse = repo.oldV2ListEvents(appGuid, cb)
	}
	return
}

func (repo CloudControllerAppEventsRepository) newV2ListEvents(appGuid string, cb ListEventsCallback) net.ApiResponse {
	path := fmt.Sprintf("/v2/events?q=%s", url.QueryEscape(fmt.Sprintf("actee:%s", appGuid)))
	return repo.listEvents(path, NewPaginatedResources(EventResourceNewV2{}), cb)
}

func (repo CloudControllerAppEventsRepository) oldV2ListEvents(appGuid string, cb ListEventsCallback) net.ApiResponse {
	path := fmt.Sprintf("/v2/apps/%s/events", appGuid)
	return repo.listEvents(path, NewPaginatedResources(EventResourceOldV2{}), cb)
}

func (repo CloudControllerAppEventsRepository) listEvents(path string, eventResources PaginatedResources, cb ListEventsCallback) (apiResponse net.ApiResponse) {
	fetchNext := true

	for fetchNext {
		var shouldFetch bool

		url := fmt.Sprintf("%s%s", repo.config.Target, path)

		apiResponse = repo.gateway.GetResource(url, repo.config.AccessToken, &eventResources)
		if apiResponse.IsNotSuccessful() {
			return
		}

		events := make([]cf.EventFields, 0, len(eventResources.Resources()))
		for _, resource := range eventResources.Resources() {
			events = append(events, resource.ToFields().(cf.EventFields))
		}

		if len(events) > 0 {
			shouldFetch = cb(events)
		}

		path = eventResources.NextURL

		fetchNext = shouldFetch && path != ""
	}

	return
}
