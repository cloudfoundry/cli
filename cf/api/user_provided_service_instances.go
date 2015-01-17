package api

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type UserProvidedServiceInstanceRepository interface {
	Create(name, drainUrl string, params map[string]interface{}) (apiErr error)
	Update(serviceInstanceFields models.ServiceInstanceFields) (apiErr error)
	GetSummaries() (models.UserProvidedServiceSummary, error)
}

type CCUserProvidedServiceInstanceRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCCUserProvidedServiceInstanceRepository(config core_config.Reader, gateway net.Gateway) (repo CCUserProvidedServiceInstanceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CCUserProvidedServiceInstanceRepository) Create(name, drainUrl string, params map[string]interface{}) (apiErr error) {
	path := "/v2/user_provided_service_instances"

	jsonBytes, err := json.Marshal(models.UserProvidedService{
		Name:           name,
		Credentials:    params,
		SpaceGuid:      repo.config.SpaceFields().Guid,
		SysLogDrainUrl: drainUrl,
	})

	if err != nil {
		apiErr = errors.NewWithError("Error parsing response", err)
		return
	}

	return repo.gateway.CreateResource(repo.config.ApiEndpoint(), path, bytes.NewReader(jsonBytes))
}

func (repo CCUserProvidedServiceInstanceRepository) Update(serviceInstanceFields models.ServiceInstanceFields) (apiErr error) {
	path := fmt.Sprintf("/v2/user_provided_service_instances/%s", serviceInstanceFields.Guid)

	reqBody := models.UserProvidedService{
		Credentials:    serviceInstanceFields.Params,
		SysLogDrainUrl: serviceInstanceFields.SysLogDrainUrl,
	}
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		apiErr = errors.NewWithError("Error parsing response", err)
		return
	}

	return repo.gateway.UpdateResource(repo.config.ApiEndpoint(), path, bytes.NewReader(jsonBytes))
}

func (repo CCUserProvidedServiceInstanceRepository) GetSummaries() (models.UserProvidedServiceSummary, error) {
	path := fmt.Sprintf("%s/v2/user_provided_service_instances", repo.config.ApiEndpoint())

	model := models.UserProvidedServiceSummary{}

	apiErr := repo.gateway.GetResource(path, &model)
	if apiErr != nil {
		return models.UserProvidedServiceSummary{}, apiErr
	}

	return model, nil
}
