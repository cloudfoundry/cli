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
	SSHOAuthClient           string `json:"app_ssh_oauth_client"`
	RoutingApiEndpoint       string `json:"routing_endpoint"`
}

func NewEndpointRepository(config core_config.ReadWriter, gateway net.Gateway) EndpointRepository {
	r := &RemoteEndpointRepository{
		config:  config,
		gateway: gateway,
	}
	return r
}

func (repo RemoteEndpointRepository) UpdateEndpoint(endpoint string) (string, error) {
	if strings.HasPrefix(endpoint, "http") {
		err := repo.attemptUpdate(endpoint)
		if err != nil {
			repo.config.SetApiEndpoint("")
			return "", err
		}

		return endpoint, nil
	}

	finalEndpoint := "https://" + endpoint
	err := repo.attemptUpdate(finalEndpoint)
	if err != nil {
		if _, ok := err.(*errors.InvalidSSLCert); ok {
			repo.config.SetApiEndpoint("")
			return "", err
		}

		finalEndpoint = "http://" + endpoint
		err = repo.attemptUpdate(finalEndpoint)
		if err != nil {
			repo.config.SetApiEndpoint("")
			return "", err
		}
	}

	return finalEndpoint, err
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
	repo.config.SetSSHOAuthClient(serverResponse.SSHOAuthClient)
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
	repo.config.SetRoutingApiEndpoint(serverResponse.RoutingApiEndpoint)

	return nil
}

// FIXME: needs semantic versioning
func defaultLoggregatorEndpoint(apiEndpoint string) string {
	matches := endpointDomainRegex.FindStringSubmatch(apiEndpoint)
	url := fmt.Sprintf("ws%s://loggregator.%s", matches[1], matches[2])
	if url[0:3] == "wss" {
		return url + ":443"
	}
	return url + ":80"
}

var endpointDomainRegex = regexp.MustCompile(`^http(s?)://[^\.]+\.([^:]+)`)
