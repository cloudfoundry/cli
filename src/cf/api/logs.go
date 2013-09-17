package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"code.google.com/p/go.net/websocket"
	"code.google.com/p/gogoprotobuf/proto"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const LOGGREGATOR_REDIRECTOR_PORT = "4443"

type LogsRepository interface {
	RecentLogsFor(app cf.Application) (logs []*logmessage.LogMessage, apiErr *net.ApiError)
	TailLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.LogMessage)) (err error)
}

type LoggregatorLogsRepository struct {
	config                  *configuration.Configuration
	gateway                 net.Gateway
	loggregatorHostResolver func(string) string
}

func NewLoggregatorLogsRepository(config *configuration.Configuration, gateway net.Gateway, loggregatorHostResolver func(string) string) (repo LoggregatorLogsRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.loggregatorHostResolver = loggregatorHostResolver
	return
}

func (repo LoggregatorLogsRepository) RecentLogsFor(app cf.Application) (logs []*logmessage.LogMessage, apiErr *net.ApiError) {
	host := repo.loggregatorHostResolver(repo.config.Target)
	request, apiErr := repo.gateway.NewRequest("GET", fmt.Sprintf("%s/dump/?app=%s", host, app.Guid), repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}
	request.Header.Del("accept")

	bytes, apiErr := repo.gateway.PerformRequestForResponseBytes(request)

	if apiErr != nil {
		return
	}

	logs, err := logmessage.ParseDumpedLogMessages(bytes)

	if err != nil {
		apiErr = net.NewApiErrorWithError("Could not parse log messages", err)
	}

	return
}

func (repo LoggregatorLogsRepository) TailLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.LogMessage)) (err error) {
	const REDIRECT_ERROR = "REDIRECTED"

	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	tr := &http.Transport{TLSClientConfig: tlsConfig, Proxy: http.ProxyFromEnvironment}

	client := http.Client{
		Transport: tr,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errors.New(REDIRECT_ERROR)
		},
	}

	host := repo.loggregatorHostResolver(repo.config.Target) + ":" + LOGGREGATOR_REDIRECTOR_PORT
	request, apiErr := repo.gateway.NewRequest("GET", fmt.Sprintf("%s/tail/?app=%s", host, app.Guid), repo.config.AccessToken, nil)
	if apiErr != nil {
		err = errors.New(apiErr.Error())
		return
	}

	resp, err := client.Do(request.Request)

	if err != nil && !strings.Contains(err.Error(), REDIRECT_ERROR) {
		return
	}

	location := resp.Header.Get("Location")
	config, err := websocket.NewConfig(location, "http://localhost")
	if err != nil {
		return
	}

	config.Header.Add("Authorization", repo.config.AccessToken)
	config.TlsConfig = tlsConfig

	ws, err := websocket.DialConfig(config)
	if err != nil {
		return
	}

	onConnect()
	go func() {
		for {
			websocket.Message.Send(ws, "I'm alive!")
			time.Sleep(25 * time.Second)
		}
	}()

	for {
		var data []byte
		err = websocket.Message.Receive(ws, &data)
		if err != nil {
			return
		}

		logMessage := new(logmessage.LogMessage)
		msgErr := proto.Unmarshal(data, logMessage)
		if msgErr != nil {
			continue
		}
		onMessage(logMessage)
	}

	return
}

func LoggregatorHost(apiHost string) string {
	re := regexp.MustCompile(`^(https?://)[^\.]+\.(.+)\/?`)
	return re.ReplaceAllString(apiHost, "${1}loggregator.${2}")
}
