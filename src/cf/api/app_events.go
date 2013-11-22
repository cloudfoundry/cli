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
	ListEvents(appGuid string) (events chan []cf.EventFields, statusChan chan net.ApiResponse)
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

func (repo CloudControllerAppEventsRepository) ListEvents(appGuid string) (eventChan chan []cf.EventFields, statusChan chan net.ApiResponse) {

	eventChan = make(chan []cf.EventFields, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		path := fmt.Sprintf("/v2/apps/%s/events", appGuid)
		for path != "" {
			url := fmt.Sprintf("%s%s", repo.config.Target, path)
			eventResources := &PaginatedEventResources{}
			apiResponse := repo.gateway.GetResource(url, repo.config.AccessToken, eventResources)
			if apiResponse.IsNotSuccessful() {
				statusChan <- apiResponse
				close(eventChan)
				close(statusChan)
				return
			}

			events := []cf.EventFields{}
			for _, resource := range eventResources.Resources {
				events = append(events, cf.EventFields{
					Timestamp:       resource.Entity.Timestamp,
					ExitDescription: resource.Entity.ExitDescription,
					ExitStatus:      resource.Entity.ExitStatus,
					InstanceIndex:   resource.Entity.InstanceIndex,
				})
			}
			if len(events) > 0 {
				eventChan <- events
			}

			path = eventResources.NextURL
		}
		close(eventChan)
		close(statusChan)
	}()

	return
}
