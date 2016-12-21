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
	"github.com/tedsuo/rata"
)

//go:generate counterfeiter . AuthenticationStore

// AuthenticationStore represents the storage the UAA client
type AuthenticationStore interface {
	UAAOAuthClient() string
	UAAOAuthClientSecret() string

	AccessToken() string
	RefreshToken() string
	SetAccessToken(token string)
	SetRefreshToken(token string)
}

// Client is the UAA client
type Client struct {
	URL       string
	store     AuthenticationStore
	userAgent string

	router     *rata.RequestGenerator
	connection Connection
}

// Config allows the Client to be configured
type Config struct {
	AppName           string
	AppVersion        string
	DialTimeout       time.Duration
	SkipSSLValidation bool
	Store             AuthenticationStore
	URL               string
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
	return &Client{
		URL:       config.URL,
		store:     config.Store,
		userAgent: userAgent,

		router:     rata.NewRequestGenerator(config.URL, internal.Routes),
		connection: NewConnection(config.SkipSSLValidation, config.DialTimeout),
	}
}
