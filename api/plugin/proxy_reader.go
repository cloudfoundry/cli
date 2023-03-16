package plugin

import "io"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ProxyReader

type ProxyReader interface {
	Wrap(io.Reader) io.ReadCloser
	Start(int64)
	Finish()
}
