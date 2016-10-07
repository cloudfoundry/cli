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

type RequestError struct {
	Err error
}

func (e RequestError) Error() string {
	return e.Err.Error()
}

type RawCCError struct {
	StatusCode  int
	RawResponse []byte
}

func (r RawCCError) Error() string {
	return fmt.Sprintf("Error Code: %i\nRaw Response: %s\n", r.StatusCode, string(r.RawResponse))
}
