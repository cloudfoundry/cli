package fakes

import (
	"errors"
	"time"

	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

type FakeLogsRepositoryWithTimeout struct{}

func (fake *FakeLogsRepositoryWithTimeout) RecentLogsFor(appGuid string) ([]*logmessage.LogMessage, error) {
	return nil, nil
}

func (fake *FakeLogsRepositoryWithTimeout) TailLogsFor(appGuid string, onConnect func(), onMessage func(*logmessage.LogMessage)) error {
	time.Sleep(150 * time.Millisecond)
	return errors.New("Fake http timeout error")
}

func (fake *FakeLogsRepositoryWithTimeout) Close() {}
