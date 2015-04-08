package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/net"
)

type EndpointRepository interface {
	UpdateEndpoint(endpoint string) (finalEndpoint string, apiErr error)
}

type RemoteEndpointRepository struct {
	config  core_config.ReadWriter
	gateway net.Gateway
}

type endpointResource struct {
	ApiVersion               string `json:"api_version"`
	AuthorizationEndpoint    string `json:"authorization_endpoint"`
	LoggregatorEndpoint      string `json:"logging_endpoint"`
	MinCliVersion            string `json:"min_cli_version"`
	MinRecommendedCliVersion string `json:"min_recommended_cli_version"`
}

func NewEndpointRepository(config core_config.ReadWriter, gateway net.Gateway) (repo RemoteEndpointRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo RemoteEndpointRepository) UpdateEndpoint(endpoint string) (finalEndpoint string, apiErr error) {
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

func (repo RemoteEndpointRepository) attemptUpdate(endpoint string) error {
	serverResponse := new(endpointResource)
	err := repo.gateway.GetResource(endpoint+"/v2/info", &serverResponse)
	if err != nil {
		return err
	}

	if endpoint != repo.config.ApiEndpoint() {
		repo.config.ClearSession()
	}

	repo.config.SetApiEndpoint(endpoint)
	repo.config.SetApiVersion(serverResponse.ApiVersion)
	repo.config.SetAuthenticationEndpoint(serverResponse.AuthorizationEndpoint)
	repo.config.SetMinCliVersion(serverResponse.MinCliVersion)
	repo.config.SetMinRecommendedCliVersion(serverResponse.MinRecommendedCliVersion)

	if serverResponse.LoggregatorEndpoint == "" {
		repo.config.SetLoggregatorEndpoint(defaultLoggregatorEndpoint(endpoint))
	} else {
		repo.config.SetLoggregatorEndpoint(serverResponse.LoggregatorEndpoint)
	}

	//* 3/5/15: loggregator endpoint will be renamed to doppler eventually,
	//          we just have to use the loggregator endpoint as doppler for now
	repo.config.SetDopplerEndpoint(strings.Replace(repo.config.LoggregatorEndpoint(), "loggregator", "doppler", 1))

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
