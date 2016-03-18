package logs_test

import (
	"errors"
	"reflect"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	noaa_errors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	authenticationfakes "github.com/cloudfoundry/cli/cf/api/authentication/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/logs/fakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"

	"sync"

	"github.com/cloudfoundry/cli/cf/api/logs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("logs with noaa repository", func() {
	var (
		fakeNoaaConsumer   *testapi.FakeNoaaConsumer
		config             core_config.ReadWriter
		fakeTokenRefresher *authenticationfakes.FakeAuthenticationRepository
		repo               *logs.NoaaLogsRepository
	)

	BeforeEach(func() {
		fakeNoaaConsumer = &testapi.FakeNoaaConsumer{}
		config = testconfig.NewRepositoryWithDefaults()
		config.SetLoggregatorEndpoint("loggregator.test.com")
		config.SetDopplerEndpoint("doppler.test.com")
		config.SetAccessToken("the-access-token")
		fakeTokenRefresher = &authenticationfakes.FakeAuthenticationRepository{}
		repo = logs.NewNoaaLogsRepository(config, fakeNoaaConsumer, fakeTokenRefresher)
	})

	Describe("RecentLogsFor", func() {
		It("refreshes token and get metric once more if token has expired.", func() {
			var recentLogsCallCount int

			fakeNoaaConsumer.RecentLogsStub = func(appGuid, authToken string) ([]*events.LogMessage, error) {
				defer func() {
					recentLogsCallCount += 1
				}()

				if recentLogsCallCount == 0 {
					return []*events.LogMessage{}, noaa_errors.NewUnauthorizedError("Unauthorized token")
				}

				return []*events.LogMessage{}, nil
			}

			repo.RecentLogsFor("app-guid")
			Expect(fakeTokenRefresher.RefreshAuthTokenCallCount()).To(Equal(1))
			Expect(fakeNoaaConsumer.RecentLogsCallCount()).To(Equal(2))
		})

		It("refreshes token and get metric once more if token has expired.", func() {
			fakeNoaaConsumer.RecentLogsReturns([]*events.LogMessage{}, errors.New("error error error"))

			_, err := repo.RecentLogsFor("app-guid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error error error"))
		})

		Context("when an error does not occur", func() {
			var msg1, msg2, msg3 *events.LogMessage

			BeforeEach(func() {
				msg1 = makeNoaaLogMessage("message 1", 1000)
				msg2 = makeNoaaLogMessage("message 2", 2000)
				msg3 = makeNoaaLogMessage("message 3", 3000)

				fakeNoaaConsumer.RecentLogsReturns([]*events.LogMessage{
					msg3,
					msg2,
					msg1,
				}, nil)
			})

			It("gets the logs for the requested app", func() {
				repo.RecentLogsFor("app-guid-1")
				arg, _ := fakeNoaaConsumer.RecentLogsArgsForCall(0)
				Expect(arg).To(Equal("app-guid-1"))
			})

			It("returns the sorted log messages", func() {
				messages, err := repo.RecentLogsFor("app-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(messages).To(Equal([]logs.Loggable{
					logs.NewNoaaLogMessage(msg1),
					logs.NewNoaaLogMessage(msg2),
					logs.NewNoaaLogMessage(msg3),
				}))
			})
		})
	})

	Describe("tailing logs", func() {
		var errChan chan error
		var logChan chan logs.Loggable

		BeforeEach(func() {
			errChan = make(chan error)
			logChan = make(chan logs.Loggable)
		})

		Context("when an error occurs", func() {
			It("returns an error when it occurs", func(done Done) {
				err := errors.New("oops")

				fakeNoaaConsumer.TailingLogsStub = func(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error) {
					errorChan <- err
				}

				go func() {
					defer GinkgoRecover()

					Eventually(errChan).Should(Receive(&err))
					close(done)
				}()

				repo.TailLogsFor("app-guid", func() {}, logChan, errChan)
			})
		})

		Context("when a noaa_errors.UnauthorizedError occurs", func() {
			It("refreshes the access token and tail logs once more", func(done Done) {
				calledOnce := false
				err := errors.New("2nd Error")
				synchronization := make(chan bool)

				fakeNoaaConsumer.CloseStub = func() error {
					synchronization <- true
					return nil
				}

				fakeNoaaConsumer.TailingLogsStub = func(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error) {
					if !calledOnce {
						calledOnce = true
						errorChan <- noaa_errors.NewUnauthorizedError("i'm sorry dave")
					} else {
						errorChan <- err
						<-synchronization
						close(errorChan)
						close(logChan)
					}
				}

				go func() {
					defer GinkgoRecover()

					Eventually(errChan).Should(Receive(&err))
					Eventually(fakeTokenRefresher.RefreshAuthTokenCallCount()).Should(Equal(1))

					close(done)
				}()

				repo.TailLogsFor("app-guid", func() {}, logChan, errChan)
			})
		})

		Context("when no error occurs", func() {
			It("asks for the logs for the given app", func(done Done) {
				fakeNoaaConsumer.TailingLogsStub = func(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error) {
					errorChan <- errors.New("quit Tailing")
				}

				go func() {
					defer GinkgoRecover()

					Eventually(fakeNoaaConsumer.TailingLogsCallCount()).Should(Equal(1))
					appGuid, token, _, _ := fakeNoaaConsumer.TailingLogsArgsForCall(0)
					Expect(appGuid).To(Equal("app-guid"))
					Expect(token).To(Equal("the-access-token"))

					close(done)
				}()

				repo.TailLogsFor("app-guid", func() {}, logChan, errChan)
			})

			It("sets the on connect callback", func() {
				fakeNoaaConsumer.TailingLogsStub = func(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error) {
					errorChan <- errors.New("quit Tailing")
				}

				var cb = func() { return }
				repo.TailLogsFor("app-guid", cb, logChan, errChan)

				Expect(fakeNoaaConsumer.SetOnConnectCallbackCallCount()).To(Equal(1))
				arg := fakeNoaaConsumer.SetOnConnectCallbackArgsForCall(0)
				Expect(reflect.ValueOf(arg).Pointer() == reflect.ValueOf(cb).Pointer()).To(BeTrue())
			})
		})

		Context("and the buffer time is sufficient for sorting", func() {
			var msg1, msg2, msg3 *events.LogMessage

			BeforeEach(func() {
				repo = logs.NewNoaaLogsRepository(config, fakeNoaaConsumer, fakeTokenRefresher)

				closeWait := sync.WaitGroup{}

				msg1 = makeNoaaLogMessage("hello1", 100)
				msg2 = makeNoaaLogMessage("hello2", 200)
				msg3 = makeNoaaLogMessage("hello3", 300)

				fakeNoaaConsumer.TailingLogsStub = func(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error) {
					closeWait.Add(1)

					go func() {
						outputChan <- msg3
						outputChan <- msg2
						outputChan <- msg1

						close(errorChan)
						close(outputChan)
					}()
				}

				fakeNoaaConsumer.CloseStub = func() error {
					closeWait.Done()

					return nil
				}
			})

			It("sorts the messages before yielding them", func() {
				receivedMessages := []logs.Loggable{}

				repo.TailLogsFor("app-guid", func() {}, logChan, errChan)
				Consistently(errChan).ShouldNot(Receive())

				m := <-logChan
				receivedMessages = append(receivedMessages, m)
				m = <-logChan
				receivedMessages = append(receivedMessages, m)
				m = <-logChan
				receivedMessages = append(receivedMessages, m)
				repo.Close()

				Expect(receivedMessages).To(Equal([]logs.Loggable{
					logs.NewNoaaLogMessage(msg1),
					logs.NewNoaaLogMessage(msg2),
					logs.NewNoaaLogMessage(msg3),
				}))
			})

			It("flushes remaining log messages when Close is called", func() {
				repo.BufferTime = 10 * time.Second

				receivedMessages := []logs.Loggable{}

				go func() {
					for msg := range logChan {
						receivedMessages = append(receivedMessages, msg)
					}
				}()

				repo.TailLogsFor("app-guid", func() {}, logChan, errChan)
				Consistently(errChan).ShouldNot(Receive())

				repo.Close()

				getReceivedMessages := func() []logs.Loggable {
					return receivedMessages
				}

				Eventually(getReceivedMessages).Should(Equal([]logs.Loggable{
					logs.NewNoaaLogMessage(msg1),
					logs.NewNoaaLogMessage(msg2),
					logs.NewNoaaLogMessage(msg3),
				}))
			})
		})
	})
})

func makeNoaaLogMessage(message string, timestamp int64) *events.LogMessage {
	messageType := events.LogMessage_OUT
	sourceName := "DEA"
	return &events.LogMessage{
		Message:     []byte(message),
		AppId:       proto.String("app-guid"),
		MessageType: &messageType,
		SourceType:  &sourceName,
		Timestamp:   proto.Int64(timestamp),
	}
}
