package shared

import (
	"io"

	pb "gopkg.in/cheggaaa/pb.v1"
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
	p.bar = pb.New(int(size)).SetUnits(pb.U_BYTES)
	p.bar.Output = p.writer
	p.bar.Start()
}

func (p ProgressBarProxyReader) Finish() {
	p.bar.Finish()
}

func NewProgressBarProxyReader(writer io.Writer) *ProgressBarProxyReader {
	return &ProgressBarProxyReader{writer: writer}
}
