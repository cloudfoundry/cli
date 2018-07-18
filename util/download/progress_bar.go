package download

import (
	"io"

	pb "gopkg.in/cheggaaa/pb.v1"
)

//go:generate counterfeiter . ProgressBar

type ProgressBar interface {
	Finish()
	NewProxyReader(r io.Reader) *pb.Reader
	SetTotal(total int) *pb.ProgressBar
	Start() *pb.ProgressBar
}
