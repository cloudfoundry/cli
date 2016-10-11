package uaa

import "fmt"

// UnverifiedServerError replaces x509.UnknownAuthorityError when the server
// has SSL but the client is unable to verify it's certificate
type UnverifiedServerError struct {
	URL string
}

func (e UnverifiedServerError) Error() string {
	return "x509: certificate signed by unknown authority"
}

// RequestError represents a connection level request error
type RequestError struct {
	Err error
}

func (e RequestError) Error() string {
	return e.Err.Error()
}

// Error is an error returned by the UAA
type Error struct {
	Type        string `json:"error"`
	Description string `json:"error_description"`
}

func (r Error) Error() string {
	return fmt.Sprintf("Error Type: %s\nDescription: %s", r.Type, r.Description)
}
