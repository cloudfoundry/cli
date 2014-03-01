package api

import (
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
	config       configuration.Reader
	endpointRepo EndpointRepository
}

func NewLoggregatorLogsRepository(config configuration.Reader, endpointRepo EndpointRepository) (repo LoggregatorLogsRepository) {
	repo.config = config
	repo.endpointRepo = endpointRepo
	return
}

func (repo LoggregatorLogsRepository) RecentLogsFor(appGuid string, onConnect func(), logChan chan *logmessage.Message) (err error) {
	host, apiErr := repo.endpointRepo.GetLoggregatorEndpoint()
	if apiErr != nil {
		err = errors.New(apiErr.Error())
		return
	}

	location := fmt.Sprintf("%s/dump/?app=%s", host, appGuid)
	stopLoggingChan := make(chan bool)
	defer close(stopLoggingChan)

	return repo.connectToWebsocket(location, onConnect, logChan, stopLoggingChan, 0*time.Nanosecond)
}

func (repo LoggregatorLogsRepository) TailLogsFor(appGuid string, onConnect func(), logChan chan *logmessage.Message, stopLoggingChan chan bool, printTimeBuffer time.Duration) error {
	host, apiErr := repo.endpointRepo.GetLoggregatorEndpoint()
	if apiErr != nil {
		return errors.New(apiErr.Error())
	}
	location := host + fmt.Sprintf("/tail/?app=%s", appGuid)
	return repo.connectToWebsocket(location, onConnect, logChan, stopLoggingChan, printTimeBuffer)
}

func (repo LoggregatorLogsRepository) connectToWebsocket(location string, onConnect func(), outputChan chan *logmessage.Message, stopLoggingChan chan bool, printTimeBuffer time.Duration) (err error) {
	trace.Logger.Printf("\n%s %s\n", terminal.HeaderColor("CONNECTING TO WEBSOCKET:"), location)

	inputChan := make(chan *logmessage.Message, LogBufferSize)
	messageQueue := NewSortedMessageQueue(printTimeBuffer, time.Now)

	wsConfig, err := websocket.NewConfig(location, "http://localhost")
	if err != nil {
		return
	}

	wsConfig.Header.Add("Authorization", repo.config.AccessToken())
	wsConfig.TlsConfig = &tls.Config{InsecureSkipVerify: true}

	ws, err := websocket.DialConfig(wsConfig)
	if err != nil {
		return
	}

	defer func() {
		ws.Close()
		repo.drainRemainingMessages(messageQueue, inputChan, outputChan)
	}()

	onConnect()

	go repo.sendKeepAlive(ws)

	go func() {
		defer close(inputChan)
		repo.listenForMessages(ws, inputChan)
	}()

	repo.processMessages(messageQueue, inputChan, outputChan, stopLoggingChan)

	return
}

func (repo LoggregatorLogsRepository) processMessages(messageQueue *SortedMessageQueue, inputChan <-chan *logmessage.Message, outputChan chan *logmessage.Message, stopLoggingChan <-chan bool) {
	for {
		select {
		case msg, ok := <-inputChan:
			if ok {
				messageQueue.PushMessage(msg)
			} else {
				return
			}
		case <-stopLoggingChan:
			return
		case <-time.After(10 * time.Millisecond):
			for messageQueue.NextTimestamp() < time.Now().UnixNano() {
				msg := messageQueue.PopMessage()
				outputChan <- msg
			}
		}
	}
}

func (repo LoggregatorLogsRepository) drainRemainingMessages(messageQueue *SortedMessageQueue, inputChan <-chan *logmessage.Message, outputChan chan *logmessage.Message) {
	for msg := range inputChan {
		messageQueue.PushMessage(msg)
	}

	for {
		msg := messageQueue.PopMessage()
		if msg == nil {
			break
		}
		outputChan <- msg
	}
}

func (repo LoggregatorLogsRepository) sendKeepAlive(ws *websocket.Conn) {
	for {
		websocket.Message.Send(ws, "I'm alive!")
		time.Sleep(25 * time.Second)
	}
}

func (repo LoggregatorLogsRepository) listenForMessages(ws *websocket.Conn, msgChan chan<- *logmessage.Message) {
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
