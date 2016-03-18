package fakes

import (
	"errors"
	"time"

	"github.com/cloudfoundry/cli/cf/api/logs"
)

type FakeLogsRepositoryWithTimeout struct{}

func (fake *FakeLogsRepositoryWithTimeout) RecentLogsFor(appGuid string) ([]logs.Loggable, error) {
	return nil, nil
}

func (fake *FakeLogsRepositoryWithTimeout) TailLogsFor(appGuid string, onConnect func()) (<-chan logs.Loggable, error) {
	time.Sleep(150 * time.Millisecond)
	return nil, errors.New("Fake http timeout error")
}

func (fake *FakeLogsRepositoryWithTimeout) Close() {}
