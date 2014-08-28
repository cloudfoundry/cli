package feature_flag

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type FeatureFlagRepository interface {
	List() ([]models.FeatureFlag, error)
}

type CloudControllerFeatureFlagRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerFeatureFlagRepository(config configuration.Reader, gateway net.Gateway) CloudControllerFeatureFlagRepository {
	return CloudControllerFeatureFlagRepository{
		config:  config,
		gateway: gateway,
	}
}

func (repo CloudControllerFeatureFlagRepository) List() ([]models.FeatureFlag, error) {
	flags := []models.FeatureFlag{}
	apiError := repo.gateway.GetResource(
		fmt.Sprintf("%s/v2/config/feature_flags", repo.config.ApiEndpoint()),
		&flags)

	if apiError != nil {
		return nil, apiError
	}

	return flags, nil
}
