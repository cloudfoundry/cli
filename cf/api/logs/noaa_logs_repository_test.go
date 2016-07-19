package logs_test

import (
	"errors"
	"reflect"
	"time"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	noaa_errors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	"code.cloudfoundry.org/cli/cf/api/authentication/authenticationfakes"
	testapi "code.cloudfoundry.org/cli/cf/api/logs/logsfakes"
	testconfig "code.cloudfoundry.org/cli/testhelpers/configuration"

	"sync"

	"code.cloudfoundry.org/cli/cf/api/logs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("logs with noaa repository", func() {
	var (
		fakeNoaaConsumer   *testapi.FakeNoaaConsumer
		config             coreconfig.ReadWriter
		fakeTokenRefresher *authenticationfakes.FakeRepository
		repo               *logs.NoaaLogsRepository
	)

	BeforeEach(func() {
		fakeNoaaConsumer = &testapi.FakeNoaaConsumer{}
		config = testconfig.NewRepositoryWithDefaults()
		config.SetLoggregatorEndpoint("loggregator.test.com")
		config.SetDopplerEndpoint("doppler.test.com")
		config.SetAccessToken("the-access-token")
		fakeTokenRefresher = &authenticationfakes.FakeRepository{}
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

	Describe("TailLogsFor", func() {
		var errChan chan error
		var logChan chan logs.Loggable

		AfterEach(func() {
			Eventually(errChan).Should(BeClosed())
			Eventually(logChan).Should(BeClosed())
		})

		Context("when an error occurs", func() {
			var e chan error
			var c chan *events.LogMessage

			BeforeEach(func() {
				errChan = make(chan error)
				logChan = make(chan logs.Loggable)

				e = make(chan error)
				c = make(chan *events.LogMessage)

				fakeNoaaConsumer.CloseStub = func() error {
					close(e)
					close(c)
					return nil
				}
			})

			It("returns an error when it occurs", func(done Done) {
				defer repo.Close()
				err := errors.New("oops")

				fakeNoaaConsumer.TailingLogsWithoutReconnectStub = func(appGuid string, authToken string) (<-chan *events.LogMessage, <-chan error) {
					go func() {
						e <- err
					}()
					return c, e
				}
				go repo.TailLogsFor("app-guid", func() {}, logChan, errChan)

				Eventually(errChan).Should(Receive(&err))

				close(done)
			})
		})

		Context("when a noaa_errors.UnauthorizedError occurs", func() {
			var e chan error
			var c chan *events.LogMessage

			BeforeEach(func() {
				errChan = make(chan error)
				logChan = make(chan logs.Loggable)

				e = make(chan error)
				c = make(chan *events.LogMessage)

				fakeNoaaConsumer.CloseStub = func() error {
					close(e)
					close(c)
					return nil
				}
			})

			It("refreshes the access token and tail logs once more", func(done Done) {
				defer repo.Close()
				calledOnce := false
				err := errors.New("2nd Error")
				synchronization := make(chan bool)

				fakeNoaaConsumer.CloseStub = func() error {
					synchronization <- true
					return nil
				}

				fakeNoaaConsumer.TailingLogsWithoutReconnectStub = func(appGuid string, authToken string) (<-chan *events.LogMessage, <-chan error) {
					ec := make(chan error)
					lc := make(chan *events.LogMessage)

					go func() {
						if !calledOnce {
							calledOnce = true
							ec <- noaa_errors.NewUnauthorizedError("i'm sorry dave")
						} else {
							ec <- err
							<-synchronization
							close(ec)
							close(lc)
						}
					}()

					return lc, ec
				}

				go repo.TailLogsFor("app-guid", func() {}, logChan, errChan)

				Eventually(errChan).Should(Receive(&err))
				Eventually(fakeTokenRefresher.RefreshAuthTokenCallCount).Should(Equal(1))

				close(done)
			})
		})

		Context("when no error occurs", func() {
			var e chan error
			var c chan *events.LogMessage

			BeforeEach(func() {
				errChan = make(chan error)
				logChan = make(chan logs.Loggable)

				e = make(chan error)
				c = make(chan *events.LogMessage)

				fakeNoaaConsumer.CloseStub = func() error {
					close(e)
					close(c)
					return nil
				}
			})

			It("asks for the logs for the given app", func(done Done) {
				defer repo.Close()

				fakeNoaaConsumer.TailingLogsWithoutReconnectReturns(c, e)

				repo.TailLogsFor("app-guid", func() {}, logChan, errChan)

				Eventually(fakeNoaaConsumer.TailingLogsWithoutReconnectCallCount).Should(Equal(1))
				appGuid, token := fakeNoaaConsumer.TailingLogsWithoutReconnectArgsForCall(0)
				Expect(appGuid).To(Equal("app-guid"))
				Expect(token).To(Equal("the-access-token"))

				close(done)
			}, 2)

			It("sets the on connect callback", func() {
				defer repo.Close()

				fakeNoaaConsumer.TailingLogsWithoutReconnectReturns(c, e)

				var cb = func() { return }
				repo.TailLogsFor("app-guid", cb, logChan, errChan)

				Expect(fakeNoaaConsumer.SetOnConnectCallbackCallCount()).To(Equal(1))
				arg := fakeNoaaConsumer.SetOnConnectCallbackArgsForCall(0)
				Expect(reflect.ValueOf(arg).Pointer() == reflect.ValueOf(cb).Pointer()).To(BeTrue())
			})
		})

		Context("and the buffer time is sufficient for sorting", func() {
			var msg1, msg2, msg3 *events.LogMessage
			var ec chan error
			var lc chan *events.LogMessage
			var syncMu sync.Mutex

			BeforeEach(func() {
				msg1 = makeNoaaLogMessage("hello1", 100)
				msg2 = makeNoaaLogMessage("hello2", 200)
				msg3 = makeNoaaLogMessage("hello3", 300)

				errChan = make(chan error)
				logChan = make(chan logs.Loggable)
				ec = make(chan error)

				syncMu.Lock()
				lc = make(chan *events.LogMessage)
				syncMu.Unlock()

				fakeNoaaConsumer.TailingLogsWithoutReconnectStub = func(string, string) (<-chan *events.LogMessage, <-chan error) {
					go func() {
						syncMu.Lock()
						lc <- msg3
						lc <- msg2
						lc <- msg1
						syncMu.Unlock()
					}()

					return lc, ec
				}
			})

			JustBeforeEach(func() {
				repo = logs.NewNoaaLogsRepository(config, fakeNoaaConsumer, fakeTokenRefresher)

				fakeNoaaConsumer.CloseStub = func() error {
					syncMu.Lock()
					close(lc)
					syncMu.Unlock()
					close(ec)

					return nil
				}
			})

			Context("when the channels are closed before reading", func() {
				It("sorts the messages before yielding them", func(done Done) {
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
					close(done)
				})
			})

			Context("when the channels are read while being written to", func() {
				It("sorts the messages before yielding them", func(done Done) {
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

					close(done)
				})

				It("flushes remaining log messages when Close is called", func() {
					repo.BufferTime = 10 * time.Second

					repo.TailLogsFor("app-guid", func() {}, logChan, errChan)
					Consistently(errChan).ShouldNot(Receive())
					Consistently(logChan).ShouldNot(Receive())

					repo.Close()

					Eventually(logChan).Should(Receive(Equal(logs.NewNoaaLogMessage(msg1))))
					Eventually(logChan).Should(Receive(Equal(logs.NewNoaaLogMessage(msg2))))
					Eventually(logChan).Should(Receive(Equal(logs.NewNoaaLogMessage(msg3))))
				})
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
