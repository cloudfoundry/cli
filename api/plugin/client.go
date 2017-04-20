package plugin

import (
	"fmt"
	"runtime"
	"time"
)

// Client is a client that can be used to make HTTP requests to plugin
// repositories.
type Client struct {
	connection Connection
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

	// SkipSSLValidation controls whether a client verifies the server's
	// certificate chain and host name. If SkipSSLValidation is true, TLS accepts
	// any certificate presented by the server and any host name in that
	// certificate for *all* client requests going forward.
	//
	// In this mode, TLS is susceptible to man-in-the-middle attacks. This should
	// be used only for testing.
	SkipSSLValidation bool
}

// NewClient returns a new plugin Client.
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
		connection: NewConnection(config.SkipSSLValidation, config.DialTimeout),
	}

	return &client
}
