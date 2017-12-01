package cloudcontroller

import (
	"io"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

// Pipebomb is a wrapper around an io.Pipe's io.ReadCloser that turns it into a
// ReadSeeker that errors on Seek calls. This is designed to prevent the caller
// from rereading the body multiple times.
type Pipebomb struct {
	io.ReadCloser
}

// Seek returns a PipeSeekError; allowing the top level calling function to
// handle the retry instead of seeking back to the beginning of the Reader.
func (*Pipebomb) Seek(offset int64, whence int) (int64, error) {
	return 0, ccerror.PipeSeekError{}
}

// NewPipeBomb returns an io.WriteCloser that can be used to stream data to a
// the Pipebomb.
func NewPipeBomb() (*Pipebomb, io.WriteCloser) {
	writerOutput, writerInput := io.Pipe()
	return &Pipebomb{ReadCloser: writerOutput}, writerInput
}
