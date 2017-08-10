// Package uaa is a GoLang library that interacts with CloudFoundry User
// Account and Authentication (UAA) Server.
//
// It is currently designed to support UAA API X.X.X. However, it may include
// features and endpoints of later API versions.
package uaa

import (
	"fmt"
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

// Client is the UAA client
type Client struct {
	id     string
	secret string

	connection Connection
	router     *internal.Router
	userAgent  string
}

// Config allows the Client to be configured
type Config struct {
	// AppName is the name of the application/process using the client.
	AppName string

	// AppVersion is the version of the application/process using the client.
	AppVersion string

	// DialTimeout is the DNS lookup timeout for the client. If not set, it is
	// infinite.
	DialTimeout time.Duration

	// ClientID is the UAA client ID the client will use.
	ClientID string

	// ClientSecret is the UAA client secret the client will use.
	ClientSecret string

	// SkipSSLValidation controls whether a client verifies the server's
	// certificate chain and host name. If SkipSSLValidation is true, TLS accepts
	// any certificate presented by the server and any host name in that
	// certificate for *all* client requests going forward.
	//
	// In this mode, TLS is susceptible to man-in-the-middle attacks. This should
	// be used only for testing.
	SkipSSLValidation bool
}

// NewClient returns a new UAA Client with the provided configuration
func NewClient(config Config) *Client {
	userAgent := fmt.Sprintf("%s/%s (%s; %s %s)",
		config.AppName,
		config.AppVersion,
		runtime.Version(),
		runtime.GOARCH,
		runtime.GOOS,
	)

	client := Client{
		id:     config.ClientID,
		secret: config.ClientSecret,

		connection: NewConnection(config.SkipSSLValidation, config.DialTimeout),
		userAgent:  userAgent,
	}
	client.WrapConnection(NewErrorWrapper())

	return &client
}
