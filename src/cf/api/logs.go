package api

import (
	"cf"
	"cf/configuration"
	"cf/terminal"
	"cf/trace"
	"code.google.com/p/go.net/websocket"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
)

const LogBufferSize = 1024

type LogsRepository interface {
	RecentLogsFor(appGuid string, onConnect func(), logChan chan *logmessage.Message) (err error)
	TailLogsFor(appGuid string, onConnect func(), logChan chan *logmessage.Message, stopLoggingChan chan bool, printInterval time.Duration) (err error)
}

type LoggregatorLogsRepository struct {
	config       *configuration.Configuration
	endpointRepo EndpointRepository
}

func NewLoggregatorLogsRepository(config *configuration.Configuration, endpointRepo EndpointRepository) (repo LoggregatorLogsRepository) {
	repo.config = config
	repo.endpointRepo = endpointRepo
	return
}

func (repo LoggregatorLogsRepository) RecentLogsFor(appGuid string, onConnect func(), logChan chan *logmessage.Message) (err error) {
	host, apiResponse := repo.endpointRepo.GetEndpoint(cf.LoggregatorEndpointKey)
	if apiResponse.IsNotSuccessful() {
		err = errors.New(apiResponse.Message)
		return
	}

	location := host + fmt.Sprintf("/dump/?app=%s", appGuid)
	stopLoggingChan := make(chan bool)
	defer close(stopLoggingChan)

	return repo.connectToWebsocket(location, onConnect, logChan, stopLoggingChan, 0*time.Nanosecond)
}

func (repo LoggregatorLogsRepository) TailLogsFor(appGuid string, onConnect func(), logChan chan *logmessage.Message, stopLoggingChan chan bool, printTimeBuffer time.Duration) error {
	host, apiResponse := repo.endpointRepo.GetEndpoint(cf.LoggregatorEndpointKey)
	if apiResponse.IsNotSuccessful() {
		return errors.New(apiResponse.Message)
	}
	location := host + fmt.Sprintf("/tail/?app=%s", appGuid)
	return repo.connectToWebsocket(location, onConnect, logChan, stopLoggingChan, printTimeBuffer)
}

func (repo LoggregatorLogsRepository) connectToWebsocket(location string, onConnect func(), outputChan chan *logmessage.Message, stopLoggingChan chan bool, printTimeBuffer time.Duration) (err error) {
	trace.Logger.Printf("\n%s %s\n", terminal.HeaderColor("CONNECTING TO WEBSOCKET:"), location)

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
	defer ws.Close()

	onConnect()

	go repo.sendKeepAlive(ws)

	inputChan := make(chan *logmessage.Message, LogBufferSize)
	stopInputChan := make(chan bool, 1)
	go func() {
		defer close(stopInputChan)
		defer close(inputChan)
		repo.listenForMessages(ws, inputChan, stopInputChan)
	}()

	messageQueue := &SortedMessageQueue{printTimeBuffer: printTimeBuffer}

	repo.processMessages(messageQueue, inputChan, outputChan, stopLoggingChan, stopInputChan)

	return
}

func (repo LoggregatorLogsRepository) processMessages(messageQueue *SortedMessageQueue, inputChan <-chan *logmessage.Message, outputChan chan *logmessage.Message, stopLoggingChan <-chan bool, stopInputChan <-chan bool) {
	flushLastMessages := func() {
		for {
			msg := messageQueue.PopMessage()
			if msg == nil {
				break
			}
			outputChan <- msg
		}
	}

	for {
		select {
		case msg, ok := <-inputChan:
			if ok {
				messageQueue.PushMessage(msg)
			}
		case <-stopInputChan:
			flushLastMessages()
			return
		case <-stopLoggingChan:
			flushLastMessages()
			return
		case <-time.After(10 * time.Millisecond):
			for messageQueue.NextTimestamp() < time.Now().UnixNano() {
				msg := messageQueue.PopMessage()
				outputChan <- msg
			}
		}
	}
}

func (repo LoggregatorLogsRepository) sendKeepAlive(ws *websocket.Conn) {
	for {
		websocket.Message.Send(ws, "I'm alive!")
		time.Sleep(25 * time.Second)
	}
}

func (repo LoggregatorLogsRepository) listenForMessages(ws *websocket.Conn, msgChan chan<- *logmessage.Message, stopInputChan chan<- bool) {
	defer func() {
		stopInputChan <- true
	}()

	for {
		var data []byte
		err := websocket.Message.Receive(ws, &data)
		if err != nil {
			break
		}

		msg, msgErr := logmessage.ParseMessage(data)
		if msgErr != nil {
			continue
		}
		msgChan <- msg
	}
}
