/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package api_test

import (
	. "cf/api"
	"cf/models"
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

		stack, apiErr := repo.FindByName("linux")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
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

		_, apiErr := repo.FindByName("linux")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(BeNil())
	})

	It("TestStacksFindAll", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     "/v2/stacks",
			Response: allStacksResponse,
		})

		ts, handler, repo := createStackRepo(req)
		defer ts.Close()

		stacks, apiErr := repo.FindAll()
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(len(stacks)).To(Equal(2))
		Expect(stacks[0].Name).To(Equal("lucid64"))
		Expect(stacks[0].Guid).To(Equal("50688ae5-9bfc-4bf6-a4bf-caadb21a32c6"))
	})

	It("TestStacksFindAll multipage", func() {
		ts, handler, repo := createStackRepo2([]testnet.TestRequest{firstreq, secondreq})
		defer ts.Close()
		var stacks []models.Stack
		stacks, apiErr := repo.FindAll()
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
		Expect(len(stacks)).To(Equal(2))
		Expect(stacks[1].Name).To(Equal("lucid64custom"))
		Expect(stacks[1].Guid).To(Equal("e8cda251-7ce8-44b9-becb-ba5f5913d8ba"))
	})

})

var firstreq = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/stacks",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
		"next_url": "/v2/stacks?page=2",
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
			}
		]}`,
	},
})

var secondreq = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/stacks",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
		"resources": [
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
		]}`,
	},
})

func createStackRepo(req testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo StackRepository) {
	return createStackRepo2([]testnet.TestRequest{req})
}

func createStackRepo2(reqs []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo StackRepository) {
	ts, handler = testnet.NewServer(reqs)

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo)
	repo = NewCloudControllerStackRepository(configRepo, gateway)
	return
}
