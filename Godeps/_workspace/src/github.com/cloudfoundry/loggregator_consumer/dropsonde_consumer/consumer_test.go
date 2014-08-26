package dropsonde_consumer_test

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"code.google.com/p/gogoprotobuf/proto"
	"crypto/tls"
	"fmt"
	"github.com/cloudfoundry/dropsonde/events"
	"github.com/cloudfoundry/loggregator_consumer/dropsonde_consumer"
	"github.com/cloudfoundry/loggregator_consumer/noaa_errors"
	"github.com/cloudfoundry/loggregatorlib/server/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

var _ = Describe("Dropsonde Consumer", func() {
	var (
		connection        dropsonde_consumer.DropsondeConsumer
		endpoint          string
		testServer        *httptest.Server
		fakeHandler       *FakeHandler
		tlsSettings       *tls.Config
		consumerProxyFunc func(*http.Request) (*url.URL, error)

		appGuid        string
		authToken      string
		incomingChan   <-chan *events.Envelope
		messagesToSend chan []byte

		err error
	)

	BeforeSuite(func() {
		buf := &bytes.Buffer{}
		log.SetOutput(buf)
	})

	BeforeEach(func() {
		messagesToSend = make(chan []byte, 256)
	})

	AfterEach(func() {
		if testServer != nil {
			testServer.Close()
		}
	})

	Describe("SetOnConnectCallback", func() {
		BeforeEach(func() {
			testServer = httptest.NewServer(handlers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond))
			endpoint = "ws://" + testServer.Listener.Addr().String()
			close(messagesToSend)
		})

		It("sets a callback and calls it when connecting", func() {
			called := false
			cb := func() { called = true }

			connection = dropsonde_consumer.NewDropsondeConsumer(endpoint, tlsSettings, nil)
			connection.SetOnConnectCallback(cb)
			connection.TailingLogs(appGuid, authToken)

			Eventually(func() bool { return called }).Should(BeTrue())
		})

		Context("when the connection fails", func() {
			It("does not call the callback", func() {
				endpoint = "!!!bad-endpoint"

				called := false
				cb := func() { called = true }

				connection = dropsonde_consumer.NewDropsondeConsumer(endpoint, tlsSettings, nil)
				connection.SetOnConnectCallback(cb)
				connection.TailingLogs(appGuid, authToken)

				Consistently(func() bool { return called }).Should(BeFalse())
			})
		})

		Context("when authorization fails", func() {
			var failer authFailer
			var endpoint string

			BeforeEach(func() {
				failer = authFailer{Message: "Helpful message"}
				testServer = httptest.NewServer(failer)
				endpoint = "ws://" + testServer.Listener.Addr().String()
			})

			It("does not call the callback", func() {
				called := false
				cb := func() { called = true }

				connection = dropsonde_consumer.NewDropsondeConsumer(endpoint, tlsSettings, nil)
				connection.SetOnConnectCallback(cb)
				connection.TailingLogs(appGuid, authToken)

				Consistently(func() bool { return called }).Should(BeFalse())
			})

		})
	})

	var startFakeTrafficController = func() {
		fakeHandler = &FakeHandler{innerHandler: handlers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond)}
		testServer = httptest.NewServer(fakeHandler)
		endpoint = "ws://" + testServer.Listener.Addr().String()
		appGuid = "app-guid"
	}

	Describe("Debug Printing", func() {
		var debugPrinter *fakeDebugPrinter

		BeforeEach(func() {
			startFakeTrafficController()

			debugPrinter = &fakeDebugPrinter{}
			connection = dropsonde_consumer.NewDropsondeConsumer(endpoint, tlsSettings, consumerProxyFunc)
			connection.SetDebugPrinter(debugPrinter)
		})

		It("includes websocket handshake", func() {
			close(messagesToSend)
			connection.TailingLogs(appGuid, authToken)

			Expect(debugPrinter.Messages[0].Body).To(ContainSubstring("Sec-WebSocket-Version: 13"))
		})

		It("does not include messages sent or received", func() {
			messagesToSend <- marshalMessage(createMessage("hello", 0))

			close(messagesToSend)
			connection.TailingLogs(appGuid, authToken)

			Expect(debugPrinter.Messages[0].Body).ToNot(ContainSubstring("hello"))
		})
	})

	Describe("TailingLogs", func() {
		perform := func() {
			connection = dropsonde_consumer.NewDropsondeConsumer(endpoint, tlsSettings, consumerProxyFunc)
			incomingChan, err = connection.TailingLogs(appGuid, authToken)
		}

		BeforeEach(func() {
			startFakeTrafficController()
		})

		Context("when there is no TLS Config or consumerProxyFunc setting", func() {
			Context("when the connection can be established", func() {
				It("receives messages on the incoming channel", func(done Done) {
					messagesToSend <- marshalMessage(createMessage("hello", 0))

					perform()
					message := <-incomingChan

					Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))
					close(messagesToSend)

					close(done)
				})

				It("closes the channel after the server closes the connection", func(done Done) {
					perform()
					close(messagesToSend)

					Eventually(incomingChan).Should(BeClosed())

					close(done)
				})

				It("sends a keepalive to the server", func() {
					messageCountingServer := &messageCountingHandler{}
					testServer := httptest.NewServer(websocket.Handler(messageCountingServer.handle))
					defer testServer.Close()

					dropsonde_consumer.KeepAlive = 10 * time.Millisecond

					connection = dropsonde_consumer.NewDropsondeConsumer("ws://"+testServer.Listener.Addr().String(), tlsSettings, consumerProxyFunc)
					incomingChan, err = connection.TailingLogs(appGuid, authToken)
					defer connection.Close()

					Eventually(messageCountingServer.count).Should(BeNumerically("~", 10, 2))
				})

				It("sends messages for a specific app", func() {
					appGuid = "the-app-guid"
					perform()
					close(messagesToSend)

					Eventually(fakeHandler.getLastURL).Should(ContainSubstring("/apps/the-app-guid/tailinglogs"))
				})

				It("sends an Authorization header with an access token", func() {
					authToken = "auth-token"
					perform()
					close(messagesToSend)

					Eventually(fakeHandler.getAuthHeader).Should(Equal("auth-token"))
				})

				Context("when the message fails to parse", func() {
					It("skips that message but continues to read messages", func(done Done) {
						messagesToSend <- []byte{0}
						messagesToSend <- marshalMessage(createMessage("hello", 0))
						perform()
						close(messagesToSend)

						message := <-incomingChan

						Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))

						close(done)
					})
				})
			})

			Context("when the connection cannot be established", func() {
				BeforeEach(func() {
					endpoint = "!!!bad-endpoint"
				})

				It("returns an error", func(done Done) {
					perform()

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Please ask your Cloud Foundry Operator"))

					close(done)
				})
			})

			Context("when the authorization fails", func() {
				var failer authFailer

				BeforeEach(func() {
					failer = authFailer{Message: "Helpful message"}
					testServer = httptest.NewServer(failer)
					endpoint = "ws://" + testServer.Listener.Addr().String()
				})

				It("it returns a helpful error message", func() {
					perform()

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
					Expect(err).To(BeAssignableToTypeOf(&noaa_errors.UnauthorizedError{}))
				})
			})
		})

		Context("when SSL settings are passed in", func() {
			BeforeEach(func() {
				//				fakeHandler = &FakeHandler{innerHandler: }
				testServer = httptest.NewTLSServer(handlers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond))
				endpoint = "wss://" + testServer.Listener.Addr().String()

				tlsSettings = &tls.Config{InsecureSkipVerify: true}
			})

			It("connects using those settings", func() {
				perform()
				close(messagesToSend)

				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Stream", func() {
		perform := func() {
			connection = dropsonde_consumer.NewDropsondeConsumer(endpoint, tlsSettings, consumerProxyFunc)
			incomingChan, err = connection.Stream(appGuid, authToken)
		}

		BeforeEach(func() {
			startFakeTrafficController()
		})

		Context("when there is no TLS Config or consumerProxyFunc setting", func() {
			Context("when the connection can be established", func() {
				It("receives messages on the incoming channel", func(done Done) {
					messagesToSend <- marshalMessage(createMessage("hello", 0))

					perform()
					message := <-incomingChan

					Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))
					close(messagesToSend)

					close(done)
				})

				It("closes the channel after the server closes the connection", func(done Done) {
					perform()
					close(messagesToSend)

					Eventually(incomingChan).Should(BeClosed())

					close(done)
				})

				It("sends a keepalive to the server", func() {
					messageCountingServer := &messageCountingHandler{}
					testServer := httptest.NewServer(websocket.Handler(messageCountingServer.handle))
					defer testServer.Close()

					dropsonde_consumer.KeepAlive = 10 * time.Millisecond

					connection = dropsonde_consumer.NewDropsondeConsumer("ws://"+testServer.Listener.Addr().String(), tlsSettings, consumerProxyFunc)
					incomingChan, err = connection.Stream(appGuid, authToken)
					defer connection.Close()

					Eventually(messageCountingServer.count).Should(BeNumerically("~", 10, 2))
				})

				It("sends messages for a specific app", func() {
					appGuid = "the-app-guid"
					perform()
					close(messagesToSend)

					Eventually(fakeHandler.getLastURL).Should(ContainSubstring("/apps/the-app-guid/stream"))
				})

				It("sends an Authorization header with an access token", func() {
					authToken = "auth-token"
					perform()
					close(messagesToSend)

					Eventually(fakeHandler.getAuthHeader).Should(Equal("auth-token"))
				})

				Context("when the message fails to parse", func() {
					It("skips that message but continues to read messages", func(done Done) {
						messagesToSend <- []byte{0}
						messagesToSend <- marshalMessage(createMessage("hello", 0))
						perform()
						close(messagesToSend)

						message := <-incomingChan

						Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))

						close(done)
					})
				})
			})

			Context("when the connection cannot be established", func() {
				BeforeEach(func() {
					endpoint = "!!!bad-endpoint"
				})

				It("returns an error", func(done Done) {
					perform()

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Please ask your Cloud Foundry Operator"))

					close(done)
				})
			})

			Context("when the authorization fails", func() {
				var failer authFailer

				BeforeEach(func() {
					failer = authFailer{Message: "Helpful message"}
					testServer = httptest.NewServer(failer)
					endpoint = "ws://" + testServer.Listener.Addr().String()
				})

				It("it returns a helpful error message", func() {
					perform()

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
					Expect(err).To(BeAssignableToTypeOf(&noaa_errors.UnauthorizedError{}))
				})
			})
		})

		Context("when SSL settings are passed in", func() {
			BeforeEach(func() {
				//				fakeHandler = &FakeHandler{innerHandler: }
				testServer = httptest.NewTLSServer(handlers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond))
				endpoint = "wss://" + testServer.Listener.Addr().String()

				tlsSettings = &tls.Config{InsecureSkipVerify: true}
			})

			It("connects using those settings", func() {
				perform()
				close(messagesToSend)

				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Close", func() {
		BeforeEach(func() {
			fakeHandler = &FakeHandler{innerHandler: handlers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond)}
			testServer = httptest.NewServer(fakeHandler)
			endpoint = "ws://" + testServer.Listener.Addr().String()
		})

		Context("when a connection is not open", func() {
			It("returns an error", func() {
				connection = dropsonde_consumer.NewDropsondeConsumer(endpoint, nil, nil)
				err := connection.Close()

				Expect(err.Error()).To(Equal("connection does not exist"))
			})
		})

		Context("when a connection is open", func() {
			It("closes any open channels", func(done Done) {
				connection = dropsonde_consumer.NewDropsondeConsumer(endpoint, nil, nil)
				incomingChan, err := connection.TailingLogs("app-guid", "auth-token")
				close(messagesToSend)

				Eventually(fakeHandler.wasCalled).Should(BeTrue())

				connection.Close()

				Expect(err).NotTo(HaveOccurred())
				Eventually(incomingChan).Should(BeClosed())

				close(done)
			})
		})
	})

	Describe("RecentLogs with http", func() {
		var (
			appGuid             = "appGuid"
			authToken           = "authToken"
			receivedLogMessages []*events.Envelope
			recentError         error
		)

		perform := func() {
			close(messagesToSend)
			connection = dropsonde_consumer.NewDropsondeConsumer(endpoint, nil, nil)
			receivedLogMessages, recentError = connection.RecentLogs(appGuid, authToken)
		}

		Context("when the connection cannot be established", func() {
			It("invalid endpoints return error", func() {
				endpoint = "invalid-endpoint"
				perform()

				Expect(recentError).ToNot(BeNil())
			})
		})

		Context("when the connection can be established", func() {

			BeforeEach(func() {
				testServer = httptest.NewServer(handlers.NewHttpHandler(messagesToSend))
				endpoint = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns messages from the server", func() {
				messagesToSend <- marshalMessage(createMessage("test-message-0", 0))
				messagesToSend <- marshalMessage(createMessage("test-message-1", 0))

				perform()

				Expect(recentError).NotTo(HaveOccurred())
				Expect(receivedLogMessages).To(HaveLen(2))
				Expect(receivedLogMessages[0].GetLogMessage().GetMessage()).To(Equal([]byte("test-message-0")))
				Expect(receivedLogMessages[1].GetLogMessage().GetMessage()).To(Equal([]byte("test-message-1")))
			})
		})

		Context("when the content type is missing", func() {
			BeforeEach(func() {

				serverMux := http.NewServeMux()
				serverMux.HandleFunc("/apps/appGuid/recentlogs", func(resp http.ResponseWriter, req *http.Request) {
					resp.Header().Set("Content-Type", "")
					resp.Write([]byte("OK"))
				})
				testServer = httptest.NewServer(serverMux)
				endpoint = "ws://" + testServer.Listener.Addr().String()
			})

			It("it returns a bad reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(dropsonde_consumer.ErrBadResponse))
			})

		})

		Context("when the content length is unknown", func() {
			BeforeEach(func() {
				fakeHandler = &FakeHandler{contentLen: "-1", innerHandler: handlers.NewHttpHandler(messagesToSend)}
				testServer = httptest.NewServer(fakeHandler)
				endpoint = "ws://" + testServer.Listener.Addr().String()
			})

			It("it handles that without throwing an error", func() {
				messagesToSend <- marshalMessage(createMessage("bad-content-length", 0))
				perform()

				Expect(recentError).NotTo(HaveOccurred())
				Expect(receivedLogMessages).To(HaveLen(1))
			})

		})

		Context("when the content type doesn't have a boundary", func() {
			BeforeEach(func() {

				serverMux := http.NewServeMux()
				serverMux.HandleFunc("/apps/appGuid/recentlogs", func(resp http.ResponseWriter, req *http.Request) {
					resp.Write([]byte("OK"))
				})
				testServer = httptest.NewServer(serverMux)
				endpoint = "ws://" + testServer.Listener.Addr().String()
			})

			It("it returns a bad reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(dropsonde_consumer.ErrBadResponse))
			})

		})

		Context("when the content type's boundary is blank", func() {
			BeforeEach(func() {

				serverMux := http.NewServeMux()
				serverMux.HandleFunc("/apps/appGuid/recentlogs", func(resp http.ResponseWriter, req *http.Request) {
					resp.Header().Set("Content-Type", "boundary=")
					resp.Write([]byte("OK"))
				})
				testServer = httptest.NewServer(serverMux)
				endpoint = "ws://" + testServer.Listener.Addr().String()
			})

			It("it returns a bad reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(dropsonde_consumer.ErrBadResponse))
			})

		})

		Context("when the path is not found", func() {
			BeforeEach(func() {

				serverMux := http.NewServeMux()
				serverMux.HandleFunc("/apps/appGuid/recentlogs", func(resp http.ResponseWriter, req *http.Request) {
					resp.WriteHeader(http.StatusNotFound)
				})
				testServer = httptest.NewServer(serverMux)
				endpoint = "ws://" + testServer.Listener.Addr().String()
			})

			It("it returns a not found reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(dropsonde_consumer.ErrNotFound))
			})

		})

		Context("when the authorization fails", func() {
			var failer authFailer

			BeforeEach(func() {
				failer = authFailer{Message: "Helpful message"}
				serverMux := http.NewServeMux()
				serverMux.Handle("/apps/appGuid/recentlogs", failer)
				testServer = httptest.NewServer(serverMux)
				endpoint = "ws://" + testServer.Listener.Addr().String()
			})

			It("it returns a helpful error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
				Expect(recentError).To(BeAssignableToTypeOf(&noaa_errors.UnauthorizedError{}))
			})
		})
	})

	Describe("RecentLogs", func() {
		var (
			appGuid     string
			authToken   string
			logMessages []*events.Envelope
			recentError error
		)

		perform := func() {
			close(messagesToSend)
			connection = dropsonde_consumer.NewDropsondeConsumer(endpoint, nil, nil)
			logMessages, recentError = connection.RecentLogs(appGuid, authToken)
		}

		BeforeEach(func() {
			fakeHandler = &FakeHandler{innerHandler: handlers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond)}
			testServer = httptest.NewServer(fakeHandler)
			endpoint = "ws://" + testServer.Listener.Addr().String()
		})

		Context("when the connection cannot be established", func() {
			It("returns an error", func() {
				endpoint = "invalid-endpoint"
				perform()

				Expect(recentError).ToNot(BeNil())
			})
		})

		Context("when the connection can be established", func() {
			It("connects to the loggregator server", func() {
				perform()

				Expect(fakeHandler.wasCalled()).To(BeTrue())
			})

			It("returns messages from the server", func() {
				messagesToSend <- marshalMessage(createMessage("test-message-0", 0))
				messagesToSend <- marshalMessage(createMessage("test-message-1", 0))
				perform()

				Expect(logMessages).To(HaveLen(2))
				Expect(logMessages[0].GetLogMessage().GetMessage()).To(Equal([]byte("test-message-0")))
				Expect(logMessages[1].GetLogMessage().GetMessage()).To(Equal([]byte("test-message-1")))
			})

			It("calls the right path on the loggregator endpoint", func() {
				appGuid = "app-guid"
				perform()

				Expect(fakeHandler.getLastURL()).To(ContainSubstring("/dump/?app=app-guid"))
			})
		})

		Context("when the authorization fails", func() {
			var failer authFailer

			BeforeEach(func() {
				failer = authFailer{Message: "Helpful message"}
				testServer = httptest.NewServer(failer)
				endpoint = "ws://" + testServer.Listener.Addr().String()
			})

			It("it returns a helpful error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
			})
		})
	})

	Describe("SortRecent", func() {
		var messages []*events.Envelope

		BeforeEach(func() {
			messages = []*events.Envelope{createMessage("hello", 2), createMessage("konnichiha", 1)}
		})

		It("sorts messages", func() {
			sortedMessages := dropsonde_consumer.SortRecent(messages)

			Expect(*sortedMessages[0].Timestamp).To(Equal(int64(1)))
			Expect(*sortedMessages[1].Timestamp).To(Equal(int64(2)))
		})

		It("sorts using a stable algorithm", func() {
			messages = append(messages, createMessage("guten tag", 1))

			sortedMessages := dropsonde_consumer.SortRecent(messages)

			Expect(sortedMessages[0].GetLogMessage().GetMessage()).To(Equal([]byte("konnichiha")))
			Expect(sortedMessages[1].GetLogMessage().GetMessage()).To(Equal([]byte("guten tag")))
			Expect(sortedMessages[2].GetLogMessage().GetMessage()).To(Equal([]byte("hello")))
		})
	})
})

func createMessage(message string, timestamp int64) *events.Envelope {
	if timestamp == 0 {
		timestamp = time.Now().UnixNano()
	}

	logMessageType := events.LogMessage_OUT
	logMessage := &events.LogMessage{
		Message:     []byte(message),
		MessageType: &logMessageType,
		AppId:       proto.String("my-app-guid"),
		SourceType:  proto.String("DEA"),
		Timestamp:   proto.Int64(timestamp),
	}

	eventType := events.Envelope_LogMessage
	return &events.Envelope{
		LogMessage: logMessage,
		EventType:  &eventType,
		Origin:     proto.String("fake-origin-1"),
		Timestamp:  proto.Int64(timestamp),
	}
}

func marshalMessage(message *events.Envelope) []byte {
	data, err := proto.Marshal(message)
	if err != nil {
		log.Println(err.Error())
	}

	return data
}

type authFailer struct {
	Message string
}

func (failer authFailer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("WWW-Authenticate", "Basic")
	rw.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(rw, "You are not authorized. %s", failer.Message)
}

type messageCountingHandler struct {
	msgCount int32
}

func (mch *messageCountingHandler) handle(conn *websocket.Conn) {
	buffer := make([]byte, 1024)
	var err error
	for err == nil {
		_, err = conn.Read(buffer)
		if err == nil {
			atomic.AddInt32(&mch.msgCount, 1)
		}
	}
}

func (mch *messageCountingHandler) count() int32 {
	return atomic.LoadInt32(&mch.msgCount)
}

type FakeHandler struct {
	innerHandler http.Handler
	called       bool
	lastURL      string
	authHeader   string
	contentLen   string
	sync.RWMutex
}

func (fh *FakeHandler) getAuthHeader() string {
	fh.RLock()
	defer fh.RUnlock()
	return fh.authHeader
}

func (fh *FakeHandler) setAuthHeader(authHeader string) {
	fh.Lock()
	defer fh.Unlock()
	fh.authHeader = authHeader
}

func (fh *FakeHandler) getLastURL() string {
	fh.RLock()
	defer fh.RUnlock()
	return fh.lastURL
}

func (fh *FakeHandler) setLastURL(url string) {
	fh.Lock()
	defer fh.Unlock()
	fh.lastURL = url
}

func (fh *FakeHandler) call() {
	fh.Lock()
	defer fh.Unlock()
	fh.called = true
}

func (fh *FakeHandler) wasCalled() bool {
	fh.RLock()
	defer fh.RUnlock()
	return fh.called
}

func (fh *FakeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fh.setLastURL(r.URL.String())
	fh.setAuthHeader(r.Header.Get("Authorization"))
	fh.call()
	if len(fh.contentLen) > 0 {
		rw.Header().Set("Content-Length", fh.contentLen)
	}
	fh.innerHandler.ServeHTTP(rw, r)
}

type fakeDebugPrinter struct {
	Messages []*fakeDebugPrinterMessage
}

type fakeDebugPrinterMessage struct {
	Title, Body string
}

func (p *fakeDebugPrinter) Print(title, body string) {
	message := &fakeDebugPrinterMessage{title, body}
	p.Messages = append(p.Messages, message)
}
