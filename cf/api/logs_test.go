package api_test

import (
	"code.google.com/p/gogoprotobuf/proto"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/loggregator_consumer"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("loggregator logs repository", func() {
	var (
		fakeConsumer       *testapi.FakeLoggregatorConsumer
		logsRepo           LogsRepository
		configRepo         configuration.ReadWriter
		fakeTokenRefresher *testapi.FakeAuthenticationRepository
	)

	BeforeEach(func() {
		BufferTime = 1 * time.Millisecond
		fakeConsumer = testapi.NewFakeLoggregatorConsumer()
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetLoggregatorEndpoint("loggregator-server.test.com")
		configRepo.SetAccessToken("the-access-token")
		fakeTokenRefresher = &testapi.FakeAuthenticationRepository{}
	})

	JustBeforeEach(func() {
		logsRepo = NewLoggregatorLogsRepository(configRepo, fakeConsumer, fakeTokenRefresher)
	})

	Describe("RecentLogsFor", func() {
		Context("when a LoggregatorConsumer.UnauthorizedError occurs", func() {
			BeforeEach(func() {
				fakeConsumer.RecentReturns.Err = []error{
					loggregator_consumer.NewUnauthorizedError("i'm sorry dave"),
					nil,
				}
			})

			It("refreshes the access token", func() {
				_, err := logsRepo.RecentLogsFor("app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeTokenRefresher.RefreshTokenCalled).To(BeTrue())
			})
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				fakeConsumer.RecentReturns.Err = []error{errors.New("oops")}
			})

			It("returns the error", func() {
				_, err := logsRepo.RecentLogsFor("app-guid")
				Expect(err).To(Equal(errors.New("oops")))
			})
		})

		Context("when an error does not occur", func() {
			BeforeEach(func() {
				fakeConsumer.RecentReturns.Messages = []*logmessage.LogMessage{
					makeLogMessage("My message 2", int64(2000)),
					makeLogMessage("My message 1", int64(1000)),
				}
			})

			It("gets the logs for the requested app", func() {
				logsRepo.RecentLogsFor("app-guid")
				Expect(fakeConsumer.RecentCalledWith.AppGuid).To(Equal("app-guid"))
			})

			It("writes the sorted log messages onto the provided channel", func() {
				messages, err := logsRepo.RecentLogsFor("app-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(string(messages[0].Message)).To(Equal("My message 1"))
				Expect(string(messages[1].Message)).To(Equal("My message 2"))
			})
		})
	})

	Describe("tailing logs", func() {
		Context("when an error occurs", func() {
			BeforeEach(func() {
				fakeConsumer.TailFunc = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					return nil, errors.New("oops")
				}
			})

			It("returns an error", func() {
				err := logsRepo.TailLogsFor("app-guid", func() {}, func(*logmessage.LogMessage) {})
				Expect(err).To(Equal(errors.New("oops")))
			})
		})

		Context("when a LoggregatorConsumer.UnauthorizedError occurs", func() {

			It("refreshes the access token", func(done Done) {
				calledOnce := false
				fakeConsumer.TailFunc = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					if !calledOnce {
						calledOnce = true
						return nil, loggregator_consumer.NewUnauthorizedError("i'm sorry dave")
					} else {
						close(done)
						return nil, nil
					}
				}

				err := logsRepo.TailLogsFor("app-guid", func() {}, func(*logmessage.LogMessage) {})
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeTokenRefresher.RefreshTokenCalled).To(BeTrue())
			})
		})

		Context("when no error occurs", func() {
			It("asks for the logs for the given app", func(done Done) {
				fakeConsumer.TailFunc = func(appGuid, token string) (<-chan *logmessage.LogMessage, error) {
					Expect(appGuid).To(Equal("app-guid"))
					Expect(token).To(Equal("the-access-token"))
					close(done)
					return nil, nil
				}

				logsRepo.TailLogsFor("app-guid", func() {}, func(msg *logmessage.LogMessage) {})
			})

			It("sets the on connect callback", func(done Done) {
				fakeConsumer.TailFunc = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					close(done)
					return nil, nil
				}

				called := false
				logsRepo.TailLogsFor("app-guid", func() { called = true }, func(msg *logmessage.LogMessage) {})
				fakeConsumer.OnConnectCallback()
				Expect(called).To(BeTrue())
			})

			Context("and the buffer time is sufficient for sorting", func() {
				BeforeEach(func() {
					BufferTime = 250 * time.Millisecond
				})

				It("sorts the messages before yielding them", func(done Done) {
					fakeConsumer.TailFunc = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
						logChan := make(chan *logmessage.LogMessage)
						go func() {
							logChan <- makeLogMessage("hello3", 300)
							logChan <- makeLogMessage("hello2", 200)
							logChan <- makeLogMessage("hello1", 100)
							fakeConsumer.WaitForClose()
							close(logChan)
						}()

						return logChan, nil
					}

					receivedMessages := []*logmessage.LogMessage{}
					err := logsRepo.TailLogsFor("app-guid", func() {}, func(msg *logmessage.LogMessage) {
						receivedMessages = append(receivedMessages, msg)
						if len(receivedMessages) >= 3 {
							logsRepo.Close()
						}
					})

					Expect(err).NotTo(HaveOccurred())

					Expect(receivedMessages).To(Equal([]*logmessage.LogMessage{
						makeLogMessage("hello1", 100),
						makeLogMessage("hello2", 200),
						makeLogMessage("hello3", 300),
					}))

					close(done)
				})
			})

			Context("and the buffer time is very long", func() {
				BeforeEach(func() {
					BufferTime = 30 * time.Second
				})

				It("flushes remaining log messages when Close is called", func(done Done) {
					synchronizationChannel := make(chan (bool))

					fakeConsumer.TailFunc = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
						fakeConsumer.OnConnectCallback()
						logChan := make(chan *logmessage.LogMessage)
						go func() {
							logChan <- makeLogMessage("One does not simply consume a log message", 1000)
							synchronizationChannel <- true
							fakeConsumer.WaitForClose()
							close(logChan)
						}()

						return logChan, nil
					}

					receivedMessages := []*logmessage.LogMessage{}

					go func() {
						defer GinkgoRecover()

						<-synchronizationChannel

						Expect(receivedMessages).To(BeEmpty())
						logsRepo.Close()
						Expect(receivedMessages).ToNot(BeEmpty())

						done <- true
					}()

					err := logsRepo.TailLogsFor("app-guid", func() {}, func(msg *logmessage.LogMessage) {
						receivedMessages = append(receivedMessages, msg)
					})

					Expect(err).NotTo(HaveOccurred())
				})
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
