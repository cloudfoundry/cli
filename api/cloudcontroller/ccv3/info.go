package ccv3

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/resources"
)

type InfoLinks struct {
	// AppSSH is the link for application ssh info.
	AppSSH resources.APILink `json:"app_ssh"`

	// CCV3 is the link to the Cloud Controller V3 API.
	CCV3 resources.APILink `json:"cloud_controller_v3"`

	// Logging is the link to the Logging API.
	Logging resources.APILink `json:"logging"`

	// Logging is the link to the Logging API.
	LogCache resources.APILink `json:"log_cache"`

	// NetworkPolicyV1 is the link to the Container to Container Networking
	// API.
	NetworkPolicyV1 resources.APILink `json:"network_policy_v1"`

	// Routing is the link to the routing API
	Routing resources.APILink `json:"routing"`

	// UAA is the link to the UAA API.
	UAA resources.APILink `json:"uaa"`

	// Login is the link to the Login API.
	Login resources.APILink `json:"login"`
}

// Info represents a GET response from the '/' endpoint of the cloud
// controller API.
type Info struct {
	// Links is a list of top level Cloud Controller APIs.
	Links   InfoLinks `json:"links"`
	CFOnK8s bool      `json:"cf_on_k8s"`
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

// Routing returns the HREF of the routing API.
func (info Info) Routing() string {
	return info.Links.Routing.HREF
}

// UAA returns the HREF of the UAA server.
func (info Info) UAA() string {
	return info.Links.UAA.HREF
}

// Login returns the HREF of the login server.
func (info Info) Login() string {
	return info.Links.Login.HREF
}

// ResourceLinks represents the information returned back from /v3.
type ResourceLinks map[string]resources.APILink

// UnmarshalJSON helps unmarshal a Cloud Controller /v3 response.
func (links ResourceLinks) UnmarshalJSON(data []byte) error {
	var ccResourceLinks struct {
		Links map[string]resources.APILink `json:"links"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccResourceLinks)
	if err != nil {
		return err
	}

	for key, val := range ccResourceLinks.Links {
		links[key] = val
	}

	return nil
}

// GetInfo returns endpoint and API information from /v3.
func (client *Client) GetInfo() (Info, Warnings, error) {
	rootResponse, warnings, err := client.RootResponse()
	if err != nil {
		return Info{}, warnings, err
	}

	return rootResponse, warnings, err
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
