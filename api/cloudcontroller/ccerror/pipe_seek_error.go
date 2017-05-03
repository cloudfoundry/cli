package ccerror

// PipeSeekError is returned by Pipebomb when a Seek is called.
type PipeSeekError struct {
}

func (e PipeSeekError) Error() string {
	return "error seeking a stream on retry"
}
