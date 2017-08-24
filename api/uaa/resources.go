package uaa

import (
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

//go:generate counterfeiter . UAAEndpointStore

type UAAEndpointStore interface {
	SetUAAEndpoint(uaaEndpoint string)
}

// SetupSettings represents configuration for establishing a connection to a UAA/Authentication server.
type SetupSettings struct {
	// DialTimeout is the DNS timeout used to make all requests to the Cloud
	// Controller.
	DialTimeout time.Duration

	// SkipSSLValidation controls whether a client verifies the server's
	// certificate chain and host name. If SkipSSLValidation is true, TLS accepts
	// any certificate presented by the server and any host name in that
	// certificate for *all* client requests going forward.
	//
	// In this mode, TLS is susceptible to man-in-the-middle attacks. This should
	// be used only for testing.
	SkipSSLValidation bool

	// BootstrapURL is a fully qualified URL to a UAA/Authentication server.
	BootstrapURL string
}

// AuthInfo represents a GET response from a login server
type AuthInfo struct {
	Links struct {
		UAA string `json:"uaa"`
	} `json:"links"`
}

// SetupResources configures the client to use the specified settings and diescopers the UAA and Authentication resources
func (client *Client) SetupResources(store UAAEndpointStore, bootstrapURL string) error {
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
	store.SetUAAEndpoint(UAALink)

	resources := map[string]string{
		"uaa": UAALink,
		"authorization_endpoint": bootstrapURL,
	}

	client.router = internal.NewRouter(internal.APIRoutes, resources)

	return nil
}
