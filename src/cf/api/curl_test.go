package api

import (
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testnet "testhelpers/net"
	"testing"
)

var jsonResponse = `
	{"resources": [
		{
		  "metadata": { "guid": "my-quota-guid" },
		  "entity": { "name": "my-remote-quota", "memory_limit": 1024 }
		}
	]}
`

func TestCurlGetRequest(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/endpoint",
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body:   jsonResponse},
	})
	ts, handler := testnet.NewTLSServer(t, []testnet.TestRequest{req})
	defer ts.Close()

	deps := newCurlDependencies()
	deps.config.Target = ts.URL

	repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
	headers, body, apiResponse := repo.Request("GET", "/v2/endpoint", "", "")

	assert.True(t, handler.AllRequestsCalled())
	assert.Contains(t, headers, "200")
	assert.Contains(t, headers, "Content-Type")
	assert.Contains(t, headers, "text/plain")
	testassert.JSONStringEquals(t, body, jsonResponse)
	assert.True(t, apiResponse.IsSuccessful())
}

func TestCurlPostRequest(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/endpoint",
		Matcher: testnet.RequestBodyMatcher(`{"key":"val"}`),
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body:   jsonResponse},
	})

	ts, handler := testnet.NewTLSServer(t, []testnet.TestRequest{req})
	defer ts.Close()

	deps := newCurlDependencies()
	deps.config.Target = ts.URL

	repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
	_, _, apiResponse := repo.Request("POST", "/v2/endpoint", "", `{"key":"val"}`)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestCurlFailingRequest(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "POST",
		Path:    "/v2/endpoint",
		Matcher: testnet.RequestBodyMatcher(`{"key":"val"}`),
		Response: testnet.TestResponse{
			Status: http.StatusBadRequest,
			Body:   jsonResponse},
	})

	ts, _ := testnet.NewTLSServer(t, []testnet.TestRequest{req})
	defer ts.Close()

	deps := newCurlDependencies()
	deps.config.Target = ts.URL

	repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
	_, body, _ := repo.Request("POST", "/v2/endpoint", "", `{"key":"val"}`)

	testassert.JSONStringEquals(t, body, jsonResponse)
}

func TestCurlWithCustomHeaders(t *testing.T) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "POST",
		Path:   "/v2/endpoint",
		Matcher: func(t *testing.T, req *http.Request) {
			assert.Equal(t, req.Header.Get("content-type"), "ascii/cats")
			assert.Equal(t, req.Header.Get("x-something-else"), "5")
		},
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body:   jsonResponse},
	})
	ts, handler := testnet.NewTLSServer(t, []testnet.TestRequest{req})
	defer ts.Close()

	deps := newCurlDependencies()
	deps.config.Target = ts.URL

	headers := "content-type: ascii/cats\nx-something-else:5"
	repo := NewCloudControllerCurlRepository(deps.config, deps.gateway)
	_, _, apiResponse := repo.Request("POST", "/v2/endpoint", headers, "")
	println(apiResponse.Message)
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

type curlDependencies struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func newCurlDependencies() (deps curlDependencies) {
	deps.config = &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
	}
	deps.gateway = net.NewCurlGateway()
	return
}
