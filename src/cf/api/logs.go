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
	"sort"
	"strings"
	"time"
)

const LOGGREGATOR_REDIRECTOR_PORT = "4443"

type LogsRepository interface {
	RecentLogsFor(app cf.Application, onConnect func(), onMessage func(logmessage.LogMessage)) (err error)
	TailLogsFor(app cf.Application, onConnect func(), onMessage func(logmessage.LogMessage), printInterval time.Duration) (err error)
}

type LoggregatorLogsRepository struct {
	config                  configuration.Configuration
	gateway                 net.Gateway
	loggregatorHostResolver func(string) string
}

func NewLoggregatorLogsRepository(config configuration.Configuration, gateway net.Gateway, loggregatorHostResolver func(string) string) (repo LoggregatorLogsRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.loggregatorHostResolver = loggregatorHostResolver
	return
}

func (repo LoggregatorLogsRepository) RecentLogsFor(app cf.Application, onConnect func(), onMessage func(logmessage.LogMessage)) (err error) {
	location, err := repo.getLocationFromRedirector(fmt.Sprintf("/dump/?app=%s", app.Guid))
	if err != nil {
		return
	}
	return repo.connectToWebsocket(location, app, onConnect, onMessage, nil)
}

func (repo LoggregatorLogsRepository) TailLogsFor(app cf.Application, onConnect func(), onMessage func(logmessage.LogMessage), printInterval time.Duration) (err error) {
	location, err := repo.getLocationFromRedirector(fmt.Sprintf("/tail/?app=%s", app.Guid))
	if err != nil {
		return
	}
	return repo.connectToWebsocket(location, app, onConnect, onMessage, time.Tick(printInterval*time.Second))
}

func (repo LoggregatorLogsRepository) connectToWebsocket(location string, app cf.Application, onConnect func(), onMessage func(logmessage.LogMessage), tickerChan <-chan time.Time) (err error) {
	const EOF_ERROR = "EOF"

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

	msgChan := make(chan logmessage.LogMessage, 1000)
	errChan := make(chan error, 0)

	go repo.listenForMessages(ws, msgChan, errChan)
	go repo.sendKeepAlive(ws)

	sortableMsg := &sortableLogMessages{}

Loop:
	for {
		select {
		case err = <-errChan:
			break Loop
		case msg, ok := <-msgChan:
			if !ok {
				break Loop
			}
			sortableMsg.Messages = append(sortableMsg.Messages, msg)
		case <-tickerChan:
			invokeCallbackWithSortedMessages(sortableMsg, onMessage)
			sortableMsg.Messages = []logmessage.LogMessage{}
		}
		if err != nil {
			break
		}
	}

	if tickerChan == nil {
		invokeCallbackWithSortedMessages(sortableMsg, onMessage)
	}

	if err.Error() == EOF_ERROR {
		err = nil
	}

	return
}

func invokeCallbackWithSortedMessages(messages *sortableLogMessages, callback func(logmessage.LogMessage)) {
	sort.Sort(messages)
	for _, msg := range messages.Messages {
		callback(msg)
	}
}

func (repo LoggregatorLogsRepository) sendKeepAlive(ws *websocket.Conn) {
	for {
		websocket.Message.Send(ws, "I'm alive!")
		time.Sleep(25 * time.Second)
	}
}

func (repo LoggregatorLogsRepository) listenForMessages(ws *websocket.Conn, msgChan chan<- logmessage.LogMessage, errChan chan<- error) {
	var err error
	defer close(msgChan)
	for {
		var data []byte
		err = websocket.Message.Receive(ws, &data)
		if err != nil {
			errChan <- err
			break
		}

		logMessage := logmessage.LogMessage{}

		msgErr := proto.Unmarshal(data, &logMessage)
		if msgErr != nil {
			continue
		}
		msgChan <- logMessage
	}
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

func (repo LoggregatorLogsRepository) getLocationFromRedirector(requestPathAndQueryParams string) (loc string, err error) {
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
	request, apiErr := repo.gateway.NewRequest("GET", host+requestPathAndQueryParams, repo.config.AccessToken, nil)
	if apiErr != nil {
		err = errors.New(apiErr.Error())
		return
	}

	resp, err := client.Do(request.Request)

	if strings.Contains(err.Error(), REDIRECT_ERROR) {
		err = nil
	}

	if err != nil {
		return
	}

	loc = resp.Header.Get("Location")
	return
}
