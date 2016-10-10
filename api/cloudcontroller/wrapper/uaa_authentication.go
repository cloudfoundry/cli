package wrapper

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

//go:generate counterfeiter . UAAClient

type UAAClient interface {
	AccessToken() string
	RefreshToken() error
}

// UAAAuthentication wraps connections and adds authentication headers to all
// requests
type UAAAuthentication struct {
	connection cloudcontroller.Connection
	client     UAAClient
}

// NewUAAAuthentication returns a pointer to a UAAAuthentication wrapper with
// the client set as the AuthenticationStore
func NewUAAAuthentication(client UAAClient) *UAAAuthentication {
	return &UAAAuthentication{
		client: client,
	}
}

// Wrap sets the connection on the UAAAuthentication and returns itself
func (t *UAAAuthentication) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	t.connection = innerconnection
	return t
}

// Make adds authentication headers to the passed in request and then calls the
// wrapped connection's Make
func (t *UAAAuthentication) Make(passedRequest cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	if passedRequest.Header == nil {
		passedRequest.Header = http.Header{}
	}

	passedRequest.Header.Set("Authorization", t.client.AccessToken())

	err := t.connection.Make(passedRequest, passedResponse)
	if _, ok := err.(ccv2.InvalidAuthTokenError); ok {
		err = t.client.RefreshToken()
		if err != nil {
			return err
		}

		passedRequest.Header.Set("Authorization", t.client.AccessToken())
		err = t.connection.Make(passedRequest, passedResponse)
	}

	return err
}
