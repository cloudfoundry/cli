package wrapper

import (
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/networkerror"
	"code.cloudfoundry.org/cli/api/uaa"
)

//go:generate counterfeiter . UAAClient

// UAAClient is the interface for getting a valid access token
type UAAClient interface {
	RefreshAccessToken(refreshToken string) (uaa.RefreshedTokens, error)
}

//go:generate counterfeiter . TokenCache

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
	connection cfnetworking.Connection
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

// Wrap sets the connection on the UAAAuthentication and returns itself
func (t *UAAAuthentication) Wrap(innerconnection cfnetworking.Connection) cfnetworking.Connection {
	t.connection = innerconnection
	return t
}

// SetClient sets the UAA client that the wrapper will use.
func (t *UAAAuthentication) SetClient(client UAAClient) {
	t.client = client
}

// Make adds authentication headers to the passed in request and then calls the
// wrapped connection's Make. If the client is not set on the wrapper, it will
// not add any header or handle any authentication errors.
func (t *UAAAuthentication) Make(request *cfnetworking.Request, passedResponse *cfnetworking.Response) error {
	request.Header.Set("Authorization", t.cache.AccessToken())

	requestErr := t.connection.Make(request, passedResponse)
	if _, ok := requestErr.(networkerror.InvalidAuthTokenError); ok {
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
