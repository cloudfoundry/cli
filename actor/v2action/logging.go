package v2action

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/SermoDigital/jose/jws"

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

func (actor Actor) ScheduleTokenRefresh(
	after func(time.Duration) <-chan time.Time,
	stop chan struct{},
	stoppedRefreshingToken chan struct{}) (<-chan error, error) {

	timeToRefresh, err := actor.refreshAccessTokenIfNecessary()
	if err != nil {
		return nil, err
	}

	refreshErrs := make(chan error)

	go func() {
		defer close(stoppedRefreshingToken)
		for {
			select {
			case <-after(*timeToRefresh):
				d, err := actor.refreshAccessTokenIfNecessary()
				if err == nil {
					timeToRefresh = d
				} else {
					refreshErrs <- err
				}
			case <-stop:
				return
			}
		}
	}()

	return refreshErrs, nil
}

func (actor Actor) refreshAccessTokenIfNecessary() (*time.Duration, error) {
	accessTokenString, err := actor.RefreshAccessToken(actor.Config.RefreshToken())
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
	return &timeToRefresh, nil
}
