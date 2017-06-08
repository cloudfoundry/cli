package plugin

import "io"

//go:generate counterfeiter . ProxyReader
type ProxyReader interface {
	Wrap(io.Reader, int64) io.ReadCloser
}
