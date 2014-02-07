package api_test

import (
	. "cf/api"
	"cf/configuration"
	"code.google.com/p/go.net/websocket"
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http/httptest"
	"strings"
	testapi "testhelpers/api"
	"time"
)

func init() {
	Describe("loggregator logs repository", func() {
		It("RecentLogsFor writes log messages onto the provided channel", func() {
			msg := marshalledLogMessageWithTime(mr.T(), "My message", int64(3000))
			expectedMessage, err := logmessage.ParseMessage(msg)
			assert.NoError(mr.T(), err)

			testServer, logsRepo := setupTestServerAndLogsRepo("/dump/", msg)
			defer testServer.Close()

			logChan := make(chan *logmessage.Message, 1000)
			err = logsRepo.RecentLogsFor("my-app-guid", func() {}, logChan)
			close(logChan)

			dumpedMessages := []*logmessage.Message{}
			for msg := range logChan {
				dumpedMessages = append(dumpedMessages, msg)
			}

			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), len(dumpedMessages), 1)
			assert.Equal(mr.T(), dumpedMessages[0].GetLogMessage().GetSourceName(), expectedMessage.GetLogMessage().GetSourceName())
			assert.Equal(mr.T(), dumpedMessages[0].GetLogMessage().GetMessage(), expectedMessage.GetLogMessage().GetMessage())
			assert.Equal(mr.T(), dumpedMessages[0].GetLogMessage().GetMessageType(), expectedMessage.GetLogMessage().GetMessageType())
		})

		It("TailLogsFor writes log messages on the channel in the correct order", func() {
			startTime := time.Now()
			messagesSent := [][]byte{
				marshalledLogMessageWithTime(mr.T(), "My message 1", startTime.UnixNano()),
				marshalledLogMessageWithTime(mr.T(), "My message 2", startTime.UnixNano()),
				marshalledLogMessageWithTime(mr.T(), "My message 3", startTime.UnixNano()),
			}

			testServer, logsRepo := setupTestServerAndLogsRepo("/tail/", messagesSent...)
			defer testServer.Close()

			logChan := make(chan *logmessage.Message, 1000)
			controlChan := make(chan bool)

			go func() {
				defer close(logChan)
				logsRepo.TailLogsFor("my-app-guid", func() {}, logChan, controlChan, time.Duration(1*time.Second))
			}()

			var messages []string
			for msg := range logChan {
				messages = append(messages, string(msg.GetLogMessage().Message))
			}

			assert.Equal(mr.T(), len(messages), 3)
			assert.Equal(mr.T(), messages, []string{"My message 1", "My message 2", "My message 3"})
		})
	})
}

func setupTestServerAndLogsRepo(requestURI string, messages ...[]byte) (testServer *httptest.Server, logsRepo *LoggregatorLogsRepository) {
	handler := func(conn *websocket.Conn) {
		request := conn.Request()
		assert.Equal(mr.T(), request.URL.Path, requestURI)
		assert.Equal(mr.T(), request.URL.RawQuery, "app=my-app-guid")
		assert.Equal(mr.T(), request.Method, "GET")
		assert.Contains(mr.T(), request.Header.Get("Authorization"), "BEARER my_access_token")

		for _, msg := range messages {
			conn.Write(msg)
		}
		time.Sleep(time.Duration(50) * time.Millisecond)
		conn.Close()
	}
	testServer = httptest.NewTLSServer(websocket.Handler(handler))

	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "https://localhost"}
	endpointRepo := &testapi.FakeEndpointRepo{}
	endpointRepo.LoggregatorEndpointReturns.Endpoint = strings.Replace(testServer.URL, "https", "wss", 1)

	repo := NewLoggregatorLogsRepository(config, endpointRepo)
	logsRepo = &repo
	return
}

func marshalledLogMessageWithTime(t mr.TestingT, messageString string, timestamp int64) []byte {
	messageType := logmessage.LogMessage_OUT
	sourceName := "DEA"
	protoMessage := &logmessage.LogMessage{
		Message:     []byte(messageString),
		AppId:       proto.String("my-app-guid"),
		MessageType: &messageType,
		SourceName:  &sourceName,
		Timestamp:   proto.Int64(timestamp),
	}

	message, err := proto.Marshal(protoMessage)
	assert.NoError(t, err)

	return message
}
