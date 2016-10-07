package wrapper

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

//go:generate counterfeiter . AuthenticationStore

type AuthenticationStore interface {
	AccessToken() string
	RefreshToken() string
	// SetAccessToken(token string)
	// SetRefreshToken(token string)

	// ClientName() string
	// ClientSecret() string
	// SkipSSLValidation() bool
}

// TokenRefreshWrapper wraps connections and adds authentication headers to all
// requests
type TokenRefreshWrapper struct {
	connection cloudcontroller.Connection
	store      AuthenticationStore
}

// NewTokenRefreshWrapper returns a pointer to a TokenRefreshWrapper with the
// store set as the AuthenticationStore
func NewTokenRefreshWrapper(store AuthenticationStore) *TokenRefreshWrapper {
	return &TokenRefreshWrapper{
		store: store,
	}
}

// Wrap sets the connection on the TokenRefreshWrapper and returns itself
func (t *TokenRefreshWrapper) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	t.connection = innerconnection
	return t
}

// Make adds authentication headers to the passed in request and then calls the
// wrapped connection's Make
func (t *TokenRefreshWrapper) Make(passedRequest cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	if passedRequest.Header == nil {
		passedRequest.Header = http.Header{}
	}

	passedRequest.Header.Add("Authorization", t.store.AccessToken())

	return t.connection.Make(passedRequest, passedResponse)
}
