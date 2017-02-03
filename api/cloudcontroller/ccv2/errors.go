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
	message := fmt.Sprintf("Unexpected Response\nResponse code: %d\nCC code:       %d\nCC error code: %s", e.ResponseCode, e.Code, e.ErrorCode)
	for _, id := range e.RequestIDs {
		message = fmt.Sprintf("%s\nRequest ID:    %s", message, id)
	}
	return fmt.Sprintf("%s\nDescription:   %s", message, e.Description)
}

// AppStoppedStatsError is returned when requesting instance information from a
// stopped app.
type AppStoppedStatsError struct {
	Message string
}

func (e AppStoppedStatsError) Error() string {
	return e.Message
}

// NotStagedError is returned when requesting instance information from a
// not staged app.
type NotStagedError struct {
	Message string
}

func (e NotStagedError) Error() string {
	return e.Message
}

// InstancesError is returned when requesting instance information encounters
// an error.
type InstancesError struct {
	Message string
}

func (e InstancesError) Error() string {
	return e.Message
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
	case http.StatusBadRequest: // 400
		return handleBadRequest(errorResponse)
	case http.StatusUnauthorized: // 401
		return handleUnauthorized(errorResponse)
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

func handleBadRequest(errorResponse CCErrorResponse) error {
	switch errorResponse.ErrorCode {
	case "CF-AppStoppedStatsError":
		return AppStoppedStatsError{Message: errorResponse.Description}
	case "CF-InstancesError":
		return InstancesError{Message: errorResponse.Description}
	case "CF-NotStaged":
		return NotStagedError{Message: errorResponse.Description}
	default:
		return cloudcontroller.BadRequestError{Message: errorResponse.Description}
	}
}

func handleUnauthorized(errorResponse CCErrorResponse) error {
	if errorResponse.ErrorCode == "CF-InvalidAuthToken" {
		return cloudcontroller.InvalidAuthTokenError{Message: errorResponse.Description}
	}

	return cloudcontroller.UnauthorizedError{Message: errorResponse.Description}
}
