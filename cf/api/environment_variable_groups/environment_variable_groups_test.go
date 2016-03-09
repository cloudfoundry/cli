package environment_variable_groups_test

import (
	"net/http"

	"github.com/cloudfoundry/cli/cf/api/environment_variable_groups"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"

	"github.com/cloudfoundry/cli/testhelpers/cloud_controller_gateway"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudControllerEnvironmentVariableGroupsRepository", func() {
	var (
		ccServer   *ghttp.Server
		configRepo core_config.ReadWriter
		repo       environment_variable_groups.CloudControllerEnvironmentVariableGroupsRepository
	)

	BeforeEach(func() {
		ccServer = ghttp.NewServer()
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetApiEndpoint(ccServer.URL())
		gateway := cloud_controller_gateway.NewTestCloudControllerGateway(configRepo)
		repo = environment_variable_groups.NewCloudControllerEnvironmentVariableGroupsRepository(configRepo, gateway)
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
				models.EnvironmentVariable{Name: "abc", Value: "123"},
				models.EnvironmentVariable{Name: "do-re-mi", Value: "fa-sol-la-ti"},
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
				models.EnvironmentVariable{Name: "abc", Value: "123"},
				models.EnvironmentVariable{Name: "do-re-mi", Value: "fa-sol-la-ti"},
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
