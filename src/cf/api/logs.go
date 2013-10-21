package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"cf/terminal"
	"code.google.com/p/go.net/websocket"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
)

type LogsRepository interface {
	RecentLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message)) (err error)
	TailLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message), printInterval time.Duration) (err error)
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

func (repo LoggregatorLogsRepository) RecentLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message)) (err error) {
	host, apiResponse := repo.endpointRepo.GetEndpoint(cf.LoggregatorEndpointKey)
	if apiResponse.IsNotSuccessful() {
		err = errors.New(apiResponse.Message)
		return
	}
	location := host + fmt.Sprintf("/dump/?app=%s", app.Guid)
	return repo.connectToWebsocket(location, app, onConnect, onMessage, 0*time.Nanosecond)
}

func (repo LoggregatorLogsRepository) TailLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message), printTimeBuffer time.Duration) error {
	host, apiResponse := repo.endpointRepo.GetEndpoint(cf.LoggregatorEndpointKey)
	if apiResponse.IsNotSuccessful() {
		return errors.New(apiResponse.Message)
	}
	location := host + fmt.Sprintf("/tail/?app=%s", app.Guid)
	return repo.connectToWebsocket(location, app, onConnect, onMessage, printTimeBuffer)
}

func (repo LoggregatorLogsRepository) connectToWebsocket(location string, app cf.Application, onConnect func(), onMessage func(*logmessage.Message), printTimeBuffer time.Duration) (err error) {
	if net.TraceEnabled() {
		fmt.Printf("\n%s %s\n", terminal.HeaderColor("CONNECTING TO WEBSOCKET:"), location)
	}

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

	msgChan := make(chan *logmessage.Message, 1000)
	outputtableMsgChan := make(chan *logmessage.Message, 1000)

	go repo.sendKeepAlive(ws)

	go repo.listenForMessages(ws, msgChan)
	go MakeAndStartMessageSorter(msgChan, outputtableMsgChan, printTimeBuffer)
	for msg := range outputtableMsgChan {
		onMessage(msg)
	}

	return
}

func MakeAndStartMessageSorter(inputChan <-chan *logmessage.Message, outputChan chan<- *logmessage.Message, printTimeBuffer time.Duration) {
	messageQueue := NewPriorityMessageQueue(printTimeBuffer)

	controlChan := make(chan bool)
	flushLastMessages := false

	go func() {
		for msg := range inputChan {
			messageQueue.PushMessage(msg)
		}
		controlChan <- true
	}()

	go func() {
	OutputLoop:
		for {
			select {
			case <-controlChan:
				flushLastMessages = true
			case <-time.After(10 * time.Millisecond):
				if flushLastMessages {
					for {
						msg := messageQueue.PopMessage()
						if msg == nil {
							break
						}
						outputChan <- msg
					}
					break OutputLoop
				}

				for messageQueue.NextTimestamp() < time.Now().UnixNano() {
					msg := messageQueue.PopMessage()
					outputChan <- msg
				}
			}
		}
		close(outputChan)
	}()
}

func (repo LoggregatorLogsRepository) sendKeepAlive(ws *websocket.Conn) {
	for {
		websocket.Message.Send(ws, "I'm alive!")
		time.Sleep(25 * time.Second)
	}
}

func (repo LoggregatorLogsRepository) listenForMessages(ws *websocket.Conn, msgChan chan<- *logmessage.Message) {
	var err error
	defer close(msgChan)
	for {
		var data []byte
		err = websocket.Message.Receive(ws, &data)
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
