package uaa

import "fmt"

// RawHTTPStatusError represents any response with a 4xx or 5xx status code.
type RawHTTPStatusError struct {
	StatusCode  int
	RawResponse []byte
}

func (r RawHTTPStatusError) Error() string {
	return fmt.Sprintf("Error Code: %d\nRaw Response: %s", r.StatusCode, r.RawResponse)
}

// UAAErrorResponse represents a generic UAA error response.
type UAAErrorResponse struct {
	Type        string `json:"error"`
	Description string `json:"error_description"`
}

func (e UAAErrorResponse) Error() string {
	return fmt.Sprintf("Error Type: %s\nDescription: %s", e.Type, e.Description)
}

// ConflictError is returned when the response status code is 409. It
// represents when there is a conflict in the state of the requested resource.
type ConflictError struct {
	Message string
}

func (e ConflictError) Error() string {
	return e.Message
}

// UnverifiedServerError replaces x509.UnknownAuthorityError when the server
// has SSL but the client is unable to verify it's certificate
type UnverifiedServerError struct {
	URL string
}

func (e UnverifiedServerError) Error() string {
	return "x509: certificate signed by unknown authority"
}

// RequestError represents a generic error encountered while performing the
// HTTP request. This generic error occurs before a HTTP response is obtained.
type RequestError struct {
	Err error
}

func (e RequestError) Error() string {
	return e.Err.Error()
}

// BadCredentialsError is returned when the credentials are rejected.
type BadCredentialsError struct {
	Message string
}

func (e BadCredentialsError) Error() string {
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

// InsufficientScopeError is returned when the client has insufficient scope
type InsufficientScopeError struct {
	Message string
}

func (e InsufficientScopeError) Error() string {
	return e.Message
}

// InvalidSCIMResourceError is returned usually when the client tries to create an inproperly formatted username
type InvalidSCIMResourceError struct {
	Message string
}

func (e InvalidSCIMResourceError) Error() string {
	return e.Message
}
