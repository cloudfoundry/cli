package logs

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf/api/authentication"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"

	"github.com/cloudfoundry/noaa"
	noaa_errors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
)

type NoaaLogsRepository struct {
	config         coreconfig.Reader
	consumer       NoaaConsumer
	tokenRefresher authentication.TokenRefresher
	messageQueue   *NoaaMessageQueue
	BufferTime     time.Duration
}

func NewNoaaLogsRepository(config coreconfig.Reader, consumer NoaaConsumer, tr authentication.TokenRefresher) *NoaaLogsRepository {
	return &NoaaLogsRepository{
		config:         config,
		consumer:       consumer,
		tokenRefresher: tr,
		messageQueue:   NewNoaaMessageQueue(),
		BufferTime:     defaultBufferTime,
	}
}

func (repo *NoaaLogsRepository) Close() {
	_ = repo.consumer.Close()
}

func loggableMessagesFromNoaaMessages(messages []*events.LogMessage) []Loggable {
	loggableMessages := make([]Loggable, len(messages))

	for i, m := range messages {
		loggableMessages[i] = NewNoaaLogMessage(m)
	}

	return loggableMessages
}

func (repo *NoaaLogsRepository) RecentLogsFor(appGUID string) ([]Loggable, error) {
	logs, err := repo.consumer.RecentLogs(appGUID, repo.config.AccessToken())

	switch err.(type) {
	case nil: // do nothing
	case *noaa_errors.UnauthorizedError:
		_, _ = repo.tokenRefresher.RefreshAuthToken()
		return repo.RecentLogsFor(appGUID)
	default:
		return loggableMessagesFromNoaaMessages(logs), err
	}

	return loggableMessagesFromNoaaMessages(noaa.SortRecent(logs)), err
}

func (repo *NoaaLogsRepository) TailLogsFor(appGUID string, onConnect func(), logChan chan<- Loggable, errChan chan<- error) {
	ticker := time.NewTicker(repo.BufferTime)
	endpoint := repo.config.DopplerEndpoint()
	if endpoint == "" {
		errChan <- errors.New(T("Loggregator endpoint missing from config file"))
		return
	}

	repo.consumer.SetOnConnectCallback(onConnect)
	c, e := repo.consumer.TailingLogsWithoutReconnect(appGUID, repo.config.AccessToken())

	go func() {
		for {
			select {
			case msg, ok := <-c:
				if !ok {
					ticker.Stop()
					repo.flushMessages(logChan)
					close(logChan)
					close(errChan)
					return
				}

				repo.messageQueue.PushMessage(msg)
			case err := <-e:
				switch err.(type) {
				case nil:
				case *noaa_errors.UnauthorizedError:
					_, _ = repo.tokenRefresher.RefreshAuthToken()
					ticker.Stop()
					repo.TailLogsFor(appGUID, onConnect, logChan, errChan)
					return
				default:
					errChan <- err

					ticker.Stop()
					close(logChan)
					close(errChan)
					return
				}
			}
		}
	}()

	go func() {
		for range ticker.C {
			repo.flushMessages(logChan)
		}
	}()
}

func (repo *NoaaLogsRepository) flushMessages(c chan<- Loggable) {
	repo.messageQueue.EnumerateAndClear(func(m *events.LogMessage) {
		c <- NewNoaaLogMessage(m)
	})
}
