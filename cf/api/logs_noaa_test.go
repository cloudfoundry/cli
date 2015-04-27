package api_test

import (
	"errors"
	"reflect"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	noaa_errors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/noaa/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("logs with noaa repository", func() {
	var (
		fakeNoaaConsumer   *testapi.FakeNoaaConsumer
		config             core_config.ReadWriter
		fakeTokenRefresher *testapi.FakeAuthenticationRepository
		repo               api.LogsNoaaRepository
	)

	BeforeEach(func() {
		fakeNoaaConsumer = &testapi.FakeNoaaConsumer{}
		config = testconfig.NewRepositoryWithDefaults()
		config.SetLoggregatorEndpoint("loggregator.test.com")
		config.SetDopplerEndpoint("doppler.test.com")
		config.SetAccessToken("the-access-token")
		fakeTokenRefresher = &testapi.FakeAuthenticationRepository{}
		repo = api.NewLogsNoaaRepository(config, fakeNoaaConsumer, fakeTokenRefresher)
	})

	Describe("RecentLogsFor", func() {

		It("refreshes token and get metric once more if token has expired.", func() {
			fakeNoaaConsumer.RecentLogsReturns([]*events.LogMessage{},
				noaa_errors.NewUnauthorizedError("Unauthorized token"))

			repo.RecentLogsFor("app-guid")
			Ω(fakeTokenRefresher.RefreshTokenCalled).To(BeTrue())
			Ω(fakeNoaaConsumer.RecentLogsCallCount()).To(Equal(2))
		})

		It("refreshes token and get metric once more if token has expired.", func() {
			fakeNoaaConsumer.RecentLogsReturns([]*events.LogMessage{}, errors.New("error error error"))

			_, err := repo.RecentLogsFor("app-guid")
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("error error error"))
		})

		Context("when an error does not occur", func() {
			BeforeEach(func() {
				l := []*events.LogMessage{
					&events.LogMessage{Message: []byte("message 3"), Timestamp: proto.Int64(3000), AppId: proto.String("app-guid-1")},
					&events.LogMessage{Message: []byte("message 2"), Timestamp: proto.Int64(2000), AppId: proto.String("app-guid-1")},
					&events.LogMessage{Message: []byte("message 1"), Timestamp: proto.Int64(1000), AppId: proto.String("app-guid-1")},
				}

				fakeNoaaConsumer.RecentLogsReturns(l, nil)
			})

			It("gets the logs for the requested app", func() {
				repo.RecentLogsFor("app-guid-1")
				arg, _ := fakeNoaaConsumer.RecentLogsArgsForCall(0)
				Expect(arg).To(Equal("app-guid-1"))
			})

			It("returns the sorted log messages", func() {
				messages, err := repo.RecentLogsFor("app-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(string(messages[0].Message)).To(Equal("message 1"))
				Expect(string(messages[1].Message)).To(Equal("message 2"))
				Expect(string(messages[2].Message)).To(Equal("message 3"))
			})
		})
	})

	Describe("tailing logs", func() {

		Context("when an error occurs", func() {
			It("returns an error when it occurs", func() {
				fakeNoaaConsumer.TailFunc = func(logChan chan<- *events.LogMessage, errChan chan<- error, stopChan chan struct{}) {
					errChan <- errors.New("oops")
				}

				err := repo.TailNoaaLogsFor("app-guid", func() {}, func(*events.LogMessage) {})
				Expect(err).To(Equal(errors.New("oops")))
			})
		})

		Context("when a noaa_errors.UnauthorizedError occurs", func() {
			It("refreshes the access token and tail logs once more", func() {
				calledOnce := false
				fakeNoaaConsumer.TailFunc = func(logChan chan<- *events.LogMessage, errChan chan<- error, stopChan chan struct{}) {
					if !calledOnce {
						calledOnce = true
						errChan <- noaa_errors.NewUnauthorizedError("i'm sorry dave")
					} else {
						errChan <- errors.New("2nd Error")
					}
				}

				err := repo.TailNoaaLogsFor("app-guid", func() {}, func(*events.LogMessage) {})
				Ω(fakeTokenRefresher.RefreshTokenCalled).To(BeTrue())
				Ω(err.Error()).To(Equal("2nd Error"))
			})
		})

		Context("when no error occurs", func() {
			It("asks for the logs for the given app", func() {
				fakeNoaaConsumer.TailFunc = func(logChan chan<- *events.LogMessage, errChan chan<- error, stopChan chan struct{}) {
					errChan <- errors.New("quit Tailing")
				}

				repo.TailNoaaLogsFor("app-guid", func() {}, func(msg *events.LogMessage) {})

				appGuid, token, _, _, _ := fakeNoaaConsumer.TailingLogsArgsForCall(0)
				Ω(appGuid).To(Equal("app-guid"))
				Ω(token).To(Equal("the-access-token"))
			})

			It("sets the on connect callback", func() {
				fakeNoaaConsumer.TailFunc = func(logChan chan<- *events.LogMessage, errChan chan<- error, stopChan chan struct{}) {
					errChan <- errors.New("quit Tailing")
				}

				var cb = func() { return }
				repo.TailNoaaLogsFor("app-guid", cb, func(msg *events.LogMessage) {})

				Ω(fakeNoaaConsumer.SetOnConnectCallbackCallCount()).To(Equal(1))
				arg := fakeNoaaConsumer.SetOnConnectCallbackArgsForCall(0)
				Ω(reflect.ValueOf(arg).Pointer() == reflect.ValueOf(cb).Pointer()).To(BeTrue())
			})
		})

		Context("and the buffer time is sufficient for sorting", func() {
			BeforeEach(func() {
				api.BufferTime = 250 * time.Millisecond
				repo = api.NewLogsNoaaRepository(config, fakeNoaaConsumer, fakeTokenRefresher)
			})

			It("sorts the messages before yielding them", func() {
				fakeNoaaConsumer.TailFunc = func(logChan chan<- *events.LogMessage, errChan chan<- error, stopChan chan struct{}) {
					logChan <- makeLogMessage("hello3", 300)
					logChan <- makeLogMessage("hello2", 200)
					logChan <- makeLogMessage("hello1", 100)
					time.Sleep(250 * time.Millisecond)
				}

				receivedMessages := []*events.LogMessage{}
				repo.TailNoaaLogsFor("app-guid", func() {}, func(msg *events.LogMessage) {
					receivedMessages = append(receivedMessages, msg)
					if len(receivedMessages) >= 3 {
						repo.Close()
					}
				})

				Expect(receivedMessages).To(Equal([]*events.LogMessage{
					makeLogMessage("hello1", 100),
					makeLogMessage("hello2", 200),
					makeLogMessage("hello3", 300),
				}))

			})
		})

		Context("and the buffer time is very long", func() {
			BeforeEach(func() {
				api.BufferTime = 30 * time.Second
				repo = api.NewLogsNoaaRepository(config, fakeNoaaConsumer, fakeTokenRefresher)
			})

			It("flushes remaining log messages when Close is called", func() {
				fakeNoaaConsumer.TailFunc = func(logChan chan<- *events.LogMessage, errChan chan<- error, stopChan chan struct{}) {
					logChan <- makeLogMessage("hello3", 300)
					logChan <- makeLogMessage("hello2", 200)
					logChan <- makeLogMessage("hello1", 100)
				}

				receivedMessages := []*events.LogMessage{}

				go func() {
					time.Sleep(500 * time.Millisecond)
					repo.Close()
				}()

				repo.TailNoaaLogsFor("app-guid", func() {}, func(msg *events.LogMessage) {
					receivedMessages = append(receivedMessages, msg)
				})

				Expect(receivedMessages).To(Equal([]*events.LogMessage{
					makeLogMessage("hello1", 100),
					makeLogMessage("hello2", 200),
					makeLogMessage("hello3", 300),
				}))

			})
		})
	})

	Describe("GetContainerMetrics()", func() {

		var (
			fakeNoaaConsumer   *testapi.FakeNoaaConsumer
			config             core_config.ReadWriter
			fakeTokenRefresher *testapi.FakeAuthenticationRepository
			repo               api.LogsNoaaRepository
		)

		BeforeEach(func() {
			fakeNoaaConsumer = &testapi.FakeNoaaConsumer{}
			config = testconfig.NewRepositoryWithDefaults()
			fakeTokenRefresher = &testapi.FakeAuthenticationRepository{}
			repo = api.NewLogsNoaaRepository(config, fakeNoaaConsumer, fakeTokenRefresher)
		})

		It("populates metrics for an app instance", func() {
			var (
				i    int32   = 2
				cpu  float64 = 50
				mem  uint64  = 128
				disk uint64  = 256
				err  error
			)

			fakeNoaaConsumer.GetContainerMetricsReturns([]*events.ContainerMetric{
				&events.ContainerMetric{
					InstanceIndex: &i,
					CpuPercentage: &cpu,
					MemoryBytes:   &mem,
					DiskBytes:     &disk,
				},
			}, nil)

			instances := []models.AppInstanceFields{
				models.AppInstanceFields{},
				models.AppInstanceFields{},
				models.AppInstanceFields{},
			}

			instances, err = repo.GetContainerMetrics("app-guid", instances)
			Ω(err).ToNot(HaveOccurred())
			Ω(instances[0].CpuUsage).To(Equal(float64(0)))
			Ω(instances[1].CpuUsage).To(Equal(float64(0)))
			Ω(instances[2].CpuUsage).To(Equal(cpu))
			Ω(instances[2].MemUsage).To(Equal(int64(mem)))
			Ω(instances[2].DiskUsage).To(Equal(int64(disk)))
		})

		It("refreshes token and get metric once more if token has expired.", func() {
			fakeNoaaConsumer.GetContainerMetricsReturns([]*events.ContainerMetric{},
				noaa_errors.NewUnauthorizedError("Unauthorized token"))

			instances := []models.AppInstanceFields{
				models.AppInstanceFields{},
			}

			instances, _ = repo.GetContainerMetrics("app-guid", instances)
			Ω(fakeTokenRefresher.RefreshTokenCalled).To(Equal(true))
			Ω(fakeNoaaConsumer.GetContainerMetricsCallCount()).To(Equal(2))
		})

	})

})

func makeLogMessage(message string, timestamp int64) *events.LogMessage {
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
