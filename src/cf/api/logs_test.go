package api_test

import (
	. "cf/api"
	"code.google.com/p/go.net/websocket"
	"code.google.com/p/gogoprotobuf/proto"
	"crypto/tls"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http/httptest"
	"strings"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
	"time"
)

var _ = Describe("loggregator logs repository", func() {
	var (
		logChan        chan *logmessage.Message
		testServer     *httptest.Server
		requestHandler *requestHandlerWithExpectedPath
		logsRepo       *LoggregatorLogsRepository
		messagesToSend [][]byte
	)

	BeforeEach(func() {
		startTime := time.Now()
		messagesToSend = [][]byte{
			marshalledLogMessageWithTime("My message 1", startTime.UnixNano()),
			marshalledLogMessageWithTime("My message 2", startTime.UnixNano()),
			marshalledLogMessageWithTime("My message 3", startTime.UnixNano()),
		}
		logChan = make(chan *logmessage.Message, 1000)
		testServer, requestHandler, logsRepo = setupTestServerAndLogsRepo(messagesToSend...)

		logsRepo.AddTrustedCerts(testServer.TLS.Certificates)
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("RecentLogsFor", func() {
		BeforeEach(func() {
			err := logsRepo.RecentLogsFor("my-app-guid", func() {}, logChan)
			Expect(err).NotTo(HaveOccurred())
			close(logChan)
		})

		It("connects to the dump endpoint", func() {
			Expect(requestHandler.lastPath).To(Equal("/dump/"))
		})

		It("writes log messages onto the provided channel", func() {
			dumpedMessages := []*logmessage.Message{}
			for msg := range logChan {
				dumpedMessages = append(dumpedMessages, msg)
			}

			Expect(len(dumpedMessages)).To(Equal(3))
			Expect(dumpedMessages[0]).To(Equal(parseMessage(messagesToSend[0])))
			Expect(dumpedMessages[1]).To(Equal(parseMessage(messagesToSend[1])))
			Expect(dumpedMessages[2]).To(Equal(parseMessage(messagesToSend[2])))
		})
	})

	Describe("TailLogsFor", func() {
		Context("when the SSL certificate is valid", func() {
			It("connects to the tailing endpoint", func() {
				err := logsRepo.TailLogsFor("my-app-guid", func() {}, logChan, make(chan bool), time.Duration(1*time.Second))
				Expect(err).NotTo(HaveOccurred())
				close(logChan)

				Expect(requestHandler.lastPath).To(Equal("/tail/"))
			})

			It("writes log messages on the channel in the correct order", func() {
				err := logsRepo.TailLogsFor("my-app-guid", func() {}, logChan, make(chan bool), time.Duration(1*time.Second))
				Expect(err).NotTo(HaveOccurred())
				close(logChan)

				var messages []string
				for msg := range logChan {
					messages = append(messages, string(msg.GetLogMessage().Message))
				}

				Expect(messages).To(Equal([]string{"My message 1", "My message 2", "My message 3"}))
			})
		})

		Context("when the SSL certificate is invalid", func() {
			BeforeEach(func() {
				testServer.TLS.Certificates = []tls.Certificate{testnet.MakeExpiredTLSCert()}
			})

			Context("when skip-validation-errors flag is set", func() {
				BeforeEach(func() {
					//					logsRepo.DisableSSL()
				})

				XIt("ignores SSL validation errors", func() {
					err := logsRepo.TailLogsFor("my-app-guid", func() {}, logChan, make(chan bool), time.Duration(1*time.Second))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when skip-validation-errors is not set", func() {
				It("fails when the server's SSL cert cannot be verified", func() {
					err := logsRepo.TailLogsFor("my-app-guid", func() {}, logChan, make(chan bool), time.Duration(1*time.Second))
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})

func parseMessage(msgBytes []byte) (msg *logmessage.Message) {
	msg, err := logmessage.ParseMessage(msgBytes)
	Expect(err).ToNot(HaveOccurred())
	return
}

type requestHandlerWithExpectedPath struct {
	handlerFunc func(conn *websocket.Conn)
	lastPath    string
}

func setupTestServerAndLogsRepo(messages ...[]byte) (testServer *httptest.Server, requestHandler *requestHandlerWithExpectedPath, logsRepo *LoggregatorLogsRepository) {
	requestHandler = new(requestHandlerWithExpectedPath)
	requestHandler.handlerFunc = func(conn *websocket.Conn) {
		request := conn.Request()
		requestHandler.lastPath = request.URL.Path
		Expect(request.URL.RawQuery).To(Equal("app=my-app-guid"))
		Expect(request.Method).To(Equal("GET"))
		Expect(request.Header.Get("Authorization")).To(ContainSubstring("BEARER my_access_token"))

		for _, msg := range messages {
			conn.Write(msg)
		}
		time.Sleep(time.Duration(50) * time.Millisecond)
		conn.Close()
	}

	testServer = httptest.NewTLSServer(websocket.Handler(requestHandler.handlerFunc))

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint("https://localhost")
	configRepo.SetLoggregatorEndpoint(strings.Replace(testServer.URL, "https", "wss", 1))

	repo := NewLoggregatorLogsRepository(configRepo)
	logsRepo = &repo
	return
}

func marshalledLogMessageWithTime(messageString string, timestamp int64) []byte {
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
	Expect(err).ToNot(HaveOccurred())

	return message
}
