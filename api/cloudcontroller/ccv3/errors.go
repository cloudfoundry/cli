package ccv3

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// CCErrorResponse represents a generic Cloud Controller V2 error response.
type CCErrorResponse struct {
	Errors []CCError
}

type CCError struct {
	Code   int    `json:"code"`
	Detail string `json:"detail"`
	Title  string `json:"title"`
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

func (e *errorWrapper) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	e.connection = innerconnection
	return e
}

func (e *errorWrapper) Make(request *http.Request, passedResponse *cloudcontroller.Response) error {
	err := e.connection.Make(request, passedResponse)

	if rawErr, ok := err.(cloudcontroller.RawCCError); ok {
		return e.convert(rawErr)
	}
	return err
}

func (e errorWrapper) convert(rawErr cloudcontroller.RawCCError) error {
	var errorResponse CCErrorResponse
	err := json.Unmarshal(rawErr.RawResponse, &errorResponse)
	if err != nil {
		return err
	}

	errors := errorResponse.Errors
	if len(errors) == 0 {

	}
	switch rawErr.StatusCode {
	case http.StatusUnauthorized:
		if errors[0].Code == 10002 {
			return InvalidAuthTokenError{Message: errors[0].Detail}
		}
		return UnauthorizedError{Message: errors[0].Detail}
	case http.StatusForbidden:
		return ForbiddenError{Message: errors[0].Detail}
	case http.StatusNotFound:
		return ResourceNotFoundError{Message: errors[0].Detail}
	default:
		return UnexpectedResponseError{
			ResponseCode:    rawErr.StatusCode,
			CCErrorResponse: errorResponse,
		}
	}
	return nil
}
