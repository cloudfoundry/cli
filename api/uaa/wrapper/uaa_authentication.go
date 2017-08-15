package wrapper

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

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
	connection uaa.Connection
	client     UAAClient
	cache      TokenCache
}

// NewUAAAuthentication returns a pointer to a UAAAuthentication wrapper with
// the client and token cache.
func NewUAAAuthentication(client UAAClient, cache TokenCache) *UAAAuthentication {
	return &UAAAuthentication{
		client: client,
		cache:  cache,
	}
}

// Wrap sets the connection on the UAAAuthentication and returns itself
func (t *UAAAuthentication) Wrap(innerconnection uaa.Connection) uaa.Connection {
	t.connection = innerconnection
	return t
}

// SetClient sets the UAA client that the wrapper will use.
func (t *UAAAuthentication) SetClient(client UAAClient) {
	t.client = client
}

// Make adds authentication headers to the passed in request and then calls the
// wrapped connection's Make
func (t *UAAAuthentication) Make(request *http.Request, passedResponse *uaa.Response) error {
	if t.client == nil {
		return t.connection.Make(request, passedResponse)
	}

	var err error
	var rawRequestBody []byte

	if request.Body != nil {
		rawRequestBody, err = ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			return err
		}

		request.Body = ioutil.NopCloser(bytes.NewBuffer(rawRequestBody))

		if skipAuthenticationHeader(request, rawRequestBody) {
			return t.connection.Make(request, passedResponse)
		}
	}

	request.Header.Set("Authorization", t.cache.AccessToken())

	err = t.connection.Make(request, passedResponse)
	if _, ok := err.(uaa.InvalidAuthTokenError); ok {
		tokens, refreshErr := t.client.RefreshAccessToken(t.cache.RefreshToken())
		if refreshErr != nil {
			return refreshErr
		}

		t.cache.SetAccessToken(tokens.AuthorizationToken())
		t.cache.SetRefreshToken(tokens.RefreshToken)

		if rawRequestBody != nil {
			request.Body = ioutil.NopCloser(bytes.NewBuffer(rawRequestBody))
		}
		request.Header.Set("Authorization", t.cache.AccessToken())
		return t.connection.Make(request, passedResponse)
	}

	return err
}

// The authentication header is not added to token refresh requests or login
// requests.
func skipAuthenticationHeader(request *http.Request, body []byte) bool {
	stringBody := string(body)

	return strings.Contains(request.URL.String(), "/oauth/token") &&
		request.Method == http.MethodPost &&
		(strings.Contains(stringBody, "grant_type=refresh_token") ||
			strings.Contains(stringBody, "grant_type=password"))
}
