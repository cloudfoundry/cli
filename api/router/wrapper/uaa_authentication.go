package wrapper

import (
	"code.cloudfoundry.org/cli/v9/api/router"
	"code.cloudfoundry.org/cli/v9/api/router/routererror"
	"code.cloudfoundry.org/cli/v9/api/uaa"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . UAAClient

// UAAClient is the interface for getting a valid access token
type UAAClient interface {
	RefreshAccessToken(refreshToken string) (uaa.RefreshedTokens, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . TokenCache

// TokenCache is where the UAA token information is stored.
type TokenCache interface {
	AccessToken() string
	RefreshToken() string
	SetAccessToken(token string)
	SetRefreshToken(token string)
}

// UAAAuthentication wraps connections and adds authentication headers to all
// requests
type UAAAuthentication struct {
	connection router.Connection
	client     UAAClient
	cache      TokenCache
}

// NewUAAAuthentication returns a pointer to a UAAAuthentication wrapper with
// the client and a token cache.
func NewUAAAuthentication(client UAAClient, cache TokenCache) *UAAAuthentication {
	return &UAAAuthentication{
		client: client,
		cache:  cache,
	}
}

// Make adds authentication headers to the passed in request and then calls the
// wrapped connection's Make. If the client is not set on the wrapper, it will
// not add any header or handle any authentication errors.
func (t *UAAAuthentication) Make(request *router.Request, passedResponse *router.Response) error {
	if t.client == nil {
		return t.connection.Make(request, passedResponse)
	}

	request.Header.Set("Authorization", t.cache.AccessToken())

	requestErr := t.connection.Make(request, passedResponse)
	if _, ok := requestErr.(routererror.InvalidAuthTokenError); ok {
		tokens, err := t.client.RefreshAccessToken(t.cache.RefreshToken())
		if err != nil {
			return err
		}

		t.cache.SetAccessToken(tokens.AuthorizationToken())
		t.cache.SetRefreshToken(tokens.RefreshToken)

		if request.Body != nil {
			err = request.ResetBody()
			if err != nil {
				return err
			}
		}
		request.Header.Set("Authorization", t.cache.AccessToken())
		requestErr = t.connection.Make(request, passedResponse)
	}

	return requestErr
}

// SetClient sets the UAA client that the wrapper will use.
func (t *UAAAuthentication) SetClient(client UAAClient) {
	t.client = client
}

// Wrap sets the connection on the UAAAuthentication and returns itself
func (t *UAAAuthentication) Wrap(innerconnection router.Connection) router.Connection {
	t.connection = innerconnection
	return t
}
