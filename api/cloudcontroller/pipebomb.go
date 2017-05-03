package cloudcontroller

import (
	"io"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

type Pipebomb struct {
	io.ReadCloser
}

func (p *Pipebomb) Seek(offset int64, whence int) (int64, error) {
	return 0, ccerror.PipeSeekError{}
}

func NewPipeBomb() (io.ReadSeeker, io.WriteCloser) {
	writerOutput, writerInput := io.Pipe()
	return &Pipebomb{ReadCloser: writerOutput}, writerInput
}
