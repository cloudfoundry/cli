package logs

import (
	"errors"
	"time"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"

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
	repo.consumer.Close()
}

func loggableMessagesFromNoaaMessages(messages []*events.LogMessage) []Loggable {
	loggableMessages := make([]Loggable, len(messages))

	for i, m := range messages {
		loggableMessages[i] = NewNoaaLogMessage(m)
	}

	return loggableMessages
}

func (repo *NoaaLogsRepository) RecentLogsFor(appGuid string) ([]Loggable, error) {
	logs, err := repo.consumer.RecentLogs(appGuid, repo.config.AccessToken())

	switch err.(type) {
	case nil: // do nothing
	case *noaa_errors.UnauthorizedError:
		repo.tokenRefresher.RefreshAuthToken()
		return repo.RecentLogsFor(appGuid)
	default:
		return loggableMessagesFromNoaaMessages(logs), err
	}

	return loggableMessagesFromNoaaMessages(noaa.SortRecent(logs)), err
}

func (repo *NoaaLogsRepository) TailLogsFor(appGuid string, onConnect func(), logChan chan<- Loggable, errChan chan<- error) {

	endpoint := repo.config.DopplerEndpoint()
	if endpoint == "" {
		errChan <- errors.New(T("Loggregator endpoint missing from config file"))
		return
	}

	repo.consumer.SetOnConnectCallback(onConnect)
	c, e := repo.consumer.TailingLogs(appGuid, repo.config.AccessToken())

	go func() {
		for {
			select {
			case msg, ok := <-c:
				if !ok {
					repo.flushMessages(logChan)
					close(logChan)
					return
				}

				repo.messageQueue.PushMessage(msg)
			case err := <-e:
				switch err.(type) {
				case nil:
				case *noaa_errors.UnauthorizedError:
					repo.tokenRefresher.RefreshAuthToken()
					repo.TailLogsFor(appGuid, onConnect, logChan, errChan)
				default:
					errChan <- err
					return
				}
			}
		}
	}()

}

func (repo *NoaaLogsRepository) flushMessages(c chan<- Loggable) {
	repo.messageQueue.EnumerateAndClear(func(m *events.LogMessage) {
		c <- NewNoaaLogMessage(m)
	})
}
