package logs

import (
	"context"
	"fmt"
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
	// }

	// func (repo *LogCacheRepository) TailLogsFor(appGUID string, onConnect func(), logChan chan<- Loggable, errChan chan<- error) {
	//  messages, logErrs, stopStreaming := sharedaction.GetStreamingLogs(appGUID, repo.logCacheClient)
	r.cancelFunc = stopStreaming

	defer close(logChan)
	defer close(errChan)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		select {
		case message, ok := <-messages:
			if !ok {
				fmt.Println("closing messages")
				return
			}
			fmt.Printf("debug.1: %+v\n", len(logChan))
			fmt.Printf("debug.2: %+v\n", logChan)
			logChan <- NewLogCacheMessage(&terminalColorLogger{}, message)
		case logErr, ok := <-logErrs:
			if !ok {
				fmt.Println("closing logErrs")
				return
			}
			fmt.Printf("debug.3: %+v\n", logErr)
			errChan <- logErr
		case <-c:
			fmt.Println("cancelFunc")
			r.cancelFunc()
		}
	}
}

func (r *logCacheRepository) Close() {
	r.cancelFunc()
}
