package ccv3

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// CCErrorResponse represents a generic Cloud Controller V3 error response.
type CCErrorResponse struct {
	Errors []CCError `json:"errors"`
}

// CCError represents a cloud controller error.
type CCError struct {
	Code   int    `json:"code"`
	Detail string `json:"detail"`
	Title  string `json:"title"`
}

// UnexpectedResponseError is returned when the client gets an error that has
// not been accounted for.
type UnexpectedResponseError struct {
	ResponseCode int
	CCErrorResponse
}

func (e UnexpectedResponseError) Error() string {
	messages := []string{
		"Unexpected Response",
		fmt.Sprintf("Response Code: %d", e.ResponseCode),
	}
	for _, ccError := range e.CCErrorResponse.Errors {
		messages = append(messages, fmt.Sprintf("Code: %d, Title: %s, Detail: %s", ccError.Code, ccError.Title, ccError.Detail))
	}

	return strings.Join(messages, "\n")
}

// errorWrapper is the wrapper that converts responses with 4xx and 5xx status
// codes to an error.
type errorWrapper struct {
	connection cloudcontroller.Connection
}

func newErrorWrapper() *errorWrapper {
	return new(errorWrapper)
}

// Wrap wraps a Cloud Controller connection in this error handling wrapper.
func (e *errorWrapper) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	e.connection = innerconnection
	return e
}

// Make creates a connection in the wrapped connection and handles errors
// that it returns.
func (e *errorWrapper) Make(request *http.Request, passedResponse *cloudcontroller.Response) error {
	err := e.connection.Make(request, passedResponse)

	if rawHTTPStatusErr, ok := err.(cloudcontroller.RawHTTPStatusError); ok {
		return convert(rawHTTPStatusErr)
	}
	return err
}

func convert(rawHTTPStatusErr cloudcontroller.RawHTTPStatusError) error {
	// Try to unmarshal the raw error into a CC error. If unmarshaling fails,
	// return the raw error.
	var errorResponse CCErrorResponse
	err := json.Unmarshal(rawHTTPStatusErr.RawResponse, &errorResponse)
	if err != nil {
		if rawHTTPStatusErr.StatusCode == http.StatusNotFound {
			return cloudcontroller.NotFoundError{string(rawHTTPStatusErr.RawResponse)}
		}
		return rawHTTPStatusErr
	}

	errors := errorResponse.Errors
	if len(errors) == 0 {
		return UnexpectedResponseError{
			ResponseCode:    rawHTTPStatusErr.StatusCode,
			CCErrorResponse: errorResponse,
		}
	}

	// There could be multiple errors in the future but for now we only convert
	// the first error.
	firstErr := errors[0]

	switch rawHTTPStatusErr.StatusCode {
	case http.StatusUnauthorized: // 401
		if firstErr.Title == "CF-InvalidAuthToken" {
			return cloudcontroller.InvalidAuthTokenError{Message: firstErr.Detail}
		}
		return cloudcontroller.UnauthorizedError{Message: firstErr.Detail}
	case http.StatusForbidden: // 403
		return cloudcontroller.ForbiddenError{Message: firstErr.Detail}
	case http.StatusNotFound: // 404
		return cloudcontroller.ResourceNotFoundError{Message: firstErr.Detail}
	case http.StatusUnprocessableEntity: // 422
		return cloudcontroller.UnprocessableEntityError{Message: firstErr.Detail}
	case http.StatusServiceUnavailable: // 503
		if firstErr.Title == "CF-TaskWorkersUnavailable" {
			return cloudcontroller.TaskWorkersUnavailableError{Message: firstErr.Detail}
		}
		return cloudcontroller.ServiceUnavailableError{Message: firstErr.Detail}
	default:
		return UnexpectedResponseError{
			ResponseCode:    rawHTTPStatusErr.StatusCode,
			CCErrorResponse: errorResponse,
		}
	}
	return nil
}
