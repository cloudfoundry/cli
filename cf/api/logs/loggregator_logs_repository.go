package logs

import (
	"errors"
	"time"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	consumer "github.com/cloudfoundry/loggregator_consumer"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	noaa_errors "github.com/cloudfoundry/noaa/errors"
)

type LoggregatorLogsRepository struct {
	consumer       consumer.LoggregatorConsumer
	config         core_config.Reader
	tokenRefresher authentication.TokenRefresher
	messageQueue   *LoggregatorMessageQueue
}

const bufferTime time.Duration = 25 * time.Millisecond

func NewLoggregatorLogsRepository(config core_config.Reader, consumer consumer.LoggregatorConsumer, refresher authentication.TokenRefresher) LogsRepository {
	return &LoggregatorLogsRepository{
		config:         config,
		consumer:       consumer,
		tokenRefresher: refresher,
		messageQueue:   NewLoggregatorMessageQueue(),
	}
}

func (repo *LoggregatorLogsRepository) Close() {
	repo.consumer.Close()
}

func loggableMessagesFromLoggregatorMessages(messages []*logmessage.LogMessage) []Loggable {
	loggableMessages := make([]Loggable, len(messages))

	for i, m := range messages {
		loggableMessages[i] = NewLoggregatorLogMessage(m)
	}

	return loggableMessages
}

func (repo *LoggregatorLogsRepository) RecentLogsFor(appGuid string) ([]Loggable, error) {
	messages, err := repo.consumer.Recent(appGuid, repo.config.AccessToken())

	switch err.(type) {
	case nil: // do nothing
	case *noaa_errors.UnauthorizedError:
		repo.tokenRefresher.RefreshAuthToken()
		return repo.RecentLogsFor(appGuid)
	default:
		return loggableMessagesFromLoggregatorMessages(messages), err
	}

	consumer.SortRecent(messages)

	return loggableMessagesFromLoggregatorMessages(messages), nil
}

func (repo *LoggregatorLogsRepository) TailLogsFor(appGuid string, onConnect func()) (<-chan Loggable, error) {
	ticker := time.NewTicker(bufferTime)

	c := make(chan Loggable)

	endpoint := repo.config.LoggregatorEndpoint()
	if endpoint == "" {
		return nil, errors.New(T("Loggregator endpoint missing from config file"))
	}

	repo.consumer.SetOnConnectCallback(onConnect)
	logChan, err := repo.consumer.Tail(appGuid, repo.config.AccessToken())
	switch err.(type) {
	case nil: // do nothing
	case *noaa_errors.UnauthorizedError:
		repo.tokenRefresher.RefreshAuthToken()
		logChan, err = repo.consumer.Tail(appGuid, repo.config.AccessToken())
	default:
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	go func() {
		for _ = range ticker.C {
			repo.flushMessageQueue(c)
		}
	}()

	go func() {
		for msg := range logChan {
			repo.messageQueue.PushMessage(msg)
		}

		repo.flushMessageQueue(c)
		close(c)
	}()

	return c, nil
}

func (repo *LoggregatorLogsRepository) flushMessageQueue(c chan Loggable) {
	repo.messageQueue.EnumerateAndClear(func(m *logmessage.LogMessage) {
		c <- NewLoggregatorLogMessage(m)
	})
}
