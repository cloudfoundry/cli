package logsfakes

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/cf/api/logs"
)

type FakeLogsRepositoryWithTimeout struct{}

func (fake *FakeLogsRepositoryWithTimeout) RecentLogsFor(appGuid string) ([]logs.Loggable, error) {
	return nil, nil
}

func (fake *FakeLogsRepositoryWithTimeout) TailLogsFor(appGuid string, onConnect func(), logChan chan<- logs.Loggable, errChan chan<- error) {
	time.Sleep(150 * time.Millisecond)
	errChan <- errors.New("Fake http timeout error")
}

func (fake *FakeLogsRepositoryWithTimeout) Close() {}

func (fake *FakeLogsRepositoryWithTimeout) FlushMessages(c chan<- logs.Loggable) {}
