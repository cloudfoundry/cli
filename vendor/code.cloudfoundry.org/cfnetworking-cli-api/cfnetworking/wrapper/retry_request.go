package wrapper

import (
	"net/http"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking"
)

// RetryRequest is a wrapper that retries failed requests if they contain a 5XX
// status code.
type RetryRequest struct {
	maxRetries int
	connection cfnetworking.Connection
}

// NewRetryRequest returns a pointer to a RetryRequest wrapper.
func NewRetryRequest(maxRetries int) *RetryRequest {
	return &RetryRequest{
		maxRetries: maxRetries,
	}
}

// Wrap sets the connection in the RetryRequest and returns itself.
func (retry *RetryRequest) Wrap(innerconnection cfnetworking.Connection) cfnetworking.Connection {
	retry.connection = innerconnection
	return retry
}

// Make retries the request if it comes back with certain status codes.
func (retry *RetryRequest) Make(request *cfnetworking.Request, passedResponse *cfnetworking.Response) error {
	var err error

	for i := 0; i < retry.maxRetries+1; i += 1 {
		err = retry.connection.Make(request, passedResponse)
		if err == nil {
			return nil
		}

		if passedResponse.HTTPResponse != nil &&
			(passedResponse.HTTPResponse.StatusCode == http.StatusBadGateway ||
				passedResponse.HTTPResponse.StatusCode == http.StatusServiceUnavailable ||
				passedResponse.HTTPResponse.StatusCode == http.StatusGatewayTimeout ||
				(passedResponse.HTTPResponse.StatusCode >= 400 && passedResponse.HTTPResponse.StatusCode < 500)) {
			break
		}

		// Reset the request body prior to the next retry
		resetErr := request.ResetBody()
		if resetErr != nil {
			return resetErr
		}
	}
	return err
}
