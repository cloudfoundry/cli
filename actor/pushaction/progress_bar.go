package pushaction

import "io"

//go:generate counterfeiter . ProgressBar

type ProgressBar interface {
	NewProgressBarWrapper(reader io.Reader, sizeOfFile int64) io.Reader
}
