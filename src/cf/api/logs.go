package api

import (
	"cf"
	"cf/configuration"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"regexp"
)

type LogsRepository interface {
	RecentLogsFor(app cf.Application) (logs []*logmessage.LogMessage, apiErr *ApiError)
}

type LoggregatorLogsRepository struct {
	config              *configuration.Configuration
	apiClient           ApiClient
	loggregatorHostName func(string) string
}

func NewLoggregatorLogsRepository(config *configuration.Configuration, client ApiClient, loggregatorHostName func(string) string) (repo LoggregatorLogsRepository) {
	repo.config = config
	repo.apiClient = client
	repo.loggregatorHostName = loggregatorHostName
	return
}

func (l LoggregatorLogsRepository) RecentLogsFor(app cf.Application) (logs []*logmessage.LogMessage, apiErr *ApiError) {
	host := l.loggregatorHostName(l.config.Target)
	request, apiErr := NewRequest("GET", fmt.Sprintf("%s/dump/?app=%s", host, app.Guid), l.config.AccessToken, nil)
	if apiErr != nil {
		return
	}
	request.Header.Del("accept")

	bytes, apiErr := l.apiClient.PerformRequestForResponseBytes(request)

	if apiErr != nil {
		return
	}

	logs, err := logmessage.ParseDumpedLogMessages(bytes)

	if err != nil {
		apiErr = NewApiErrorWithError("Could not parse log messages", err)
	}

	return
}

func LoggregatorHost(apiHost string) string {
	re := regexp.MustCompile(`^(https?://)[^\.]+\.(.+)\/?`)
	return re.ReplaceAllString(apiHost, "${1}loggregator.${2}")
}
