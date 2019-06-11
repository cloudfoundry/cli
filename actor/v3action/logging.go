package v3action

import (
	"code.cloudfoundry.org/cli/actor/loggingaction"
	"context"
)

func (actor Actor) GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client loggingaction.LogCacheClient) (<-chan loggingaction.LogMessage, <-chan error, Warnings, error, context.CancelFunc) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, nil, allWarnings, err, func() {}
	}

	messages, logErrs, cancelFunc := loggingaction.GetStreamingLogs(app.GUID, client)

	return messages, logErrs, allWarnings, err, cancelFunc
}
