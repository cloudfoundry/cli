package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var jsonResponse = `
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
	deps.gateway = net.NewCloudControllerGateway()
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestCurlGetRequest", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/endpoint",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body:   jsonResponse},
		})
		ts, handler := testnet.NewTLSServer([]testnet.TestRequest{req})
		defer ts.Close()

		deps := newCurlDependencies()
		deps.config.SetApiEndpoint(ts.URL)

		repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
		headers, body, apiResponse := repo.Request("GET", "/v2/endpoint", "", "")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(headers).To(ContainSubstring("200"))
		Expect(headers).To(ContainSubstring("Content-Type"))
		Expect(headers).To(ContainSubstring("text/plain"))
		testassert.JSONStringEquals(body, jsonResponse)
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})

	It("TestCurlPostRequest", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/endpoint",
			Matcher: testnet.RequestBodyMatcher(`{"key":"val"}`),
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body:   jsonResponse},
		})

		ts, handler := testnet.NewTLSServer([]testnet.TestRequest{req})
		defer ts.Close()

		deps := newCurlDependencies()
		deps.config.SetApiEndpoint(ts.URL)

		repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
		_, _, apiResponse := repo.Request("POST", "/v2/endpoint", "", `{"key":"val"}`)

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})

	It("TestCurlFailingRequest", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:  "POST",
			Path:    "/v2/endpoint",
			Matcher: testnet.RequestBodyMatcher(`{"key":"val"}`),
			Response: testnet.TestResponse{
				Status: http.StatusBadRequest,
				Body:   jsonResponse},
		})

		ts, _ := testnet.NewTLSServer([]testnet.TestRequest{req})
		defer ts.Close()

		deps := newCurlDependencies()
		deps.config.SetApiEndpoint(ts.URL)

		repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
		_, body, _ := repo.Request("POST", "/v2/endpoint", "", `{"key":"val"}`)

		testassert.JSONStringEquals(body, jsonResponse)
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
				Body:   jsonResponse},
		})
		ts, handler := testnet.NewTLSServer([]testnet.TestRequest{req})
		defer ts.Close()

		deps := newCurlDependencies()
		deps.config.SetApiEndpoint(ts.URL)

		headers := "content-type: ascii/cats\nx-something-else:5"
		repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
		_, _, apiResponse := repo.Request("POST", "/v2/endpoint", headers, "")
		println(apiResponse.Message)
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})

	It("TestCurlWithInvalidHeaders", func() {
		deps := newCurlDependencies()
		repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
		_, _, apiResponse := repo.Request("POST", "/v2/endpoint", "not-valid", "")
		Expect(apiResponse.IsError()).To(BeTrue())
		Expect(apiResponse.Message).To(ContainSubstring("headers"))
	})
})
