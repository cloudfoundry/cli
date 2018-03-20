package progressbar

import (
	"io"
	"time"

	log "github.com/sirupsen/logrus"
	pb "gopkg.in/cheggaaa/pb.v1"
)

type ProgressBar struct {
	ready chan bool
	bar   *pb.ProgressBar
}

func NewProgressBar() *ProgressBar {
	return &ProgressBar{
		ready: make(chan bool),
	}
}

func (p *ProgressBar) NewProgressBarWrapper(reader io.Reader, sizeOfFile int64) io.Reader {
	log.WithField("file_size", sizeOfFile).Debug("new progress bar")

	ready, ok := <-p.ready
	if !ready || !ok {
		return nil
	}

	log.Debug("progress bar ready")
	p.bar = pb.New(int(sizeOfFile)).SetUnits(pb.U_BYTES)
	p.bar.ShowTimeLeft = false
	p.bar.Start()
	return p.bar.NewProxyReader(reader)
}

func (p *ProgressBar) Ready() {
	p.ready <- true
}

func (p *ProgressBar) Complete() {
	// Adding sleep to ensure UI has finished drawing
	time.Sleep(time.Second)
	p.bar.Finish()
}
