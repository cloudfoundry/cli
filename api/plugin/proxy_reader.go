package plugin

import "io"

//go:generate counterfeiter . ProxyReader
type ProxyReader interface {
	Wrap(io.Reader) io.ReadCloser
	Start(int64)
	Finish()
}
