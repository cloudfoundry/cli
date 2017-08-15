// Package nooabridge wraps a UAA client and a tokenCache to support the
// TokenRefresher interface for noaa/consumer.
package noaabridge

import "code.cloudfoundry.org/cli/api/uaa"

//go:generate counterfeiter . UAAClient

// UAAClient is the interface for getting a valid access token
type UAAClient interface {
	RefreshAccessToken(refreshToken string) (uaa.RefreshedTokens, error)
}

//go:generate counterfeiter . TokenCache

// TokenCache is where the UAA token information is stored.
type TokenCache interface {
	RefreshToken() string
	SetAccessToken(token string)
	SetRefreshToken(token string)
}

// TokenRefresher implements the TokenRefresher interface. It requires a UAA
// client and a token cache for storing the access and refresh tokens.
type TokenRefresher struct {
	uaaClient UAAClient
	cache     TokenCache
}

// NewTokenRefresher returns back a pointer to a TokenRefresher.
func NewTokenRefresher(uaaClient UAAClient, cache TokenCache) *TokenRefresher {
	return &TokenRefresher{
		uaaClient: uaaClient,
		cache:     cache,
	}
}

// RefreshAuthToken refreshes the current Authorization Token and stores the
// Access and Refresh token in it's cache. The returned Authorization Token
// includes the type prefixed by a space.
func (t *TokenRefresher) RefreshAuthToken() (string, error) {
	tokens, err := t.uaaClient.RefreshAccessToken(t.cache.RefreshToken())
	if err != nil {
		return "", err
	}

	t.cache.SetAccessToken(tokens.AuthorizationToken())
	t.cache.SetRefreshToken(tokens.RefreshToken)
	return tokens.AuthorizationToken(), nil
}
