package wrapper

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/api/uaa"
)

type errorWrapper struct {
	connection uaa.Connection
}

func NewErrorWrapper() *errorWrapper {
	return new(errorWrapper)
}

func (e *errorWrapper) Wrap(innerconnection uaa.Connection) uaa.Connection {
	e.connection = innerconnection
	return e
}

func (e *errorWrapper) Make(request *http.Request, passedResponse *uaa.Response) error {
	err := e.connection.Make(request, passedResponse)

	if rawHTTPStatusErr, ok := err.(uaa.RawHTTPStatusError); ok {
		return convert(rawHTTPStatusErr)
	}

	return err
}

func convert(rawHTTPStatusErr uaa.RawHTTPStatusError) error {
	// Try to unmarshal the raw http status error into a UAA error. If
	// unmarshaling fails, return the raw error.
	var uaaErrorResponse uaa.UAAErrorResponse
	err := json.Unmarshal(rawHTTPStatusErr.RawResponse, &uaaErrorResponse)
	if err != nil {
		return rawHTTPStatusErr
	}

	switch rawHTTPStatusErr.StatusCode {
	case http.StatusBadRequest: // 400
		if uaaErrorResponse.Type == "invalid_scim_resource" {
			return uaa.InvalidSCIMResourceError{Message: uaaErrorResponse.Description}
		}
		return rawHTTPStatusErr
	case http.StatusUnauthorized: // 401
		if uaaErrorResponse.Type == "invalid_token" {
			return uaa.InvalidAuthTokenError{Message: uaaErrorResponse.Description}
		}
		return rawHTTPStatusErr
	case http.StatusForbidden: // 403
		if uaaErrorResponse.Type == "insufficient_scope" {
			return uaa.InsufficientScopeError{Message: uaaErrorResponse.Description}
		}
		return rawHTTPStatusErr
	case http.StatusConflict: // 409
		return uaa.ConflictError{Message: uaaErrorResponse.Description}
	default:
		return rawHTTPStatusErr
	}
}
