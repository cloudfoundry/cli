package pushaction

import "io"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ProgressBar

type ProgressBar interface {
	NewProgressBarWrapper(reader io.Reader, sizeOfFile int64) io.Reader
}
