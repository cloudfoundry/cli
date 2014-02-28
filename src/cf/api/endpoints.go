package api

import (
	"cf/configuration"
	"cf/errors"
	"cf/net"
	"regexp"
	"strings"
)

type EndpointRepository interface {
	UpdateEndpoint(endpoint string) (finalEndpoint string, apiResponse errors.Error)
	GetLoggregatorEndpoint() (endpoint string, apiResponse errors.Error)
	GetUAAEndpoint() (endpoint string, apiResponse errors.Error)
	GetCloudControllerEndpoint() (endpoint string, apiResponse errors.Error)
}

type RemoteEndpointRepository struct {
	config  configuration.ReadWriter
	gateway net.Gateway
}

func NewEndpointRepository(config configuration.ReadWriter, gateway net.Gateway) (repo RemoteEndpointRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo RemoteEndpointRepository) UpdateEndpoint(endpoint string) (finalEndpoint string, apiResponse errors.Error) {
	endpointMissingScheme := !strings.HasPrefix(endpoint, "https://") && !strings.HasPrefix(endpoint, "http://")

	if endpointMissingScheme {
		finalEndpoint = "https://" + endpoint
		apiResponse = repo.attemptUpdate(finalEndpoint)

		if apiResponse != nil {
			finalEndpoint = "http://" + endpoint
			apiResponse = repo.attemptUpdate(finalEndpoint)
		}
		return
	}

	finalEndpoint = endpoint

	apiResponse = repo.attemptUpdate(finalEndpoint)

	return
}

func (repo RemoteEndpointRepository) attemptUpdate(endpoint string) (apiResponse errors.Error) {
	request, apiResponse := repo.gateway.NewRequest("GET", endpoint+"/v2/info", "", nil)
	if apiResponse != nil {
		return
	}

	serverResponse := new(struct {
		ApiVersion            string `json:"api_version"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		LoggregatorEndpoint   string `json:"logging_endpoint"`
	})
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, &serverResponse)
	if apiResponse != nil {
		return
	}

	if endpoint != repo.config.ApiEndpoint() {
		repo.config.ClearSession()
	}

	repo.config.SetApiEndpoint(endpoint)
	repo.config.SetApiVersion(serverResponse.ApiVersion)
	repo.config.SetAuthorizationEndpoint(serverResponse.AuthorizationEndpoint)
	repo.config.SetLoggregatorEndpoint(serverResponse.LoggregatorEndpoint)

	return
}

func (repo RemoteEndpointRepository) GetLoggregatorEndpoint() (endpoint string, apiResponse errors.Error) {
	if repo.config.LoggregatorEndpoint() == "" {
		if repo.config.ApiEndpoint() == "" {
			apiResponse = errors.NewErrorWithMessage("Loggregator endpoint missing from config file")
		} else {
			endpoint = defaultLoggregatorEndpoint(repo.config.ApiEndpoint())
		}
	} else {
		endpoint = repo.config.LoggregatorEndpoint()
	}

	return
}

func (repo RemoteEndpointRepository) GetCloudControllerEndpoint() (endpoint string, apiResponse errors.Error) {
	if repo.config.ApiEndpoint() == "" {
		apiResponse = errors.NewErrorWithMessage("Target endpoint missing from config file")
		return
	}

	endpoint = repo.config.ApiEndpoint()
	return
}

func (repo RemoteEndpointRepository) GetUAAEndpoint() (endpoint string, apiResponse errors.Error) {
	if repo.config.AuthorizationEndpoint() == "" {
		apiResponse = errors.NewErrorWithMessage("UAA endpoint missing from config file")
		return
	}

	endpoint = strings.Replace(repo.config.AuthorizationEndpoint(), "login", "uaa", 1)

	return
}

// FIXME: needs semantic versioning
func defaultLoggregatorEndpoint(apiEndpoint string) string {
	url := endpointDomainRegex.ReplaceAllString(apiEndpoint, "ws${1}://loggregator.${2}")
	if url[0:3] == "wss" {
		return url + ":4443"
	} else {
		return url + ":80"
	}
}

var endpointDomainRegex = regexp.MustCompile(`^http(s?)://[^\.]+\.(.+)\/?`)
