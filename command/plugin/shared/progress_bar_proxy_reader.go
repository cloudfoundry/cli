package shared

import (
	"io"

	pb "gopkg.in/cheggaaa/pb.v1"
)

// ProgressBarProxyReader wraps a progress bar in a ProxyReader interface.
type ProgressBarProxyReader struct {
	bar *pb.ProgressBar
}

func (p ProgressBarProxyReader) Wrap(reader io.Reader, size int64) io.ReadCloser {
	p.bar.Total = size
	return p.bar.NewProxyReader(reader)
}

func NewProgressBarProxyReader(bar *pb.ProgressBar) *ProgressBarProxyReader {
	return &ProgressBarProxyReader{bar: bar}
}
