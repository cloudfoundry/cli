// Package router is a GoLang library that interacts with CloudFoundry Go Router
package router

import (
	"fmt"
	"runtime"

	"code.cloudfoundry.org/cli/api/router/internal"

	"github.com/tedsuo/rata"
)

// Client is a client that can be used to talk to a Cloud Controller's V2
// Endpoints.
type Client struct {
	routerGroupEndpoint string

	connection Connection
	router     *rata.RequestGenerator
	userAgent  string
}

// Config allows the Client to be configured
type Config struct {
	// AppName is the name of the application/process using the client.
	AppName string

	// AppVersion is the version of the application/process using the client.
	AppVersion string

	// ConnectionConfig is the configuration for the client connection.
	ConnectionConfig

	// RoutingEndpoint is the url of the router API.
	RoutingEndpoint string

	// Wrappers that apply to the client connection.
	Wrappers []ConnectionWrapper
}

// NewClient returns a new Router Client.
func NewClient(config Config) *Client {
	userAgent := fmt.Sprintf("%s/%s (%s; %s %s)",
		config.AppName,
		config.AppVersion,
		runtime.Version(),
		runtime.GOARCH,
		runtime.GOOS,
	)

	client := Client{
		userAgent:  userAgent,
		router:     rata.NewRequestGenerator(config.RoutingEndpoint, internal.APIRoutes),
		connection: NewConnection(config.ConnectionConfig),
	}

	for _, wrapper := range config.Wrappers {
		client.connection = wrapper.Wrap(client.connection)
	}

	return &client
}
