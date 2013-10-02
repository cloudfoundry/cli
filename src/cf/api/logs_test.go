package api_test

import (
	"bytes"
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"code.google.com/p/go.net/websocket"
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	messagetesthelpers "github.com/cloudfoundry/loggregatorlib/logmessage/testhelpers"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testhelpers"
	"testing"
	"time"
)

var recentLogsEndpoint = func(message *logmessage.Message) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authMatches := request.Header.Get("authorization") == "BEARER my_access_token"
		methodMatches := request.Method == "GET"

		path := "/dump/?app=my-app-guid"

		paths := strings.Split(path, "?")
		pathMatches := request.URL.Path == paths[0]
		if len(paths) > 1 {
			queryStringMatches := strings.Contains(request.RequestURI, paths[1])
			pathMatches = pathMatches && queryStringMatches
		}

		if !(authMatches && methodMatches && pathMatches) {
			fmt.Printf("One of the matchers did not match. Auth [%t] Method [%t] Path [%t]", authMatches, methodMatches, pathMatches)

			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		var response *bytes.Buffer
		response = new(bytes.Buffer)
		logmessage.DumpMessage(*message, response)

		writer.Write(response.Bytes())
	}
}

func TestRecentLogsFor(t *testing.T) {

	// out of order messages we will send
	messagesSent := [][]byte{
		messagetesthelpers.MarshalledLogMessage(t, "My message", "my-app-id"),
	}

	websocketEndpoint := func(conn *websocket.Conn) {
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

	logRedirectEndpoint := func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, request.URL.Path, "/dump/")
		assert.Equal(t, request.URL.RawQuery, "app=my-app-guid")
		assert.Equal(t, request.Method, "GET")
		assert.Contains(t, request.Header.Get("Authorization"), "BEARER my_access_token")

		writer.Header().Set("Location", strings.Replace(websocketServer.URL, "https", "wss", 1))
		writer.WriteHeader(http.StatusFound)
	}

	http.HandleFunc("/dump/", logRedirectEndpoint)
	go http.ListenAndServe(":"+LOGGREGATOR_REDIRECTOR_PORT, nil)

	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "http://localhost"}
	loggregatorHostResolver := func(hostname string) string { return hostname }

	logsRepo := NewLoggregatorLogsRepository(config, gateway, loggregatorHostResolver)

	connected := false
	onConnect := func() {
		connected = true
	}

	// ordered messages we expect to receive
	dumpedMessages := []logmessage.LogMessage{}
	onMessage := func(message logmessage.LogMessage) {
		dumpedMessages = append(dumpedMessages, message)
	}

	// method under test
	err = logsRepo.RecentLogsFor(app, onConnect, onMessage)
	assert.NoError(t, err)

	assert.Equal(t, len(dumpedMessages), 1)
	assert.Equal(t, dumpedMessages[0].GetMessage(), expectedMessage.GetLogMessage().GetMessage())
	assert.Equal(t, dumpedMessages[0].GetMessageType(), expectedMessage.GetLogMessage().GetMessageType())
}

func TestTailsLogsFor(t *testing.T) {

	// out of order messages we will send
	messagesSent := [][]byte{
		marshalledLogMessageWithTime(t, "My message 3", int64(3000)),
		marshalledLogMessageWithTime(t, "My message 1", int64(1000)),
		marshalledLogMessageWithTime(t, "My message 2", int64(2000)),
	}

	websocketEndpoint := func(conn *websocket.Conn) {
		for _, msg := range messagesSent {
			conn.Write(msg)
		}
		time.Sleep(time.Duration(2) * time.Second)
		conn.Close()
	}
	websocketServer := httptest.NewTLSServer(websocket.Handler(websocketEndpoint))
	defer websocketServer.Close()

	var logRedirectEndpoint = func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, request.URL.Path, "/tail/")
		assert.Equal(t, request.URL.RawQuery, "app=my-app-guid")
		assert.Equal(t, request.Method, "GET")
		assert.Contains(t, request.Header.Get("Authorization"), "BEARER my_access_token")

		writer.Header().Set("Location", strings.Replace(websocketServer.URL, "https", "wss", 1))
		writer.WriteHeader(http.StatusFound)
	}
	http.HandleFunc("/tail/", logRedirectEndpoint)
	go http.ListenAndServe(":"+LOGGREGATOR_REDIRECTOR_PORT, nil)

	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "http://localhost"}
	loggregatorHostResolver := func(hostname string) string { return hostname }

	logsRepo := NewLoggregatorLogsRepository(config, gateway, loggregatorHostResolver)

	connected := false
	onConnect := func() {
		connected = true
	}

	// ordered messages we expect to receive
	tailedMessages := []logmessage.LogMessage{}
	onMessage := func(message logmessage.LogMessage) {
		tailedMessages = append(tailedMessages, message)
	}

	// method under test
	logsRepo.TailLogsFor(app, onConnect, onMessage, time.Duration(1))

	assert.True(t, connected)

	assert.Equal(t, len(tailedMessages), 3)

	actualMessage, err := proto.Marshal(&tailedMessages[0])
	assert.NoError(t, err)
	assert.Equal(t, actualMessage, messagesSent[1])

	actualMessage, err = proto.Marshal(&tailedMessages[1])
	assert.NoError(t, err)
	assert.Equal(t, actualMessage, messagesSent[2])

	actualMessage, err = proto.Marshal(&tailedMessages[2])
	assert.NoError(t, err)
	assert.Equal(t, actualMessage, messagesSent[0])

}

func TestLoggregatorHost(t *testing.T) {
	apiHost := "https://api.run.pivotal.io"
	loggregatorHost := LoggregatorHost(apiHost)

	assert.Equal(t, loggregatorHost, "https://loggregator.run.pivotal.io")
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
