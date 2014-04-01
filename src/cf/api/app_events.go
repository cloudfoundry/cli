package api

import (
	"cf/api/resources"
	"cf/api/strategy"
	"cf/configuration"
	"cf/models"
	"cf/net"
)

type AppEventsRepository interface {
	RecentEvents(appGuid string, limit uint64) ([]models.EventFields, error)
}

type CloudControllerAppEventsRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerAppEventsRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerAppEventsRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppEventsRepository) RecentEvents(appGuid string, limit uint64) ([]models.EventFields, error) {
	count := uint64(0)
	events := make([]models.EventFields, 0, limit)
	apiErr := repo.listEvents(appGuid, limit, func(eventField models.EventFields) bool {
		count++
		events = append(events, eventField)
		return count < limit
	})

	return events, apiErr
}

func (repo CloudControllerAppEventsRepository) listEvents(appGuid string, limit uint64, cb func(models.EventFields) bool) error {
	endpointStrategy, err := strategy.NewEndpointStrategy(repo.config.ApiVersion())
	if err != nil {
		return err
	}

	url := endpointStrategy.EventsURL(appGuid, limit)
	resource := endpointStrategy.EventsResource()

	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		url,
		resource,
		func(resource interface{}) bool {
			return cb(resource.(resources.EventResource).ToFields())
		},
	)
}
