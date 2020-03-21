package logs

import (
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
	client         sharedaction.LogCacheClient
	recentLogsFunc func(appGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, error)
}

func NewLogCacheRepository(
	client sharedaction.LogCacheClient,
	recentLogsFunc func(appGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, error),
) *logCacheRepository {
	return &logCacheRepository{client: client, recentLogsFunc: recentLogsFunc}
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
