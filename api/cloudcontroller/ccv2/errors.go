package ccv2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// CCErrorResponse represents a generic Cloud Controller V2 error response.
type CCErrorResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	ErrorCode   string `json:"error_code"`
}

// UnauthorizedError is returned when the client does not have the correct
// permissions to execute the request.
type UnauthorizedError struct {
	Message string
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

// InvalidAuthTokenError is returned when the client has an invalid
// authorization header.
type InvalidAuthTokenError struct {
	Message string
}

func (e InvalidAuthTokenError) Error() string {
	return e.Message
}

// ForbiddenError is returned when the client is forbidden from executing the
// request.
type ForbiddenError struct {
	Message string
}

func (e ForbiddenError) Error() string {
	return e.Message
}

// ResourceNotFoundError is returned when the client requests a resource that
// does not exist or does not have permissions to see.
type ResourceNotFoundError struct {
	Message string
}

func (e ResourceNotFoundError) Error() string {
	return e.Message
}

// UnexpectedResponseError is returned when the client gets an error that has
// not been accounted for.
type UnexpectedResponseError struct {
	ResponseCode int
	CCErrorResponse
}

func (e UnexpectedResponseError) Error() string {
	return fmt.Sprintf("Unexpected Response\nResponse Code: %s\nCC Code: %i\nCC ErrorCode: %s\nDescription: %s", e.ResponseCode, e.Code, e.ErrorCode, e.Description)
}

type errorWrapper struct {
	connection cloudcontroller.Connection
}

func newErrorWrapper() *errorWrapper {
	return new(errorWrapper)
}

func (e *errorWrapper) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	e.connection = innerconnection
	return e
}

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

	switch rawHTTPStatusErr.StatusCode {
	case http.StatusUnauthorized:
		if errorResponse.ErrorCode == "CF-InvalidAuthToken" {
			return InvalidAuthTokenError{Message: errorResponse.Description}
		}
		return UnauthorizedError{Message: errorResponse.Description}
	case http.StatusForbidden:
		return ForbiddenError{Message: errorResponse.Description}
	case http.StatusNotFound:
		return ResourceNotFoundError{Message: errorResponse.Description}
	default:
		return UnexpectedResponseError{
			ResponseCode:    rawHTTPStatusErr.StatusCode,
			CCErrorResponse: errorResponse,
		}
	}
	return nil
}
