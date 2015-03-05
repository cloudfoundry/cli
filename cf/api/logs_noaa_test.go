package api_test

import (
	"github.com/cloudfoundry/cli/cf/api"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/loggregator_consumer/noaa_errors"
	"github.com/cloudfoundry/noaa/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("logs with noaa repository", func() {

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
