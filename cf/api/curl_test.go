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

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package api_test

import (
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

var expectedJSONResponse = `
	{"resources": [
		{
		  "metadata": { "guid": "my-quota-guid" },
		  "entity": { "name": "my-remote-quota", "memory_limit": 1024 }
		}
	]}
`

type curlDependencies struct {
	config  configuration.ReadWriter
	gateway net.Gateway
}

func newCurlDependencies() (deps curlDependencies) {
	deps.config = testconfig.NewRepository()
	deps.config.SetAccessToken("BEARER my_access_token")
	deps.gateway = net.NewCloudControllerGateway(deps.config)
	return
}

var _ = Describe("curl command", func() {
	It("TestCurlGetRequest", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/endpoint",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body:   expectedJSONResponse},
		})
		ts, handler := testnet.NewServer([]testnet.TestRequest{req})
		defer ts.Close()

		deps := newCurlDependencies()
		deps.config.SetApiEndpoint(ts.URL)

		repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
		headers, body, apiErr := repo.Request("GET", "/v2/endpoint", "", "")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(headers).To(ContainSubstring("200"))
		Expect(headers).To(ContainSubstring("Content-Type"))
		Expect(headers).To(ContainSubstring("text/plain"))
		testassert.JSONStringEquals(body, expectedJSONResponse)
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestCurlPostRequest", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/endpoint",
			Matcher: testnet.RequestBodyMatcher(`{"key":"val"}`),
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body:   expectedJSONResponse},
		})

		ts, handler := testnet.NewServer([]testnet.TestRequest{req})
		defer ts.Close()

		deps := newCurlDependencies()
		deps.config.SetApiEndpoint(ts.URL)

		repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
		_, _, apiErr := repo.Request("POST", "/v2/endpoint", "", `{"key":"val"}`)

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestCurlFailingRequest", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/endpoint",
			Matcher: testnet.RequestBodyMatcher(`{"key":"val"}`),
			Response: testnet.TestResponse{
				Status: http.StatusBadRequest,
				Body:   expectedJSONResponse},
		})

		ts, _ := testnet.NewServer([]testnet.TestRequest{req})
		defer ts.Close()

		deps := newCurlDependencies()
		deps.config.SetApiEndpoint(ts.URL)

		repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
		_, body, apiErr := repo.Request("POST", "/v2/endpoint", "", `{"key":"val"}`)

		Expect(apiErr).NotTo(HaveOccurred())
		testassert.JSONStringEquals(body, expectedJSONResponse)
	})

	It("TestCurlWithCustomHeaders", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "POST",
			Path:   "/v2/endpoint",
			Matcher: func(req *http.Request) {
				Expect(req.Header.Get("content-type")).To(Equal("ascii/cats"))
				Expect(req.Header.Get("x-something-else")).To(Equal("5"))
			},
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body:   expectedJSONResponse},
		})
		ts, handler := testnet.NewServer([]testnet.TestRequest{req})
		defer ts.Close()

		deps := newCurlDependencies()
		deps.config.SetApiEndpoint(ts.URL)

		headers := "content-type: ascii/cats\nx-something-else:5"
		repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
		_, _, apiErr := repo.Request("POST", "/v2/endpoint", headers, "")
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("TestCurlWithInvalidHeaders", func() {
		deps := newCurlDependencies()
		repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
		_, _, apiErr := repo.Request("POST", "/v2/endpoint", "not-valid", "")
		Expect(apiErr).To(HaveOccurred())
		Expect(apiErr.Error()).To(ContainSubstring("headers"))
	})
})
