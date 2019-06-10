package v2action

import (
	"code.cloudfoundry.org/cli/actor/loggingaction"
	"context"
)

func (actor Actor) GetRecentLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client loggingaction.LogCacheClient) ([]loggingaction.LogMessage, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	reorderedLogMessages, err := loggingaction.GetRecentLogs(app.GUID, client)
	return reorderedLogMessages, allWarnings, err
}

func (actor Actor) GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client loggingaction.LogCacheClient) (<-chan loggingaction.LogMessage, <-chan error, Warnings, error, context.CancelFunc) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, nil, allWarnings, err, func() {}
	}

	messages, logErrs, cancel := loggingaction.GetStreamingLogs(app.GUID, client)

	return messages, logErrs, allWarnings, err, cancel
}
