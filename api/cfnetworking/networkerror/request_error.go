package networkerror

// RequestError represents a generic error encountered while performing the
// HTTP request. This generic error occurs before a HTTP response is obtained.
type RequestError struct {
	Err error
}

func (e RequestError) Error() string {
	return e.Err.Error()
}
