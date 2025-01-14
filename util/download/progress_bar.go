package download

import (
	"io"

	"github.com/cheggaaa/pb/v3"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ProgressBar

type ProgressBar interface {
	Finish()
	NewProxyReader(r io.Reader) *pb.Reader
	SetTotal(total int) *pb.ProgressBar
	Start() *pb.ProgressBar
}
