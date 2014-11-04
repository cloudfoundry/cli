package net

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ProgressReader struct {
	ioReadSeeker io.ReadSeeker
	bytesRead    int64
	total        int64
	quit         chan bool
	ui           terminal.UI
}

func NewProgressReader(readSeeker io.ReadSeeker, ui terminal.UI) *ProgressReader {
	return &ProgressReader{
		ioReadSeeker: readSeeker,
		ui:           ui,
	}
}

func (progressReader *ProgressReader) Read(p []byte) (int, error) {
	if progressReader.ioReadSeeker == nil {
		return 0, os.ErrInvalid
	}

	n, err := progressReader.ioReadSeeker.Read(p)

	if progressReader.total > int64(0) {
		if n > 0 {
			if progressReader.quit == nil {
				progressReader.quit = make(chan bool)
				go progressReader.printProgress(progressReader.quit)
			}

			progressReader.bytesRead += int64(n)

			if progressReader.total == progressReader.bytesRead {
				progressReader.quit <- true
				return n, err
			}
		}
	}

	return n, err
}

func (progressReader *ProgressReader) Seek(offset int64, whence int) (int64, error) {
	return progressReader.ioReadSeeker.Seek(offset, whence)
}

func (progressReader *ProgressReader) printProgress(quit chan bool) {
	timer := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-quit:
			progressReader.ui.Say("\rDone uploading")
			return
		case <-timer.C:
			fmt.Println("\r%s uploaded...", formatters.ByteSize(progressReader.bytesRead))
		}
	}
}

func (progressReader *ProgressReader) SetTotalSize(size int64) {
	progressReader.total = size
}
