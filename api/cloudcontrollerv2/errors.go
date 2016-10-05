package cloudcontrollerv2

import "fmt"

type CCErrorResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	ErrorCode   string `json:"error_code"`
}

type UnexpectedResponseError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e UnexpectedResponseError) Error() string {
	return fmt.Sprintf("Unexpected Response\nStatus: %s\nBody:\n%s", e.Status, e.Body)
}

type ResourceNotFoundError struct {
	CCErrorResponse
}

func (e ResourceNotFoundError) Error() string {
	return e.Description
}

type UnauthorizedError struct {
}

func (e UnauthorizedError) Error() string {
	return "unauthorized"
}

type ForbiddenError struct {
}

func (e ForbiddenError) Error() string {
	return "forbidden"
}

// UnverifiedServerError replaces x509.UnknownAuthorityError when the server
// has SSL but the client is unable to verify it's certificate
type UnverifiedServerError struct {
	URL string
}

func (e UnverifiedServerError) Error() string {
	return "x509: certificate signed by unknown authority"
}

type RequestError struct {
	Err error
}

func (e RequestError) Error() string {
	return e.Err.Error()
}
