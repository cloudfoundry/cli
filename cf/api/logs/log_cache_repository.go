package logs

import (
	"context"
	"os"
	"os/signal"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type terminalColorLogger struct {
}

func (t terminalColorLogger) LogSysHeaderColor(message string) string {
	return terminal.LogSysHeaderColor(message)
}

func (t terminalColorLogger) LogStdoutColor(message string) string {
	return terminal.LogStdoutColor(message)
}

func (t terminalColorLogger) LogStderrColor(message string) string {
	return terminal.LogStderrColor(message)
}

type logCacheRepository struct {
	client               sharedaction.LogCacheClient
	cancelFunc           context.CancelFunc
	recentLogsFunc       func(appGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, error)
	getStreamingLogsFunc func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc)
}

func NewLogCacheRepository(
	client sharedaction.LogCacheClient,
	recentLogsFunc func(appGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, error),
	getStreamingLogsFunc func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc),
) *logCacheRepository {
	return &logCacheRepository{client: client, recentLogsFunc: recentLogsFunc, getStreamingLogsFunc: getStreamingLogsFunc}
}

func (r *logCacheRepository) RecentLogsFor(appGUID string) ([]Loggable, error) {
	logs, err := r.recentLogsFunc(appGUID, r.client)
	if err != nil {
		return nil, err
	}

	loggables := make([]Loggable, len(logs))
	for i, v := range logs {
		loggables[i] = NewLogCacheMessage(&terminalColorLogger{}, v)
	}

	return loggables, nil
}

func (r *logCacheRepository) TailLogsFor(appGUID string, onConnect func(), logChan chan<- Loggable, errChan chan<- error) {
	messages, logErrs, stopStreaming := r.getStreamingLogsFunc(appGUID, r.client)

	r.cancelFunc = stopStreaming

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
			logChan <- NewLogCacheMessage(&terminalColorLogger{}, message)
		case logErr, ok := <-logErrs:
			if !ok {
				return
			}
			errChan <- logErr
		case <-c:
			r.cancelFunc()
		}
	}
}

func (r *logCacheRepository) Close() {
	r.cancelFunc()
}
