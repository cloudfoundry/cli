package ccv3

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

type InfoLinks struct {
	// AppSSH is the link for application ssh info.
	AppSSH APILink `json:"app_ssh"`

	// CCV3 is the link to the Cloud Controller V3 API.
	CCV3 APILink `json:"cloud_controller_v3"`

	// Logging is the link to the Logging API.
	Logging APILink `json:"logging"`

	// Logging is the link to the Logging API.
	LogCache APILink `json:"log_cache"`

	// NetworkPolicyV1 is the link to the Container to Container Networking
	// API.
	NetworkPolicyV1 APILink `json:"network_policy_v1"`

	// Routing is the link to the routing API
	Routing APILink `json:"routing"`

	// UAA is the link to the UAA API.
	UAA APILink `json:"uaa"`
}

// Info represents a GET response from the '/' endpoint of the cloud
// controller API.
type Info struct {
	// Links is a list of top level Cloud Controller APIs.
	Links InfoLinks `json:"links"`
}

// AppSSHEndpoint returns the HREF for SSHing into an app container.
func (info Info) AppSSHEndpoint() string {
	return info.Links.AppSSH.HREF
}

// AppSSHHostKeyFingerprint returns the SSH key fingerprint of the SSH proxy
// that brokers connections to application instances.
func (info Info) AppSSHHostKeyFingerprint() string {
	return info.Links.AppSSH.Meta.HostKeyFingerprint
}

// CloudControllerAPIVersion returns the version of the CloudController.
func (info Info) CloudControllerAPIVersion() string {
	return info.Links.CCV3.Meta.Version
}

// LogCache returns the HREF of the Loggregator Traffic Controller.
func (info Info) LogCache() string {
	return info.Links.LogCache.HREF
}

// Logging returns the HREF of the Loggregator Traffic Controller.
func (info Info) Logging() string {
	return info.Links.Logging.HREF
}

// NetworkPolicyV1 returns the HREF of the Container Networking v1 Policy API
func (info Info) NetworkPolicyV1() string {
	return info.Links.NetworkPolicyV1.HREF
}

// OAuthClient returns the oauth client ID of the SSH proxy that brokers
// connections to application instances.
func (info Info) OAuthClient() string {
	return info.Links.AppSSH.Meta.OAuthClient
}

func (info Info) Routing() string {
	return info.Links.Routing.HREF
}

// UAA returns the HREF of the UAA server.
func (info Info) UAA() string {
	return info.Links.UAA.HREF
}

// ccv3Link returns the HREF of the CloudController v3 API.
func (info Info) ccV3Link() string {
	return info.Links.CCV3.HREF
}

// ResourceLinks represents the information returned back from /v3.
type ResourceLinks map[string]APILink

// UnmarshalJSON helps unmarshal a Cloud Controller /v3 response.
func (resources ResourceLinks) UnmarshalJSON(data []byte) error {
	var ccResourceLinks struct {
		Links map[string]APILink `json:"links"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccResourceLinks)
	if err != nil {
		return err
	}

	for key, val := range ccResourceLinks.Links {
		resources[key] = val
	}

	return nil
}

// GetInfo returns endpoint and API information from /v3.
func (client *Client) GetInfo() (Info, ResourceLinks, Warnings, error) {
	rootResponse, warnings, err := client.RootResponse()
	if err != nil {
		return Info{}, ResourceLinks{}, warnings, err
	}

	info := ResourceLinks{}

	_, v3Warnings, err := client.MakeRequest(RequestParams{
		URL:          rootResponse.ccV3Link(),
		ResponseBody: &info,
	})
	warnings = append(warnings, v3Warnings...)

	// root response -> targetcf as apiInfo
	// info -> targetcf as resourcelinks
	return rootResponse, info, warnings, err
}

// rootResponse returns the CC API root document.
func (client *Client) RootResponse() (Info, Warnings, error) {
	var responseBody Info

	_, warnings, err := client.MakeRequest(RequestParams{
		URL:          client.CloudControllerURL,
		ResponseBody: &responseBody,
	})

	unknownSourceErr, ok := err.(ccerror.UnknownHTTPSourceError)
	if ok && unknownSourceErr.StatusCode == http.StatusNotFound {
		return Info{}, nil, ccerror.APINotFoundError{URL: client.CloudControllerURL}
	}

	return responseBody, warnings, err
}
