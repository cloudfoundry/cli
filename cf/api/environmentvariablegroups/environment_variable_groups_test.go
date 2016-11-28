package environmentvariablegroups_test

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/cf/api/environmentvariablegroups"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"

	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"

	"github.com/onsi/gomega/ghttp"

	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudControllerRepository", func() {
	var (
		ccServer   *ghttp.Server
		configRepo coreconfig.ReadWriter
		repo       environmentvariablegroups.CloudControllerRepository
	)

	BeforeEach(func() {
		ccServer = ghttp.NewServer()
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetAPIEndpoint(ccServer.URL())
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		repo = environmentvariablegroups.NewCloudControllerRepository(configRepo, gateway)
	})

	AfterEach(func() {
		ccServer.Close()
	})

	Describe("ListRunning", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/config/environment_variable_groups/running"),
					ghttp.RespondWith(http.StatusOK, `{ "abc": 123, "do-re-mi": "fa-sol-la-ti" }`),
				),
			)
		})

		It("lists the environment variables in the running group", func() {
			envVars, err := repo.ListRunning()
			Expect(err).NotTo(HaveOccurred())

			Expect(envVars).To(ConsistOf([]models.EnvironmentVariable{
				{Name: "abc", Value: "123"},
				{Name: "do-re-mi", Value: "fa-sol-la-ti"},
			}))
		})
	})

	Describe("ListStaging", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/config/environment_variable_groups/staging"),
					ghttp.RespondWith(http.StatusOK, `{ "abc": 123, "do-re-mi": "fa-sol-la-ti" }`),
				),
			)
		})

		It("lists the environment variables in the staging group", func() {
			envVars, err := repo.ListStaging()
			Expect(err).NotTo(HaveOccurred())
			Expect(envVars).To(ConsistOf([]models.EnvironmentVariable{
				{Name: "abc", Value: "123"},
				{Name: "do-re-mi", Value: "fa-sol-la-ti"},
			}))
		})
	})

	Describe("SetStaging", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/v2/config/environment_variable_groups/staging"),
					ghttp.VerifyJSON(`{ "abc": "one-two-three", "def": 456 }`),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		})

		It("sets the environment variables in the staging group", func() {
			err := repo.SetStaging(`{"abc": "one-two-three", "def": 456}`)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})
	})

	Describe("SetRunning", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/v2/config/environment_variable_groups/running"),
					ghttp.VerifyJSON(`{ "abc": "one-two-three", "def": 456 }`),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		})

		It("sets the environment variables in the running group", func() {
			err := repo.SetRunning(`{"abc": "one-two-three", "def": 456}`)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})
	})
})
