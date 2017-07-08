package ccerror

// UnverifiedServerError replaces x509.UnknownAuthorityError when the server
// has SSL but the client is unable to verify it's certificate
type UnverifiedServerError struct {
	URL string
}

func (UnverifiedServerError) Error() string {
	return "x509: certificate signed by unknown authority"
}
