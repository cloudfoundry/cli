package progressbar

import (
	"io"
	"time"

	"github.com/cheggaaa/pb/v3"
	log "github.com/sirupsen/logrus"
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

func (p *ProgressBar) Complete() {
	// Adding sleep to ensure UI has finished drawing
	time.Sleep(time.Second)
	p.bar.Finish()
}

func (p *ProgressBar) NewProgressBarWrapper(reader io.Reader, sizeOfFile int64) io.Reader {
	log.WithField("file_size", sizeOfFile).Debug("new progress bar")

	ready, ok := <-p.ready
	if !ready || !ok {
		return nil
	}

	log.Debug("progress bar ready")
	p.bar = pb.New(int(sizeOfFile))
	p.bar.Set(pb.Bytes, true)
	p.bar.Start()
	return p.bar.NewProxyReader(reader)
}

func (p *ProgressBar) Ready() {
	p.ready <- true
}
