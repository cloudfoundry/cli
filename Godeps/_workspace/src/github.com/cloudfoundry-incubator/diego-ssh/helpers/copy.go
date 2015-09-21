package helpers

import (
	"io"
	"sync"

	"github.com/pivotal-golang/lager"
)

func Copy(logger lager.Logger, wg *sync.WaitGroup, dest io.Writer, src io.Reader) {
	logger = logger.Session("copy")
	logger.Info("started")

	io.Copy(dest, src)

	logger.Info("completed")

	if wg != nil {
		wg.Done()
	}
}

func CopyAndClose(logger lager.Logger, wg *sync.WaitGroup, dest io.WriteCloser, src io.Reader) {
	logger = logger.Session("copy")
	logger.Info("started")

	io.Copy(dest, src)
	dest.Close()

	logger.Info("completed")

	if wg != nil {
		wg.Done()
	}
}
