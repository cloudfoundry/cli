package wrapper

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// RetryRequest is a wrapper that retries failed requests if they contain a 5XX
// status code.
type RetryRequest struct {
	maxRetries int
	connection cloudcontroller.Connection
}

// NewRetryRequest returns a pointer to a RetryRequest wrapper.
func NewRetryRequest(maxRetries int) *RetryRequest {
	return &RetryRequest{
		maxRetries: maxRetries,
	}
}

// Wrap sets the connection in the RetryRequest and returns itself.
func (retry *RetryRequest) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	retry.connection = innerconnection
	return retry
}

// Make retries the request if it comes back with a 5XX status code.
func (retry *RetryRequest) Make(request *http.Request, passedResponse *cloudcontroller.Response) error {
	var err error
	for i := 0; i < retry.maxRetries+1; i += 1 {
		err = retry.connection.Make(request, passedResponse)
		if err == nil {
			return nil
		}

		if passedResponse.HTTPResponse == nil || passedResponse.HTTPResponse.StatusCode < 500 {
			break
		}
	}
	return err
}
