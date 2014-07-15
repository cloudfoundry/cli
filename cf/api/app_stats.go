package api

import (
	"fmt"
	"strconv"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type AppStatsRepository interface {
	GetStats(appGuid string) (stats []models.AppStatsFields, apiErr error)
}

type CloudControllerAppStatsRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerAppStatsRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerAppStatsRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppStatsRepository) GetStats(guid string) (stats []models.AppStatsFields, apiErr error) {
	path := fmt.Sprintf("%s/v2/apps/%s/stats", repo.config.ApiEndpoint(), guid)
	statsResponse := map[string]models.AppStatsFields{}
	apiErr = repo.gateway.GetResource(path, &statsResponse)
	if apiErr != nil {
		return
	}

	stats = make([]models.AppStatsFields, len(statsResponse), len(statsResponse))
	for key, value := range statsResponse {
		index, err := strconv.Atoi(key)
		if err != nil {
			continue
		}

		stats[index] = value
	}
	return
}
