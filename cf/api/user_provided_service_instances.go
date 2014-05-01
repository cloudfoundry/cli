package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type UserProvidedServiceInstanceRepository interface {
	Create(name, drainUrl string, params map[string]string) (apiErr error)
	Update(serviceInstanceFields models.ServiceInstanceFields) (apiErr error)
}

type CCUserProvidedServiceInstanceRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCCUserProvidedServiceInstanceRepository(config configuration.Reader, gateway net.Gateway) (repo CCUserProvidedServiceInstanceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CCUserProvidedServiceInstanceRepository) Create(name, drainUrl string, params map[string]string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/user_provided_service_instances", repo.config.ApiEndpoint())

	type RequestBody struct {
		Name           string            `json:"name"`
		Credentials    map[string]string `json:"credentials"`
		SpaceGuid      string            `json:"space_guid"`
		SysLogDrainUrl string            `json:"syslog_drain_url"`
	}

	jsonBytes, err := json.Marshal(RequestBody{
		Name:           name,
		Credentials:    params,
		SpaceGuid:      repo.config.SpaceFields().Guid,
		SysLogDrainUrl: drainUrl,
	})

	if err != nil {
		apiErr = errors.NewWithError("Error parsing response", err)
		return
	}

	return repo.gateway.CreateResource(path, bytes.NewReader(jsonBytes))
}

func (repo CCUserProvidedServiceInstanceRepository) Update(serviceInstanceFields models.ServiceInstanceFields) (apiErr error) {
	path := fmt.Sprintf("%s/v2/user_provided_service_instances/%s", repo.config.ApiEndpoint(), serviceInstanceFields.Guid)

	type RequestBody struct {
		Credentials    map[string]string `json:"credentials,omitempty"`
		SysLogDrainUrl string            `json:"syslog_drain_url,omitempty"`
	}

	reqBody := RequestBody{serviceInstanceFields.Params, serviceInstanceFields.SysLogDrainUrl}
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		apiErr = errors.NewWithError("Error parsing response", err)
		return
	}

	return repo.gateway.UpdateResource(path, bytes.NewReader(jsonBytes))
}
