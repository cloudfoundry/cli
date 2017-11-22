package consumer_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/noaa/test_helpers"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Consumer (Asynchronous)", func() {
	var (
		cnsmr                *consumer.Consumer
		trafficControllerURL string
		testServer           *httptest.Server
		fakeHandler          *test_helpers.FakeHandler
		tlsSettings          *tls.Config
		maxRetryCount        int

		appGuid        string
		authToken      string
		messagesToSend chan []byte

		streamPathBuilder consumer.StreamPathBuilder
	)

	BeforeEach(func() {
		trafficControllerURL = ""
		testServer = nil
		fakeHandler = nil
		tlsSettings = nil
		maxRetryCount = consumer.DefaultMaxRetryCount

		appGuid = ""
		authToken = ""
		messagesToSend = make(chan []byte, 256)

		streamPathBuilder = nil
	})

	JustBeforeEach(func() {
		cnsmr = consumer.New(trafficControllerURL, tlsSettings, nil)
		cnsmr.SetMinRetryDelay(100 * time.Millisecond)
		cnsmr.SetMaxRetryDelay(500 * time.Millisecond)
		cnsmr.SetMaxRetryCount(maxRetryCount)

		if streamPathBuilder != nil {
			cnsmr.SetStreamPathBuilder(streamPathBuilder)
		}
	})

	AfterEach(func() {
		cnsmr.Close()
		if testServer != nil {
			testServer.Close()
		}
	})

	Describe("SetOnConnectCallback", func() {
		BeforeEach(func() {
			testServer = httptest.NewServer(NewWebsocketHandler(messagesToSend, 100*time.Millisecond))
			trafficControllerURL = "ws://" + testServer.Listener.Addr().String()
			close(messagesToSend)
		})

		It("sets a callback and calls it when connecting", func() {
			called := make(chan bool)
			cb := func() { called <- true }

			cnsmr.SetOnConnectCallback(cb)

			cnsmr.TailingLogsWithoutReconnect(appGuid, authToken)

			Eventually(called).Should(Receive())
		})

		Context("when the connection fails", func() {
			BeforeEach(func() {
				trafficControllerURL = "!!!bad-url"
			})

			It("does not call the callback", func() {
				called := make(chan bool)
				cb := func() { called <- true }

				cnsmr.SetOnConnectCallback(cb)
				cnsmr.TailingLogsWithoutReconnect(appGuid, authToken)

				Consistently(called).ShouldNot(Receive())
			})
		})

		Context("when authorization fails", func() {
			var (
				failure test_helpers.AuthFailureHandler
			)

			BeforeEach(func() {
				failure = test_helpers.AuthFailureHandler{Message: "Helpful message"}
				testServer = httptest.NewServer(failure)
				trafficControllerURL = "ws://" + testServer.Listener.Addr().String()
			})

			It("does not call the callback", func() {
				called := false
				cb := func() { called = true }

				cnsmr.SetOnConnectCallback(cb)
				_, errs := cnsmr.TailingLogsWithoutReconnect(appGuid, authToken)

				Consistently(func() bool { return called }).Should(BeFalse())
				Eventually(errs).Should(Receive(Not(BeRetryable())))
			})

		})

	})

	var startFakeTrafficController = func() {
		fakeHandler = &test_helpers.FakeHandler{
			InputChan: make(chan []byte, 10),
			GenerateHandler: func(input chan []byte) http.Handler {
				return NewWebsocketHandler(input, 100*time.Millisecond)
			},
		}

		testServer = httptest.NewServer(fakeHandler)
		trafficControllerURL = "ws://" + testServer.Listener.Addr().String()
		appGuid = "app-guid"
	}

	Describe("Debug Printing", func() {
		var debugPrinter *mockDebugPrinter

		BeforeEach(func() {
			startFakeTrafficController()
		})

		JustBeforeEach(func() {
			debugPrinter = newMockDebugPrinter()
			cnsmr.SetDebugPrinter(debugPrinter)
		})

		It("includes websocket handshake", func() {
			fakeHandler.Close()

			cnsmr.TailingLogsWithoutReconnect(appGuid, authToken)

			var body string
			Eventually(debugPrinter.PrintInput.Dump).Should(Receive(&body))
			Expect(body).To(ContainSubstring("Sec-WebSocket-Version: 13"))
		})

		It("does not include messages sent or received", func() {
			fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

			fakeHandler.Close()
			cnsmr.TailingLogsWithoutReconnect(appGuid, authToken)

			var body string
			Eventually(debugPrinter.PrintInput.Dump).Should(Receive(&body))
			Expect(body).ToNot(ContainSubstring("hello"))
		})
	})

	Describe("TailingLogsWithoutReconnect", func() {
		var (
			logMessages <-chan *events.LogMessage
			errors      <-chan error
		)

		BeforeEach(func() {
			startFakeTrafficController()
		})

		JustBeforeEach(func() {
			logMessages, errors = cnsmr.TailingLogsWithoutReconnect(appGuid, authToken)
		})

		AfterEach(func() {
			cnsmr.Close()
			Eventually(logMessages).Should(BeClosed())
		})

		Context("when there is no TLS Config or consumerProxyFunc setting", func() {
			Context("when the connection can be established", func() {
				It("returns a read only LogMessage chan and error chan", func() {
					fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

					var message *events.LogMessage
					Eventually(logMessages).Should(Receive(&message))
					Expect(message.GetMessage()).To(Equal([]byte("hello")))
					Consistently(errors).ShouldNot(Receive())
				})

				It("receives messages on the incoming channel", func(done Done) {
					fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

					message := <-logMessages

					Expect(message.GetMessage()).To(Equal([]byte("hello")))
					fakeHandler.Close()

					close(done)
				})

				It("does not include metrics", func(done Done) {
					fakeHandler.InputChan <- marshalMessage(createContainerMetric(int32(1), int64(2)))
					fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

					message := <-logMessages

					Expect(message.GetMessage()).To(Equal([]byte("hello")))
					fakeHandler.Close()

					close(done)
				})

				Context("with a specific app", func() {
					BeforeEach(func() {
						appGuid = "the-app-guid"
					})

					It("sends messages for a specific app", func() {
						fakeHandler.Close()

						Eventually(fakeHandler.GetLastURL).Should(ContainSubstring("/apps/the-app-guid/stream"))
					})

					Context("when a stream path builder is provided", func() {
						BeforeEach(func() {
							streamPathBuilder = func(appGuid string) string {
								return fmt.Sprintf("/logs/%s/stream", appGuid)
							}
						})

						It("uses the stream path from the builder", func() {
							fakeHandler.Close()

							Eventually(fakeHandler.GetLastURL).Should(ContainSubstring("/logs/the-app-guid/stream"))
						})
					})
				})

				Context("with an access token", func() {
					BeforeEach(func() {
						authToken = "auth-token"
					})

					It("sends an Authorization header with an access token", func() {
						fakeHandler.Close()

						Eventually(fakeHandler.GetAuthHeader).Should(Equal("auth-token"))
					})
				})

				Context("when remote connection dies unexpectedly", func() {
					It("receives a message on the error channel", func(done Done) {
						fakeHandler.Close()

						var err error
						Eventually(errors).Should(Receive(&err))
						Expect(err).ToNot(BeRetryable())
						Expect(err.Error()).To(ContainSubstring("websocket: close 1000"))

						close(done)
					})
				})

				Context("when the message fails to parse", func() {
					It("skips that message but continues to read messages", func(done Done) {
						fakeHandler.InputChan <- []byte{0}
						fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))
						fakeHandler.Close()

						message := <-logMessages

						Expect(message.GetMessage()).To(Equal([]byte("hello")))

						close(done)
					})
				})
			})

			Context("when the connection cannot be established", func() {
				BeforeEach(func() {
					trafficControllerURL = "!!!bad-url"
				})

				It("receives an error on errChan", func() {
					var err error
					Eventually(errors).Should(Receive(&err))
					Expect(err).ToNot(BeRetryable())
					Expect(err.Error()).To(ContainSubstring("Please ask your Cloud Foundry Operator"))
				})
			})

			Context("when the authorization fails", func() {
				var failure test_helpers.AuthFailureHandler

				BeforeEach(func() {
					failure = test_helpers.AuthFailureHandler{Message: "Helpful message"}
					testServer = httptest.NewServer(failure)
					trafficControllerURL = "ws://" + testServer.Listener.Addr().String()
				})

				It("it returns a helpful error message", func() {
					var err error
					Eventually(errors).Should(Receive(&err))
					Expect(err).ToNot(BeRetryable())
					Expect(err.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
				})
			})
		})

		Context("when SSL settings are passed in", func() {
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(NewWebsocketHandler(messagesToSend, 100*time.Millisecond))
				trafficControllerURL = "wss://" + testServer.Listener.Addr().String()

				tlsSettings = &tls.Config{InsecureSkipVerify: true}
			})

			It("connects using those settings", func() {
				Consistently(errors).ShouldNot(Receive())

				close(messagesToSend)
				Eventually(errors).Should(Receive())
			})
		})

		Context("when error source is not NOAA", func() {
			It("does not pass on the error", func(done Done) {
				fakeHandler.InputChan <- marshalMessage(createError("foreign error"))

				Consistently(errors).Should(BeEmpty())
				fakeHandler.Close()

				close(done)
			})

			It("continues to process log messages", func() {
				fakeHandler.InputChan <- marshalMessage(createError("foreign error"))
				fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

				fakeHandler.Close()

				Eventually(logMessages).Should(Receive())
			})
		})
	})

	Describe("TailingLogs", func() {
		var (
			logMessages <-chan *events.LogMessage
			errors      <-chan error
			retries     uint
		)

		BeforeEach(func() {
			retries = 5
			startFakeTrafficController()
		})

		JustBeforeEach(func() {
			logMessages, errors = cnsmr.TailingLogs(appGuid, authToken)
		})

		It("resets the delay after a successful connection", func() {
			fakeHandler.InputChan <- marshalMessage(createMessage("message 1", 0))
			Eventually(logMessages).Should(Receive())

			fakeHandler.Close()
			expectedErrorCount := 4
			for i := 0; i < expectedErrorCount; i++ {
				Eventually(errors, time.Second).Should(Receive(HaveOccurred()))
			}
			fakeHandler.Reset()

			fakeHandler.InputChan <- marshalMessage(createMessage("message 2", 0))

			Eventually(logMessages).Should(Receive())

			fakeHandler.Close()
			for i := uint(0); i < retries; i++ {
				Eventually(errors).Should(Receive(BeRetryable()))
			}
		})

		Context("when maxRetryCount is set", func() {
			BeforeEach(func() {
				maxRetryCount = 3
			})

			It("resets the count after a successful connection", func() {
				fakeHandler.InputChan <- marshalMessage(createMessage("message 1", 0))
				Eventually(logMessages).Should(Receive())

				fakeHandler.Close()
				for i := 0; i < maxRetryCount-1; i++ {
					Eventually(errors, time.Second).Should(Receive(HaveOccurred()))
				}
				fakeHandler.Reset()

				fakeHandler.InputChan <- marshalMessage(createMessage("message 2", 0))

				Eventually(logMessages).Should(Receive())

				fakeHandler.Close()
				for i := 0; i < maxRetryCount-1; i++ {
					Eventually(errors).Should(Receive(BeRetryable()))
				}
			})
		})

		Context("with multiple connections", func() {
			var (
				moreLogMessages <-chan *events.LogMessage
				moreErrors      <-chan error
			)

			JustBeforeEach(func() {
				moreLogMessages, moreErrors = cnsmr.TailingLogs(appGuid, authToken)
			})

			It("closes all channels", func() {
				cnsmr.Close()
				Eventually(logMessages).Should(BeClosed())
				Eventually(errors).Should(BeClosed())
				Eventually(moreLogMessages).Should(BeClosed())
				Eventually(moreErrors).Should(BeClosed())
			})
		})

		Context("when the connection cannot be established", func() {
			BeforeEach(func() {
				trafficControllerURL = `https://invalid`
			})

			It("returns an error", func() {
				var err error
				Eventually(errors).Should(Receive(&err))
				Expect(err).To(HaveOccurred())
				Expect(err).ToNot(BeRetryable())
				Expect(err.Error()).To(ContainSubstring("Please ask your Cloud Foundry Operator"))
			})
		})

		Context("with a failing handler", func() {
			BeforeEach(func() {
				fakeHandler.Fail = true
			})

			It("exponentially backs off", func() {
				ts := make(chan time.Duration, 100)
				done := make(chan struct{})

				var wg sync.WaitGroup
				wg.Add(1)
				defer wg.Wait()
				defer close(done)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					last := time.Now()
					for {
						select {
						case err := <-errors:
							Expect(err).To(BeRetryable())
							ts <- time.Since(last)
							last = time.Now()
						case <-done:
							return
						}
					}
				}()

				Eventually(ts).Should(Receive(BeNumerically("<", 50*time.Millisecond)))
				Eventually(ts).Should(Receive(BeNumerically("~", 100*time.Millisecond, 50*time.Millisecond)))
				Eventually(ts).Should(Receive(BeNumerically("~", 200*time.Millisecond, 50*time.Millisecond)))
				Eventually(ts).Should(Receive(BeNumerically("~", 400*time.Millisecond, 50*time.Millisecond)))
			})

			It("doesn't go beyond max sleep", func() {
				ts := make(chan time.Duration, 100)
				done := make(chan struct{})

				var wg sync.WaitGroup
				wg.Add(1)
				defer wg.Wait()
				defer close(done)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					last := time.Now()
					for {
						select {
						case err := <-errors:
							Expect(err).To(BeRetryable())
							ts <- time.Since(last)
							last = time.Now()
						case <-done:
							return
						}
					}
				}()

				Eventually(ts).Should(Receive())
				timeout := time.After(time.Second)
				for {
					select {
					case delay := <-ts:
						Expect(delay).To(BeNumerically("<", 550*time.Millisecond))
					case <-timeout:
						return
					}
				}
			})

			Context("when maxRetryCount is set", func() {
				BeforeEach(func() {
					maxRetryCount = 3
				})

				It("doesn't go beyond maxRetryCount", func() {
					for i := 0; i < maxRetryCount; i++ {
						Eventually(errors).Should(Receive(BeRetryable()))
					}
					Eventually(errors).Should(Receive(Equal(consumer.ErrMaxRetriesReached)))
					Eventually(logMessages).Should(BeClosed())
					Eventually(errors).Should(BeClosed())
				})
			})

			It("will not attempt reconnect if consumer is closed", func() {
				Eventually(errors).Should(Receive())
				Expect(fakeHandler.WasCalled()).To(BeTrue())
				fakeHandler.Reset()
				cnsmr.Close()

				Eventually(errors).Should(BeClosed())
				Consistently(fakeHandler.WasCalled, 2).Should(BeFalse())
			})
		})
	})

	Describe("StreamWithoutReconnect", func() {
		var (
			incoming <-chan *events.Envelope
			errors   <-chan error
		)

		BeforeEach(func() {
			startFakeTrafficController()
		})

		JustBeforeEach(func() {
			incoming, errors = cnsmr.StreamWithoutReconnect(appGuid, authToken)
		})

		Context("when there is no TLS Config or consumerProxyFunc setting", func() {
			Context("when the connection can be established", func() {
				It("receives messages on the incoming channel", func(done Done) {
					defer close(done)

					fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))

					var message *events.Envelope
					Eventually(incoming).Should(Receive(&message))
					Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))
					fakeHandler.Close()

				})

				Context("with a specific app ID", func() {
					BeforeEach(func() {
						appGuid = "the-app-guid"
					})

					It("sends messages for a specific app", func() {
						fakeHandler.Close()

						Eventually(fakeHandler.GetLastURL).Should(ContainSubstring("/apps/the-app-guid/stream"))
					})
				})

				Context("with an access token", func() {
					BeforeEach(func() {
						authToken = "auth-token"
					})

					It("sends an Authorization header with an access token", func() {
						fakeHandler.Close()

						Eventually(fakeHandler.GetAuthHeader).Should(Equal("auth-token"))
					})
				})

				Context("when the message fails to parse", func() {
					It("skips that message but continues to read messages", func(done Done) {
						fakeHandler.InputChan <- []byte{0}
						fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))
						fakeHandler.Close()

						message := <-incoming

						Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))

						close(done)
					})
				})
			})

			Context("when the connection cannot be established", func() {
				BeforeEach(func() {
					trafficControllerURL = `http://invalid`
				})

				It("returns an error", func(done Done) {
					var err error
					Eventually(errors).Should(Receive(&err))
					Expect(err).To(HaveOccurred())
					Expect(err).ToNot(BeRetryable())
					Expect(err.Error()).To(ContainSubstring("Please ask your Cloud Foundry Operator"))

					close(done)
				})
			})

			Context("when the authorization fails", func() {
				var failer test_helpers.AuthFailureHandler

				BeforeEach(func() {
					failer = test_helpers.AuthFailureHandler{Message: "Helpful message"}
					testServer = httptest.NewServer(failer)
					trafficControllerURL = "ws://" + testServer.Listener.Addr().String()
				})

				It("it returns a helpful error message", func() {
					var err error
					Eventually(errors).Should(Receive(&err))
					Expect(err).To(HaveOccurred())
					Expect(err).ToNot(BeRetryable())
					Expect(err.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
				})
			})
		})

		Context("when SSL settings are passed in", func() {
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(NewWebsocketHandler(messagesToSend, 100*time.Millisecond))
				trafficControllerURL = "wss://" + testServer.Listener.Addr().String()

				tlsSettings = &tls.Config{InsecureSkipVerify: true}
			})

			It("connects using those settings", func() {
				Consistently(errors).ShouldNot(Receive())

				close(messagesToSend)
				Eventually(errors).Should(Receive())
			})
		})
	})

	Describe("Stream", func() {
		var (
			envelopes <-chan *events.Envelope
			errors    <-chan error
		)

		BeforeEach(func() {
			startFakeTrafficController()
		})

		JustBeforeEach(func() {
			envelopes, errors = cnsmr.Stream(appGuid, authToken)
		})

		Context("when the connection fails", func() {
			BeforeEach(func() {
				fakeHandler.Fail = true
			})

			It("exponentially backs off", func() {
				ts := make(chan time.Duration, 100)
				done := make(chan struct{})

				var wg sync.WaitGroup
				wg.Add(1)
				defer wg.Wait()
				defer close(done)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					last := time.Now()
					for {
						select {
						case err := <-errors:
							Expect(err).To(BeRetryable())
							ts <- time.Since(last)
							last = time.Now()
						case <-done:
							return
						}
					}
				}()

				Eventually(ts).Should(Receive(BeNumerically("<", 50*time.Millisecond)))
				Eventually(ts).Should(Receive(BeNumerically("~", 100*time.Millisecond, 50*time.Millisecond)))
				Eventually(ts).Should(Receive(BeNumerically("~", 200*time.Millisecond, 50*time.Millisecond)))
				Eventually(ts).Should(Receive(BeNumerically("~", 400*time.Millisecond, 50*time.Millisecond)))
			})
		})

		It("resets the delay after a successful connection", func() {
			fakeHandler.InputChan <- marshalMessage(createMessage("message 1", 0))
			Eventually(envelopes).Should(Receive())

			fakeHandler.Close()
			expectedErrorCount := 4
			for i := 0; i < expectedErrorCount; i++ {
				Eventually(errors, time.Second).Should(Receive(BeRetryable()))
			}
			fakeHandler.Reset()

			fakeHandler.InputChan <- marshalMessage(createMessage("message 2", 0))

			Eventually(envelopes).Should(Receive())

			fakeHandler.Close()
			for i := uint(0); i < 3; i++ {
				Eventually(errors).Should(Receive(BeRetryable()))
			}
		})
	})

	Describe("Close", func() {
		var (
			incomings    <-chan *events.Envelope
			streamErrors <-chan error
		)

		BeforeEach(func() {
			startFakeTrafficController()
		})

		Context("when a connection is not open", func() {
			It("returns an error", func() {
				err := cnsmr.Close()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("connection does not exist"))
				fakeHandler.Close()
			})
		})

		Context("when a connection is open", func() {
			JustBeforeEach(func() {
				incomings, streamErrors = cnsmr.StreamWithoutReconnect(appGuid, authToken)
			})

			It("closes the outputs", func() {
				Eventually(fakeHandler.WasCalled).Should(BeTrue())

				Expect(cnsmr.Close()).To(Succeed())
				Eventually(incomings).Should(BeClosed())
				Eventually(streamErrors).Should(BeClosed())

				fakeHandler.Close()
			})

			Context("and the server is closed", func() {
				JustBeforeEach(func() {
					fakeHandler.Close()
				})

				It("returns errors", func() {
					var err error
					Eventually(streamErrors).Should(Receive(&err))
					Expect(err.Error()).To(ContainSubstring("websocket: close 1000"))

					connErr := cnsmr.Close()
					Expect(connErr).To(HaveOccurred())
					Expect(connErr.Error()).To(ContainSubstring("close sent"))
				})
			})
		})
	})

	Describe("Firehose", func() {
		var (
			envelopes <-chan *events.Envelope
			errors    <-chan error
		)

		BeforeEach(func() {
			startFakeTrafficController()
		})

		Context("with data in the server", func() {
			BeforeEach(func() {
				fakeHandler.InputChan <- marshalMessage(createMessage("message 1", 0))
			})

			JustBeforeEach(func() {
				envelopes, errors = cnsmr.Firehose("subscription-id", authToken)
			})

			It("resets the delay after a successful connection", func(done Done) {
				defer close(done)
				Eventually(envelopes).Should(Receive())

				fakeHandler.Close()

				expectedErrorCount := 4
				for i := 0; i < expectedErrorCount; i++ {
					Eventually(errors).Should(Receive(BeRetryable()))
				}
				fakeHandler.Reset()

				fakeHandler.InputChan <- marshalMessage(createMessage("message 2", 0))

				Eventually(envelopes).Should(Receive())
				fakeHandler.Close()
				for i := 0; i < 5; i++ {
					Eventually(errors).Should(Receive(BeRetryable()))
				}
			}, 10)
		})

		Context("when the connection read takes too long", func() {
			var (
				idleTimeout time.Duration
			)

			JustBeforeEach(func() {
				idleTimeout = 500 * time.Millisecond
				cnsmr.SetIdleTimeout(idleTimeout)
				envelopes, errors = cnsmr.Firehose("subscription-id", authToken)
			})

			It("returns an error when the idle timeout expires", func() {
				var err error
				Eventually(errors).Should(Receive(&err))
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeRetryable())
				Expect(err.Error()).To(ContainSubstring("i/o timeout"))
			})

			It("reestablishes the connection", func() {
				Eventually(fakeHandler.WasCalled).Should(BeTrue())
				fakeHandler.Reset()
				Eventually(fakeHandler.WasCalled).Should(BeTrue())
			})
		})
	})

	Describe("FirehoseWithoutReconnect", func() {
		var (
			incomings    <-chan *events.Envelope
			streamErrors <-chan error
			idleTimeout  time.Duration
		)

		BeforeEach(func() {
			incomings = make(chan *events.Envelope)
			startFakeTrafficController()
		})

		JustBeforeEach(func() {
			cnsmr.SetIdleTimeout(idleTimeout)
			incomings, streamErrors = cnsmr.FirehoseWithoutReconnect("subscription-id", authToken)
		})

		Context("when there is no TLS Config or consumerProxyFunc setting", func() {
			Context("when the connection can be established", func() {
				It("receives messages from the full firehose", func() {
					fakeHandler.Close()

					Eventually(fakeHandler.GetLastURL).Should(ContainSubstring("/firehose/subscription-id"))
				})

				Context("with a message", func() {
					BeforeEach(func() {
						fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))
					})

					It("receives messages on the incoming channel", func(done Done) {
						defer close(done)

						message := <-incomings

						Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))
						fakeHandler.Close()
					})
				})

				Context("with an authorization token", func() {
					BeforeEach(func() {
						authToken = "auth-token"
					})

					It("sends an Authorization header with an access token", func() {
						fakeHandler.Close()
						Eventually(fakeHandler.GetAuthHeader).Should(Equal("auth-token"))
					})
				})

				Context("when the message fails to parse", func() {
					BeforeEach(func() {
						fakeHandler.InputChan <- []byte{0}
						fakeHandler.InputChan <- marshalMessage(createMessage("hello", 0))
					})

					It("skips that message but continues to read messages", func(done Done) {
						defer close(done)
						fakeHandler.Close()

						message := <-incomings
						Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))
					})
				})
			})

			Context("when the connection cannot be established", func() {
				BeforeEach(func() {
					trafficControllerURL = "https://invalid"
				})

				It("returns an error", func(done Done) {
					defer close(done)

					var err error
					Eventually(streamErrors).Should(Receive(&err))
					Expect(err).To(HaveOccurred())
					Expect(err).ToNot(BeRetryable())
					Expect(err.Error()).To(ContainSubstring("Please ask your Cloud Foundry Operator"))
				})
			})

			Context("when the authorization fails", func() {
				var failer test_helpers.AuthFailureHandler

				BeforeEach(func() {
					failer = test_helpers.AuthFailureHandler{Message: "Helpful message"}
					testServer = httptest.NewServer(failer)
					trafficControllerURL = "ws://" + testServer.Listener.Addr().String()
				})

				It("it returns a helpful error message", func() {
					var err error
					Eventually(streamErrors).Should(Receive(&err))
					Expect(err).To(HaveOccurred())
					Expect(err).ToNot(BeRetryable())
					Expect(err.Error()).To(ContainSubstring("You are not authorized. Helpful message"))
				})
			})
		})

		Context("when the connection read takes too long", func() {
			BeforeEach(func() {
				idleTimeout = 500 * time.Millisecond
			})

			It("returns an error when the idle timeout expires", func() {
				var err error
				Eventually(streamErrors).Should(Receive(&err))
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeRetryable())
				Expect(err.Error()).To(ContainSubstring("i/o timeout"))
			})
		})

		Context("when SSL settings are passed in", func() {
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(NewWebsocketHandler(messagesToSend, 100*time.Millisecond))
				trafficControllerURL = "wss://" + testServer.Listener.Addr().String()

				tlsSettings = &tls.Config{InsecureSkipVerify: true}
			})

			It("connects using those settings", func() {
				Consistently(streamErrors).ShouldNot(Receive())

				close(messagesToSend)
				Eventually(streamErrors).Should(Receive())
			})
		})
	})

	Describe("FilteredFirehose", func() {
		BeforeEach(func() {
			startFakeTrafficController()
		})

		It("appends the correct query string for filtering log messages", func() {
			cnsmr.FilteredFirehose("subscription-id", authToken, consumer.LogMessages)
			fakeHandler.Close()
			Eventually(fakeHandler.GetLastURL).Should(ContainSubstring("/firehose/subscription-id?filter-type=logs"))
		})

		It("appends the correct query string for filtering metrics", func() {
			cnsmr.FilteredFirehose("subscription-id", authToken, consumer.Metrics)
			fakeHandler.Close()
			Eventually(fakeHandler.GetLastURL).Should(ContainSubstring("/firehose/subscription-id?filter-type=metrics"))
		})
	})
})

func BeRetryable() types.GomegaMatcher {
	return BeAssignableToTypeOf(errors.NewRetryError(fmt.Errorf("some-error")))
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
