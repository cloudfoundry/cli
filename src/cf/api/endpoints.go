package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"regexp"
	"strings"
)

const (
	authEndpointPrefix         = "login"
	uaaEndpointPrefix          = "uaa"
	loggregatorEndpointKey     = "loggregator"
	cloudControllerEndpointKey = "api"
)

type EndpointRepository interface {
	UpdateEndpoint(endpoint string) (apiResponse net.ApiResponse)
	GetEndpoint(name cf.EndpointType) (endpoint string, apiResponse net.ApiResponse)
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
		apiResponse = net.NewApiResponseWithMessage("API endpoints should start with https:// or http://")
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
		apiResponse = net.NewApiResponseWithMessage(err.Error())
	}

	return
}

func (repo RemoteEndpointRepository) GetEndpoint(name cf.EndpointType) (endpoint string, apiResponse net.ApiResponse) {
	switch name {
	case cf.CloudControllerEndpointKey:
		return repo.cloudControllerEndpoint()
	case cf.UaaEndpointKey:
		return repo.uaaControllerEndpoint()
	case cf.LoggregatorEndpointKey:
		return repo.loggregatorEndpoint()
	}

	apiResponse = net.NewNotFoundApiResponse("Endpoint type %s is unkown", string(name))

	return
}

func (repo RemoteEndpointRepository) cloudControllerEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	if repo.config.Target == "" {
		apiResponse = net.NewApiResponseWithMessage("Endpoint missing from config file")
		return
	}

	endpoint = repo.config.Target
	return
}

func (repo RemoteEndpointRepository) uaaControllerEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	if repo.config.AuthorizationEndpoint == "" {
		apiResponse = net.NewApiResponseWithMessage("Endpoint missing from config file")
		return
	}

	endpoint = strings.Replace(repo.config.AuthorizationEndpoint, authEndpointPrefix, uaaEndpointPrefix, 1)

	return
}

func (repo RemoteEndpointRepository) loggregatorEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	if repo.config.AuthorizationEndpoint == "" {
		apiResponse = net.NewApiResponseWithMessage("Endpoint missing from config file")
		return
	}

	re := regexp.MustCompile(`^http(s?)://(.+)\/?`)

	endpoint = re.ReplaceAllString(repo.config.AuthorizationEndpoint, "ws${1}://loggregator.${2}")
	endpoint = strings.Replace(endpoint, uaaEndpointPrefix+".", "", 1)
	endpoint = endpoint + ":4443"

	return
}
