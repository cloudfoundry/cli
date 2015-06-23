package fakes

import (
	"errors"
	"time"

	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

type FakeOldLogsRepositoryWithTimeout struct {
}

func (fake *FakeOldLogsRepositoryWithTimeout) RecentLogsFor(appGuid string) ([]*logmessage.LogMessage, error) {
	return nil, nil
}

func (fake *FakeOldLogsRepositoryWithTimeout) TailLogsFor(appGuid string, onConnect func(), onMessage func(*logmessage.LogMessage)) error {
	time.Sleep(150 * time.Millisecond)
	return errors.New("Fake http timeout error")
}

func (fake *FakeOldLogsRepositoryWithTimeout) Close() {
}

// var _ api.OldLogsRepository = new(FakeOldLogsRepositoryWithTimeout)
