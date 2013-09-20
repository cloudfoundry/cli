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
	expectedMessage := messagetesthelpers.MarshalledLogMessage(t, "My message", "my-app-id")
	message, err := logmessage.ParseMessage(expectedMessage)
	assert.NoError(t, err)

	ts := httptest.NewTLSServer(http.HandlerFunc(recentLogsEndpoint(message)))
	defer ts.Close()

	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	loggregatorHostResolver := func(hostname string) string { return hostname }
	logsRepo := NewLoggregatorLogsRepository(config, gateway, loggregatorHostResolver)

	logs, err := logsRepo.RecentLogsFor(app)
	assert.NoError(t, err)

	assert.Equal(t, len(logs), 1)
	actualMessage, err := proto.Marshal(logs[0])
	assert.NoError(t, err)
	assert.Equal(t, actualMessage, expectedMessage)
}

func TestTailsLogsFor(t *testing.T) {
	expectedMessages := [][]byte{
		marshalledLogMessageWithTime(t, "My message 3", int64(3000)),
		marshalledLogMessageWithTime(t, "My message 1", int64(1000)),
		marshalledLogMessageWithTime(t, "My message 2", int64(2000)),
	}

	websocketEndpoint := func(conn *websocket.Conn) {
		for _, msg := range expectedMessages {
			conn.Write(msg)
		}
		conn.Close()
	}

	websocketServer := httptest.NewTLSServer(websocket.Handler(websocketEndpoint))
	defer websocketServer.Close()

	var redirectEndpoint = func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, request.URL.Path, "/tail/")
		assert.Equal(t, request.URL.RawQuery, "app=my-app-guid")
		assert.Equal(t, request.Method, "GET")
		assert.Contains(t, request.Header.Get("Authorization"), "BEARER my_access_token")

		writer.Header().Set("Location", strings.Replace(websocketServer.URL, "https", "wss", 1))
		writer.WriteHeader(http.StatusFound)
	}

	http.HandleFunc("/", redirectEndpoint)

	go http.ListenAndServe(":"+LOGGREGATOR_REDIRECTOR_PORT, nil)

	redirectServer := httptest.NewTLSServer(http.HandlerFunc(redirectEndpoint))
	defer redirectServer.Close()

	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "http://localhost"}

	loggregatorHostResolver := func(hostname string) string { return hostname }
	logsRepo := NewLoggregatorLogsRepository(config, gateway, loggregatorHostResolver)

	connected := false

	onConnect := func() {
		connected = true
	}

	tailedMessages := []logmessage.LogMessage{}

	onMessage := func(message logmessage.LogMessage) {
		tailedMessages = append(tailedMessages, message)
	}

	logsRepo.TailLogsFor(app, onConnect, onMessage)

	assert.Equal(t, len(tailedMessages), 1)

	actualMessage, err := proto.Marshal(&tailedMessages[0])
	assert.NoError(t, err)
	assert.Equal(t, actualMessage, expectedMessages[1])

	actualMessage, err = proto.Marshal(&tailedMessages[1])
	assert.NoError(t, err)
	assert.Equal(t, actualMessage, expectedMessages[2])

	actualMessage, err = proto.Marshal(&tailedMessages[2])
	assert.NoError(t, err)
	assert.Equal(t, actualMessage, expectedMessages[0])

	assert.True(t, connected)
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
