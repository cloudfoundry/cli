package ccv2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// APIInformation represents the information returned back from /v2/info
type APIInformation struct {
	APIVersion                   string `json:"api_version"`
	AuthorizationEndpoint        string `json:"authorization_endpoint"`
	DopplerEndpoint              string `json:"doppler_logging_endpoint"`
	LoggregatorEndpoint          string `json:"logging_endpoint"`
	MinimumCLIVersion            string `json:"min_cli_version"`
	MinimumRecommendedCLIVersion string `json:"min_recommended_cli_version"`
	Name                         string `json:"name"`
	RoutingEndpoint              string `json:"routing_endpoint"`
	TokenEndpoint                string `json:"token_endpoint"`
}

// API returns the Cloud Controller API URL for the targeted Cloud Controller.
func (client *CloudControllerClient) API() string {
	return client.cloudControllerURL
}

// APIVersion returns Cloud Controller API Version for the targeted Cloud
// Controller.
func (client *CloudControllerClient) APIVersion() string {
	return client.cloudControllerAPIVersion
}

// AuthorizationEndpoint returns the authorization endpoint for the targeted
// Cloud Controller.
func (client *CloudControllerClient) AuthorizationEndpoint() string {
	return client.authorizationEndpoint
}

// DopplerEndpoint returns the Doppler endpoint for the targetd Cloud
// Controller.
func (client *CloudControllerClient) DopplerEndpoint() string {
	return client.dopplerEndpoint
}

// LoggregatorEndpoint returns the Loggregator endpoint for the targeted Cloud
// Controller.
func (client *CloudControllerClient) LoggregatorEndpoint() string {
	return client.loggregatorEndpoint
}

// RoutingEndpoint returns the Routing endpoint for the targeted Cloud
// Controller.
func (client *CloudControllerClient) RoutingEndpoint() string {
	return client.routingEndpoint
}

// TokenEndpoint returns the Token endpoint for the targeted Cloud Controller.
func (client *CloudControllerClient) TokenEndpoint() string {
	return client.tokenEndpoint
}

// Info returns back endpoint and API information from /v2/info.
func (client *CloudControllerClient) Info() (APIInformation, Warnings, error) {
	response, err := http.Get(fmt.Sprintf("http://%s/v2/info", client.cloudControllerURL))
	if err != nil {
		return APIInformation{}, nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return APIInformation{}, nil, err
	}

	var info APIInformation
	err = json.Unmarshal(body, &info)
	if err != nil {
		return APIInformation{}, nil, err
	}

	return info, nil, nil
}
