package uaa

import "code.cloudfoundry.org/cli/api/uaa/internal"

//go:generate counterfeiter . AuthenticationStore

// AuthenticationStore represents the storage and configuration for the UAA
// client
type AuthenticationStore interface {
	ClientID() string
	ClientSecret() string
	SkipSSLValidation() bool

	AccessToken() string
	RefreshToken() string
	SetAccessToken(token string)
}

// Client is the UAA client
type Client struct {
	store      AuthenticationStore
	URL        string
	connection Connection
}

// NewClient returns a new UAA client
func NewClient(URL string, store AuthenticationStore) *Client {
	return &Client{
		store:      store,
		URL:        URL,
		connection: NewConnection(URL, internal.Routes, store.SkipSSLValidation()),
	}
}

// AccessToken returns the implicit grant access token
func (client *Client) AccessToken() string {
	return client.store.AccessToken()
}
