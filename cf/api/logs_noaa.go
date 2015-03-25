package api

import (
	"errors"
	"time"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"

	"github.com/cloudfoundry/noaa"
	noaa_errors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/noaa/events"
)

type LogsNoaaRepository interface {
	GetContainerMetrics(string, []models.AppInstanceFields) ([]models.AppInstanceFields, error)
	RecentLogsFor(appGuid string) ([]*events.LogMessage, error)
	TailNoaaLogsFor(appGuid string, onConnect func(), onMessage func(*events.LogMessage)) error
	Close()
}

type logNoaaRepository struct {
	config         core_config.Reader
	consumer       NoaaConsumer
	tokenRefresher authentication.TokenRefresher
	messageQueue   *SortedMessageQueue
	onMessage      func(*events.LogMessage)
}

var BufferTime time.Duration = 5 * time.Second

func NewLogsNoaaRepository(config core_config.Reader, consumer NoaaConsumer, tr authentication.TokenRefresher) LogsNoaaRepository {
	return &logNoaaRepository{
		config:         config,
		consumer:       consumer,
		tokenRefresher: tr,
		messageQueue:   NewSortedMessageQueue(BufferTime, time.Now),
	}
}

func (l *logNoaaRepository) Close() {
	l.consumer.Close()
	l.flushMessageQueue()
}

func (l *logNoaaRepository) GetContainerMetrics(appGuid string, instances []models.AppInstanceFields) ([]models.AppInstanceFields, error) {
	metrics, err := l.consumer.GetContainerMetrics(appGuid, l.config.AccessToken())
	switch err.(type) {
	case nil: // do nothing
	case *noaa_errors.UnauthorizedError:
		l.tokenRefresher.RefreshAuthToken()
		metrics, err = l.consumer.GetContainerMetrics(appGuid, l.config.AccessToken())
	default:
		return instances, err
	}

	for _, m := range metrics {
		instances[int(*m.InstanceIndex)].MemUsage = int64(m.GetMemoryBytes())
		instances[int(*m.InstanceIndex)].CpuUsage = m.GetCpuPercentage()
		instances[int(*m.InstanceIndex)].DiskUsage = int64(m.GetDiskBytes())
	}

	return instances, nil
}

func (l *logNoaaRepository) RecentLogsFor(appGuid string) ([]*events.LogMessage, error) {
	logs, err := l.consumer.RecentLogs(appGuid, l.config.AccessToken())

	switch err.(type) {
	case nil: // do nothing
	case *noaa_errors.UnauthorizedError:
		l.tokenRefresher.RefreshAuthToken()
		logs, err = l.consumer.RecentLogs(appGuid, l.config.AccessToken())
	default:
		return logs, err
	}

	return noaa.SortRecent(logs), err
}

func (l *logNoaaRepository) TailNoaaLogsFor(appGuid string, onConnect func(), onMessage func(*events.LogMessage)) error {
	l.onMessage = onMessage
	endpoint := l.config.DopplerEndpoint()
	if endpoint == "" {
		return errors.New(T("Loggregator endpoint missing from config file"))
	}

	l.consumer.SetOnConnectCallback(onConnect)

	logChan := make(chan *events.LogMessage)
	errChan := make(chan error)
	closeChan := make(chan struct{})
	go l.consumer.TailingLogs(appGuid, l.config.AccessToken(), logChan, errChan, closeChan)

	for {
		sendNoaaMessages(l.messageQueue, onMessage)

		select {
		case err := <-errChan:
			switch err.(type) {
			case nil: // do nothing
			case *noaa_errors.UnauthorizedError:
				if closeChan != nil {
					l.tokenRefresher.RefreshAuthToken()
					close(closeChan)
					closeChan = nil
					go l.consumer.TailingLogs(appGuid, l.config.AccessToken(), logChan, errChan, make(chan struct{}))
				} else {
					l.Close()
					return err
				}
			default:
				l.Close()
				return err
			}
		case log := <-logChan:
			l.messageQueue.PushMessage(log)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func sendNoaaMessages(queue *SortedMessageQueue, onMessage func(*events.LogMessage)) {
	for queue.NextTimestamp() < time.Now().UnixNano() {
		msg := queue.PopMessage()
		onMessage(msg)
	}
}

func (l *logNoaaRepository) flushMessageQueue() {
	if l.onMessage == nil {
		return
	}

	for {
		message := l.messageQueue.PopMessage()
		if message == nil {
			break
		}

		l.onMessage(message)
	}

	l.onMessage = nil
}
