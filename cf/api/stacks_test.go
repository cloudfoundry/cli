package api_test

import (
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("StacksRepo", func() {
	var (
		testServer  *httptest.Server
		testHandler *testnet.TestHandler
		configRepo  configuration.ReadWriter
		repo        StackRepository
	)

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, testHandler = testnet.NewServer(reqs)
		configRepo.SetApiEndpoint(testServer.URL)
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetAccessToken("BEARER my_access_token")

		gateway := net.NewCloudControllerGateway(configRepo)
		repo = NewCloudControllerStackRepository(configRepo, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("FindByName", func() {
		Context("when a stack exists", func() {
			BeforeEach(func() {
				setupTestServer(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/stacks?q=name%3Alinux",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body: `
				{
					"resources": [
						{
						  "metadata": { "guid": "custom-linux-guid" },
						  "entity": { "name": "custom-linux" }
						}
					]
				}`}})
			})

			It("finds the stack", func() {
				stack, err := repo.FindByName("linux")

				Expect(testHandler).To(testnet.HaveAllRequestsCalled())
				Expect(err).NotTo(HaveOccurred())
				Expect(stack).To(Equal(models.Stack{
					Name: "custom-linux",
					Guid: "custom-linux-guid",
				}))
			})
		})

		Context("when a stack does not exist", func() {
			BeforeEach(func() {
				setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/stacks?q=name%3Alinux",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body:   ` { "resources": []}`,
					}}))
			})

			It("returns an error", func() {
				_, err := repo.FindByName("linux")

				Expect(testHandler).To(testnet.HaveAllRequestsCalled())
				Expect(err).To(BeAssignableToTypeOf(&errors.ModelNotFoundError{}))
			})
		})
	})

	Describe("FindAll", func() {
		BeforeEach(func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/stacks",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body: `{
							"next_url": "/v2/stacks?page=2",
							"resources": [
								{
									"metadata": {
										"guid": "stack-guid-1",
										"url": "/v2/stacks/stack-guid-1",
										"created_at": "2013-08-31 01:32:40 +0000",
										"updated_at": "2013-08-31 01:32:40 +0000"
									},
									"entity": {
										"name": "lucid64",
										"description": "Ubuntu 10.04"
									}
								}
							]
						}`}}),

				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/stacks",
					Response: testnet.TestResponse{
						Status: http.StatusOK,
						Body: `
						{
							"resources": [
								{
									"metadata": {
										"guid": "stack-guid-2",
										"url": "/v2/stacks/stack-guid-2",
										"created_at": "2013-08-31 01:32:40 +0000",
										"updated_at": "2013-08-31 01:32:40 +0000"
									},
									"entity": {
										"name": "lucid64custom",
										"description": "Fake Ubuntu 10.04"
									}
								}
							]
						}`}}))
		})

		It("finds all the stacks", func() {
			stacks, err := repo.FindAll()

			Expect(testHandler).To(testnet.HaveAllRequestsCalled())
			Expect(err).NotTo(HaveOccurred())
			Expect(stacks).To(Equal([]models.Stack{
				{
					Guid:        "stack-guid-1",
					Name:        "lucid64",
					Description: "Ubuntu 10.04",
				},
				{
					Guid:        "stack-guid-2",
					Name:        "lucid64custom",
					Description: "Fake Ubuntu 10.04",
				},
			}))
		})
	})

})
