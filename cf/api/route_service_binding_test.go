package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouteServiceBindingsRepository", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  core_config.ReadWriter
		repo        CloudControllerRouteServiceBindingRepository
	)

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()

		gateway := net.NewCloudControllerGateway(configRepo, time.Now, &testterm.FakeUI{})
		repo = NewCloudControllerRouteServiceBindingRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("Bind", func() {
		var (
			serviceInstanceGuid string
			routeGuid           string
			upsi                bool
		)

		Context("when the service instance is managed", func() {
			BeforeEach(func() {
				serviceInstanceGuid = "service-instance-guid"
				routeGuid = "route-guid"
				upsi = false
			})

			JustBeforeEach(func() {
				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid),
					Response: testnet.TestResponse{Status: http.StatusCreated},
				}))
			})

			It("creates the service binding", func() {
				apiErr := repo.Bind(serviceInstanceGuid, routeGuid, upsi)

				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

		Context("when the service instance is user provided", func() {
			BeforeEach(func() {
				serviceInstanceGuid = "service-instance-guid"
				routeGuid = "route-guid"
				upsi = true
			})

			JustBeforeEach(func() {
				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "PUT",
					Path:     fmt.Sprintf("/v2/user_provided_service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid),
					Response: testnet.TestResponse{Status: http.StatusCreated},
				}))
			})

			It("creates the service binding", func() {
				apiErr := repo.Bind(serviceInstanceGuid, routeGuid, upsi)

				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

		Context("when an API error occurs", func() {
			BeforeEach(func() {
				serviceInstanceGuid = "service-instance-guid"
				routeGuid = "route-guid"
				upsi = false

				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "PUT",
					Path:   fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid),
					Response: testnet.TestResponse{
						Status: http.StatusBadRequest,
						Body:   `{"code":61003,"description":"Route does not exist"}`,
					},
				}))
			})

			It("returns an error", func() {
				apiErr := repo.Bind(serviceInstanceGuid, routeGuid, upsi)
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(apiErr).To(HaveOccurred())
				Expect(apiErr.(errors.HttpError).ErrorCode()).To(Equal("61003"))
				Expect(apiErr.Error()).To(ContainSubstring("Route does not exist"))
			})
		})
	})

	Describe("Unbind", func() {
		var (
			serviceInstanceGuid string
			routeGuid           string
			upsi                bool
		)

		Context("when the service instance is managed", func() {
			BeforeEach(func() {
				serviceInstanceGuid = "service-instance-guid"
				routeGuid = "route-guid"
				upsi = false
			})

			JustBeforeEach(func() {
				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "DELETE",
					Path:     fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid),
					Response: testnet.TestResponse{Status: http.StatusNoContent},
				}))
			})

			It("deletes the service binding", func() {
				apiErr := repo.Unbind(serviceInstanceGuid, routeGuid, upsi)

				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

		Context("when the service instance is user provided", func() {
			BeforeEach(func() {
				serviceInstanceGuid = "service-instance-guid"
				routeGuid = "route-guid"
				upsi = true
			})

			JustBeforeEach(func() {
				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:   "DELETE",
					Path:     fmt.Sprintf("/v2/user_provided_service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid),
					Response: testnet.TestResponse{Status: http.StatusNoContent},
				}))
			})

			It("deletes the service binding", func() {
				apiErr := repo.Unbind(serviceInstanceGuid, routeGuid, upsi)

				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

		Context("when an API error occurs", func() {
			BeforeEach(func() {
				serviceInstanceGuid = "service-instance-guid"
				routeGuid = "route-guid"
				upsi = false

				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "DELETE",
					Path:   fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid),
					Response: testnet.TestResponse{
						Status: http.StatusBadRequest,
						Body:   `{"code":61003,"description":"Route does not exist"}`,
					},
				}))
			})

			It("returns an error", func() {
				apiErr := repo.Unbind(serviceInstanceGuid, routeGuid, upsi)
				Expect(testHandler).To(HaveAllRequestsCalled())
				Expect(apiErr).To(HaveOccurred())
				Expect(apiErr.(errors.HttpError).ErrorCode()).To(Equal("61003"))
				Expect(apiErr.Error()).To(ContainSubstring("Route does not exist"))
			})
		})
	})
})
