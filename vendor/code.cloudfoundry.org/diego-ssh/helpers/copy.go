package helpers

import (
	"io"
	"sync"

	"code.cloudfoundry.org/lager"
)

func Copy(logger lager.Logger, wg *sync.WaitGroup, dest io.Writer, src io.Reader) {
	logger = logger.Session("copy")
	logger.Info("started")
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

	n, err := io.Copy(dest, src)
	if err != nil {
		logger.Error("copy-error", err)
	}

	logger.Info("completed", lager.Data{"bytes-copied": n})
}

func CopyAndClose(logger lager.Logger, wg *sync.WaitGroup, dest io.WriteCloser, src io.Reader, closeFunc func()) {
	logger = logger.Session("copy-and-close")
	logger.Info("started")

	defer func() {
		closeFunc()

		if wg != nil {
			wg.Done()
		}
	}()

	n, err := io.Copy(dest, src)
	if err != nil {
		logger.Error("copy-error", err)
	}

	logger.Info("completed", lager.Data{"bytes-copied": n})
}
