package api

import (
	"cf/api/resources"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
	"strconv"
)

type AppEventsRepository interface {
	RecentEvents(appGuid string, limit uint) ([]models.EventFields, error)
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

func (repo CloudControllerAppEventsRepository) RecentEvents(appGuid string, limit uint) ([]models.EventFields, error) {
	count := uint(0)
	events := make([]models.EventFields, 0, limit)
	apiErr := repo.listEvents(appGuid, url.Values{
		"order-direction":  []string{"desc"},
		"results-per-page": []string{strconv.FormatUint(uint64(limit), 10)},
	}, func(eventField models.EventFields) bool {
		count++
		events = append(events, eventField)
		return count < limit
	})

	return events, apiErr
}

func (repo CloudControllerAppEventsRepository) listEvents(appGuid string, queryParams url.Values, cb func(models.EventFields) bool) error {
	queryParams.Set("q", "actee:"+appGuid)
	apiErr := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		fmt.Sprintf("/v2/events?%s", queryParams.Encode()),
		resources.EventResourceNewV2{},
		func(resource interface{}) bool {
			return cb(resource.(resources.EventResourceNewV2).ToFields())
		})

	// FIXME: needs semantic API version
	switch apiErr.(type) {
	case *errors.HttpNotFoundError:
		apiErr = repo.gateway.ListPaginatedResources(
			repo.config.ApiEndpoint(),
			repo.config.AccessToken(),
			fmt.Sprintf("/v2/apps/%s/events", appGuid),
			resources.EventResourceOldV2{},
			func(resource interface{}) bool {
				return cb(resource.(resources.EventResourceOldV2).ToFields())
			})
	}

	return apiErr
}
