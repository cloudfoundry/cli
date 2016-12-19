package cloudcontroller

import "fmt"

// UnverifiedServerError replaces x509.UnknownAuthorityError when the server
// has SSL but the client is unable to verify it's certificate
type UnverifiedServerError struct {
	URL string
}

func (e UnverifiedServerError) Error() string {
	return "x509: certificate signed by unknown authority"
}

// SSLValidationHostnameError replaces x509.HostnameError when the server has
// SSL certificate that does not match the hostname.
type SSLValidationHostnameError struct {
	Message string
}

func (e SSLValidationHostnameError) Error() string {
	return fmt.Sprintf("Hostname does not match SSL Certificate (%s)", e.Message)
}

// RequestError represents a generic error encountered while performing the
// HTTP request. This generic error occurs before a HTTP response is obtained.
type RequestError struct {
	Err error
}

func (e RequestError) Error() string {
	return e.Err.Error()
}

// RawHTTPStatusError represents any response with a 4xx or 5xx status code.
type RawHTTPStatusError struct {
	StatusCode  int
	RawResponse []byte
}

func (r RawHTTPStatusError) Error() string {
	return fmt.Sprintf("Error Code: %d\nRaw Response: %s", r.StatusCode, r.RawResponse)
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

// UnprocessableEntityError is returned when the request cannot be processed by
// the cloud controller.
type UnprocessableEntityError struct {
	Message string
}

func (e UnprocessableEntityError) Error() string {
	return e.Message
}

// TaskWorkersUnavailableError represents the case when no Diego workers are
// available.
type TaskWorkersUnavailableError struct {
	Message string
}

func (e TaskWorkersUnavailableError) Error() string {
	return e.Message
}

// ServiceUnavailableError wraps a http 503 error.
type ServiceUnavailableError struct {
	Message string
}

func (e ServiceUnavailableError) Error() string {
	return e.Message
}

// NotFoundError wraps a generic 404 error.
type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return e.Message
}

// APINotFoundError is returned when the API endpoint is not found.
type APINotFoundError struct {
	URL string
}

func (e APINotFoundError) Error() string {
	return fmt.Sprintf("Unable to find API at %s", e.URL)
}
