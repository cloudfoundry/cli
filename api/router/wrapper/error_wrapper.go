package wrapper

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/v7/api/router"
	"code.cloudfoundry.org/cli/v7/api/router/routererror"
)

const expiredTokenMessage = "Token is expired"
const unauthorizedMessage = "You are not authorized to perform the requested action"

// ErrorWrapper is the wrapper that converts responses with 4xx and 5xx status
// codes to an error.
type ErrorWrapper struct {
	connection router.Connection
}

func NewErrorWrapper() *ErrorWrapper {
	return new(ErrorWrapper)
}

func (e *ErrorWrapper) Make(request *router.Request, passedResponse *router.Response) error {
	err := e.connection.Make(request, passedResponse)

	if rawHTTPStatusErr, ok := err.(routererror.RawHTTPStatusError); ok {
		if rawHTTPStatusErr.StatusCode == http.StatusNotFound {
			var resourceNotFoundError routererror.ResourceNotFoundError
			_ = json.Unmarshal(rawHTTPStatusErr.RawResponse, &resourceNotFoundError)
			return resourceNotFoundError
		}

		if rawHTTPStatusErr.StatusCode == http.StatusUnauthorized {
			var routingAPIErrorBody routererror.ErrorResponse

			_ = json.Unmarshal(rawHTTPStatusErr.RawResponse, &routingAPIErrorBody)

			if routingAPIErrorBody.Message == expiredTokenMessage {
				return routererror.InvalidAuthTokenError{Message: "Token is expired"}
			}

			if routingAPIErrorBody.Message == unauthorizedMessage {
				return routererror.UnauthorizedError{Message: routingAPIErrorBody.Message}
			}
		}
	}

	return err
}

func (e *ErrorWrapper) Wrap(connection router.Connection) router.Connection {
	e.connection = connection
	return e
}
