package shared

import (
	"io"

	"github.com/cheggaaa/pb/v3"
)

// ProgressBarProxyReader wraps a progress bar in a ProxyReader interface.
type ProgressBarProxyReader struct {
	writer io.Writer
	bar    *pb.ProgressBar
}

func (p ProgressBarProxyReader) Wrap(reader io.Reader) io.ReadCloser {
	return p.bar.NewProxyReader(reader)
}

func (p *ProgressBarProxyReader) Start(size int64) {
	p.bar = pb.New(int(size))
	p.bar.Set(pb.Bytes, true)
	p.bar.SetWriter(p.writer)
	p.bar.Start()
}

func (p ProgressBarProxyReader) Finish() {
	p.bar.Finish()
}

func NewProgressBarProxyReader(writer io.Writer) *ProgressBarProxyReader {
	return &ProgressBarProxyReader{writer: writer}
}
