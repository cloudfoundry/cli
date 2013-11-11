package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
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

type EventEntity struct {
	Timestamp       time.Time
	ExitDescription string `json:"exit_description"`
	ExitStatus      int    `json:"exit_status"`
	InstanceIndex   int    `json:"instance_index"`
}

type AppEventsRepository interface {
	ListEvents(app cf.Application) (events chan []cf.Event, errorChan chan net.ApiResponse)
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

func (repo CloudControllerAppEventsRepository) ListEvents(app cf.Application) (eventChan chan []cf.Event, errorChan chan net.ApiResponse) {

	eventChan = make(chan []cf.Event, 4)
	errorChan = make(chan net.ApiResponse, 1)

	go func() {
		path := fmt.Sprintf("/v2/apps/%s/events", app.Guid)
		for path != "" {
			url := fmt.Sprintf("%s%s", repo.config.Target, path)
			eventResources := &PaginatedEventResources{}
			apiResponse := repo.gateway.GetResource(url, repo.config.AccessToken, eventResources)
			if apiResponse.IsNotSuccessful() {
				errorChan <- apiResponse
				close(eventChan)
				close(errorChan)
				return
			}

			events := []cf.Event{}
			for _, resource := range eventResources.Resources {
				events = append(events, cf.Event{
					Timestamp:       resource.Entity.Timestamp,
					ExitDescription: resource.Entity.ExitDescription,
					ExitStatus:      resource.Entity.ExitStatus,
					InstanceIndex:   resource.Entity.InstanceIndex,
				})
			}
			eventChan <- events
			path = eventResources.NextURL
		}
		close(eventChan)
		close(errorChan)
	}()

	return
}
