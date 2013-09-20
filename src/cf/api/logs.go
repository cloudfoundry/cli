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
	"sort"
)

const LOGGREGATOR_REDIRECTOR_PORT = "4443"

type LogsRepository interface {
	RecentLogsFor(app cf.Application) (logs []*logmessage.LogMessage, apiErr *net.ApiError)
	TailLogsFor(app cf.Application, onConnect func(), onMessage func(logmessage.LogMessage)) (err error)
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

func (repo LoggregatorLogsRepository) TailLogsFor(app cf.Application, onConnect func(), onMessage func(logmessage.LogMessage)) (err error) {
	location, err := repo.getLocationFromRedirector(app)

	config, err := websocket.NewConfig(location, "http://localhost")
	if err != nil {
		return
	}

	config.Header.Add("Authorization", repo.config.AccessToken)
	config.TlsConfig = &tls.Config{InsecureSkipVerify: true}

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

	ticks := time.Duration(5)
	c := make(chan logmessage.LogMessage, 100*ticks)

	go func() {
		for {
			var data []byte
			err = websocket.Message.Receive(ws, &data)
			if err != nil {
				return
			}

			logMessage := logmessage.LogMessage{}

			msgErr := proto.Unmarshal(data, &logMessage)
			if msgErr != nil {
				continue
			}
			c <- logMessage
		}
	}()

	tickerChan := time.Tick(ticks *time.Second)
	sortableMsg := &sortableLogMessages{}
	var msg logmessage.LogMessage

	for {
		select {
		case msg = <- c:
			sortableMsg.Messages = append(sortableMsg.Messages, msg)
		case <- tickerChan:
			sort.Sort(sortableMsg)
			for _, msg = range sortableMsg.Messages {
				onMessage(msg)
			}
			sortableMsg.Messages = []logmessage.LogMessage{}
		}
	}

	return
}

type sortableLogMessages struct {
	Messages []logmessage.LogMessage
}

func (sort *sortableLogMessages) Len() int {
	return len(sort.Messages)
}

func (sort *sortableLogMessages) Less(i, j int) bool {
	msgI := sort.Messages[i]
	msgJ := sort.Messages[j]
	return *msgI.Timestamp < *msgJ.Timestamp
}

func (sort *sortableLogMessages) Swap(i, j int) {
	sort.Messages[i], sort.Messages[j] = sort.Messages[j], sort.Messages[i]
}

func LoggregatorHost(apiHost string) string {
	re := regexp.MustCompile(`^(https?://)[^\.]+\.(.+)\/?`)
	return re.ReplaceAllString(apiHost, "${1}loggregator.${2}")
}

func (repo LoggregatorLogsRepository) getLocationFromRedirector(app cf.Application) (loc string, err error) {
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

	loc = resp.Header.Get("Location")
	return
}
