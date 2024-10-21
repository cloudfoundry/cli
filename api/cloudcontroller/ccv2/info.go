package ccv2

import (
	"net/http"

	"code.cloudfoundry.org/cli/v7/api/cloudcontroller"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv2/internal"
)

// APIInformation represents the information returned back from /v2/info
type APIInformation struct {

	// APIVersion is the Cloud Controller API version number.
	APIVersion string `json:"api_version"`

	// AuthorizationEndpoint is the authorization endpoint for the targeted Cloud
	// Controller.
	AuthorizationEndpoint string `json:"authorization_endpoint"`

	DopplerEndpoint string `json:"doppler_logging_endpoint"`

	LogCacheEndpoint string `json:"log_cache_endpoint"`

	MinCLIVersion string `json:"min_cli_version"`

	// MinimumRecommendedCLIVersion is the minimum CLI version number recommended
	// for the targeted Cloud Controller.
	MinimumRecommendedCLIVersion string `json:"min_recommended_cli_version"`

	// Name is the name given to the targeted Cloud Controller.
	Name string `json:"name"`

	// RoutingEndpoint is the Routing endpoint for the targeted Cloud Controller.
	RoutingEndpoint string `json:"routing_endpoint"`
}

// API returns the Cloud Controller API URL for the targeted Cloud Controller.
func (client *Client) API() string {
	return client.cloudControllerURL
}

// APIVersion returns Cloud Controller API Version for the targeted Cloud
// Controller.
func (client *Client) APIVersion() string {
	return client.cloudControllerAPIVersion
}

// AuthorizationEndpoint returns the authorization endpoint for the targeted
// Cloud Controller.
func (client *Client) AuthorizationEndpoint() string {
	return client.authorizationEndpoint
}

// DopplerEndpoint returns the Doppler endpoint for the targetd Cloud
// Controller.
func (client *Client) DopplerEndpoint() string {
	return client.dopplerEndpoint
}

func (client *Client) LogCacheEndpoint() string {
	return client.logCacheEndpoint
}

// Info returns back endpoint and API information from /v2/info.
func (client *Client) Info() (APIInformation, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetInfoRequest,
	})
	if err != nil {
		return APIInformation{}, nil, err
	}

	var info APIInformation
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &info,
	}

	err = client.connection.Make(request, &response)
	if unknownSourceErr, ok := err.(ccerror.UnknownHTTPSourceError); ok && unknownSourceErr.StatusCode == http.StatusNotFound {
		return APIInformation{}, nil, ccerror.APINotFoundError{URL: client.cloudControllerURL}
	}
	return info, response.Warnings, err
}

// MinCLIVersion returns the minimum CLI version required for the targeted
// Cloud Controller
func (client *Client) MinCLIVersion() string {
	return client.minCLIVersion
}

// RoutingEndpoint returns the Routing endpoint for the targeted Cloud
// Controller.
func (client *Client) RoutingEndpoint() string {
	return client.routingEndpoint
}

// Info represents a GET response from the '/' endpoint of the cloud
// controller API.
type Info struct {
	// Links is a list of top level Cloud Controller APIs.
	Links InfoLinks `json:"links"`
}

type InfoLinks struct {
	LogCache APILink `json:"log_cache"`
}

// rootResponse returns the CC API root document.
func (client *Client) RootResponse() (Info, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRootRequest,
	})
	if err != nil {
		return Info{}, nil, err
	}

	var rootResult Info
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &rootResult,
	}

	err = client.connection.Make(request, &response)
	if unknownSourceErr, ok := err.(ccerror.UnknownHTTPSourceError); ok && unknownSourceErr.StatusCode == http.StatusNotFound {
		return Info{}, nil, ccerror.APINotFoundError{URL: client.cloudControllerURL}
	}
	return rootResult, response.Warnings, err
}
