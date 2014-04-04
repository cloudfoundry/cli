package api

import (
	"cf/configuration"
	"crypto/tls"
	"errors"
	consumer "github.com/cloudfoundry/loggregator_consumer"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
)

type LogsRepository interface {
	RecentLogsFor(appGuid string) ([]*logmessage.LogMessage, error)
	TailLogsFor(appGuid string, bufferTime time.Duration, onConnect func(), onMessage func(*logmessage.LogMessage)) error
	Close()
}

type LoggregatorLogsRepository struct {
	consumer     consumer.LoggregatorConsumer
	config       configuration.Reader
	TrustedCerts []tls.Certificate
}

func NewLoggregatorLogsRepository(config configuration.Reader, consumer consumer.LoggregatorConsumer) LoggregatorLogsRepository {
	return LoggregatorLogsRepository{config: config, consumer: consumer}
}

func (repo LoggregatorLogsRepository) Close() {
	repo.consumer.Close()
}

func (repo LoggregatorLogsRepository) RecentLogsFor(appGuid string) ([]*logmessage.LogMessage, error) {
	messages, err := repo.consumer.Recent(appGuid, repo.config.AccessToken())
	consumer.SortRecent(messages)
	return messages, err
}

func (repo LoggregatorLogsRepository) TailLogsFor(appGuid string, bufferTime time.Duration, onConnect func(), onMessage func(*logmessage.LogMessage)) error {
	endpoint := repo.config.LoggregatorEndpoint()
	if endpoint == "" {
		return errors.New("Loggregator endpoint missing from config file")
	}

	repo.consumer.SetOnConnectCallback(onConnect)
	logChan, err := repo.consumer.Tail(appGuid, repo.config.AccessToken())
	if err != nil {
		return err
	}

	bufferMessages(logChan, onMessage, bufferTime)
	return nil
}

func bufferMessages(logChan <-chan *logmessage.LogMessage, onMessage func(*logmessage.LogMessage), bufferTime time.Duration) {
	messageQueue := NewSortedMessageQueue(bufferTime, func() time.Time {
		return time.Now()
	})

	for {
		sendMessages(messageQueue, onMessage)

		select {
		case msg, ok := <-logChan:
			if !ok {
				return
			}
			messageQueue.PushMessage(msg)
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func sendMessages(queue *SortedMessageQueue, onMessage func(*logmessage.LogMessage)) {
	for queue.NextTimestamp() < time.Now().UnixNano() {
		msg := queue.PopMessage()
		onMessage(msg)
	}
}
