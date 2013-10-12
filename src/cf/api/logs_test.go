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

	// ordered messages we expect to receive
	dumpedMessages := []*logmessage.Message{}
	onMessage := func(message *logmessage.Message) {
		dumpedMessages = append(dumpedMessages, message)
	}

	// method under test
	err = logsRepo.RecentLogsFor(app, onConnect, onMessage)
	assert.NoError(t, err)

	assert.Equal(t, len(dumpedMessages), 1)
	assert.Equal(t, dumpedMessages[0].GetShortSourceTypeName(), expectedMessage.GetShortSourceTypeName())
	assert.Equal(t, dumpedMessages[0].GetLogMessage().GetMessage(), expectedMessage.GetLogMessage().GetMessage())
	assert.Equal(t, dumpedMessages[0].GetLogMessage().GetMessageType(), expectedMessage.GetLogMessage().GetMessageType())
}

func TestTailsLogsFor(t *testing.T) {
	// out of order messages we will send
	messagesSent := [][]byte{
		marshalledLogMessageWithTime(t, "My message 3", int64(3000)),
		marshalledLogMessageWithTime(t, "My message 1", int64(1000)),
		marshalledLogMessageWithTime(t, "My message 2", int64(2000)),
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
		time.Sleep(time.Duration(2) * time.Second)
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
	onMessage := func(message *logmessage.Message) {
		tailedMessages = append(tailedMessages, message)
	}

	// method under test
	logsRepo.TailLogsFor(app, onConnect, onMessage, time.Duration(1))

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
