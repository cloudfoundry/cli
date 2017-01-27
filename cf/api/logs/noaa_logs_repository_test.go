package logs_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	noaaerrors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	"code.cloudfoundry.org/cli/cf/api/authentication/authenticationfakes"
	testapi "code.cloudfoundry.org/cli/cf/api/logs/logsfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"

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
		retryTimeout       time.Duration
		repo               *logs.NoaaLogsRepository
	)

	BeforeEach(func() {
		fakeNoaaConsumer = &testapi.FakeNoaaConsumer{}
		config = testconfig.NewRepositoryWithDefaults()
		config.SetDopplerEndpoint("doppler.test.com")
		config.SetAccessToken("the-access-token")
		fakeTokenRefresher = &authenticationfakes.FakeRepository{}
		retryTimeout = time.Second + 500*time.Millisecond
		repo = logs.NewNoaaLogsRepository(config, fakeNoaaConsumer, fakeTokenRefresher, retryTimeout)
	})

	Describe("Authentication Token Refresh", func() {
		It("sets the noaa token refresher", func() {
			Expect(fakeNoaaConsumer.RefreshTokenFromCallCount()).To(Equal(1))
			Expect(fakeNoaaConsumer.RefreshTokenFromArgsForCall(0)).To(Equal(fakeTokenRefresher))
		})
	})

	Describe("RecentLogsFor", func() {
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
			var (
				e       chan error
				c       chan *events.LogMessage
				closeWg *sync.WaitGroup
			)

			BeforeEach(func() {
				closeWg = new(sync.WaitGroup)
				errChan = make(chan error)
				logChan = make(chan logs.Loggable)

				e = make(chan error, 1)
				c = make(chan *events.LogMessage)

				closeWg.Add(1)
				fakeNoaaConsumer.CloseStub = func() error {
					defer closeWg.Done()
					close(e)
					close(c)
					return nil
				}
			})

			AfterEach(func() {
				closeWg.Wait()
			})

			It("returns an error when it occurs", func(done Done) {
				defer repo.Close()
				err := errors.New("oops")

				fakeNoaaConsumer.TailingLogsStub = func(appGuid string, authToken string) (<-chan *events.LogMessage, <-chan error) {
					e <- err
					return c, e
				}

				var wg sync.WaitGroup
				wg.Add(1)
				defer wg.Wait()
				go func() {
					defer wg.Done()
					repo.TailLogsFor("app-guid", func() {}, logChan, errChan)
				}()

				Eventually(errChan).Should(Receive(&err))

				close(done)
			})

			It("does not return a RetryError before RetryTimeout", func(done Done) {
				defer repo.Close()
				err := noaaerrors.NewRetryError(errors.New("oops"))

				fakeNoaaConsumer.TailingLogsStub = func(appGuid string, authToken string) (<-chan *events.LogMessage, <-chan error) {
					e <- err
					return c, e
				}

				var wg sync.WaitGroup
				wg.Add(1)
				defer wg.Wait()
				go func() {
					defer wg.Done()
					repo.TailLogsFor("app-guid", func() {}, logChan, errChan)
				}()

				Consistently(errChan).ShouldNot(Receive())
				close(done)
			})

			It("returns a RetryError if no data is received before RetryTimeout", func() {
				defer repo.Close()
				err := noaaerrors.NewRetryError(errors.New("oops"))

				fakeNoaaConsumer.TailingLogsStub = func(appGuid string, authToken string) (<-chan *events.LogMessage, <-chan error) {
					e <- err
					return c, e
				}

				var wg sync.WaitGroup
				wg.Add(1)
				defer wg.Wait()
				go func() {
					defer wg.Done()
					repo.TailLogsFor("app-guid", func() {}, logChan, errChan)
				}()

				Consistently(errChan, time.Second).ShouldNot(Receive())
				expectedErr := errors.New("Timed out waiting for connection to Loggregator (doppler.test.com).")
				Eventually(errChan, time.Second).Should(Receive(Equal(expectedErr)))
			})

			It("Resets the retry timeout after a successful reconnection", func() {
				defer repo.Close()
				err := noaaerrors.NewRetryError(errors.New("oops"))

				fakeNoaaConsumer.TailingLogsStub = func(appGuid string, authToken string) (<-chan *events.LogMessage, <-chan error) {
					e <- err
					return c, e
				}

				var wg sync.WaitGroup
				wg.Add(1)
				defer wg.Wait()
				go func() {
					defer wg.Done()
					repo.TailLogsFor("app-guid", func() {}, logChan, errChan)
				}()

				Consistently(errChan, time.Second).ShouldNot(Receive())
				fakeNoaaConsumer.SetOnConnectCallbackArgsForCall(0)()

				c <- makeNoaaLogMessage("foo", 100)
				Eventually(logChan).Should(Receive())
				Consistently(errChan, time.Second).ShouldNot(Receive())

				e <- err
				expectedErr := errors.New("Timed out waiting for connection to Loggregator (doppler.test.com).")
				Eventually(errChan, 2*time.Second).Should(Receive(Equal(expectedErr)))
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

				fakeNoaaConsumer.TailingLogsReturns(c, e)

				repo.TailLogsFor("app-guid", func() {}, logChan, errChan)

				Eventually(fakeNoaaConsumer.TailingLogsCallCount).Should(Equal(1))
				appGuid, token := fakeNoaaConsumer.TailingLogsArgsForCall(0)
				Expect(appGuid).To(Equal("app-guid"))
				Expect(token).To(Equal("the-access-token"))

				close(done)
			}, 2)

			It("sets the on connect callback", func() {
				defer repo.Close()

				fakeNoaaConsumer.TailingLogsReturns(c, e)

				callbackCalled := make(chan struct{})
				var cb = func() {
					close(callbackCalled)
					return
				}
				repo.TailLogsFor("app-guid", cb, logChan, errChan)

				Expect(fakeNoaaConsumer.SetOnConnectCallbackCallCount()).To(Equal(1))
				arg := fakeNoaaConsumer.SetOnConnectCallbackArgsForCall(0)
				arg()
				Expect(callbackCalled).To(BeClosed())
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

				fakeNoaaConsumer.TailingLogsStub = func(string, string) (<-chan *events.LogMessage, <-chan error) {
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
				repo = logs.NewNoaaLogsRepository(config, fakeNoaaConsumer, fakeTokenRefresher, retryTimeout)

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
