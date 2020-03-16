package logs

import (
	"context"
	"os"
	"os/signal"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
)

type LogCacheRepository struct {
	logCacheClient *logcache.Client
	cancelFunc     context.CancelFunc
}

func NewLogCacheRepository(logCacheClient *logcache.Client) *LogCacheRepository {
	return &LogCacheRepository{
		logCacheClient: logCacheClient,
	}
}

func (repo *LogCacheRepository) RecentLogsFor(appGUID string) ([]Loggable, error) {
	logs, err := sharedaction.GetRecentLogs(appGUID, repo.logCacheClient)

	if err != nil {
		return loggableMessagesFromLogCacheMessages(logs), err
	}

	return loggableMessagesFromLogCacheMessages(logs), err
}

func loggableMessagesFromLogCacheMessages(messages []sharedaction.LogMessage) []Loggable {

	loggableMessages := make([]Loggable, len(messages))

	for i, m := range messages {
		loggableMessages[i] = NewLogCacheMessage(&m)
	}

	return loggableMessages
}

func (repo *LogCacheRepository) TailLogsFor(appGUID string, onConnect func(), logChan chan<- Loggable, errChan chan<- error) {
	messages, logErrs, stopStreaming := sharedaction.GetStreamingLogs(appGUID, repo.logCacheClient)
	repo.cancelFunc = stopStreaming

	defer close(logChan)
	defer close(errChan)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		select {
		case message, ok := <-messages:
			if !ok {
				return
			}
			logChan <- NewLogCacheMessage(&message)
		case logErr, ok := <-logErrs:
			if !ok {
				return
			}
			errChan <- logErr
		case <-c:
			repo.cancelFunc()
		}
	}
}

func (repo *LogCacheRepository) Close() {
	repo.cancelFunc()
}
