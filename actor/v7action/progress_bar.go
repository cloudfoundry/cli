package v7action

import (
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"os"
	"time"
)

//go:generate counterfeiter . SimpleProgressBar

type SimpleProgressBar interface {
	Initialize(path string) (io.Reader, int64, error)
	Terminate()
}

type ProgressBar struct {
	bar *pb.ProgressBar
}

func NewProgressBar() *ProgressBar {
	return &ProgressBar{}
}

func (p *ProgressBar) Initialize(path string) (io.Reader, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}

	p.bar = pb.New(int(fileInfo.Size())).SetUnits(pb.U_BYTES)
	p.bar.ShowTimeLeft = false
	p.bar.Start()
	return p.bar.NewProxyReader(file), fileInfo.Size(), nil

}

func (p *ProgressBar) Terminate() {
	// Adding sleep to ensure UI has finished drawing
	time.Sleep(time.Second)
	p.bar.Finish()
}
