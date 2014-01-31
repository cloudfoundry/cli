package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
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
	config  *configuration.Configuration
	gateway net.Gateway
}

func newCurlDependencies() (deps curlDependencies) {
	deps.config = &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
	}
	deps.gateway = net.NewCloudControllerGateway()
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCurlGetRequest", func() {
			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/endpoint",
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body:   jsonResponse},
			})
			ts, handler := testnet.NewTLSServer(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			deps := newCurlDependencies()
			deps.config.Target = ts.URL

			repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
			headers, body, apiResponse := repo.Request("GET", "/v2/endpoint", "", "")

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.Contains(mr.T(), headers, "200")
			assert.Contains(mr.T(), headers, "Content-Type")
			assert.Contains(mr.T(), headers, "text/plain")
			testassert.JSONStringEquals(mr.T(), body, jsonResponse)
			assert.True(mr.T(), apiResponse.IsSuccessful())
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

			ts, handler := testnet.NewTLSServer(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			deps := newCurlDependencies()
			deps.config.Target = ts.URL

			repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
			_, _, apiResponse := repo.Request("POST", "/v2/endpoint", "", `{"key":"val"}`)

			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
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

			ts, _ := testnet.NewTLSServer(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			deps := newCurlDependencies()
			deps.config.Target = ts.URL

			repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
			_, body, _ := repo.Request("POST", "/v2/endpoint", "", `{"key":"val"}`)

			testassert.JSONStringEquals(mr.T(), body, jsonResponse)
		})
		It("TestCurlWithCustomHeaders", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "POST",
				Path:   "/v2/endpoint",
				Matcher: func(t mr.TestingT, req *http.Request) {
					assert.Equal(mr.T(), req.Header.Get("content-type"), "ascii/cats")
					assert.Equal(mr.T(), req.Header.Get("x-something-else"), "5")
				},
				Response: testnet.TestResponse{
					Status: http.StatusOK,
					Body:   jsonResponse},
			})
			ts, handler := testnet.NewTLSServer(mr.T(), []testnet.TestRequest{req})
			defer ts.Close()

			deps := newCurlDependencies()
			deps.config.Target = ts.URL

			headers := "content-type: ascii/cats\nx-something-else:5"
			repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
			_, _, apiResponse := repo.Request("POST", "/v2/endpoint", headers, "")
			println(apiResponse.Message)
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
		It("TestCurlWithInvalidHeaders", func() {

			deps := newCurlDependencies()
			repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
			_, _, apiResponse := repo.Request("POST", "/v2/endpoint", "not-valid", "")
			assert.True(mr.T(), apiResponse.IsError())
			assert.Contains(mr.T(), apiResponse.Message, "headers")
		})
	})
}
