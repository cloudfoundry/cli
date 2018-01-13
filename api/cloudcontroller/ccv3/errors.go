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
		if rawHTTPStatusErr.StatusCode >= http.StatusInternalServerError {
			return convert500(rawHTTPStatusErr)
		}
		return convert400(rawHTTPStatusErr)
	}
	return err
}

func convert400(rawHTTPStatusErr ccerror.RawHTTPStatusError) error {
	firstErr, errorResponse, err := unmarshalFirstV3Error(rawHTTPStatusErr)
	if err != nil {
		return err
	}

	switch rawHTTPStatusErr.StatusCode {
	case http.StatusUnauthorized: // 401
		if firstErr.Title == "CF-InvalidAuthToken" {
			return ccerror.InvalidAuthTokenError{Message: firstErr.Detail}
		}
		return ccerror.UnauthorizedError{Message: firstErr.Detail}
	case http.StatusForbidden: // 403
		return ccerror.ForbiddenError{Message: firstErr.Detail}
	case http.StatusNotFound: // 404
		return handleNotFound(firstErr)
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

func convert500(rawHTTPStatusErr ccerror.RawHTTPStatusError) error {
	switch rawHTTPStatusErr.StatusCode {
	case http.StatusServiceUnavailable: // 503
		firstErr, _, err := unmarshalFirstV3Error(rawHTTPStatusErr)
		if err != nil {
			return err
		}
		if firstErr.Title == "CF-TaskWorkersUnavailable" {
			return ccerror.TaskWorkersUnavailableError{Message: firstErr.Detail}
		}
		return ccerror.ServiceUnavailableError{Message: firstErr.Detail}
	default:
		return ccerror.V3UnexpectedResponseError{
			ResponseCode: rawHTTPStatusErr.StatusCode,
			RequestIDs:   rawHTTPStatusErr.RequestIDs,
			V3ErrorResponse: ccerror.V3ErrorResponse{
				Errors: []ccerror.V3Error{{
					Detail: string(rawHTTPStatusErr.RawResponse),
				}},
			},
		}
	}
}

func unmarshalFirstV3Error(rawHTTPStatusErr ccerror.RawHTTPStatusError) (ccerror.V3Error, ccerror.V3ErrorResponse, error) {
	// Try to unmarshal the raw error into a CC error. If unmarshaling fails,
	// return the raw error.
	var errorResponse ccerror.V3ErrorResponse
	err := json.Unmarshal(rawHTTPStatusErr.RawResponse, &errorResponse)
	// error parsing json
	if err != nil {
		return ccerror.V3Error{}, errorResponse, ccerror.UnknownHTTPSourceError{
			StatusCode:  rawHTTPStatusErr.StatusCode,
			RawResponse: rawHTTPStatusErr.RawResponse,
		}
	}

	errors := errorResponse.Errors
	if len(errors) == 0 {
		return ccerror.V3Error{}, errorResponse, ccerror.V3UnexpectedResponseError{
			ResponseCode:    rawHTTPStatusErr.StatusCode,
			V3ErrorResponse: errorResponse,
		}
	}
	// There could be multiple errors in the future but for now we only convert
	// the first error.
	firstErr := errors[0]
	return firstErr, errorResponse, nil
}

func handleNotFound(errorResponse ccerror.V3Error) error {
	switch errorResponse.Detail {
	case "App not found":
		return ccerror.ApplicationNotFoundError{}
	case "Droplet not found":
		return ccerror.DropletNotFoundError{}
	case "Instance not found":
		return ccerror.InstanceNotFoundError{}
	case "Process not found":
		return ccerror.ProcessNotFoundError{}
	default:
		return ccerror.ResourceNotFoundError{Message: errorResponse.Detail}
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
