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

// UnexpectedResponseError is returned when the client gets an error that has
// not been accounted for.
type UnexpectedResponseError struct {
	CCErrorResponse

	RequestIDs   []string
	ResponseCode int
}

func (e UnexpectedResponseError) Error() string {
	message := fmt.Sprintf("Unexpected Response\nResponse Code: %d\nCC Code:       %d\nCC ErrorCode:  %s\nDescription:   %s", e.ResponseCode, e.Code, e.ErrorCode, e.Description)
	for _, id := range e.RequestIDs {
		message = fmt.Sprintf("%s\nRequest ID:    %s", message, id)
	}
	return message
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

// Make converts RawHTTPStatusError, which represents responses with 4xx and
// 5xx status codes, to specific errors.
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
			return cloudcontroller.NotFoundError{Message: string(rawHTTPStatusErr.RawResponse)}
		}
		return rawHTTPStatusErr
	}

	switch rawHTTPStatusErr.StatusCode {
	case http.StatusUnauthorized: // 401
		if errorResponse.ErrorCode == "CF-InvalidAuthToken" {
			return cloudcontroller.InvalidAuthTokenError{Message: errorResponse.Description}
		}
		return cloudcontroller.UnauthorizedError{Message: errorResponse.Description}
	case http.StatusForbidden: // 403
		return cloudcontroller.ForbiddenError{Message: errorResponse.Description}
	case http.StatusNotFound: // 404
		return cloudcontroller.ResourceNotFoundError{Message: errorResponse.Description}
	case http.StatusUnprocessableEntity: // 422
		return cloudcontroller.UnprocessableEntityError{Message: errorResponse.Description}
	default:
		return UnexpectedResponseError{
			CCErrorResponse: errorResponse,
			RequestIDs:      rawHTTPStatusErr.RequestIDs,
			ResponseCode:    rawHTTPStatusErr.StatusCode,
		}
	}
	return nil
}
