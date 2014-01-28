package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strconv"
	"time"
)

const APP_EVENT_TIMESTAMP_FORMAT = "2006-01-02T15:04:05-07:00"

type PaginatedEventResources struct {
	Resources []EventResource
	NextURL   string `json:"next_url"`
}

type EventResource struct {
	Resource
	Entity EventEntity
}

func (resource EventResource) ToFields() (event cf.EventFields) {
	description := fmt.Sprintf("reason: %s, exit_status: %s", resource.Entity.ExitDescription, strconv.Itoa(resource.Entity.ExitStatus))
	event.Name = "app crashed"
	event.Timestamp = resource.Entity.Timestamp
	event.Description = description
	event.InstanceIndex = resource.Entity.InstanceIndex
	return
}

type EventEntity struct {
	Timestamp       time.Time
	ExitDescription string `json:"exit_description"`
	ExitStatus      int    `json:"exit_status"`
	InstanceIndex   int    `json:"instance_index"`
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
	fetchNext := true
	path := fmt.Sprintf("/v2/apps/%s/events", appGuid)

	for fetchNext {
		var shouldFetch bool

		url := fmt.Sprintf("%s%s", repo.config.Target, path)
		eventResources := &PaginatedEventResources{}

		apiResponse = repo.gateway.GetResource(url, repo.config.AccessToken, eventResources)
		if apiResponse.IsNotSuccessful() {
			return
		}

		events := []cf.EventFields{}
		for _, resource := range eventResources.Resources {
			events = append(events, resource.ToFields())
		}

		if len(events) > 0 {
			shouldFetch = cb(events)
		}

		path = eventResources.NextURL

		fetchNext = shouldFetch && path != ""
	}

	return
}
