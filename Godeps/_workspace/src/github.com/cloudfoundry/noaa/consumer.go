package noaa

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	noaa_errors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
)

var (
	// KeepAlive sets the interval between keep-alive messages sent by the client to loggregator.
	KeepAlive         = 25 * time.Second
	reconnectTimeout  = 500 * time.Millisecond
	boundaryRegexp    = regexp.MustCompile("boundary=(.*)")
	ErrNotOK          = errors.New("unknown issue when making HTTP request to Loggregator")
	ErrNotFound       = ErrNotOK // NotFound isn't an accurate description of how this is used; please use ErrNotOK instead
	ErrBadResponse    = errors.New("bad server response")
	ErrBadRequest     = errors.New("bad client request")
	ErrLostConnection = errors.New("remote server terminated connection unexpectedly")
)

// Consumer represents the actions that can be performed against traffic controller.
type Consumer struct {
	trafficControllerUrl string
	tlsConfig            *tls.Config
	ws                   *websocket.Conn
	callback             func()
	proxy                func(*http.Request) (*url.URL, error)
	debugPrinter         DebugPrinter
	sync.RWMutex
	stopChan chan struct{}
}

// NewConsumer creates a new consumer to a traffic controller.
func NewConsumer(trafficControllerUrl string, tlsConfig *tls.Config, proxy func(*http.Request) (*url.URL, error)) *Consumer {
	return &Consumer{trafficControllerUrl: trafficControllerUrl, tlsConfig: tlsConfig, proxy: proxy, debugPrinter: nullDebugPrinter{}, stopChan: make(chan struct{})}
}

// TailingLogs behaves exactly as TailingLogsWithoutReconnect, except that it retries 5 times if the connection
// to the remote server is lost and returns all errors from each attempt on errorChan.
func (cnsmr *Consumer) TailingLogs(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error) {
	action := func() error {
		return cnsmr.TailingLogsWithoutReconnect(appGuid, authToken, outputChan)
	}

	cnsmr.retryAction(action, errorChan)
}

// TailingLogsWithoutReconnect listens indefinitely for log messages only; other event types are dropped.
//
// If you wish to be able to terminate the listen early, run TailingLogsWithoutReconnect in a Goroutine and
// call Close() when you are finished listening.
//
// Messages are presented in the order received from the loggregator server. Chronological or
// other ordering is not guaranteed. It is the responsibility of the consumer of these channels
// to provide any desired sorting mechanism.
func (cnsmr *Consumer) TailingLogsWithoutReconnect(appGuid string, authToken string, outputChan chan<- *events.LogMessage) error {
	allEvents := make(chan *events.Envelope)

	streamPath := fmt.Sprintf("/apps/%s/stream", appGuid)
	errChan := make(chan error)
	go func() {
		err := cnsmr.stream(streamPath, authToken, allEvents)
		errChan <- err
		close(errChan)
	}()

	go func() {
		defer close(allEvents)

		for event := range allEvents {
			if *event.EventType == events.Envelope_LogMessage {
				outputChan <- event.GetLogMessage()
			}
		}
	}()

	return <-errChan
}

// Stream behaves exactly as StreamWithoutReconnect, except that it retries 5 times if the connection
// to the remote server is lost.
func (cnsmr *Consumer) Stream(appGuid string, authToken string, outputChan chan<- *events.Envelope, errorChan chan<- error) {
	action := func() error {
		return cnsmr.StreamWithoutReconnect(appGuid, authToken, outputChan)
	}

	cnsmr.retryAction(action, errorChan)
}

// StreamWithoutReconnect listens indefinitely for all log and event messages.
//
// If you wish to be able to terminate the listen early, run StreamWithoutReconnect in a Goroutine and
// call Close() when you are finished listening.
//
// Messages are presented in the order received from the loggregator server. Chronological or other ordering
// is not guaranteed. It is the responsibility of the consumer of these channels to provide any desired sorting
// mechanism.
func (cnsmr *Consumer) StreamWithoutReconnect(appGuid string, authToken string, outputChan chan<- *events.Envelope) error {
	streamPath := fmt.Sprintf("/apps/%s/stream", appGuid)
	return cnsmr.stream(streamPath, authToken, outputChan)
}

// Firehose behaves exactly as FirehoseWithoutReconnect, except that it retries 5 times if the connection
// to the remote server is lost.
func (cnsmr *Consumer) Firehose(subscriptionId string, authToken string, outputChan chan<- *events.Envelope, errorChan chan<- error) {
	action := func() error {
		return cnsmr.FirehoseWithoutReconnect(subscriptionId, authToken, outputChan)
	}

	cnsmr.retryAction(action, errorChan)
}

// FirehoseWithoutReconnect streams all data. All clients with the same subscriptionId will receive a proportionate share of the
// message stream. Each pool of clients will receive the entire stream.
//
// If you wish to be able to terminate the listen early, run FirehoseWithoutReconnect in a Goroutine and
// call Close() when you are finished listening.
//
// Messages are presented in the order received from the loggregator server. Chronological or other ordering
// is not guaranteed. It is the responsibility of the consumer of these channels to provide any desired sorting
// mechanism.
func (cnsmr *Consumer) FirehoseWithoutReconnect(subscriptionId string, authToken string, outputChan chan<- *events.Envelope) error {
	streamPath := "/firehose/" + subscriptionId
	return cnsmr.stream(streamPath, authToken, outputChan)
}

func (cnsmr *Consumer) stream(streamPath string, authToken string, outputChan chan<- *events.Envelope) error {
	var err error

	cnsmr.Lock()
	cnsmr.ws, err = cnsmr.establishWebsocketConnection(streamPath, authToken)
	cnsmr.Unlock()

	if err != nil {
		return err
	}

	return cnsmr.listenForMessages(outputChan)
}

func makeError(err error, code int32) *events.Envelope {
	return &events.Envelope{
		EventType: events.Envelope_Error.Enum(),
		Error: &events.Error{
			Source:  proto.String("NOAA"),
			Code:    &code,
			Message: proto.String(err.Error()),
		},
	}
}

// RecentLogs connects to traffic controller via its 'recentlogs' http(s) endpoint and returns a slice of recent messages.
// It does not guarantee any order of the messages; they are in the order returned by traffic controller.
//
// The SortRecent method is provided to sort the data returned by this method.
func (cnsmr *Consumer) RecentLogs(appGuid string, authToken string) ([]*events.LogMessage, error) {
	envelopes, err := cnsmr.readEnvelopesFromTrafficController(appGuid, authToken, "recentlogs")

	if err != nil {
		return nil, err
	}

	messages := make([]*events.LogMessage, 0, 200)
	for _, envelope := range envelopes {
		messages = append(messages, envelope.GetLogMessage())
	}

	return messages, err
}

// ContainerMetrics connects to traffic controller via its 'containermetrics' http(s) endpoint and returns the most recent messages for an app.
// The returned metrics will be sorted by InstanceIndex.
func (cnsmr *Consumer) ContainerMetrics(appGuid string, authToken string) ([]*events.ContainerMetric, error) {
	envelopes, err := cnsmr.readEnvelopesFromTrafficController(appGuid, authToken, "containermetrics")

	if err != nil {
		return nil, err
	}

	messages := make([]*events.ContainerMetric, 0, 200)

	for _, envelope := range envelopes {
		if envelope.GetEventType() == events.Envelope_LogMessage {
			return []*events.ContainerMetric{}, errors.New(fmt.Sprintf("Upstream error: %s", envelope.GetLogMessage().GetMessage()))
		}

		messages = append(messages, envelope.GetContainerMetric())
	}

	SortContainerMetrics(messages)

	return messages, err
}

func (cnsmr *Consumer) readEnvelopesFromTrafficController(appGuid string, authToken string, endpoint string) ([]*events.Envelope, error) {
	trafficControllerUrl, err := url.ParseRequestURI(cnsmr.trafficControllerUrl)
	if err != nil {
		return nil, err
	}

	scheme := "https"

	if trafficControllerUrl.Scheme == "ws" {
		scheme = "http"
	}

	recentPath := fmt.Sprintf("%s://%s/apps/%s/%s", scheme, trafficControllerUrl.Host, appGuid, endpoint)
	transport := &http.Transport{Proxy: cnsmr.proxy, TLSClientConfig: cnsmr.tlsConfig}
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequest("GET", recentPath, nil)
	req.Header.Set("Authorization", authToken)

	resp, err := client.Do(req)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error dialing traffic controller server: %s.\nPlease ask your Cloud Foundry Operator to check the platform configuration (traffic controller endpoint is %s).", err.Error(), cnsmr.trafficControllerUrl))
	}

	defer resp.Body.Close()

	err = checkForErrors(resp)
	if err != nil {
		return nil, err
	}

	reader, err := getMultipartReader(resp)
	if err != nil {
		return nil, err
	}

	var envelopes []*events.Envelope
	var buffer bytes.Buffer

	for part, loopErr := reader.NextPart(); loopErr == nil; part, loopErr = reader.NextPart() {
		buffer.Reset()

		_, err := buffer.ReadFrom(part)
		if err != nil {
			break
		}

		envelope := new(events.Envelope)
		proto.Unmarshal(buffer.Bytes(), envelope)

		envelopes = append(envelopes, envelope)
	}

	return envelopes, nil
}

func checkForErrors(resp *http.Response) error {
	if resp.StatusCode == http.StatusUnauthorized {
		data, _ := ioutil.ReadAll(resp.Body)
		return noaa_errors.NewUnauthorizedError(string(data))
	}

	if resp.StatusCode == http.StatusBadRequest {
		return ErrBadRequest
	}

	if resp.StatusCode != http.StatusOK {
		return ErrNotOK
	}
	return nil
}

func getMultipartReader(resp *http.Response) (*multipart.Reader, error) {
	contentType := resp.Header.Get("Content-Type")

	if len(strings.TrimSpace(contentType)) == 0 {
		return nil, ErrBadResponse
	}

	matches := boundaryRegexp.FindStringSubmatch(contentType)

	if len(matches) != 2 || len(strings.TrimSpace(matches[1])) == 0 {
		return nil, ErrBadResponse
	}
	reader := multipart.NewReader(resp.Body, matches[1])
	return reader, nil
}

// Close terminates the websocket connection to traffic controller.
func (cnsmr *Consumer) Close() error {
	cnsmr.Lock()
	defer cnsmr.Unlock()
	defer close(cnsmr.stopChan)
	if cnsmr.ws == nil {
		return errors.New("connection does not exist")
	}

	cnsmr.ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Time{})
	return cnsmr.ws.Close()
}

// SetOnConnectCallback sets a callback function to be called with the websocket connection is established.
func (cnsmr *Consumer) SetOnConnectCallback(cb func()) {
	cnsmr.callback = cb
}

// SetDebugPrinter enables logging of the websocket handshake.
func (cnsmr *Consumer) SetDebugPrinter(debugPrinter DebugPrinter) {
	cnsmr.debugPrinter = debugPrinter
}

func (cnsmr *Consumer) listenForMessages(msgChan chan<- *events.Envelope) error {
	defer cnsmr.ws.Close()

	for {
		_, data, err := cnsmr.ws.ReadMessage()
		if err != nil {
			return err
		}

		envelope := &events.Envelope{}
		err = proto.Unmarshal(data, envelope)
		if err != nil {
			continue
		}

		msgChan <- envelope
	}
}

func headersString(header http.Header) string {
	var result string
	for name, values := range header {
		result += name + ": " + strings.Join(values, ", ") + "\n"
	}
	return result
}

func (cnsmr *Consumer) establishWebsocketConnection(path string, authToken string) (*websocket.Conn, error) {
	header := http.Header{"Origin": []string{"http://localhost"}, "Authorization": []string{authToken}}

	dialer := websocket.Dialer{NetDial: cnsmr.proxyDial, TLSClientConfig: cnsmr.tlsConfig}

	url := cnsmr.trafficControllerUrl + path

	cnsmr.debugPrinter.Print("WEBSOCKET REQUEST:",
		"GET "+path+" HTTP/1.1\n"+
			"Host: "+cnsmr.trafficControllerUrl+"\n"+
			"Upgrade: websocket\nConnection: Upgrade\nSec-WebSocket-Version: 13\nSec-WebSocket-Key: [HIDDEN]\n"+
			headersString(header))

	ws, resp, err := dialer.Dial(url, header)

	if resp != nil {
		cnsmr.debugPrinter.Print("WEBSOCKET RESPONSE:",
			resp.Proto+" "+resp.Status+"\n"+
				headersString(resp.Header))
	}

	if resp != nil && resp.StatusCode == http.StatusUnauthorized {
		bodyData, _ := ioutil.ReadAll(resp.Body)
		err = noaa_errors.NewUnauthorizedError(string(bodyData))
		return ws, err
	}

	if err == nil && cnsmr.callback != nil {
		cnsmr.callback()
	}

	if err != nil {

		return nil, errors.New(fmt.Sprintf("Error dialing traffic controller server: %s.\nPlease ask your Cloud Foundry Operator to check the platform configuration (traffic controller is %s).", err.Error(), cnsmr.trafficControllerUrl))
	}

	return ws, err
}

func (cnsmr *Consumer) proxyDial(network, addr string) (net.Conn, error) {
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

func (cnsmr *Consumer) retryAction(action func() error, errorChan chan<- error) {
	reconnectAttempts := 0

	oldConnectCallback := cnsmr.callback
	defer func() { cnsmr.callback = oldConnectCallback }()

	defer close(errorChan)

	cnsmr.callback = func() {
		reconnectAttempts = 0
		if oldConnectCallback != nil {
			oldConnectCallback()
		}
	}

	for ; reconnectAttempts < 5; reconnectAttempts++ {
		select {
		case <-cnsmr.stopChan:
			return
		default:
		}

		errorChan <- action()
		time.Sleep(reconnectTimeout)
	}
}
