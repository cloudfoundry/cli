package consumer

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cloudfoundry/noaa"
	noaa_errors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
)

// TailingLogs listens indefinitely for log messages only; other event types
// are dropped.
// Whenever an error is encountered, the error will be sent down the error
// channel and TailingLogs will attempt to reconnect up to 5 times.  After
// five failed reconnection attempts, TailingLogs will give up and close the
// error and LogMessage channels.
//
// If c is closed, the returned channels will both be closed.
//
// Errors must be drained from the returned error channel for it to continue
// retrying; if they are not drained, the connection attempts will hang.
func (c *Consumer) TailingLogs(appGuid, authToken string) (<-chan *events.LogMessage, <-chan error) {
	return c.tailingLogs(appGuid, authToken, maxRetries)
}

// TailingLogsWithoutReconnect functions identically to TailingLogs but without
// any reconnect attempts when errors occur.
func (c *Consumer) TailingLogsWithoutReconnect(appGuid string, authToken string) (<-chan *events.LogMessage, <-chan error) {
	return c.tailingLogs(appGuid, authToken, 0)
}

// Stream listens indefinitely for all log and event messages.
//
// Messages are presented in the order received from the loggregator server.
// Chronological or other ordering is not guaranteed. It is the responsibility
// of the consumer of these channels to provide any desired sorting mechanism.
//
// Whenever an error is encountered, the error will be sent down the error
// channel and Stream will attempt to reconnect up to 5 times.  After five
// failed reconnection attempts, Stream will give up and close the error and
// Envelope channels.
func (c *Consumer) Stream(appGuid string, authToken string) (outputChan <-chan *events.Envelope, errorChan <-chan error) {
	return c.runStream(appGuid, authToken, maxRetries)
}

// StreamWithoutReconnect functions identically to Stream but without any
// reconnect attempts when errors occur.
func (c *Consumer) StreamWithoutReconnect(appGuid string, authToken string) (<-chan *events.Envelope, <-chan error) {
	return c.runStream(appGuid, authToken, 0)
}

// Firehose streams all data. All clients with the same subscriptionId will
// receive a proportionate share of the message stream.  Each pool of clients
// will receive the entire stream.
//
// Messages are presented in the order received from the loggregator server.
// Chronological or other ordering is not guaranteed. It is the responsibility
// of the consumer of these channels to provide any desired sorting mechanism.
//
// Whenever an error is encountered, the error will be sent down the error
// channel and Firehose will attempt to reconnect up to 5 times.  After five
// failed reconnection attempts, Firehose will give up and close the error and
// Envelope channels.
func (c *Consumer) Firehose(subscriptionId string, authToken string) (<-chan *events.Envelope, <-chan error) {
	return c.firehose(subscriptionId, authToken, 5)
}

// FirehoseWithoutReconnect functions identically to Firehose but without any
// reconnect attempts when errors occur.
func (c *Consumer) FirehoseWithoutReconnect(subscriptionId string, authToken string) (<-chan *events.Envelope, <-chan error) {
	return c.firehose(subscriptionId, authToken, 0)
}

// SetDebugPrinter sets the websocket connection to write debug information to
// debugPrinter.
func (c *Consumer) SetDebugPrinter(debugPrinter noaa.DebugPrinter) {
	c.debugPrinter = debugPrinter
}

// SetOnConnectCallback sets a callback function to be called with the
// websocket connection is established.
func (c *Consumer) SetOnConnectCallback(cb func()) {
	c.callbackLock.Lock()
	defer c.callbackLock.Unlock()
	c.callback = cb
}

// Close terminates the websocket connection to traffic controller.  It will
// return an error if it has problems closing the connection.  If there is no
// connection to close, the consumer will be closed so that no further
// operations will be performed (including any pending retries), and an error
// declaring "connection does not exist" will be returned.
func (c *Consumer) Close() error {
	c.conLock.Lock()
	defer c.conLock.Unlock()
	defer c.close()
	if c.ws == nil {
		return errors.New("connection does not exist")
	}

	c.ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Time{})
	return c.ws.Close()
}

// Closed returns whether or not Close has been called.
func (c *Consumer) Closed() bool {
	c.closedLock.Lock()
	defer c.closedLock.Unlock()
	return c.closed
}

func (c *Consumer) SetIdleTimeout(idleTimeout time.Duration) {
	c.idleTimeout = idleTimeout
}

func (c *Consumer) close() {
	c.closedLock.Lock()
	defer c.closedLock.Unlock()
	c.closed = true
}

func (c *Consumer) onConnectCallback() func() {
	c.callbackLock.RLock()
	defer c.callbackLock.RUnlock()
	return c.callback
}

func (c *Consumer) tailingLogs(appGuid, authToken string, retries uint) (<-chan *events.LogMessage, <-chan error) {
	outputs := make(chan *events.LogMessage)
	errors := make(chan error, 1)
	callback := func(env *events.Envelope) {
		if env.GetEventType() == events.Envelope_LogMessage {
			outputs <- env.GetLogMessage()
		}
	}
	go func() {
		defer close(errors)
		defer close(outputs)
		c.streamAppData(appGuid, authToken, callback, errors, retries)
	}()
	return outputs, errors
}

func (c *Consumer) runStream(appGuid, authToken string, retries uint) (<-chan *events.Envelope, <-chan error) {
	outputs := make(chan *events.Envelope)
	errors := make(chan error, 1)

	callback := func(env *events.Envelope) {
		outputs <- env
	}

	go func() {
		defer close(errors)
		defer close(outputs)
		c.streamAppData(appGuid, authToken, callback, errors, retries)
	}()
	return outputs, errors
}

func (c *Consumer) streamAppData(appGuid, authToken string, callback func(*events.Envelope), errors chan<- error, retries uint) {
	streamPath := fmt.Sprintf("/apps/%s/stream", appGuid)
	action := func() error {
		return c.stream(streamPath, authToken, callback)
	}
	c.retryAction(action, errors, retries)
}

func (c *Consumer) firehose(subID, authToken string, retries uint) (<-chan *events.Envelope, <-chan error) {
	outputs := make(chan *events.Envelope)
	errors := make(chan error, 1)
	callback := func(env *events.Envelope) {
		outputs <- env
	}

	streamPath := "/firehose/" + subID
	action := func() error {
		return c.stream(streamPath, authToken, callback)
	}
	go func() {
		defer close(errors)
		defer close(outputs)
		c.retryAction(action, errors, retries)
	}()
	return outputs, errors
}

func (c *Consumer) stream(streamPath string, authToken string, callback func(*events.Envelope)) error {
	var err error

	c.conLock.Lock()
	c.ws, err = c.establishWebsocketConnection(streamPath, authToken)
	c.conLock.Unlock()

	if err != nil {
		return err
	}

	return c.listenForMessages(callback)
}

func (c *Consumer) listenForMessages(callback func(*events.Envelope)) error {
	defer c.ws.Close()

	for {
		if c.idleTimeout != 0 {
			c.ws.SetReadDeadline(time.Now().Add(c.idleTimeout))
		}
		_, data, err := c.ws.ReadMessage()

		// If the connection was closed (i.e. if c.Close() was called), we
		// will have a non-nil error, but we want to return a nil error.
		if c.Closed() {
			return nil
		}

		if err != nil {
			return err
		}

		envelope := &events.Envelope{}
		err = proto.Unmarshal(data, envelope)
		if err != nil {
			continue
		}

		callback(envelope)
	}
}

func (c *Consumer) retryAction(action func() error, errors chan<- error, retries uint) {
	reconnectAttempts := uint(0)

	oldConnectCallback := c.onConnectCallback()
	defer c.SetOnConnectCallback(oldConnectCallback)

	c.callback = func() {
		reconnectAttempts = 0
		if oldConnectCallback != nil {
			oldConnectCallback()
		}
	}

	for ; reconnectAttempts <= retries; reconnectAttempts++ {
		if c.Closed() {
			return
		}

		errors <- action()
		time.Sleep(reconnectTimeout)
	}
}

func (c *Consumer) establishWebsocketConnection(path string, authToken string) (*websocket.Conn, error) {
	header := http.Header{"Origin": []string{"http://localhost"}, "Authorization": []string{authToken}}
	url := c.trafficControllerUrl + path

	c.debugPrinter.Print("WEBSOCKET REQUEST:",
		"GET "+path+" HTTP/1.1\n"+
			"Host: "+c.trafficControllerUrl+"\n"+
			"Upgrade: websocket\nConnection: Upgrade\nSec-WebSocket-Version: 13\nSec-WebSocket-Key: [HIDDEN]\n"+
			headersString(header))

	ws, resp, err := c.dialer.Dial(url, header)
	if resp != nil {
		c.debugPrinter.Print("WEBSOCKET RESPONSE:",
			resp.Proto+" "+resp.Status+"\n"+
				headersString(resp.Header))
	}

	if resp != nil && resp.StatusCode == http.StatusUnauthorized {
		bodyData, _ := ioutil.ReadAll(resp.Body)
		err = noaa_errors.NewUnauthorizedError(string(bodyData))
		return ws, err
	}

	if err == nil && c.callback != nil {
		c.callback()
	}

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error dialing traffic controller server: %s.\nPlease ask your Cloud Foundry Operator to check the platform configuration (traffic controller is %s).", err.Error(), c.trafficControllerUrl))
	}

	return ws, err
}

func (c *Consumer) proxyDial(network, addr string) (net.Conn, error) {
	targetUrl, err := url.Parse("http://" + addr)
	if err != nil {
		return nil, err
	}

	proxy := c.proxy
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

func headersString(header http.Header) string {
	var result string
	for name, values := range header {
		result += name + ": " + strings.Join(values, ", ") + "\n"
	}
	return result
}
