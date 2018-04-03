package director

import (
	"github.com/cloudfoundry/bosh-cli/ui"
	"io"
)

type NoopFileReporter struct{}
type NoopReadSeekCloser struct {
	Reader io.ReadCloser
}

func (nrsc NoopReadSeekCloser) Close() error {
	return nrsc.Reader.Close()
}

func (nrsc NoopReadSeekCloser) Read(p []byte) (n int, err error) {
	return nrsc.Reader.Read(p)
}

func (nrsc NoopReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	switch reader := nrsc.Reader.(type) {
	case io.Seeker:
		return reader.Seek(offset, whence)
	default:
		return 0, nil
	}
}

func NewNoopFileReporter() NoopFileReporter {
	return NoopFileReporter{}
}

func (r NoopFileReporter) TrackUpload(size int64, reader io.ReadCloser) ui.ReadSeekCloser {
	return NoopReadSeekCloser{reader}
}
func (r NoopFileReporter) TrackDownload(size int64, writer io.Writer) io.Writer { return writer }

type NoopTaskReporter struct{}

func NewNoopTaskReporter() NoopTaskReporter {
	return NoopTaskReporter{}
}

func (r NoopTaskReporter) TaskStarted(id int)                   {}
func (r NoopTaskReporter) TaskFinished(id int, state string)    {}
func (r NoopTaskReporter) TaskOutputChunk(id int, chunk []byte) {}
