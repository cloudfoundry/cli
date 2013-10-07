package api

import (
	"cf/configuration"
	"cf/net"
)

type EndpointRepository interface {
	UpdateEndpoint(endpoint string) (apiStatus net.ApiStatus)
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

func (repo RemoteEndpointRepository) UpdateEndpoint(endpoint string) (apiStatus net.ApiStatus) {
	request, apiStatus := repo.gateway.NewRequest("GET", endpoint+"/v2/info", "", nil)
	if apiStatus.IsNotSuccessful() {
		return
	}

	scheme := request.URL.Scheme
	if scheme != "http" && scheme != "https" {
		apiStatus = net.NewApiStatusWithMessage("API endpoints should start with https:// or http://")
		return
	}

	type infoResponse struct {
		ApiVersion            string `json:"api_version"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
	}

	serverResponse := new(infoResponse)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, &serverResponse)
	if apiStatus.IsNotSuccessful() {
		return
	}

	repo.configRepo.ClearSession()
	repo.config.Target = endpoint
	repo.config.ApiVersion = serverResponse.ApiVersion
	repo.config.AuthorizationEndpoint = serverResponse.AuthorizationEndpoint

	err := repo.configRepo.Save()
	if err != nil {
		apiStatus = net.NewApiStatusWithMessage(err.Error())
	}

	return
}
