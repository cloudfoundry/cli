package api_test

import (
	. "cf/api"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("Testing with ginkgo", func() {
	var allStacksResponse testnet.TestResponse

	BeforeEach(func() {
		allStacksResponse = testnet.TestResponse{Status: http.StatusOK, Body: `
			{
			  "resources": [
				{
				  "metadata": {
					"guid": "50688ae5-9bfc-4bf6-a4bf-caadb21a32c6",
					"url": "/v2/stacks/50688ae5-9bfc-4bf6-a4bf-caadb21a32c6",
					"created_at": "2013-08-31 01:32:40 +0000",
					"updated_at": "2013-08-31 01:32:40 +0000"
				  },
				  "entity": {
					"name": "lucid64",
					"description": "Ubuntu 10.04"
				  }
				},
				{
				  "metadata": {
					"guid": "e8cda251-7ce8-44b9-becb-ba5f5913d8ba",
					"url": "/v2/stacks/e8cda251-7ce8-44b9-becb-ba5f5913d8ba",
					"created_at": "2013-08-31 01:32:40 +0000",
					"updated_at": "2013-08-31 01:32:40 +0000"
				  },
				  "entity": {
					"name": "lucid64custom",
					"description": "Fake Ubuntu 10.04"
				  }
				}
			  ]
		}`}
	})

	It("TestStacksFindByName", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/stacks?q=name%3Alinux",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: ` { "resources": [
				{
				  "metadata": { "guid": "custom-linux-guid" },
				  "entity": { "name": "custom-linux" }
				}
			]}`},
		})
		ts, handler, repo := createStackRepo(req)
		defer ts.Close()

		stack, apiResponse := repo.FindByName("linux")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
		Expect(stack.Name).To(Equal("custom-linux"))
		Expect(stack.Guid).To(Equal("custom-linux-guid"))
	})

	It("TestStacksFindByNameNotFound", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/stacks?q=name%3Alinux",
			Response: testnet.TestResponse{Status: http.StatusOK, Body: ` { "resources": []}`},
		})
		ts, handler, repo := createStackRepo(req)
		defer ts.Close()

		_, apiResponse := repo.FindByName("linux")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
	})

	It("TestStacksFindAll", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/stacks",
			Response: allStacksResponse,
		})

		ts, handler, repo := createStackRepo(req)
		defer ts.Close()

		stacks, apiResponse := repo.FindAll()
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())
		Expect(len(stacks)).To(Equal(2))
		Expect(stacks[0].Name).To(Equal("lucid64"))
		Expect(stacks[0].Guid).To(Equal("50688ae5-9bfc-4bf6-a4bf-caadb21a32c6"))
	})
})

func createStackRepo(req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo StackRepository) {
	ts, handler = testnet.NewServer([]testnet.TestRequest{req})

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerStackRepository(configRepo, gateway)
	return
}
