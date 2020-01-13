package v7action

import (
	"context"
	"errors"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/loggingaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"github.com/SermoDigital/jose/jws"
)

func (actor Actor) GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, nil, nil, allWarnings, err
	}

	messages, logErrs, cancelFunc := sharedaction.GetStreamingLogs(app.GUID, client)

	return messages, logErrs, cancelFunc, allWarnings, err
}

func (actor Actor) GetRecentLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	logCacheMessages, err := loggingaction.GetRecentLogs(app.GUID, client)
	if err != nil {
		return nil, allWarnings, err
	}

	//TODO: Messages need sorting for most recent?
	// logCacheMessages = client.SortRecent(logCacheMessages)

	var logMessages []sharedaction.LogMessage

	for _, message := range logCacheMessages {
		logMessages = append(logMessages, *sharedaction.NewLogMessage(
			message.Message,
			message.MessageType,
			message.Timestamp, // time.Unix(0, message.Timestamp),
			message.SourceType,
			message.SourceInstance,
		))
	}

	return logMessages, allWarnings, nil
}

func (actor Actor) ScheduleTokenRefresh() (chan bool, error) {
	accessTokenString, err := actor.RefreshAccessToken()
	if err != nil {
		return nil, err
	}
	accessTokenString = strings.TrimPrefix(accessTokenString, "bearer ")
	token, err := jws.ParseJWT([]byte(accessTokenString))
	if err != nil {
		return nil, err
	}

	var expiresIn time.Duration
	expiration, ok := token.Claims().Expiration()
	if ok {
		expiresIn = time.Until(expiration)

		// When we refresh exactly every EXPIRY_DURATION nanoseconds usually the auth token
		// ends up expiring on the log-cache client. Better to refresh a little more often
		// to avoid log outage
		expiresIn = expiresIn * 9 / 10
	} else {
		return nil, errors.New("Failed to get an expiry time from the current access token")
	}
	quitNowChannel := make(chan bool, 1)

	go func() {
		ticker := time.NewTicker(expiresIn)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, err := actor.RefreshAccessToken()
				if err != nil {
					panic(err)
				}
			case <-quitNowChannel:
				return
			}
		}
	}()

	return quitNowChannel, nil
}
