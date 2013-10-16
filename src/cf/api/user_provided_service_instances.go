package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"fmt"
)

type UserProvidedServiceInstanceRepository interface {
	Create(name string, params map[string]string) (apiResponse net.ApiResponse)
	Update(serviceInstance cf.ServiceInstance, params map[string]string) (apiResponse net.ApiResponse)
}

type CCUserProvidedServiceInstanceRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCCUserProvidedServiceInstanceRepository(config *configuration.Configuration, gateway net.Gateway) (repo CCUserProvidedServiceInstanceRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CCUserProvidedServiceInstanceRepository) Create(name string, params map[string]string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/user_provided_service_instances", repo.config.Target)

	type RequestBody struct {
		Name        string            `json:"name"`
		Credentials map[string]string `json:"credentials"`
		SpaceGuid   string            `json:"space_guid"`
	}

	reqBody := RequestBody{name, params, repo.config.Space.Guid}
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error parsing response", err)
		return
	}

	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, bytes.NewReader(jsonBytes))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CCUserProvidedServiceInstanceRepository) Update(serviceInstance cf.ServiceInstance, params map[string]string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/user_provided_service_instances/%s", repo.config.Target, serviceInstance.Guid)

	type RequestBody struct {
		Credentials map[string]string `json:"credentials"`
	}

	reqBody := RequestBody{params}
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error parsing response", err)
		return
	}

	request, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(jsonBytes))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}
