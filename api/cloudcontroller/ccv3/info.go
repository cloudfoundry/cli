package ccv3

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

// APIInfo represents a GET response from the '/' endpoint of the cloud
// controller API.
type APIInfo struct {
	// Links is a list of top level Cloud Controller APIs.
	Links struct {
		// AppSSH is the link for application ssh info
		AppSSH APILink `json:"app_ssh"`

		// CCV3 is the link to the Cloud Controller V3 API
		CCV3 APILink `json:"cloud_controller_v3"`

		// Logging is the link to the Logging API
		Logging APILink `json:"logging"`

		NetworkPolicyV1 APILink `json:"network_policy_v1"`

		// UAA is the link to the UAA API
		UAA APILink `json:"uaa"`
	} `json:"links"`
}

func (info APIInfo) AppSSHHostKeyFingerprint() string {
	return info.Links.AppSSH.Meta.HostKeyFingerprint
}

func (info APIInfo) AppSSHEndpoint() string {
	return info.Links.AppSSH.HREF
}

func (info APIInfo) OAuthClient() string {
	return info.Links.AppSSH.Meta.OAuthClient
}

// Logging returns the HREF for Logging.
func (info APIInfo) Logging() string {
	return info.Links.Logging.HREF
}

func (info APIInfo) NetworkPolicyV1() string {
	return info.Links.NetworkPolicyV1.HREF
}

// UAA returns the HREF for the UAA.
func (info APIInfo) UAA() string {
	return info.Links.UAA.HREF
}

// CloudControllerAPIVersion returns the version for the CloudController.
func (info APIInfo) CloudControllerAPIVersion() string {
	return info.Links.CCV3.Meta.Version
}

func (info APIInfo) ccV3Link() string {
	return info.Links.CCV3.HREF
}

// ResourceLinks represents the information returned back from /v3.
type ResourceLinks map[string]APILink

// UnmarshalJSON helps unmarshal a Cloud Controller /v3 response.
func (resources ResourceLinks) UnmarshalJSON(data []byte) error {
	var ccResourceLinks struct {
		Links map[string]APILink `json:"links"`
	}
	if err := json.Unmarshal(data, &ccResourceLinks); err != nil {
		return err
	}

	for key, val := range ccResourceLinks.Links {
		resources[key] = val
	}

	return nil
}

// Info returns endpoint and API information from /v3.
func (client *Client) Info() (APIInfo, ResourceLinks, Warnings, error) {
	rootResponse, warnings, err := client.rootResponse()
	if err != nil {
		return APIInfo{}, ResourceLinks{}, warnings, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		Method: http.MethodGet,
		URL:    rootResponse.ccV3Link(),
	})
	if err != nil {
		return APIInfo{}, ResourceLinks{}, warnings, err
	}

	info := ResourceLinks{} // Explicitly initializing
	response := cloudcontroller.Response{
		Result: &info,
	}

	err = client.connection.Make(request, &response)
	warnings = append(warnings, response.Warnings...)

	if err != nil {
		return APIInfo{}, ResourceLinks{}, warnings, err
	}

	return rootResponse, info, warnings, nil
}

// rootResponse returns the CC API root document.
func (client *Client) rootResponse() (APIInfo, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		Method: http.MethodGet,
		URL:    client.cloudControllerURL,
	})
	if err != nil {
		return APIInfo{}, nil, err
	}

	var rootResponse APIInfo
	response := cloudcontroller.Response{
		Result: &rootResponse,
	}

	err = client.connection.Make(request, &response)
	if unknownSourceErr, ok := err.(ccerror.UnknownHTTPSourceError); ok && unknownSourceErr.StatusCode == http.StatusNotFound {
		return APIInfo{}, nil, ccerror.APINotFoundError{URL: client.cloudControllerURL}
	}

	return rootResponse, response.Warnings, err
}
