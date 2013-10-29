package api

import (
	"cf"
	"cf/configuration"
	"code.google.com/p/go.net/websocket"
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	testapi "testhelpers/api"
	"testing"
	"time"
)

func TestRecentLogsFor(t *testing.T) {
	// out of order messages we will send
	messagesSent := [][]byte{
		marshalledLogMessageWithTime(t, "My message", int64(3000)),
	}

	websocketEndpoint := func(conn *websocket.Conn) {
		request := conn.Request()
		assert.Equal(t, request.URL.Path, "/dump/")
		assert.Equal(t, request.URL.RawQuery, "app=my-app-guid")
		assert.Equal(t, request.Method, "GET")
		assert.Contains(t, request.Header.Get("Authorization"), "BEARER my_access_token")

		for _, msg := range messagesSent {
			conn.Write(msg)
		}
		time.Sleep(time.Duration(2) * time.Second)
		conn.Close()
	}
	websocketServer := httptest.NewTLSServer(websocket.Handler(websocketEndpoint))
	defer websocketServer.Close()

	expectedMessage, err := logmessage.ParseMessage(messagesSent[0])
	assert.NoError(t, err)

	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "https://localhost"}

	endpointRepo := &testapi.FakeEndpointRepo{GetEndpointEndpoints: map[cf.EndpointType]string{
		cf.LoggregatorEndpointKey: strings.Replace(websocketServer.URL, "https", "wss", 1),
	}}

	logsRepo := NewLoggregatorLogsRepository(config, endpointRepo)

	connected := false
	onConnect := func() {
		connected = true
	}

	logChan := make(chan *logmessage.Message, 1000)
	err = logsRepo.RecentLogsFor(app, onConnect, logChan)

	// ordered messages we expect to receive
	dumpedMessages := []*logmessage.Message{}
	for msg := range logChan {
		dumpedMessages = append(dumpedMessages, msg)
	}

	assert.NoError(t, err)

	assert.Equal(t, len(dumpedMessages), 1)
	assert.Equal(t, dumpedMessages[0].GetShortSourceTypeName(), expectedMessage.GetShortSourceTypeName())
	assert.Equal(t, dumpedMessages[0].GetLogMessage().GetMessage(), expectedMessage.GetLogMessage().GetMessage())
	assert.Equal(t, dumpedMessages[0].GetLogMessage().GetMessageType(), expectedMessage.GetLogMessage().GetMessageType())
}

func TestTailsLogsFor(t *testing.T) {
	// out of order messages we will send
	messagesSent := [][]byte{
		marshalledLogMessageWithTime(t, "My message 3", int64(300000)),
		marshalledLogMessageWithTime(t, "My message 1", int64(100000)),
		marshalledLogMessageWithTime(t, "My message 2", int64(200000)),
	}

	websocketEndpoint := func(conn *websocket.Conn) {
		request := conn.Request()
		assert.Equal(t, request.URL.Path, "/tail/")
		assert.Equal(t, request.URL.RawQuery, "app=my-app-guid")
		assert.Equal(t, request.Method, "GET")
		assert.Contains(t, request.Header.Get("Authorization"), "BEARER my_access_token")

		for _, msg := range messagesSent {
			conn.Write(msg)
		}
		time.Sleep(time.Duration(200) * time.Millisecond)
		conn.Close()
	}
	websocketServer := httptest.NewTLSServer(websocket.Handler(websocketEndpoint))
	defer websocketServer.Close()

	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "https://localhost"}
	endpointRepo := &testapi.FakeEndpointRepo{GetEndpointEndpoints: map[cf.EndpointType]string{
		cf.LoggregatorEndpointKey: strings.Replace(websocketServer.URL, "https", "wss", 1),
	}}

	logsRepo := NewLoggregatorLogsRepository(config, endpointRepo)

	connected := false
	onConnect := func() {
		connected = true
	}

	// ordered messages we expect to receive
	tailedMessages := []*logmessage.Message{}

	logChan := make(chan *logmessage.Message, 1000)

	controlChan := make(chan bool)

	// method under test
	logsRepo.TailLogsFor(app, onConnect, logChan, controlChan, time.Duration(1))

	for msg := range logChan {
		tailedMessages = append(tailedMessages, msg)
	}

	assert.True(t, connected)

	assert.Equal(t, len(tailedMessages), 3)

	tailedMessage := tailedMessages[0]
	actualMessage, err := proto.Marshal(tailedMessage.GetLogMessage())
	assert.NoError(t, err)
	assert.Equal(t, actualMessage, messagesSent[1])

	tailedMessage = tailedMessages[1]
	actualMessage, err = proto.Marshal(tailedMessage.GetLogMessage())
	assert.NoError(t, err)
	assert.Equal(t, actualMessage, messagesSent[2])

	tailedMessage = tailedMessages[2]
	actualMessage, err = proto.Marshal(tailedMessage.GetLogMessage())
	assert.NoError(t, err)
	assert.Equal(t, actualMessage, messagesSent[0])
}

func TestMessageOutputTimesDuringNormalFlow(t *testing.T) {
	// out of order messages we will send
	startTime := time.Now()
	messagesSent := [][]byte{
		marshalledLogMessageWithTime(t, "My message 1", startTime.Add(-9*time.Second).UnixNano()), //really really late message
		marshalledLogMessageWithTime(t, "My message 2", startTime.Add(-2*time.Second).UnixNano()),
		marshalledLogMessageWithTime(t, "My message 3", startTime.Add(-1*time.Second).UnixNano()),
	}

	websocketEndpoint := func(conn *websocket.Conn) {
		request := conn.Request()
		assert.Equal(t, request.URL.Path, "/tail/")
		assert.Equal(t, request.URL.RawQuery, "app=my-app-guid")
		assert.Equal(t, request.Method, "GET")
		assert.Contains(t, request.Header.Get("Authorization"), "BEARER my_access_token")

		for _, msg := range messagesSent {
			conn.Write(msg)
			time.Sleep(200 * time.Millisecond)
		}
		time.Sleep(1 * time.Second)
		conn.Close()
	}
	websocketServer := httptest.NewTLSServer(websocket.Handler(websocketEndpoint))
	defer websocketServer.Close()

	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "https://localhost"}
	endpointRepo := &testapi.FakeEndpointRepo{GetEndpointEndpoints: map[cf.EndpointType]string{
		cf.LoggregatorEndpointKey: strings.Replace(websocketServer.URL, "https", "wss", 1),
	}}

	logsRepo := NewLoggregatorLogsRepository(config, endpointRepo)

	logChan := make(chan *logmessage.Message, 1000)
	controlChan := make(chan bool)

	logsRepo.TailLogsFor(app, func() {}, logChan, controlChan, time.Duration(1*time.Second))

	for msg := range logChan {
		//assertions about the arrival times of the messages
		timeWhenOutputtable := startTime.Add(1 * time.Second).UnixNano()
		timeNow := time.Now().UnixNano()

		switch string(msg.GetLogMessage().Message) {
		case "My message 1":
			assert.True(t, (timeNow-timeWhenOutputtable) < (50*time.Millisecond).Nanoseconds())
			assert.True(t, (timeNow-timeWhenOutputtable) > (10*time.Millisecond).Nanoseconds())
		case "My message 2":
			assert.True(t, (timeNow-timeWhenOutputtable) < (250*time.Millisecond).Nanoseconds())
			assert.True(t, (timeNow-timeWhenOutputtable) > (200*time.Millisecond).Nanoseconds())
		case "My message 3":
			assert.True(t, (timeNow-timeWhenOutputtable) < (450*time.Millisecond).Nanoseconds())
			assert.True(t, (timeNow-timeWhenOutputtable) > (400*time.Millisecond).Nanoseconds())
		}
	}
}

func TestMessageOutputWhenFlushingAfterServerDeath(t *testing.T) {
	// out of order messages we will send
	startTime := time.Now()
	messagesSent := [][]byte{
		marshalledLogMessageWithTime(t, "My message 1", startTime.Add(-9*time.Second).UnixNano()), //really really late message
		marshalledLogMessageWithTime(t, "My message 2", startTime.Add(-2*time.Second).UnixNano()),
		marshalledLogMessageWithTime(t, "My message 3", startTime.Add(-1*time.Second).UnixNano()),
	}

	websocketEndpoint := func(conn *websocket.Conn) {
		request := conn.Request()
		assert.Equal(t, request.URL.Path, "/tail/")
		assert.Equal(t, request.URL.RawQuery, "app=my-app-guid")
		assert.Equal(t, request.Method, "GET")
		assert.Contains(t, request.Header.Get("Authorization"), "BEARER my_access_token")

		for _, msg := range messagesSent {
			conn.Write(msg)
			time.Sleep(200 * time.Millisecond)
		}
		conn.Close()
	}
	websocketServer := httptest.NewTLSServer(websocket.Handler(websocketEndpoint))
	defer websocketServer.Close()

	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "https://localhost"}
	endpointRepo := &testapi.FakeEndpointRepo{GetEndpointEndpoints: map[cf.EndpointType]string{
		cf.LoggregatorEndpointKey: strings.Replace(websocketServer.URL, "https", "wss", 1),
	}}

	logsRepo := NewLoggregatorLogsRepository(config, endpointRepo)

	firstMessageTime := time.Now().Add(-10 * time.Second).UnixNano()

	logChan := make(chan *logmessage.Message, 1000)
	controlChan := make(chan bool)
	logsRepo.TailLogsFor(app, func() {}, logChan, controlChan, time.Duration(1*time.Second))

	for msg := range logChan {
		switch string(msg.GetLogMessage().Message) {
		case "My message 1":
			firstMessageTime = time.Now().UnixNano()
		case "My message 2":
			timeNow := time.Now().UnixNano()
			delta := timeNow - firstMessageTime
			assert.True(t, delta < (5*time.Millisecond).Nanoseconds())
			assert.True(t, delta >= 0)
		case "My message 3":
			timeNow := time.Now().UnixNano()
			delta := timeNow - firstMessageTime
			assert.True(t, delta < (5*time.Millisecond).Nanoseconds())
			assert.True(t, delta >= 0)
		}
	}
}

func marshalledLogMessageWithTime(t *testing.T, messageString string, timestamp int64) []byte {
	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_DEA
	protoMessage := &logmessage.LogMessage{
		Message:     []byte(messageString),
		AppId:       proto.String("my-app-guid"),
		MessageType: &messageType,
		SourceType:  &sourceType,
		Timestamp:   proto.Int64(timestamp),
	}

	message, err := proto.Marshal(protoMessage)
	assert.NoError(t, err)

	return message
}
