package appevents

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/cf/api/resources"
	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/cf/models"
	"code.cloudfoundry.org/cli/v8/cf/net"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Repository

type Repository interface {
	RecentEvents(appGUID string, limit int64) ([]models.EventFields, error)
}

type CloudControllerAppEventsRepository struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

func NewCloudControllerAppEventsRepository(config coreconfig.Reader, gateway net.Gateway) CloudControllerAppEventsRepository {
	return CloudControllerAppEventsRepository{
		config:  config,
		gateway: gateway,
	}
}

func (repo CloudControllerAppEventsRepository) RecentEvents(appGUID string, limit int64) ([]models.EventFields, error) {
	count := int64(0)
	events := make([]models.EventFields, 0, limit)
	apiErr := repo.listEvents(appGUID, limit, func(eventField models.EventFields) bool {
		count++
		events = append(events, eventField)
		return count < limit
	})

	return events, apiErr
}

func (repo CloudControllerAppEventsRepository) listEvents(appGUID string, limit int64, cb func(models.EventFields) bool) error {
	path := fmt.Sprintf("/v2/events?results-per-page=%d&order-direction=desc&q=actee:%s", limit, appGUID)
	return repo.gateway.ListPaginatedResources(
		repo.config.APIEndpoint(),
		path,
		resources.EventResourceNewV2{},

		func(resource interface{}) bool {
			return cb(resource.(resources.EventResource).ToFields())
		})
}
