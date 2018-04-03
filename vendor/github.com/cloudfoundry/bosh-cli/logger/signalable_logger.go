package logger

import (
	"fmt"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

func NewSignalableLogger(logger boshlog.Logger, sigCh chan os.Signal) (boshlog.Logger, chan bool) {
	doneChannel := make(chan bool, 1)

	go func() {
		for {
			<-sigCh
			fmt.Println("Received SIGHUP - toggling debug output")
			logger.ToggleForcedDebug()
			doneChannel <- true
		}
	}()

	return logger, doneChannel
}
