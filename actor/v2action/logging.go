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
		close(stoppedRefreshingToken)
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

func (actor Actor) tokenExpiryTime(accessToken string) (*time.Duration, error) {
	var expiresIn time.Duration

	accessTokenString := strings.TrimPrefix(accessToken, "bearer ")
	token, err := jws.ParseJWT([]byte(accessTokenString))
	if err != nil {
		return nil, err
	}

	expiration, ok := token.Claims().Expiration()
	if ok {
		expiresIn = time.Until(expiration)
	}
	return &expiresIn, nil
}

func (actor Actor) refreshAccessTokenIfNecessary() (*time.Duration, error) {
	accessToken := actor.Config.AccessToken()

	duration, err := actor.tokenExpiryTime(accessToken)
	if err != nil || *duration < time.Minute {
		accessToken, err = actor.RefreshAccessToken(actor.Config.RefreshToken())
		if err != nil {
			return nil, err
		}
	}

	accessToken = strings.TrimPrefix(accessToken, "bearer ")
	token, err := jws.ParseJWT([]byte(accessToken))
	if err != nil {
		return nil, err
	}

	var timeToRefresh time.Duration
	expiration, ok := token.Claims().Expiration()
	if !ok {
		return nil, errors.New("Failed to get an expiry time from the current access token")
	}
	expiresIn := time.Until(expiration)
	if expiresIn >= 2*time.Minute {
		timeToRefresh = expiresIn - time.Minute
	} else {
		timeToRefresh = expiresIn * 9 / 10
	}
	return &timeToRefresh, nil
}
