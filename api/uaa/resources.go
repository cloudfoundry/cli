package uaa

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// SetupResources configures the client to use the specified settings and diescopers the UAA and Authentication resources
func (client *Client) SetupResources(bootstrapURL string) error {
	request, err := client.newRequest(requestOptions{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("%s/login", bootstrapURL),
	})

	if err != nil {
		return err
	}

	info := NewInfo(bootstrapURL)
	response := Response{
		Result: &info,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return err
	}

	resources := map[string]string{
		"uaa":                    info.UAALink(),
		"authorization_endpoint": bootstrapURL,
	}

	client.router = internal.NewRouter(internal.APIRoutes, resources)
	client.Info = info

	client.config.SetUAAEndpoint(info.UAALink())

	return nil
}
