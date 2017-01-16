package wrapper

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/uaa"
)

//go:generate counterfeiter . UAAClient

// UAAClient is the interface for getting a valid access token
type UAAClient interface {
	RefreshAccessToken(refreshToken string) (uaa.RefreshToken, error)
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
	connection cloudcontroller.Connection
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
func (t *UAAAuthentication) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	t.connection = innerconnection
	return t
}

// Make adds authentication headers to the passed in request and then calls the
// wrapped connection's Make
func (t *UAAAuthentication) Make(request *http.Request, passedResponse *cloudcontroller.Response) error {
	var (
		err            error
		rawRequestBody []byte
	)

	if request.Body != nil {
		rawRequestBody, err = ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			return err
		}
		request.Body = ioutil.NopCloser(bytes.NewBuffer(rawRequestBody))
	}

	request.Header.Set("Authorization", t.cache.AccessToken())

	err = t.connection.Make(request, passedResponse)
	if _, ok := err.(cloudcontroller.InvalidAuthTokenError); ok {
		var token uaa.RefreshToken
		token, err = t.client.RefreshAccessToken(t.cache.RefreshToken())
		if err != nil {
			return err
		}

		t.cache.SetAccessToken(token.AuthorizationToken())
		t.cache.SetRefreshToken(token.RefreshToken)

		if rawRequestBody != nil {
			request.Body = ioutil.NopCloser(bytes.NewBuffer(rawRequestBody))
		}
		request.Header.Set("Authorization", t.cache.AccessToken())
		err = t.connection.Make(request, passedResponse)
	}

	return err
}
