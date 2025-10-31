package uaa

import (
	"code.cloudfoundry.org/cli/v9/api/uaa/internal"
)

// SetupResources configures the client to use the specified settings and diescopers the UAA and Authentication resources
func (client *Client) SetupResources(uaaURL string, loginURL string) error {
	info := NewInfo(uaaURL, loginURL)

	resources := map[string]string{
		"uaa":                    uaaURL,
		"authorization_endpoint": loginURL,
	}

	client.router = internal.NewRouter(internal.APIRoutes, resources)
	client.Info = info

	client.config.SetUAAEndpoint(uaaURL)

	return nil
}
