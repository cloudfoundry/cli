package v2action

import (
	"context"

	"code.cloudfoundry.org/cli/actor/sharedaction"
)

func (actor Actor) GetStreamingLogs(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc) {
	return sharedaction.GetStreamingLogs(appGUID, client)
}

func (actor Actor) GetRecentLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	logCacheMessages, err := sharedaction.GetRecentLogs(app.GUID, client)
	if err != nil {
		return nil, allWarnings, err
	}

	var logMessages []sharedaction.LogMessage

	for _, message := range logCacheMessages {
		logMessages = append(logMessages, *sharedaction.NewLogMessage(
			message.Message(),
			message.Type(),
			message.Timestamp(),
			message.SourceType(),
			message.SourceInstance(),
		))
	}

	return logMessages, allWarnings, nil
}

func (actor Actor) GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, nil, func() {}, allWarnings, err
	}

	messages, logErrs, stopStreaming := actor.GetStreamingLogs(app.GUID, client)

	return messages, logErrs, stopStreaming, allWarnings, err
}
