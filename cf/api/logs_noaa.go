package api

import (
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
}

type logNoaaRepository struct {
	config         core_config.Reader
	consumer       NoaaConsumer
	tokenRefresher authentication.TokenRefresher
}

func NewLogsNoaaRepository(config core_config.Reader, consumer NoaaConsumer, tr authentication.TokenRefresher) LogsNoaaRepository {
	return &logNoaaRepository{
		config:         config,
		consumer:       consumer,
		tokenRefresher: tr,
	}
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

func (*logNoaaRepository) TailNoaaLogsFor(appGuid string, onConnect func(), onMessage func(*events.LogMessage)) error {
	return nil
}
