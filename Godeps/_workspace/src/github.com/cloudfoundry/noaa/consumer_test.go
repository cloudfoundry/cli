package noaa_test

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"time"

	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/cloudfoundry/loggregatorlib/server/handlers"
	"github.com/cloudfoundry/noaa"
	noaa_errors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Noaa", func() {
	var (
		cnsmr                *noaa.Consumer
		trafficControllerUrl string
		testServer           *httptest.Server
		fakeHandler          *FakeHandler
		tlsSettings          *tls.Config
		consumerProxyFunc    func(*http.Request) (*url.URL, error)

		appGuid        string
		authToken      string
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
			testServer = httptest.NewServer(handlers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond, loggertesthelper.Logger()))
			trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			close(messagesToSend)
		})

		It("sets a callback and calls it when connecting", func() {
			called := false
			cb := func() { called = true }

			cnsmr = noaa.NewConsumer(trafficControllerUrl, tlsSettings, nil)
			cnsmr.SetOnConnectCallback(cb)

			logChan := make(chan *events.LogMessage, 100)
			cnsmr.TailingLogsWithoutReconnect(appGuid, authToken, logChan)

			Eventually(func() bool { return called }).Should(BeTrue())
		})

		Context("when the connection fails", func() {
			It("does not call the callback", func() {
				trafficControllerUrl = "!!!bad-url"

				called := false
				cb := func() { called = true }

				cnsmr = noaa.NewConsumer(trafficControllerUrl, tlsSettings, nil)
				cnsmr.SetOnConnectCallback(cb)
				logChan := make(chan *events.LogMessage, 100)
				cnsmr.TailingLogsWithoutReconnect(appGuid, authToken, logChan)

				Consistently(func() bool { return called }).Should(BeFalse())
			})
		})

		Context("when authorization fails", func() {
			var failer authFailer
			var trafficControllerUrl string

			BeforeEach(func() {
				failer = authFailer{Message: "Helpful message"}
				testServer = httptest.NewServer(failer)
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("does not call the callback", func() {
				called := false
				cb := func() { called = true }

				cnsmr = noaa.NewConsumer(trafficControllerUrl, tlsSettings, nil)
				cnsmr.SetOnConnectCallback(cb)
				logChan := make(chan *events.LogMessage, 100)
				cnsmr.TailingLogsWithoutReconnect(appGuid, authToken, logChan)

				Consistently(func() bool { return called }).Should(BeFalse())
			})

		})
	})

	var startFakeTrafficController = func() {
		fakeHandler = &FakeHandler{
			InputChan: make(chan []byte, 10),
			GenerateHandler: func(input chan []byte) http.Handler {
				return handlers.NewWebsocketHandler(input, 100*time.Millisecond, loggertesthelper.Logger())
			},
		}

		testServer = httptest.NewServer(fakeHandler)
		trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
		appGuid = "app-guid"
	}

	Describe("Debug Printing", func() {
		var debugPrinter *fakeDebugPrinter

		BeforeEach(func() {
			startFakeTrafficController()

			debugPrinter = &fakeDebugPrinter{}
			cnsmr = noaa.NewConsumer(trafficControllerUrl, tlsSettings, consumerProxyFunc)
			cnsmr.SetDebugPrinter(debugPrinter)
		})

		It("includes websocket handshake", func() {
			fakeHandler.Close()

			logChan := make(chan *events.LogMessage, 100)
			cnsmr.TailingLogsWithoutReconnect(appGuid, authToken, logChan)

			Eventually(func() int { return len(debugPrinter.Messages) }).Should(BeNumerically(">=", 1))
			Expect(debugPrinter.Messages[0].Body).To(ContainSubstring("Sec-WebSocket-Version: 13"))
		})

		It("does not include messages sent or received", func() {
			fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

			fakeHandler.Close()
			logChan := make(chan *events.LogMessage, 100)
			cnsmr.TailingLogsWithoutReconnect(appGuid, authToken, logChan)

			Eventually(func() int { return len(debugPrinter.Messages) }).Should(BeNumerically(">=", 1))
			Expect(debugPrinter.Messages[0].Body).ToNot(ContainSubstring("hello"))
		})
	})

	Describe("TailingLogsWithoutReconnect", func() {
		var logMessageChan chan *events.LogMessage
		var errorChan chan error
		var finishedChan chan struct{}

		perform := func() {
			cnsmr = noaa.NewConsumer(trafficControllerUrl, tlsSettings, consumerProxyFunc)

			errorChan = make(chan error, 10)
			logMessageChan = make(chan *events.LogMessage)
			go func() {
				errorChan <- cnsmr.TailingLogsWithoutReconnect(appGuid, authToken, logMessageChan)
				close(finishedChan)
			}()
		}

		BeforeEach(func() {
			finishedChan = make(chan struct{})
			startFakeTrafficController()
		})

		AfterEach(func() {
			cnsmr.Close()
			<-finishedChan
		})

		Context("when there is no TLS Config or consumerProxyFunc setting", func() {
			Context("when the connection can be established", func() {
				It("receives messages on the incoming channel", func(done Done) {
					fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

					perform()
					message := <-logMessageChan

					Expect(message.GetMessage()).To(Equal([]byte("hello")))
					fakeHandler.Close()

					close(done)
				})

				It("does not include metrics", func(done Done) {
					fakeHandler.InputChan <- marshalMessage(createContainerMetric(int32(1), int64(2)))
					fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

					perform()
					message := <-logMessageChan

					Expect(message.GetMessage()).To(Equal([]byte("hello")))
					fakeHandler.Close()

					close(done)
				})

				It("sends messages for a specific app", func() {
					appGuid = "the-app-guid"
					perform()
					fakeHandler.Close()

					Eventually(fakeHandler.getLastURL).Should(ContainSubstring("/apps/the-app-guid/stream"))
				})

				It("sends an Authorization header with an access token", func() {
					authToken = "auth-token"
					perform()
					fakeHandler.Close()

					Eventually(fakeHandler.getAuthHeader).Should(Equal("auth-token"))
				})

				Context("when remote connection dies unexpectedly", func() {
					It("receives a message on the error channel", func(done Done) {
						perform()
						fakeHandler.Close()

						var err error
						Eventually(errorChan).Should(Receive(&err))
						Expect(err.Error()).To(ContainSubstring("websocket: close 1000"))

						close(done)
					})
				})

				Context("when the message fails to parse", func() {
					It("skips that message but continues to read messages", func(done Done) {
						fakeHandler.InputChan <- []byte{0}
						fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))
						perform()
						fakeHandler.Close()

						message := <-logMessageChan

						Expect(message.GetMessage()).To(Equal([]byte("hello")))

						close(done)
					})
				})
			})

			Context("when the connection cannot be established", func() {
				BeforeEach(func() {
					trafficControllerUrl = "!!!bad-url"
				})

				It("receives an error on errChan", func(done Done) {
					perform()

					var err error
					Eventually(errorChan).Should(Receive(&err))
					Expect(err.Error()).To(ContainSubstring("Please ask your Cloud Foundry Operator"))

					close(done)
				})
			})

			Context("when the authorization fails", func() {
				var failer authFailer

				BeforeEach(func() {
					failer = authFailer{Message: "Helpful message"}
					testServer = httptest.NewServer(failer)
					trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
				})

				It("it returns a helpful error message", func() {
					perform()

					var err error
					Eventually(errorChan).Should(Receive(&err))
					Expect(err.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
				})
			})
		})

		Context("when SSL settings are passed in", func() {
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(handlers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond, loggertesthelper.Logger()))
				trafficControllerUrl = "wss://" + testServer.Listener.Addr().String()

				tlsSettings = &tls.Config{InsecureSkipVerify: true}
			})

			It("connects using those settings", func() {
				perform()
				close(messagesToSend)

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when error source is not NOAA", func() {
			It("does not pass on the error", func(done Done) {
				fakeHandler.InputChan <- marshalMessage(createError("foreign error"))

				perform()

				Consistently(errorChan).Should(BeEmpty())
				fakeHandler.Close()

				close(done)
			})

			It("continues to process log messages", func() {
				fakeHandler.InputChan <- marshalMessage(createError("foreign error"))
				fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

				perform()
				fakeHandler.Close()

				Eventually(logMessageChan).Should(Receive())
			})
		})
	})

	Describe("TailingLogs", func() {
		var logMessageChan chan *events.LogMessage
		var errorChan chan error
		var doneChan chan struct{}

		perform := func() {
			cnsmr = noaa.NewConsumer(trafficControllerUrl, tlsSettings, consumerProxyFunc)
			logMessageChan = make(chan *events.LogMessage)
			errorChan = make(chan error, 10)
			doneChan = make(chan struct{})
			go func() {
				cnsmr.TailingLogs(appGuid, authToken, logMessageChan, errorChan)
				close(doneChan)
			}()
		}

		BeforeEach(func() {
			startFakeTrafficController()
		})

		It("attempts to connect five times", func() {
			fakeHandler.fail = true
			perform()

			fakeHandler.Close()

			Eventually(errorChan, 3).Should(HaveLen(5))
			Eventually(doneChan, 10).Should(BeClosed())
		})

		It("waits 500ms before reconnecting", func() {
			perform()

			fakeHandler.Close()
			start := time.Now()
			Eventually(errorChan, 3).Should(HaveLen(5))
			end := time.Now()
			Expect(end).To(BeTemporally(">=", start.Add(4*500*time.Millisecond)))
			cnsmr.Close()
			Eventually(doneChan, 5).Should(BeClosed())
		})

		It("resets the attempt counter after a successful connection", func(done Done) {
			perform()

			fakeHandler.InputChan <- marshalMessage(createMessage("message 1", 0))
			Eventually(logMessageChan).Should(Receive())

			fakeHandler.Close()

			expectedErrorCount := 4
			Eventually(errorChan, 3*time.Second).Should(HaveLen(expectedErrorCount))
			fakeHandler.Reset()

			for i := 0; i < expectedErrorCount; i++ {
				<-errorChan
			}

			fakeHandler.InputChan <- marshalMessage(createMessage("message 2", 0))

			Eventually(logMessageChan).Should(Receive())
			fakeHandler.Close()
			Eventually(errorChan, 3).Should(HaveLen(5))

			cnsmr.Close()
			Eventually(doneChan, 5).Should(BeClosed())
			close(done)
		}, 10)

		It("will not attempt reconnect if consumer is closed", func() {
			fakeHandler.fail = true

			perform()
			Eventually(errorChan).Should(Receive())
			Expect(fakeHandler.wasCalled()).To(BeTrue())
			fakeHandler.Reset()
			cnsmr.Close()

			Eventually(errorChan).Should(BeClosed())
			Consistently(fakeHandler.wasCalled, 2).Should(BeFalse())
			Eventually(doneChan, 5).Should(BeClosed())
		})
	})

	Describe("StreamWithoutReconnect", func() {
		var incomingChan chan *events.Envelope
		var streamErrorChan chan error
		var finishedChan chan struct{}

		perform := func() {
			streamErrorChan = make(chan error, 10)
			cnsmr = noaa.NewConsumer(trafficControllerUrl, tlsSettings, consumerProxyFunc)
			go func() {
				streamErrorChan <- cnsmr.StreamWithoutReconnect(appGuid, authToken, incomingChan)
				close(finishedChan)
			}()
		}

		BeforeEach(func() {
			incomingChan = make(chan *events.Envelope)
			finishedChan = make(chan struct{})
			startFakeTrafficController()
		})

		AfterEach(func() {
			cnsmr.Close()
			<-finishedChan
		})

		Context("when there is no TLS Config or consumerProxyFunc setting", func() {
			Context("when the connection can be established", func() {
				It("receives messages on the incoming channel", func(done Done) {
					fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

					perform()
					message := <-incomingChan

					Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))
					fakeHandler.Close()

					close(done)
				})

				It("sends messages for a specific app", func() {
					appGuid = "the-app-guid"
					perform()
					fakeHandler.Close()

					Eventually(fakeHandler.getLastURL).Should(ContainSubstring("/apps/the-app-guid/stream"))
				})

				It("sends an Authorization header with an access token", func() {
					authToken = "auth-token"
					perform()
					fakeHandler.Close()

					Eventually(fakeHandler.getAuthHeader).Should(Equal("auth-token"))
				})

				Context("when the message fails to parse", func() {
					It("skips that message but continues to read messages", func(done Done) {
						fakeHandler.InputChan <- []byte{0}
						fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))
						perform()
						fakeHandler.Close()

						message := <-incomingChan

						Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))

						close(done)
					})
				})
			})

			Context("when the connection cannot be established", func() {
				BeforeEach(func() {
					trafficControllerUrl = "!!!bad-url"
				})

				It("returns an error", func(done Done) {
					perform()

					var err error
					Eventually(streamErrorChan).Should(Receive(&err))
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
					trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
				})

				It("it returns a helpful error message", func() {
					perform()

					var err error
					Eventually(streamErrorChan).Should(Receive(&err))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
				})
			})
		})

		Context("when SSL settings are passed in", func() {
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(handlers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond, loggertesthelper.Logger()))
				trafficControllerUrl = "wss://" + testServer.Listener.Addr().String()

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
		var envelopeChan chan *events.Envelope
		var errorChan chan error
		var doneChan chan struct{}

		perform := func() {
			cnsmr = noaa.NewConsumer(trafficControllerUrl, tlsSettings, consumerProxyFunc)
			envelopeChan = make(chan *events.Envelope)
			errorChan = make(chan error, 10)
			doneChan = make(chan struct{})
			go func() {
				cnsmr.Stream(appGuid, authToken, envelopeChan, errorChan)
				close(doneChan)
			}()
		}

		BeforeEach(func() {
			startFakeTrafficController()
		})

		It("attempts to connect five times", func() {
			fakeHandler.fail = true
			perform()

			fakeHandler.Close()

			Eventually(errorChan, 3).Should(HaveLen(5))
			Eventually(doneChan, 10).Should(BeClosed())
		})

		It("waits 500ms before reconnecting", func() {
			perform()

			fakeHandler.Close()
			start := time.Now()
			Eventually(errorChan, 3).Should(HaveLen(5))
			end := time.Now()
			Expect(end).To(BeTemporally(">=", start.Add(4*500*time.Millisecond)))
			cnsmr.Close()
			Eventually(doneChan).Should(BeClosed())
		})

		It("resets the attempt counter after a successful connection", func(done Done) {
			perform()

			fakeHandler.InputChan <- marshalMessage(createMessage("message 1", 0))
			Eventually(envelopeChan).Should(Receive())

			fakeHandler.Close()

			expectedErrorCount := 4
			Eventually(errorChan, 3*time.Second).Should(HaveLen(expectedErrorCount))
			fakeHandler.Reset()

			for i := 0; i < expectedErrorCount; i++ {
				<-errorChan
			}

			fakeHandler.InputChan <- marshalMessage(createMessage("message 2", 0))

			Eventually(envelopeChan).Should(Receive())
			fakeHandler.Close()
			Eventually(errorChan, 3).Should(HaveLen(5))

			cnsmr.Close()
			Eventually(doneChan).Should(BeClosed())
			close(done)
		}, 10)
	})

	Describe("Close", func() {
		var incomingChan chan *events.Envelope
		var streamErrorChan chan error

		perform := func() {
			streamErrorChan = make(chan error, 10)
			cnsmr = noaa.NewConsumer(trafficControllerUrl, nil, nil)
			go func() {
				streamErrorChan <- cnsmr.StreamWithoutReconnect(appGuid, authToken, incomingChan)
			}()
		}

		BeforeEach(func() {
			incomingChan = make(chan *events.Envelope)
			startFakeTrafficController()
		})

		Context("when a connection is not open", func() {
			It("returns an error", func() {
				cnsmr = noaa.NewConsumer(trafficControllerUrl, nil, nil)
				err := cnsmr.Close()

				Expect(err.Error()).To(Equal("connection does not exist"))
			})
		})

		Context("when a connection is open", func() {
			It("terminates the blocking function call", func(done Done) {
				perform()
				fakeHandler.Close()

				Eventually(fakeHandler.wasCalled).Should(BeTrue())
				connErr := cnsmr.Close()
				Expect(connErr.Error()).To(ContainSubstring("use of closed network connection"))

				var err error
				Eventually(streamErrorChan).Should(Receive(&err))
				Expect(err.Error()).To(ContainSubstring("websocket: close 1000"))

				close(done)
			})
		})
	})

	Describe("RecentLogs", func() {
		var (
			appGuid             = "appGuid"
			authToken           = "authToken"
			receivedLogMessages []*events.LogMessage
			recentError         error
		)

		perform := func() {
			close(messagesToSend)
			cnsmr = noaa.NewConsumer(trafficControllerUrl, nil, nil)
			receivedLogMessages, recentError = cnsmr.RecentLogs(appGuid, authToken)
		}

		Context("when the connection cannot be established", func() {
			It("invalid urls return error", func() {
				trafficControllerUrl = "invalid-url"
				perform()

				Expect(recentError).ToNot(BeNil())
			})
		})

		Context("when the connection can be established", func() {
			BeforeEach(func() {
				testServer = httptest.NewServer(handlers.NewHttpHandler(messagesToSend, loggertesthelper.Logger()))
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns messages from the server", func() {
				messagesToSend <- marshalMessage(createMessage("test-message-0", 0))
				messagesToSend <- marshalMessage(createMessage("test-message-1", 0))

				perform()

				Expect(recentError).NotTo(HaveOccurred())
				Expect(receivedLogMessages).To(HaveLen(2))
				Expect(receivedLogMessages[0].GetMessage()).To(Equal([]byte("test-message-0")))
				Expect(receivedLogMessages[1].GetMessage()).To(Equal([]byte("test-message-1")))
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
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns a bad reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(noaa.ErrBadResponse))
			})
		})

		Context("when the content length is unknown", func() {
			BeforeEach(func() {
				fakeHandler = &FakeHandler{
					contentLen: "-1",
					InputChan:  make(chan []byte, 10),
					GenerateHandler: func(input chan []byte) http.Handler {
						return handlers.NewHttpHandler(input, loggertesthelper.Logger())
					},
				}
				testServer = httptest.NewServer(fakeHandler)
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("does not throw an error", func() {
				fakeHandler.InputChan <- marshalMessage(createMessage("bad-content-length", 0))
				fakeHandler.Close()
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
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns a bad reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(noaa.ErrBadResponse))
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
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns a bad reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(noaa.ErrBadResponse))
			})

		})

		Context("when the path is not found", func() {
			BeforeEach(func() {

				serverMux := http.NewServeMux()
				serverMux.HandleFunc("/apps/appGuid/recentlogs", func(resp http.ResponseWriter, req *http.Request) {
					resp.WriteHeader(http.StatusNotFound)
				})
				testServer = httptest.NewServer(serverMux)
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns a not found reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(noaa.ErrNotFound))
			})

		})

		Context("when the authorization fails", func() {
			var failer authFailer

			BeforeEach(func() {
				failer = authFailer{Message: "Helpful message"}
				serverMux := http.NewServeMux()
				serverMux.Handle("/apps/appGuid/recentlogs", failer)
				testServer = httptest.NewServer(serverMux)
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns a helpful error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
				Expect(recentError).To(BeAssignableToTypeOf(&noaa_errors.UnauthorizedError{}))
			})
		})
	})

	Describe("ContainerMetrics", func() {
		var (
			appGuid                  = "appGuid"
			authToken                = "authToken"
			receivedContainerMetrics []*events.ContainerMetric
			recentError              error
		)

		perform := func() {
			close(messagesToSend)
			cnsmr = noaa.NewConsumer(trafficControllerUrl, nil, nil)
			receivedContainerMetrics, recentError = cnsmr.ContainerMetrics(appGuid, authToken)
		}

		Context("when the connection cannot be established", func() {
			It("invalid urls return error", func() {
				trafficControllerUrl = "invalid-url"
				perform()

				Expect(recentError).ToNot(BeNil())
			})
		})

		Context("when the connection can be established", func() {
			BeforeEach(func() {
				testServer = httptest.NewServer(handlers.NewHttpHandler(messagesToSend, loggertesthelper.Logger()))
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			Context("with a successful connection", func() {
				It("returns messages from the server", func() {
					messagesToSend <- marshalMessage(createContainerMetric(2, 2000))
					messagesToSend <- marshalMessage(createContainerMetric(1, 1000))

					perform()

					Expect(recentError).NotTo(HaveOccurred())
					Expect(receivedContainerMetrics).To(HaveLen(2))
					Expect(receivedContainerMetrics[0].GetInstanceIndex()).To(Equal(int32(1)))
					Expect(receivedContainerMetrics[1].GetInstanceIndex()).To(Equal(int32(2)))
				})
			})

			Context("when trafficcontroller returns an error as a log message", func() {
				It("returns the error", func() {
					messagesToSend <- marshalMessage(createContainerMetric(2, 2000))
					messagesToSend <- marshalMessage(createMessage("an error occurred", 2000))

					perform()

					Expect(recentError).To(HaveOccurred())
					Expect(recentError).To(MatchError("Upstream error: an error occurred"))
				})
			})
		})

		Context("when the content type is missing", func() {
			BeforeEach(func() {
				serverMux := http.NewServeMux()
				serverMux.HandleFunc("/apps/appGuid/containermetrics", func(resp http.ResponseWriter, req *http.Request) {
					resp.Header().Set("Content-Type", "")
					resp.Write([]byte("OK"))
				})
				testServer = httptest.NewServer(serverMux)
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns a bad reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(noaa.ErrBadResponse))
			})
		})

		Context("when the content length is unknown", func() {
			BeforeEach(func() {
				fakeHandler = &FakeHandler{
					contentLen: "-1",
					InputChan:  make(chan []byte, 10),
					GenerateHandler: func(input chan []byte) http.Handler {
						return handlers.NewHttpHandler(input, loggertesthelper.Logger())
					},
				}
				testServer = httptest.NewServer(fakeHandler)
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("does not throw an error", func() {
				fakeHandler.InputChan <- marshalMessage(createContainerMetric(2, 2000))
				fakeHandler.Close()
				perform()

				Expect(recentError).NotTo(HaveOccurred())
				Expect(receivedContainerMetrics).To(HaveLen(1))
			})

		})

		Context("when the content type doesn't have a boundary", func() {
			BeforeEach(func() {

				serverMux := http.NewServeMux()
				serverMux.HandleFunc("/apps/appGuid/containermetrics", func(resp http.ResponseWriter, req *http.Request) {
					resp.Write([]byte("OK"))
				})
				testServer = httptest.NewServer(serverMux)
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns a bad reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(noaa.ErrBadResponse))
			})

		})

		Context("when the content type's boundary is blank", func() {
			BeforeEach(func() {

				serverMux := http.NewServeMux()
				serverMux.HandleFunc("/apps/appGuid/containermetrics", func(resp http.ResponseWriter, req *http.Request) {
					resp.Header().Set("Content-Type", "boundary=")
					resp.Write([]byte("OK"))
				})
				testServer = httptest.NewServer(serverMux)
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns a bad reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(noaa.ErrBadResponse))
			})

		})

		Context("when the path is not found", func() {
			BeforeEach(func() {

				serverMux := http.NewServeMux()
				serverMux.HandleFunc("/apps/appGuid/containermetrics", func(resp http.ResponseWriter, req *http.Request) {
					resp.WriteHeader(http.StatusNotFound)
				})
				testServer = httptest.NewServer(serverMux)
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns a not found reponse error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError).To(Equal(noaa.ErrNotFound))
			})

		})

		Context("when the authorization fails", func() {
			var failer authFailer

			BeforeEach(func() {
				failer = authFailer{Message: "Helpful message"}
				serverMux := http.NewServeMux()
				serverMux.Handle("/apps/appGuid/containermetrics", failer)
				testServer = httptest.NewServer(serverMux)
				trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
			})

			It("returns a helpful error message", func() {
				perform()

				Expect(recentError).To(HaveOccurred())
				Expect(recentError.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
				Expect(recentError).To(BeAssignableToTypeOf(&noaa_errors.UnauthorizedError{}))
			})
		})
	})

	Describe("Firehose", func() {
		var envelopeChan chan *events.Envelope
		var errorChan chan error
		var doneChan chan struct{}

		perform := func() {
			cnsmr = noaa.NewConsumer(trafficControllerUrl, tlsSettings, consumerProxyFunc)
			envelopeChan = make(chan *events.Envelope)
			errorChan = make(chan error, 10)
			doneChan = make(chan struct{})
			go func() {
				cnsmr.Firehose("subscription-id", authToken, envelopeChan, errorChan)
				close(doneChan)
			}()
		}

		BeforeEach(func() {
			startFakeTrafficController()
		})

		It("attempts to connect five times", func() {
			fakeHandler.fail = true
			perform()

			fakeHandler.Close()

			Eventually(errorChan, 3).Should(HaveLen(5))
			Eventually(doneChan, 10).Should(BeClosed())
		})

		It("waits 500ms before reconnecting", func() {
			perform()

			fakeHandler.Close()
			start := time.Now()
			Eventually(errorChan, 3).Should(HaveLen(5))
			end := time.Now()
			Expect(end).To(BeTemporally(">=", start.Add(4*500*time.Millisecond)))
			cnsmr.Close()
			Eventually(doneChan).Should(BeClosed())
		})

		It("resets the attempt counter after a successful connection", func(done Done) {
			perform()

			fakeHandler.InputChan <- marshalMessage(createMessage("message 1", 0))
			Eventually(envelopeChan).Should(Receive())

			fakeHandler.Close()

			expectedErrorCount := 4
			Eventually(errorChan, 3*time.Second).Should(HaveLen(expectedErrorCount))
			fakeHandler.Reset()

			for i := 0; i < expectedErrorCount; i++ {
				<-errorChan
			}

			fakeHandler.InputChan <- marshalMessage(createMessage("message 2", 0))

			Eventually(envelopeChan).Should(Receive())
			fakeHandler.Close()
			Eventually(errorChan, 3).Should(HaveLen(5))

			cnsmr.Close()
			Eventually(doneChan).Should(BeClosed())
			close(done)
		}, 10)
	})

	Describe("FirehoseWithoutReconnect", func() {
		var incomingChan chan *events.Envelope
		var streamErrorChan chan error
		var finishedChan chan struct{}

		perform := func() {
			streamErrorChan = make(chan error, 10)
			cnsmr = noaa.NewConsumer(trafficControllerUrl, tlsSettings, consumerProxyFunc)
			go func() {
				streamErrorChan <- cnsmr.FirehoseWithoutReconnect("subscription-id", authToken, incomingChan)
				close(finishedChan)
			}()
		}

		BeforeEach(func() {
			incomingChan = make(chan *events.Envelope)
			finishedChan = make(chan struct{})
			startFakeTrafficController()
		})

		AfterEach(func() {
			<-finishedChan
		})

		Context("when there is no TLS Config or consumerProxyFunc setting", func() {
			Context("when the connection can be established", func() {
				It("receives messages on the incoming channel", func(done Done) {
					fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

					perform()
					message := <-incomingChan

					Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))
					fakeHandler.Close()

					close(done)
				})

				It("receives messages from the full firehose", func() {
					perform()
					fakeHandler.Close()

					Eventually(fakeHandler.getLastURL).Should(ContainSubstring("/firehose/subscription-id"))
				})

				It("sends an Authorization header with an access token", func() {
					authToken = "auth-token"
					perform()
					fakeHandler.Close()

					Eventually(fakeHandler.getAuthHeader).Should(Equal("auth-token"))
				})

				Context("when the message fails to parse", func() {
					It("skips that message but continues to read messages", func(done Done) {
						fakeHandler.InputChan <- []byte{0}
						fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))
						perform()
						fakeHandler.Close()

						message := <-incomingChan

						Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))

						close(done)
					})
				})
			})

			Context("when the connection cannot be established", func() {
				BeforeEach(func() {
					trafficControllerUrl = "!!!bad-url"
				})

				It("returns an error", func(done Done) {
					perform()

					var err error
					Eventually(streamErrorChan).Should(Receive(&err))
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
					trafficControllerUrl = "ws://" + testServer.Listener.Addr().String()
				})

				It("it returns a helpful error message", func() {
					perform()

					var err error
					Eventually(streamErrorChan).Should(Receive(&err))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
				})
			})
		})

		Context("when SSL settings are passed in", func() {
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(handlers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond, loggertesthelper.Logger()))
				trafficControllerUrl = "wss://" + testServer.Listener.Addr().String()

				tlsSettings = &tls.Config{InsecureSkipVerify: true}
			})

			It("connects using those settings", func() {
				perform()
				close(messagesToSend)

				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

func createMessage(message string, timestamp int64) *events.Envelope {
	if timestamp == 0 {
		timestamp = time.Now().UnixNano()
	}

	logMessage := createLogMessage(message, timestamp)

	return &events.Envelope{
		LogMessage: logMessage,
		EventType:  events.Envelope_LogMessage.Enum(),
		Origin:     proto.String("fake-origin-1"),
		Timestamp:  proto.Int64(timestamp),
	}
}

func createContainerMetric(instanceIndex int32, timestamp int64) *events.Envelope {
	if timestamp == 0 {
		timestamp = time.Now().UnixNano()
	}

	cm := &events.ContainerMetric{
		ApplicationId: proto.String("appId"),
		InstanceIndex: proto.Int32(instanceIndex),
		CpuPercentage: proto.Float64(1),
		MemoryBytes:   proto.Uint64(2),
		DiskBytes:     proto.Uint64(3),
	}

	return &events.Envelope{
		ContainerMetric: cm,
		EventType:       events.Envelope_ContainerMetric.Enum(),
		Origin:          proto.String("fake-origin-1"),
		Timestamp:       proto.Int64(timestamp),
	}
}

func createError(message string) *events.Envelope {
	timestamp := time.Now().UnixNano()

	err := &events.Error{
		Message: &message,
		Source:  proto.String("foreign"),
		Code:    proto.Int32(42),
	}

	return &events.Envelope{
		Error:     err,
		EventType: events.Envelope_Error.Enum(),
		Origin:    proto.String("fake-origin-1"),
		Timestamp: proto.Int64(timestamp),
	}
}

func createLogMessage(message string, timestamp int64) *events.LogMessage {
	return &events.LogMessage{
		Message:     []byte(message),
		MessageType: events.LogMessage_OUT.Enum(),
		AppId:       proto.String("my-app-guid"),
		SourceType:  proto.String("DEA"),
		Timestamp:   proto.Int64(timestamp),
	}
}

func marshalMessage(message *events.Envelope) []byte {
	data, err := proto.Marshal(message)
	if err != nil {
		println(err.Error())
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

type FakeHandler struct {
	GenerateHandler func(chan []byte) http.Handler
	InputChan       chan []byte
	called          bool
	lastURL         string
	authHeader      string
	contentLen      string
	fail            bool
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

	fh.Lock()
	defer fh.Unlock()

	if fh.fail {
		return
	}

	handler := fh.GenerateHandler(fh.InputChan)
	handler.ServeHTTP(rw, r)
}

func (fh *FakeHandler) Close() {
	close(fh.InputChan)
}

func (fh *FakeHandler) Reset() {
	fh.Lock()
	defer fh.Unlock()

	fh.InputChan = make(chan []byte)
	fh.called = false
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
