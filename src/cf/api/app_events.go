package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
)

const APP_EVENT_TIMESTAMP_FORMAT = "2006-01-02T15:04:05-07:00"

type AppEventsRepository interface {
	ListEvents(app cf.Application) (events []cf.Event, apiResponse net.ApiResponse)
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

func (repo CloudControllerAppEventsRepository) ListEvents(app cf.Application) (events []cf.Event, apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/apps/%s/events", repo.config.Target, app.Guid)
	eventResources := repo.doRequest(url)

	for {
		for _, resource := range eventResources.Resources {
			events = append(events, cf.Event{
				Timestamp:       resource.Entity.Timestamp,
				ExitDescription: resource.Entity.ExitDescription,
				ExitStatus:      resource.Entity.ExitStatus,
				InstanceIndex:   resource.Entity.InstanceIndex,
			})
		}

		if eventResources.NextURL == "" {
			break
		}
		url = fmt.Sprintf("%s%s", repo.config.Target, eventResources.NextURL)

		eventResources = repo.doRequest(url)
	}

	return
}

func (repo CloudControllerAppEventsRepository) doRequest(url string) (eventResources *PaginatedEventResources) {
	request, apiResponse := repo.gateway.NewRequest("GET", url, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	eventResources = &PaginatedEventResources{}
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, eventResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	return
}
