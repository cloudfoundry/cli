package ccerror

import "fmt"

// PipeSeekError is returned by Pipebomb when a Seek is called.
type PipeSeekError struct {
	// Err is the error that caused the Seek to be called.
	Err error
}

func (e PipeSeekError) Error() string {
	return fmt.Sprintf("error seeking a stream on retry: %s", e.Err)
}
