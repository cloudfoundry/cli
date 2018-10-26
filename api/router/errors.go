package router

// errorWrapper is the wrapper that converts responses with 4xx and 5xx status
// codes to an error.
type errorWrapper struct {
	connection Connection
}

func newErrorWrapper() *errorWrapper {
	return new(errorWrapper)
}
