package ccv3

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

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
func (e *errorWrapper) Make(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	err := e.connection.Make(request, passedResponse)

	if rawHTTPStatusErr, ok := err.(ccerror.RawHTTPStatusError); ok {
		return convert(rawHTTPStatusErr)
	}
	return err
}

func convert(rawHTTPStatusErr ccerror.RawHTTPStatusError) error {
	// Try to unmarshal the raw error into a CC error. If unmarshaling fails,
	// return the raw error.
	var errorResponse ccerror.V3ErrorResponse
	err := json.Unmarshal(rawHTTPStatusErr.RawResponse, &errorResponse)

	// error parsing json
	if err != nil {
		if rawHTTPStatusErr.StatusCode == http.StatusNotFound {
			return ccerror.NotFoundError{Message: string(rawHTTPStatusErr.RawResponse)}
		}
		return rawHTTPStatusErr
	}

	errors := errorResponse.Errors
	if len(errors) == 0 {
		return ccerror.V3UnexpectedResponseError{
			ResponseCode:    rawHTTPStatusErr.StatusCode,
			V3ErrorResponse: errorResponse,
		}
	}

	// There could be multiple errors in the future but for now we only convert
	// the first error.
	firstErr := errors[0]

	switch rawHTTPStatusErr.StatusCode {
	case http.StatusUnauthorized: // 401
		if firstErr.Title == "CF-InvalidAuthToken" {
			return ccerror.InvalidAuthTokenError{Message: firstErr.Detail}
		}
		return ccerror.UnauthorizedError{Message: firstErr.Detail}
	case http.StatusForbidden: // 403
		return ccerror.ForbiddenError{Message: firstErr.Detail}
	case http.StatusNotFound: // 404
		return ccerror.ResourceNotFoundError{Message: firstErr.Detail}
	case http.StatusUnprocessableEntity: // 422
		return handleUnprocessableEntity(firstErr)
	case http.StatusServiceUnavailable: // 503
		if firstErr.Title == "CF-TaskWorkersUnavailable" {
			return ccerror.TaskWorkersUnavailableError{Message: firstErr.Detail}
		}
		return ccerror.ServiceUnavailableError{Message: firstErr.Detail}
	default:
		return ccerror.V3UnexpectedResponseError{
			ResponseCode:    rawHTTPStatusErr.StatusCode,
			RequestIDs:      rawHTTPStatusErr.RequestIDs,
			V3ErrorResponse: errorResponse,
		}
	}
}

func handleUnprocessableEntity(errorResponse ccerror.V3Error) error {
	switch errorResponse.Detail {
	case "name must be unique in space":
		return ccerror.NameNotUniqueInSpaceError{}
	case "Buildpack must be an existing admin buildpack or a valid git URI":
		return ccerror.InvalidBuildpackError{}
	default:
		return ccerror.UnprocessableEntityError{Message: errorResponse.Detail}
	}
}
