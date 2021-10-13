package cfnetworking

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/networkerror"
)

// errorWrapper is the wrapper that converts responses with 4xx and 5xx status
// codes to an error.
type errorWrapper struct {
	connection Connection
}

func NewErrorWrapper() *errorWrapper {
	return new(errorWrapper)
}

// Wrap wraps a Cloud Controller connection in this error handling wrapper.
func (e *errorWrapper) Wrap(innerconnection Connection) Connection {
	e.connection = innerconnection
	return e
}

// Make converts RawHTTPStatusError, which represents responses with 4xx and
// 5xx status codes, to specific errors.
func (e *errorWrapper) Make(request *Request, passedResponse *Response) error {
	err := e.connection.Make(request, passedResponse)

	if rawHTTPStatusErr, ok := err.(networkerror.RawHTTPStatusError); ok {
		return convert(rawHTTPStatusErr)
	}
	return err
}

func convert(rawHTTPStatusErr networkerror.RawHTTPStatusError) error {
	// Try to unmarshal the raw error into a CC error. If unmarshaling fails,
	// return the raw error.
	var errorResponse networkerror.ErrorResponse
	err := json.Unmarshal(rawHTTPStatusErr.RawResponse, &errorResponse)
	if err != nil {
		return rawHTTPStatusErr
	}

	switch rawHTTPStatusErr.StatusCode {
	case http.StatusBadRequest: // 400
		return networkerror.BadRequestError(errorResponse)
	case http.StatusUnauthorized: // 401
		return networkerror.UnauthorizedError(errorResponse)
	case http.StatusForbidden: // 403
		return networkerror.ForbiddenError(errorResponse)
	case http.StatusNotAcceptable: // 406
		return networkerror.NotAcceptableError(errorResponse)
	case http.StatusConflict: // 409
		return networkerror.ConflictError(errorResponse)
	default:
		return networkerror.UnexpectedResponseError{
			ErrorResponse: errorResponse,
			RequestIDs:    rawHTTPStatusErr.RequestIDs,
			ResponseCode:  rawHTTPStatusErr.StatusCode,
		}
	}
}
