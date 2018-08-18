// Package uaa is a GoLang library that interacts with CloudFoundry User
// Account and Authentication (UAA) Server.
//
// It is currently designed to support UAA API X.X.X. However, it may include
// features and endpoints of later API versions.
package uaa

import (
	"fmt"
	"runtime"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// Client is the UAA client
type Client struct {
	Info

	config Config

	connection Connection
	router     *internal.Router
	userAgent  string
}

// NewClient returns a new UAA Client with the provided configuration
func NewClient(config Config) *Client {
	userAgent := fmt.Sprintf("%s/%s (%s; %s %s)",
		config.BinaryName(),
		config.BinaryVersion(),
		runtime.Version(),
		runtime.GOARCH,
		runtime.GOOS,
	)

	client := Client{
		config: config,

		connection: NewConnection(config.SkipSSLValidation(), config.UAADisableKeepAlives(), config.DialTimeout()),
		userAgent:  userAgent,
	}
	client.WrapConnection(NewErrorWrapper())

	return &client
}
