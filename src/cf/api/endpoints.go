package api

import (
	"cf/configuration"
	"cf/errors"
	"cf/net"
	"regexp"
	"strings"
)

type EndpointRepository interface {
	UpdateEndpoint(endpoint string) (finalEndpoint string, apiErr errors.Error)
	GetLoggregatorEndpoint() (endpoint string, apiErr errors.Error)
	GetUAAEndpoint() (endpoint string, apiErr errors.Error)
	GetCloudControllerEndpoint() (endpoint string, apiErr errors.Error)
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

func (repo RemoteEndpointRepository) UpdateEndpoint(endpoint string) (finalEndpoint string, apiErr errors.Error) {
	endpointMissingScheme := !strings.HasPrefix(endpoint, "https://") && !strings.HasPrefix(endpoint, "http://")

	if endpointMissingScheme {
		finalEndpoint = "https://" + endpoint
		apiErr = repo.attemptUpdate(finalEndpoint)

		if apiErr != nil {
			finalEndpoint = "http://" + endpoint
			apiErr = repo.attemptUpdate(finalEndpoint)
		}
		return
	}

	finalEndpoint = endpoint

	apiErr = repo.attemptUpdate(finalEndpoint)

	return
}

func (repo RemoteEndpointRepository) attemptUpdate(endpoint string) (apiErr errors.Error) {
	request, apiErr := repo.gateway.NewRequest("GET", endpoint+"/v2/info", "", nil)
	if apiErr != nil {
		return
	}

	serverResponse := new(struct {
		ApiVersion            string `json:"api_version"`
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		LoggregatorEndpoint   string `json:"logging_endpoint"`
	})
	_, apiErr = repo.gateway.PerformRequestForJSONResponse(request, &serverResponse)
	if apiErr != nil {
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

func (repo RemoteEndpointRepository) GetLoggregatorEndpoint() (endpoint string, apiErr errors.Error) {
	if repo.config.LoggregatorEndpoint() == "" {
		if repo.config.ApiEndpoint() == "" {
			apiErr = errors.NewErrorWithMessage("Loggregator endpoint missing from config file")
		} else {
			endpoint = defaultLoggregatorEndpoint(repo.config.ApiEndpoint())
		}
	} else {
		endpoint = repo.config.LoggregatorEndpoint()
	}

	return
}

func (repo RemoteEndpointRepository) GetCloudControllerEndpoint() (endpoint string, apiErr errors.Error) {
	if repo.config.ApiEndpoint() == "" {
		apiErr = errors.NewErrorWithMessage("Target endpoint missing from config file")
		return
	}

	endpoint = repo.config.ApiEndpoint()
	return
}

func (repo RemoteEndpointRepository) GetUAAEndpoint() (endpoint string, apiErr errors.Error) {
	if repo.config.AuthorizationEndpoint() == "" {
		apiErr = errors.NewErrorWithMessage("UAA endpoint missing from config file")
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
