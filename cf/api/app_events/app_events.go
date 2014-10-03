package app_events

import (
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/api/strategy"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type AppEventsRepository interface {
	RecentEvents(appGuid string, limit int64) ([]models.EventFields, error)
}

type CloudControllerAppEventsRepository struct {
	config   core_config.Reader
	gateway  net.Gateway
	strategy strategy.EndpointStrategy
}

func NewCloudControllerAppEventsRepository(config core_config.Reader, gateway net.Gateway, strategy strategy.EndpointStrategy) CloudControllerAppEventsRepository {
	return CloudControllerAppEventsRepository{
		config:   config,
		gateway:  gateway,
		strategy: strategy,
	}
}

func (repo CloudControllerAppEventsRepository) RecentEvents(appGuid string, limit int64) ([]models.EventFields, error) {
	count := int64(0)
	events := make([]models.EventFields, 0, limit)
	apiErr := repo.listEvents(appGuid, limit, func(eventField models.EventFields) bool {
		count++
		events = append(events, eventField)
		return count < limit
	})

	return events, apiErr
}

func (repo CloudControllerAppEventsRepository) listEvents(appGuid string, limit int64, cb func(models.EventFields) bool) error {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.strategy.EventsURL(appGuid, limit),
		repo.strategy.EventsResource(),

		func(resource interface{}) bool {
			return cb(resource.(resources.EventResource).ToFields())
		})
}
