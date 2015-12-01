package api

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

//go:generate counterfeiter -o fakes/fake_logs_repository.go . LogsRepository
type LogsRepository interface {
	RecentLogsFor(appGuid string) ([]*logmessage.LogMessage, error)
	TailLogsFor(appGuid string, onConnect func(), onMessage func(*logmessage.LogMessage)) error
	Close()
}

type LoggregatorLogsRepository struct {
	consumer consumer.LoggregatorConsumer
	config   core_config.Reader
	// TrustedCerts   []tls.Certificate
	tokenRefresher authentication.TokenRefresher
	messageQueue   *Loggregator_SortedMessageQueue

	onMessage func(*logmessage.LogMessage)
}

var BufferTime time.Duration = 5 * time.Second

func NewLoggregatorLogsRepository(config core_config.Reader, consumer consumer.LoggregatorConsumer, refresher authentication.TokenRefresher) LogsRepository {
	return &LoggregatorLogsRepository{
		config:         config,
		consumer:       consumer,
		tokenRefresher: refresher,
		messageQueue:   NewLoggregator_SortedMessageQueue(BufferTime, time.Now),
	}
}

func (repo *LoggregatorLogsRepository) Close() {
	repo.consumer.Close()
	repo.flushMessageQueue()
}

func (repo *LoggregatorLogsRepository) RecentLogsFor(appGuid string) ([]*logmessage.LogMessage, error) {
	messages, err := repo.consumer.Recent(appGuid, repo.config.AccessToken())

	switch err.(type) {
	case nil: // do nothing
	case *noaa_errors.UnauthorizedError:
		repo.tokenRefresher.RefreshAuthToken()
		messages, err = repo.consumer.Recent(appGuid, repo.config.AccessToken())
	default:
		return messages, err
	}

	consumer.SortRecent(messages)
	return messages, err
}

func (repo *LoggregatorLogsRepository) TailLogsFor(appGuid string, onConnect func(), onMessage func(*logmessage.LogMessage)) error {
	repo.onMessage = onMessage

	endpoint := repo.config.LoggregatorEndpoint()
	if endpoint == "" {
		return errors.New(T("Loggregator endpoint missing from config file"))
	}

	repo.consumer.SetOnConnectCallback(onConnect)
	logChan, err := repo.consumer.Tail(appGuid, repo.config.AccessToken())
	switch err.(type) {
	case nil: // do nothing
	case *noaa_errors.UnauthorizedError:
		repo.tokenRefresher.RefreshAuthToken()
		logChan, err = repo.consumer.Tail(appGuid, repo.config.AccessToken())
	default:
		return err
	}

	if err != nil {
		return err
	}

	repo.bufferMessages(logChan, onMessage)
	return nil
}

func (repo *LoggregatorLogsRepository) bufferMessages(logChan <-chan *logmessage.LogMessage, onMessage func(*logmessage.LogMessage)) {

	for {
		sendMessages(repo.messageQueue, onMessage)

		select {
		case msg, ok := <-logChan:
			if !ok {
				return
			}
			repo.messageQueue.PushMessage(msg)
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func (repo *LoggregatorLogsRepository) flushMessageQueue() {
	if repo.onMessage == nil {
		return
	}

	for {
		message := repo.messageQueue.PopMessage()
		if message == nil {
			break
		}

		repo.onMessage(message)
	}

	repo.onMessage = nil
}

func sendMessages(queue *Loggregator_SortedMessageQueue, onMessage func(*logmessage.LogMessage)) {
	for queue.NextTimestamp() < time.Now().UnixNano() {
		msg := queue.PopMessage()
		onMessage(msg)
	}
}
