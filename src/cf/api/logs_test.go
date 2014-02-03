package api_test

import (
	. "cf/api"
	"cf/configuration"
	"code.google.com/p/go.net/websocket"
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	testapi "testhelpers/api"

	. "github.com/onsi/ginkgo"
	mr "github.com/tjarratt/mr_t"
	"time"
)

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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestRecentLogsFor", func() {
			messagesSent := [][]byte{
				marshalledLogMessageWithTime(mr.T(), "My message", int64(3000)),
			}

			websocketEndpoint := func(conn *websocket.Conn) {
				request := conn.Request()
				assert.Equal(mr.T(), request.URL.Path, "/dump/")
				assert.Equal(mr.T(), request.URL.RawQuery, "app=my-app-guid")
				assert.Equal(mr.T(), request.Method, "GET")
				assert.Contains(mr.T(), request.Header.Get("Authorization"), "BEARER my_access_token")

				for _, msg := range messagesSent {
					conn.Write(msg)
				}
				time.Sleep(time.Duration(20) * time.Millisecond)
				conn.Close()
			}
			websocketServer := httptest.NewTLSServer(websocket.Handler(websocketEndpoint))
			defer websocketServer.Close()

			expectedMessage, err := logmessage.ParseMessage(messagesSent[0])
			assert.NoError(mr.T(), err)

			config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "https://localhost"}

			endpointRepo := &testapi.FakeEndpointRepo{}
			endpointRepo.LoggregatorEndpointReturns.Endpoint = strings.Replace(websocketServer.URL, "https", "wss", 1)

			logsRepo := NewLoggregatorLogsRepository(config, endpointRepo)

			connected := false
			onConnect := func() {
				connected = true
			}

			logChan := make(chan *logmessage.Message, 1000)

			err = logsRepo.RecentLogsFor("my-app-guid", onConnect, logChan)
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
		It("TestRecentLogsSendsAllMessages", func() {

			websocketEndpoint := func(conn *websocket.Conn) {
				conn.Write(marshalledLogMessageWithTime(mr.T(), "My message", int64(3000)))
				time.Sleep(time.Duration(1) * time.Millisecond)
				conn.Close()
			}
			websocketServer := httptest.NewTLSServer(websocket.Handler(websocketEndpoint))
			defer websocketServer.Close()

			config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "https://localhost"}
			endpointRepo := &testapi.FakeEndpointRepo{}
			endpointRepo.LoggregatorEndpointReturns.Endpoint = strings.Replace(websocketServer.URL, "https", "wss", 1)

			logsRepo := NewLoggregatorLogsRepository(config, endpointRepo)
			logChan := make(chan *logmessage.Message, 1000)

			logsRepo.RecentLogsFor("my-app-guid", func() {}, logChan)
			assert.Equal(mr.T(), len(logChan), 1)
		})
		It("TestTailsLogsFor", func() {

			messagesSent := [][]byte{
				marshalledLogMessageWithTime(mr.T(), "My message 3", int64(300000)),
				marshalledLogMessageWithTime(mr.T(), "My message 1", int64(100000)),
				marshalledLogMessageWithTime(mr.T(), "My message 2", int64(200000)),
			}

			websocketEndpoint := func(conn *websocket.Conn) {
				request := conn.Request()
				assert.Equal(mr.T(), request.URL.Path, "/tail/")
				assert.Equal(mr.T(), request.URL.RawQuery, "app=my-app-guid")
				assert.Equal(mr.T(), request.Method, "GET")
				assert.Contains(mr.T(), request.Header.Get("Authorization"), "BEARER my_access_token")

				for _, msg := range messagesSent {
					conn.Write(msg)
				}
				time.Sleep(time.Duration(1) * time.Millisecond)
				conn.Close()
			}
			websocketServer := httptest.NewTLSServer(websocket.Handler(websocketEndpoint))
			defer websocketServer.Close()

			config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "https://localhost"}
			endpointRepo := &testapi.FakeEndpointRepo{}
			endpointRepo.LoggregatorEndpointReturns.Endpoint = strings.Replace(websocketServer.URL, "https", "wss", 1)

			logsRepo := NewLoggregatorLogsRepository(config, endpointRepo)

			connected := false
			onConnect := func() {
				connected = true
			}

			tailedMessages := []*logmessage.Message{}

			logChan := make(chan *logmessage.Message, 1000)

			controlChan := make(chan bool)

			logsRepo.TailLogsFor("my-app-guid", onConnect, logChan, controlChan, time.Duration(1))
			close(logChan)

			for msg := range logChan {
				tailedMessages = append(tailedMessages, msg)
			}

			assert.True(mr.T(), connected)

			assert.Equal(mr.T(), len(tailedMessages), 3)

			tailedMessage := tailedMessages[0]
			actualMessage, err := proto.Marshal(tailedMessage.GetLogMessage())
			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), actualMessage, messagesSent[1])

			tailedMessage = tailedMessages[1]
			actualMessage, err = proto.Marshal(tailedMessage.GetLogMessage())
			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), actualMessage, messagesSent[2])

			tailedMessage = tailedMessages[2]
			actualMessage, err = proto.Marshal(tailedMessage.GetLogMessage())
			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), actualMessage, messagesSent[0])
		})
		It("TestMessageOutputOrder", func() {

			startTime := time.Now()
			messagesSent := [][]byte{
				marshalledLogMessageWithTime(mr.T(), "My message 1", startTime.UnixNano()),
				marshalledLogMessageWithTime(mr.T(), "My message 2", startTime.UnixNano()),
				marshalledLogMessageWithTime(mr.T(), "My message 3", startTime.UnixNano()),
			}

			websocketEndpoint := func(conn *websocket.Conn) {
				request := conn.Request()
				assert.Equal(mr.T(), request.URL.Path, "/tail/")
				assert.Equal(mr.T(), request.URL.RawQuery, "app=my-app-guid")
				assert.Equal(mr.T(), request.Method, "GET")
				assert.Contains(mr.T(), request.Header.Get("Authorization"), "BEARER my_access_token")

				for _, msg := range messagesSent {
					conn.Write(msg)
				}
				time.Sleep(time.Duration(1) * time.Millisecond)
				conn.Close()
			}
			websocketServer := httptest.NewTLSServer(websocket.Handler(websocketEndpoint))
			defer websocketServer.Close()

			config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "https://localhost"}
			endpointRepo := &testapi.FakeEndpointRepo{}
			endpointRepo.LoggregatorEndpointReturns.Endpoint = strings.Replace(websocketServer.URL, "https", "wss", 1)

			logsRepo := NewLoggregatorLogsRepository(config, endpointRepo)

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

			assert.Equal(mr.T(), messages, []string{"My message 1", "My message 2", "My message 3"})
		})
		It("TestMessageOutputWhenFlushingAfterServerDeath", func() {

			startTime := time.Now()
			messagesSent := [][]byte{
				marshalledLogMessageWithTime(mr.T(), "My message 1", startTime.UnixNano()),
				marshalledLogMessageWithTime(mr.T(), "My message 2", startTime.UnixNano()),
				marshalledLogMessageWithTime(mr.T(), "My message 3", startTime.UnixNano()),
			}

			websocketEndpoint := func(conn *websocket.Conn) {
				request := conn.Request()
				assert.Equal(mr.T(), request.URL.Path, "/tail/")
				assert.Equal(mr.T(), request.URL.RawQuery, "app=my-app-guid")
				assert.Equal(mr.T(), request.Method, "GET")
				assert.Contains(mr.T(), request.Header.Get("Authorization"), "BEARER my_access_token")

				for _, msg := range messagesSent {
					conn.Write(msg)
				}
				conn.Close()
			}
			websocketServer := httptest.NewTLSServer(websocket.Handler(websocketEndpoint))
			defer websocketServer.Close()

			config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: "https://localhost"}
			endpointRepo := &testapi.FakeEndpointRepo{}
			endpointRepo.LoggregatorEndpointReturns.Endpoint = strings.Replace(websocketServer.URL, "https", "wss", 1)

			logsRepo := NewLoggregatorLogsRepository(config, endpointRepo)

			firstMessageTime := time.Now().Add(-10 * time.Second).UnixNano()

			logChan := make(chan *logmessage.Message, 1000)
			controlChan := make(chan bool)

			go func() {
				defer close(logChan)
				logsRepo.TailLogsFor("my-app-guid", func() {}, logChan, controlChan, time.Duration(1*time.Second))
			}()

			for msg := range logChan {
				switch string(msg.GetLogMessage().Message) {
				case "My message 1":
					firstMessageTime = time.Now().UnixNano()
				case "My message 2":
					timeNow := time.Now().UnixNano()
					delta := timeNow - firstMessageTime
					assert.True(mr.T(), delta < (5*time.Millisecond).Nanoseconds())
					assert.True(mr.T(), delta >= 0)
				case "My message 3":
					timeNow := time.Now().UnixNano()
					delta := timeNow - firstMessageTime
					assert.True(mr.T(), delta < (5*time.Millisecond).Nanoseconds())
					assert.True(mr.T(), delta >= 0)
				}
			}
		})
	})
}
