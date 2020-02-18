package v7action

import (
	"context"
	"errors"
	"strings"
	"time"

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

	var timeToRefresh time.Duration
	expiration, ok := token.Claims().Expiration()
	if ok {
		expiresIn := time.Until(expiration)
		timeToRefresh = expiresIn * 9 / 10
	} else {
		return nil, errors.New("Failed to get an expiry time from the current access token")
	}
	quitNowChannel := make(chan bool, 1)

	go func() {
		ticker := time.NewTicker(timeToRefresh)
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
