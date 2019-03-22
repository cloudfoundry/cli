package wrapper

import (
	"strings"
	"time"

	"github.com/SermoDigital/jose/jws"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/uaa"
)

//go:generate counterfeiter . UAAClient

const accessTokenExpirationMargin = time.Minute

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

// Make adds authentication headers to the passed in request and then calls the
// wrapped connection's Make. If the client is not set on the wrapper, it will
// not add any header or handle any authentication errors.
func (t *UAAAuthentication) Make(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	if t.client == nil {
		return t.connection.Make(request, passedResponse)
	}
	err := t.refreshToken()
	if nil != err {
		return err
	}
	request.Header.Set("Authorization", t.cache.AccessToken())
	err = t.connection.Make(request, passedResponse)
	return err
}

// SetClient sets the UAA client that the wrapper will use.
func (t *UAAAuthentication) SetClient(client UAAClient) {
	t.client = client
}

// Wrap sets the connection on the UAAAuthentication and returns itself
func (t *UAAAuthentication) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	t.connection = innerconnection
	return t
}

// refreshToken refreshes the JWT access token if it is expired or about to expire.
// If the access token is not yet expired, no action is performed.
func (t *UAAAuthentication) refreshToken() error {
	tokenStr := strings.TrimPrefix(t.cache.AccessToken(), "bearer ")
	token, err := jws.ParseJWT([]byte(tokenStr))
	if err != nil {
		return err
	}
	expiration, _ := token.Claims().Expiration()
	if expiration.Sub(time.Now()) < accessTokenExpirationMargin {
		tokens, err := t.client.RefreshAccessToken(t.cache.RefreshToken())
		if err != nil {
			return err
		}
		t.cache.SetAccessToken(tokens.AuthorizationToken())
		t.cache.SetRefreshToken(tokens.RefreshToken)
	}
	return nil
}
