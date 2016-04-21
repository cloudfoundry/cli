package apifakes

import (
	"errors"
	"time"

	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

type OldFakeLogsRepositoryWithTimeout struct{}

func (fake *OldFakeLogsRepositoryWithTimeout) RecentLogsFor(appGUID string) ([]*logmessage.LogMessage, error) {
	return nil, nil
}

func (fake *OldFakeLogsRepositoryWithTimeout) TailLogsFor(appGUID string, onConnect func()) (<-chan *logmessage.LogMessage, error) {
	time.Sleep(150 * time.Millisecond)
	return nil, errors.New("Fake http timeout error")
}

func (fake *OldFakeLogsRepositoryWithTimeout) Close() {}
