// Package loggregator_consumer provides a simple, channel-based API for clients to communicate with
// loggregator servers.
package loggregator_consumer

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var (
	// KeepAlive sets the interval between keep-alive messages sent by the client to loggregator.
	KeepAlive = 25 * time.Second
)

/* LoggregatorConsumer represents the actions that can be performed against a loggregator server.
 */
type LoggregatorConsumer interface {

	//	Tail listens indefinitely for log messages. It returns two channels; the first is populated
	//	with log messages, while the second contains errors (e.g. from parsing messages). It returns
	//	immediately. Call Close() to terminate the connection when you are finished listening.
	//
	//	Messages are presented in the order received from the loggregator server. Chronological or
	//	other ordering is not guaranteed. It is the responsibility of the consumer of these channels
	//	to provide any desired sorting mechanism.
	Tail(appGuid string, authToken string) (<-chan *logmessage.LogMessage, error)

	//	Recent connects to loggregator via its 'dump' endpoint and returns a slice of recent messages.
	//	It does not guarantee any order of the messages; they are in the order returned by loggregator.
	//
	//	The SortRecent method is provided to sort the data returned by this method.
	Recent(appGuid string, authToken string) ([]*logmessage.LogMessage, error)

	// Close terminates the websocket connection to loggregator.
	Close() error

	// SetOnConnectCallback sets a callback function to be called with the websocket connection is established.
	SetOnConnectCallback(func())
}

type consumer struct {
	endpoint  string
	tlsConfig *tls.Config
	ws        *websocket.Conn
	callback  func()
	proxy     func(*http.Request) (*url.URL, error)
}

/* New creates a new consumer to a loggregator endpoint.
 */
func New(endpoint string, tlsConfig *tls.Config, proxy func(*http.Request) (*url.URL, error)) LoggregatorConsumer {
	return &consumer{endpoint: endpoint, tlsConfig: tlsConfig, proxy: proxy}
}

/*
Tail listens indefinitely for log messages. It returns two channels; the first is populated
with log messages, while the second contains errors (e.g. from parsing messages). It returns immediately.
Call Close() to terminate the connection when you are finished listening.

Messages are presented in the order received from the loggregator server. Chronological or other ordering
is not guaranteed. It is the responsibility of the consumer of these channels to provide any desired sorting
mechanism.
*/
func (cnsmr *consumer) Tail(appGuid string, authToken string) (<-chan *logmessage.LogMessage, error) {
	incomingChan := make(chan *logmessage.LogMessage)
	var err error

	tailPath := fmt.Sprintf("/tail/?app=%s", appGuid)
	cnsmr.ws, err = cnsmr.establishWebsocketConnection(tailPath, authToken)

	if err == nil {
		go cnsmr.sendKeepAlive(KeepAlive)

		go func() {
			defer close(incomingChan)
			cnsmr.listenForMessages(incomingChan)
		}()
	}

	return incomingChan, err
}

/*
Recent connects to loggregator via its 'dump' endpoint and returns a slice of recent messages. It does not
guarantee any order of the messages; they are in the order returned by loggregator.

The SortRecent method is provided to sort the data returned by this method.
*/
func (cnsmr *consumer) Recent(appGuid string, authToken string) ([]*logmessage.LogMessage, error) {
	var err error

	dumpPath := fmt.Sprintf("/dump/?app=%s", appGuid)
	cnsmr.ws, err = cnsmr.establishWebsocketConnection(dumpPath, authToken)

	if err != nil {
		return nil, err
	}

	messages := []*logmessage.LogMessage{}
	messageChan := make(chan *logmessage.LogMessage)

	go func() {
		err = cnsmr.listenForMessages(messageChan)
		close(messageChan)
	}()

drainLoop:
	for {
		select {
		case msg, ok := <-messageChan:
			if !ok {
				break drainLoop
			}

			messages = append(messages, msg)
		}
	}

	return messages, nil
}

/* Close terminates the websocket connection to loggregator.
 */
func (cnsmr *consumer) Close() error {
	if cnsmr.ws == nil {
		return errors.New("connection does not exist")
	}

	return cnsmr.ws.Close()
}

func (cnsmr *consumer) SetOnConnectCallback(cb func()) {
	cnsmr.callback = cb
}

/*
SortRecent sorts a slice of LogMessages by timestamp. The sort is stable, so
messages with the same timestamp are sorted in the order that they are received.

The input slice is sorted; the return value is simply a pointer to the same slice.
*/
func SortRecent(messages []*logmessage.LogMessage) []*logmessage.LogMessage {
	sort.Stable(logMessageSlice(messages))
	return messages
}

type logMessageSlice []*logmessage.LogMessage

func (lms logMessageSlice) Len() int {
	return len(lms)
}

func (lms logMessageSlice) Less(i, j int) bool {
	return *(lms[i]).Timestamp < *(lms[j]).Timestamp
}

func (lms logMessageSlice) Swap(i, j int) {
	lms[i], lms[j] = lms[j], lms[i]
}

func (cnsmr *consumer) sendKeepAlive(interval time.Duration) {
	for {
		err := cnsmr.ws.WriteMessage(websocket.TextMessage, []byte("I'm alive!"))
		if err != nil {
			return
		}
		time.Sleep(interval)
	}
}

func (cnsmr *consumer) listenForMessages(msgChan chan<- *logmessage.LogMessage) error {
	defer cnsmr.ws.Close()

	for {
		var data []byte

		_, data, err := cnsmr.ws.ReadMessage()
		if err != nil {
			return err
		}

		msg, msgErr := logmessage.ParseMessage(data)
		if msgErr != nil {
			continue
		}

		msgChan <- msg.GetLogMessage()
	}
}

func (cnsmr *consumer) establishWebsocketConnection(path string, authToken string) (*websocket.Conn, error) {
	header := http.Header{"Origin": []string{"http://localhost"}, "Authorization": []string{authToken}}
	dialer := websocket.Dialer{NetDial: cnsmr.proxyDial, TLSClientConfig: cnsmr.tlsConfig}

	ws, resp, err := dialer.Dial(cnsmr.endpoint+path, header)

	if err == nil && cnsmr.callback != nil {
		cnsmr.callback()
	}
	if resp != nil && resp.StatusCode == http.StatusUnauthorized {
		bodyData := make([]byte, 4096)
		resp.Body.Read(bodyData)
		resp.Body.Close()
		err = NewUnauthorizedError(string(bodyData))
	}

	return ws, err
}

func (cnsmr *consumer) proxyDial(network, addr string) (net.Conn, error) {
	targetUrl, err := url.Parse("http://" + addr)
	if err != nil {
		return nil, err
	}

	proxy := cnsmr.proxy
	if proxy == nil {
		proxy = http.ProxyFromEnvironment
	}

	proxyUrl, err := proxy(&http.Request{URL: targetUrl})
	if err != nil {
		return nil, err
	}
	if proxyUrl == nil {
		return net.Dial(network, addr)
	}

	proxyConn, err := net.Dial(network, proxyUrl.Host)
	if err != nil {
		return nil, err
	}

	connectReq := &http.Request{
		Method: "CONNECT",
		URL:    targetUrl,
		Host:   targetUrl.Host,
		Header: make(http.Header),
	}
	connectReq.Write(proxyConn)

	connectResp, err := http.ReadResponse(bufio.NewReader(proxyConn), connectReq)
	if err != nil {
		proxyConn.Close()
		return nil, err
	}
	if connectResp.StatusCode != http.StatusOK {
		f := strings.SplitN(connectResp.Status, " ", 2)
		proxyConn.Close()
		return nil, errors.New(f[1])
	}

	return proxyConn, nil
}
