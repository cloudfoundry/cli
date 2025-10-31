package networkerror

// UnverifiedServerError replaces x509.UnknownAuthorityError when the server
// has SSL but the client is unable to verify its certificate
type UnverifiedServerError struct {
	URL string
}

func (UnverifiedServerError) Error() string {
	return "x509: certificate signed by unknown authority"
}
