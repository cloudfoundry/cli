package uaa

// Response contains the result of a UAA request
type Response struct {
	// Result is the unserialized response of the UAA request
	Result interface{}

	// RawResponse is the raw bytes of the HTTP Response
	RawResponse []byte
}
