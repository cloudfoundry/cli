package api

import (
	"cf/configuration"
	"cf/net"
	"regexp"
	"strings"
)

type EndpointRepository interface {
	UpdateEndpoint(endpoint string) (finalEndpoint string, apiResponse net.ApiResponse)
	GetLoggregatorEndpoint() (endpoint string, apiResponse net.ApiResponse)
	GetUAAEndpoint() (endpoint string, apiResponse net.ApiResponse)
	GetCloudControllerEndpoint() (endpoint string, apiResponse net.ApiResponse)
}

type RemoteEndpointRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewEndpointRepository(config *configuration.Configuration, gateway net.Gateway) (repo RemoteEndpointRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo RemoteEndpointRepository) UpdateEndpoint(endpoint string) (finalEndpoint string, apiResponse net.ApiResponse) {
	endpointMissingScheme := !strings.HasPrefix(endpoint, "https://") && !strings.HasPrefix(endpoint, "http://")

	if endpointMissingScheme {
		finalEndpoint = "https://" + endpoint
		apiResponse = repo.attemptUpdate(finalEndpoint)

		if apiResponse.IsNotSuccessful() {
			finalEndpoint = "http://" + endpoint
			apiResponse = repo.attemptUpdate(finalEndpoint)
		}
		return
	}

	finalEndpoint = endpoint

	apiResponse = repo.attemptUpdate(finalEndpoint)

	return
}

func (repo RemoteEndpointRepository) attemptUpdate(endpoint string) (apiResponse net.ApiResponse) {
	request, apiResponse := repo.gateway.NewRequest("GET", endpoint+"/v2/info", "", nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	serverResponse := new(struct {
		ApiVersion            string `json:"api_version"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		LoggregatorEndpoint   string `json:"logging_endpoint"`
	})
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, &serverResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if endpoint != repo.config.Target {
		repo.config.ClearSession()
	}

	repo.config.Target = endpoint
	repo.config.ApiVersion = serverResponse.ApiVersion
	repo.config.AuthorizationEndpoint = serverResponse.AuthorizationEndpoint
	repo.config.LoggregatorEndPoint = serverResponse.LoggregatorEndpoint

	return
}

func (repo RemoteEndpointRepository) GetLoggregatorEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	if repo.config.LoggregatorEndPoint == "" {
		if repo.config.Target == "" {
			apiResponse = net.NewApiResponseWithMessage("Loggregator endpoint missing from config file")
		} else {
			endpoint = defaultLoggregatorEndpoint(repo.config.Target)
		}
	} else {
		endpoint = repo.config.LoggregatorEndPoint
	}

	return
}

func (repo RemoteEndpointRepository) GetCloudControllerEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	if repo.config.Target == "" {
		apiResponse = net.NewApiResponseWithMessage("Target endpoint missing from config file")
		return
	}

	endpoint = repo.config.Target
	return
}

func (repo RemoteEndpointRepository) GetUAAEndpoint() (endpoint string, apiResponse net.ApiResponse) {
	if repo.config.AuthorizationEndpoint == "" {
		apiResponse = net.NewApiResponseWithMessage("UAA endpoint missing from config file")
		return
	}

	endpoint = strings.Replace(repo.config.AuthorizationEndpoint, "login", "uaa", 1)

	return
}

// TODO - remove
func defaultLoggregatorEndpoint(apiEndpoint string) string {
	url := endpointDomainRegex.ReplaceAllString(apiEndpoint, "ws${1}://loggregator.${2}")
	if url[0:3] == "wss" {
		return url + ":4443"
	} else {
		return url + ":80"
	}
}

var endpointDomainRegex = regexp.MustCompile(`^http(s?)://[^\.]+\.(.+)\/?`)
