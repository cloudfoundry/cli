package wrapper

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cli/api/uaa"
)

// RetryRequest is a wrapper that retries failed requests if they contain a 5XX
// status code.
type RetryRequest struct {
	maxRetries int
	connection uaa.Connection
}

// NewRetryRequest returns a pointer to a RetryRequest wrapper.
func NewRetryRequest(maxRetries int) *RetryRequest {
	return &RetryRequest{
		maxRetries: maxRetries,
	}
}

// Make retries the request if it comes back with a 5XX status code.
func (retry *RetryRequest) Make(request *http.Request, passedResponse *uaa.Response) error {
	var err error
	var rawRequestBody []byte

	if request.Body != nil {
		rawRequestBody, err = ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			return err
		}
	}

	for i := 0; i < retry.maxRetries+1; i++ {
		if rawRequestBody != nil {
			request.Body = ioutil.NopCloser(bytes.NewBuffer(rawRequestBody))
		}
		err = retry.connection.Make(request, passedResponse)
		if err == nil {
			return nil
		}

		if retry.skipRetry(request.Method, passedResponse.HTTPResponse) {
			break
		}
	}
	return err
}

// Wrap sets the connection in the RetryRequest and returns itself.
func (retry *RetryRequest) Wrap(innerconnection uaa.Connection) uaa.Connection {
	retry.connection = innerconnection
	return retry
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
