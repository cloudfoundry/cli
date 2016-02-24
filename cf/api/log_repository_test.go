package api_test

import (
	authenticationfakes "github.com/cloudfoundry/cli/cf/api/authentication/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	noaa_errors "github.com/cloudfoundry/noaa/errors"
	"github.com/gogo/protobuf/proto"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("loggregator logs repository", func() {
	var (
		fakeConsumer *testapi.FakeLoggregatorConsumer
		logsRepo     LogsRepository
		configRepo   core_config.ReadWriter
		authRepo     *authenticationfakes.FakeAuthenticationRepository
	)

	BeforeEach(func() {
		fakeConsumer = testapi.NewFakeLoggregatorConsumer()
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetLoggregatorEndpoint("loggregator-server.test.com")
		configRepo.SetAccessToken("the-access-token")
		authRepo = &authenticationfakes.FakeAuthenticationRepository{}
	})

	JustBeforeEach(func() {
		logsRepo = NewLoggregatorLogsRepository(configRepo, fakeConsumer, authRepo)
	})

	Describe("RecentLogsFor", func() {
		Context("when a noaa_errors.UnauthorizedError occurs", func() {
			BeforeEach(func() {
				fakeConsumer.RecentReturns.Err = []error{
					noaa_errors.NewUnauthorizedError("i'm sorry dave"),
					nil,
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
				_, err := logsRepo.TailLogsFor("app-guid", func() {})
				Expect(err).To(Equal(errors.New("oops")))
			})
		})

		Context("when a LoggregatorConsumer.UnauthorizedError occurs", func() {

			It("refreshes the access token", func(done Done) {
				calledOnce := false
				fakeConsumer.TailFunc = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					if !calledOnce {
						calledOnce = true
						return nil, noaa_errors.NewUnauthorizedError("i'm sorry dave")
					} else {
						close(done)
						return nil, nil
					}
				}

				_, err := logsRepo.TailLogsFor("app-guid", func() {})
				Expect(err).ToNot(HaveOccurred())
				Expect(authRepo.RefreshAuthTokenCallCount()).To(Equal(1))
			})

			Context("when LoggregatorConsumer.UnauthorizedError occurs again", func() {
				It("returns an error", func(done Done) {
					fakeConsumer.TailFunc = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
						return nil, noaa_errors.NewUnauthorizedError("All the errors")
					}

					_, err := logsRepo.TailLogsFor("app-guid", func() {})
					Expect(err).To(HaveOccurred())
					close(done)
				})
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

				logsRepo.TailLogsFor("app-guid", func() {})
			})

			It("sets the on connect callback", func(done Done) {
				fakeConsumer.TailFunc = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					close(done)
					return nil, nil
				}

				called := false
				logsRepo.TailLogsFor("app-guid", func() { called = true })
				fakeConsumer.OnConnectCallback()
				Expect(called).To(BeTrue())
			})

			It("sorts the messages before yielding them", func(done Done) {
				var receivedMessages []*logmessage.LogMessage
				msg3 := makeLogMessage("hello3", 300)
				msg2 := makeLogMessage("hello2", 200)
				msg1 := makeLogMessage("hello1", 100)

				fakeConsumer.TailFunc = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					logChan := make(chan *logmessage.LogMessage)
					go func() {
						logChan <- msg3
						logChan <- msg2
						logChan <- msg1
						fakeConsumer.WaitForClose()
						close(logChan)
					}()

					return logChan, nil
				}

				c, err := logsRepo.TailLogsFor("app-guid", func() {})

				for msg := range c {
					receivedMessages = append(receivedMessages, msg)
					if len(receivedMessages) >= 3 {
						logsRepo.Close()
					}
				}

				Expect(err).NotTo(HaveOccurred())

				Expect(receivedMessages).To(Equal([]*logmessage.LogMessage{
					msg1,
					msg2,
					msg3,
				}))

				close(done)
			})

			It("flushes remaining log messages and closes the returned channel when Close is called", func() {
				synchronizationChannel := make(chan (bool))
				var receivedMessages []*logmessage.LogMessage
				msg3 := makeLogMessage("hello3", 300)
				msg2 := makeLogMessage("hello2", 200)
				msg1 := makeLogMessage("hello1", 100)

				fakeConsumer.TailFunc = func(_, _ string) (<-chan *logmessage.LogMessage, error) {
					logChan := make(chan *logmessage.LogMessage)
					go func() {
						logChan <- msg3
						logChan <- msg2
						logChan <- msg1
						fakeConsumer.WaitForClose()
						close(logChan)
					}()

					return logChan, nil
				}

				Expect(fakeConsumer.IsClosed).To(BeFalse())

				channel, err := logsRepo.TailLogsFor("app-guid", func() {})
				Expect(err).NotTo(HaveOccurred())
				Expect(channel).NotTo(BeClosed())

				go func() {
					for msg := range channel {
						receivedMessages = append(receivedMessages, msg)
					}

					synchronizationChannel <- true
				}()

				logsRepo.Close()

				Expect(fakeConsumer.IsClosed).To(BeTrue())

				<-synchronizationChannel

				Expect(channel).To(BeClosed())

				Expect(receivedMessages).To(Equal([]*logmessage.LogMessage{
					msg1,
					msg2,
					msg3,
				}))
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
