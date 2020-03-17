package logs

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
)

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

		loggables[i] = NewLogCacheMessage(v)
	}

	return loggables, nil
}
