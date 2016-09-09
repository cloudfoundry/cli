package cloudcontrollerv2

import "fmt"

type UnexpectedResponseError struct {
	error
	StatusCode int
	Status     string
	Body       string
}

func (e UnexpectedResponseError) Error() string {
	return fmt.Sprintf("Unexpected Response\nStatus: %s\nBody:\n%s", e.Status, e.Body)
}

type ResourceNotFoundError struct {
	error
}

func (e ResourceNotFoundError) Error() string {
	return "resource not found"
}

type UnauthorizedError struct {
	error
}

func (e UnauthorizedError) Error() string {
	return "unauthorized"
}

type ForbiddenError struct {
	error
}

func (e ForbiddenError) Error() string {
	return "forbidden"
}

// UnverifiedServerError replaces x509.UnknownAuthorityError when the server
// has SSL but the client is unable to verify it's certificate
type UnverifiedServerError struct {
	error
}

func (e UnverifiedServerError) Error() string {
	return "x509: certificate signed by unknown authority"
}

type RequestError error
