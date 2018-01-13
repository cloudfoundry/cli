package wrapper

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
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
func (retry *RetryRequest) Make(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	var err error

	for i := 0; i < retry.maxRetries+1; i++ {
		err = retry.connection.Make(request, passedResponse)
		if err == nil {
			return nil
		}

		if retry.skipRetry(request.Method, passedResponse.HTTPResponse) {
			break
		}

		// Reset the request body prior to the next retry
		resetErr := request.ResetBody()
		if resetErr != nil {
			if _, ok := resetErr.(ccerror.PipeSeekError); ok {
				return ccerror.PipeSeekError{Err: err}
			}
			return resetErr
		}
	}
	return err
}

// skipRetry will skip retry if the request method is POST or contains a status
// code that is not one of following http status codes: 500, 502, 503, 504.
func (*RetryRequest) skipRetry(httpMethod string, response *http.Response) bool {
	return httpMethod == http.MethodPost ||
		response != nil &&
			response.StatusCode != http.StatusInternalServerError &&
			response.StatusCode != http.StatusBadGateway &&
			response.StatusCode != http.StatusServiceUnavailable &&
			response.StatusCode != http.StatusGatewayTimeout
}
