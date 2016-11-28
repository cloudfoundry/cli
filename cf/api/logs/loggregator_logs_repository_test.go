package logs_test

import (
	"code.cloudfoundry.org/cli/cf/api/authentication/authenticationfakes"
	"code.cloudfoundry.org/cli/cf/api/logs/logsfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	noaa_errors "github.com/cloudfoundry/noaa/errors"
	"github.com/gogo/protobuf/proto"

	"time"

	. "code.cloudfoundry.org/cli/cf/api/logs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("loggregator logs repository", func() {
	var (
		fakeConsumer *logsfakes.FakeLoggregatorConsumer
		logsRepo     *LoggregatorLogsRepository
		configRepo   coreconfig.ReadWriter
		authRepo     *authenticationfakes.FakeRepository
	)

	BeforeEach(func() {
		fakeConsumer = new(logsfakes.FakeLoggregatorConsumer)
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetLoggregatorEndpoint("loggregator-server.test.com")
		configRepo.SetAccessToken("the-access-token")
		authRepo = &authenticationfakes.FakeRepository{}
	})

	JustBeforeEach(func() {
		logsRepo = NewLoggregatorLogsRepository(configRepo, fakeConsumer, authRepo)
	})

	Describe("RecentLogsFor", func() {
		Context("when a noaa_errors.UnauthorizedError occurs", func() {
			var recentCalled bool
			BeforeEach(func() {
				fakeConsumer.RecentStub = func(string, string) ([]*logmessage.LogMessage, error) {
					if recentCalled {
						return nil, nil
					}
					recentCalled = true
					return nil, noaa_errors.NewUnauthorizedError("i'm sorry dave")
				}
			})

			It("refreshes the access token", func() {
				_, err := logsRepo.RecentLogsFor("app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(authRepo.RefreshAuthTokenCallCount()).To(Equal(1))
			})
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				fakeConsumer.RecentReturns(nil, errors.New("oops"))
			})

			It("returns the error", func() {
				_, err := logsRepo.RecentLogsFor("app-guid")
				Expect(err).To(Equal(errors.New("oops")))
			})
		})

		Context("when an error does not occur", func() {
			var msg1, msg2 *logmessage.LogMessage

			BeforeEach(func() {
				msg1 = makeLogMessage("My message 2", int64(2000))
				msg2 = makeLogMessage("My message 1", int64(1000))

				fakeConsumer.RecentReturns([]*logmessage.LogMessage{
					msg1,
					msg2,
				}, nil)
			})

			It("gets the logs for the requested app", func() {
				logsRepo.RecentLogsFor("app-guid")
				appGuid, _ := fakeConsumer.RecentArgsForCall(0)
				Expect(appGuid).To(Equal("app-guid"))
			})

			It("writes the sorted log messages onto the provided channel", func() {
				messages, err := logsRepo.RecentLogsFor("app-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(messages).To(Equal([]Loggable{
					NewLoggregatorLogMessage(msg2),
					NewLoggregatorLogMessage(msg1),
				}))
			})
		})
	})

	Describe("tailing logs", func() {
		var logChan chan Loggable
		var errChan chan error

		BeforeEach(func() {
			logChan = make(chan Loggable)
			errChan = make(chan error)
		})

		Context("when an error occurs", func() {
			e := errors.New("oops")

			BeforeEach(func() {
				fakeConsumer.TailStub = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					return nil, e
				}
			})

			It("returns an error", func(done Done) {
				go func() {
					Eventually(errChan).Should(Receive(&e))

					close(done)
				}()

				logsRepo.TailLogsFor("app-guid", func() {}, logChan, errChan)
			})
		})

		Context("when a LoggregatorConsumer.UnauthorizedError occurs", func() {
			It("refreshes the access token", func(done Done) {
				calledOnce := false

				fakeConsumer.TailStub = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					if !calledOnce {
						calledOnce = true
						return nil, noaa_errors.NewUnauthorizedError("i'm sorry dave")
					} else {
						return nil, nil
					}
				}

				go func() {
					defer GinkgoRecover()

					Eventually(authRepo.RefreshAuthTokenCallCount).Should(Equal(1))
					Consistently(errChan).ShouldNot(Receive())

					close(done)
				}()

				logsRepo.TailLogsFor("app-guid", func() {}, logChan, errChan)
			})

			Context("when LoggregatorConsumer.UnauthorizedError occurs again", func() {
				It("returns an error", func(done Done) {
					err := noaa_errors.NewUnauthorizedError("All the errors")

					fakeConsumer.TailStub = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
						return nil, err
					}

					go func() {
						defer GinkgoRecover()

						// Not equivalent to ShouldNot(Receive(BeNil()))
						// Should receive something, but it shouldn't be nil
						Eventually(errChan).Should(Receive(&err))
						close(done)
					}()

					logsRepo.TailLogsFor("app-guid", func() {}, logChan, errChan)
				})
			})
		})

		Context("when no error occurs", func() {
			It("asks for the logs for the given app", func() {
				fakeConsumer.TailStub = func(appGuid, token string) (<-chan *logmessage.LogMessage, error) {
					Expect(appGuid).To(Equal("app-guid"))
					Expect(token).To(Equal("the-access-token"))
					return nil, nil
				}

				logsRepo.TailLogsFor("app-guid", func() {}, logChan, errChan)
			})

			It("sets the on connect callback", func() {
				fakeConsumer.TailStub = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					return nil, nil
				}

				called := false
				logsRepo.TailLogsFor("app-guid", func() { called = true }, logChan, errChan)

				Expect(fakeConsumer.SetOnConnectCallbackCallCount()).To(Equal(1))
				// best way we could come up with to match on a callback function
				callbackFunc := fakeConsumer.SetOnConnectCallbackArgsForCall(0)
				callbackFunc()
				Expect(called).To(Equal(true))
			})

			It("sorts the messages before yielding them", func() {
				var receivedMessages []Loggable
				msg3 := makeLogMessage("hello3", 300)
				msg2 := makeLogMessage("hello2", 200)
				msg1 := makeLogMessage("hello1", 100)

				fakeConsumer.TailStub = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					consumerLogChan := make(chan *logmessage.LogMessage)
					go func() {
						consumerLogChan <- msg3
						consumerLogChan <- msg2
						consumerLogChan <- msg1

						close(consumerLogChan)
					}()

					return consumerLogChan, nil
				}

				logsRepo.TailLogsFor("app-guid", func() {}, logChan, errChan)

				for msg := range logChan {
					receivedMessages = append(receivedMessages, msg)

					if len(receivedMessages) == 3 {
						break
					}
				}

				Consistently(errChan).ShouldNot(Receive())

				Expect(receivedMessages).To(Equal([]Loggable{
					NewLoggregatorLogMessage(msg1),
					NewLoggregatorLogMessage(msg2),
					NewLoggregatorLogMessage(msg3),
				}))
			})

			It("flushes remaining log messages and closes the returned channel when Close is called", func() {
				logsRepo.BufferTime = 10 * time.Second

				msg3 := makeLogMessage("hello3", 300)
				msg2 := makeLogMessage("hello2", 200)
				msg1 := makeLogMessage("hello1", 100)

				fakeConsumer.TailStub = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					messageChan := make(chan *logmessage.LogMessage)
					go func() {
						messageChan <- msg3
						messageChan <- msg2
						messageChan <- msg1
						close(messageChan)
					}()

					return messageChan, nil
				}

				Expect(fakeConsumer.CloseCallCount()).To(Equal(0))

				logsRepo.TailLogsFor("app-guid", func() {}, logChan, errChan)
				Consistently(errChan).ShouldNot(Receive())

				logsRepo.Close()

				Expect(fakeConsumer.CloseCallCount()).To(Equal(1))

				Eventually(logChan).Should(Receive(Equal(NewLoggregatorLogMessage(msg1))))
				Eventually(logChan).Should(Receive(Equal(NewLoggregatorLogMessage(msg2))))
				Eventually(logChan).Should(Receive(Equal(NewLoggregatorLogMessage(msg3))))
			})
		})
	})
})

func makeLogMessage(message string, timestamp int64) *logmessage.LogMessage {
	messageType := logmessage.LogMessage_OUT
	sourceName := "DEA"
	return &logmessage.LogMessage{
		Message:     []byte(message),
		AppId:       proto.String("my-app-guid"),
		MessageType: &messageType,
		SourceName:  &sourceName,
		Timestamp:   proto.Int64(timestamp),
	}
}
