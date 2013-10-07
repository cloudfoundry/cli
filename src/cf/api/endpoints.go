package api

import (
	"cf/configuration"
	"cf/net"
)

type EndpointRepository interface {
	UpdateEndpoint(endpoint string) (apiResponse net.ApiResponse)
}

type RemoteEndpointRepository struct {
	config     *configuration.Configuration
	gateway    net.Gateway
	configRepo configuration.ConfigurationRepository
}

func NewEndpointRepository(config *configuration.Configuration, gateway net.Gateway, configRepo configuration.ConfigurationRepository) (repo RemoteEndpointRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.configRepo = configRepo
	return
}

func (repo RemoteEndpointRepository) UpdateEndpoint(endpoint string) (apiResponse net.ApiResponse) {
	request, apiResponse := repo.gateway.NewRequest("GET", endpoint+"/v2/info", "", nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	scheme := request.URL.Scheme
	if scheme != "http" && scheme != "https" {
		apiResponse = net.NewApiStatusWithMessage("API endpoints should start with https:// or http://")
		return
	}

	type infoResponse struct {
		ApiVersion            string `json:"api_version"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
	}

	serverResponse := new(infoResponse)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, &serverResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	repo.configRepo.ClearSession()
	repo.config.Target = endpoint
	repo.config.ApiVersion = serverResponse.ApiVersion
	repo.config.AuthorizationEndpoint = serverResponse.AuthorizationEndpoint

	err := repo.configRepo.Save()
	if err != nil {
		apiResponse = net.NewApiStatusWithMessage(err.Error())
	}

	return
}
