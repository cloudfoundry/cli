package uaa

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// AuthInfo represents a GET response from a login server
type AuthInfo struct {
	Links struct {
		UAA string `json:"uaa"`
	} `json:"links"`
}

// SetupResources configures the client to use the specified settings and diescopers the UAA and Authentication resources
func (client *Client) SetupResources(bootstrapURL string) error {
	request, err := client.newRequest(requestOptions{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("%s/login", bootstrapURL),
	})

	if err != nil {
		return err
	}

	info := AuthInfo{} // Explicitly initializing
	response := Response{
		Result: &info,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return err
	}

	UAALink := info.Links.UAA
	if UAALink == "" {
		UAALink = bootstrapURL
	}
	client.config.SetUAAEndpoint(UAALink)

	resources := map[string]string{
		"uaa": UAALink,
		"authorization_endpoint": bootstrapURL,
	}

	client.router = internal.NewRouter(internal.APIRoutes, resources)

	return nil
}
