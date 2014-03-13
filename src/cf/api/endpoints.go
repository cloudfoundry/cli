package api

import (
	"cf/configuration"
	"cf/errors"
	"cf/net"
	"fmt"
	"regexp"
	"strings"
)

type EndpointRepository interface {
	UpdateEndpoint(endpoint string) (finalEndpoint string, apiErr errors.Error)
}

type RemoteEndpointRepository struct {
	config  configuration.ReadWriter
	gateway net.Gateway
}

type endpointResource struct {
	ApiVersion            string `json:"api_version"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	LoggregatorEndpoint   string `json:"logging_endpoint"`
}

func NewEndpointRepository(config configuration.ReadWriter, gateway net.Gateway) (repo RemoteEndpointRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo RemoteEndpointRepository) UpdateEndpoint(endpoint string) (finalEndpoint string, apiErr errors.Error) {
	defer func() {
		if apiErr != nil {
			repo.config.SetApiEndpoint("")
		}
	}()

	endpointMissingScheme := !strings.HasPrefix(endpoint, "https://") && !strings.HasPrefix(endpoint, "http://")

	if endpointMissingScheme {
		finalEndpoint := "https://" + endpoint
		apiErr := repo.attemptUpdate(finalEndpoint)

		switch apiErr.(type) {
		case nil:
		case *errors.InvalidSSLCert:
			return endpoint, apiErr
		default:
			finalEndpoint = "http://" + endpoint
			apiErr = repo.attemptUpdate(finalEndpoint)
		}

		return finalEndpoint, apiErr
	} else {
		apiErr := repo.attemptUpdate(endpoint)
		return endpoint, apiErr
	}
}

func (repo RemoteEndpointRepository) attemptUpdate(endpoint string) errors.Error {
	request, err := repo.gateway.NewRequest("GET", endpoint+"/v2/info", "", nil)
	if err != nil {
		return err
	}

	serverResponse := new(endpointResource)
	_, err = repo.gateway.PerformRequestForJSONResponse(request, &serverResponse)
	if err != nil {
		return err
	}

	if endpoint != repo.config.ApiEndpoint() {
		repo.config.ClearSession()
	}

	repo.config.SetApiEndpoint(endpoint)
	repo.config.SetApiVersion(serverResponse.ApiVersion)
	repo.config.SetAuthenticationEndpoint(serverResponse.AuthorizationEndpoint)
	repo.config.SetUaaEndpoint(serverResponse.AuthorizationEndpoint)

	if serverResponse.LoggregatorEndpoint == "" {
		repo.config.SetLoggregatorEndpoint(defaultLoggregatorEndpoint(endpoint))
	} else {
		repo.config.SetLoggregatorEndpoint(serverResponse.LoggregatorEndpoint)
	}

	return nil
}

// FIXME: needs semantic versioning
func defaultLoggregatorEndpoint(apiEndpoint string) string {
	matches := endpointDomainRegex.FindStringSubmatch(apiEndpoint)
	url := fmt.Sprintf("ws%s://loggregator.%s", matches[1], matches[2])
	if url[0:3] == "wss" {
		return url + ":443"
	} else {
		return url + ":80"
	}
}

var endpointDomainRegex = regexp.MustCompile(`^http(s?)://[^\.]+\.([^:]+)`)
