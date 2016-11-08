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
	switch rawHTTPStatusErr.StatusCode {
	case http.StatusUnauthorized:
		if errors[0].Title == "CF-InvalidAuthToken" {
			return cloudcontroller.InvalidAuthTokenError{Message: errors[0].Detail}
		}
		return cloudcontroller.UnauthorizedError{Message: errors[0].Detail}
	case http.StatusForbidden:
		return cloudcontroller.ForbiddenError{Message: errors[0].Detail}
	case http.StatusNotFound:
		return cloudcontroller.ResourceNotFoundError{Message: errors[0].Detail}
	default:
		return UnexpectedResponseError{
			ResponseCode:    rawHTTPStatusErr.StatusCode,
			CCErrorResponse: errorResponse,
		}
	}
	return nil
}
