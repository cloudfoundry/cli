package net_test

import (
	. "cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var failingCloudControllerRequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
	jsonResponse := `{ "code": 210003, "description": "The host is taken: test1" }`
	fmt.Fprintln(writer, jsonResponse)
}

func TestCloudControllerGatewayErrorHandling(t *testing.T) {
	gateway := NewCloudControllerGateway()

	ts := httptest.NewTLSServer(http.HandlerFunc(failingCloudControllerRequest))
	defer ts.Close()

	request, apiResponse := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
	assert.False(t, apiResponse.IsNotSuccessful())

	apiResponse = gateway.PerformRequest(request)

	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Contains(t, apiResponse.Message, "The host is taken: test1")
	assert.Contains(t, apiResponse.ErrorCode, "210003")
}

var invalidTokenCloudControllerRequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
	jsonResponse := `{ "code": 1000, "description": "The token is invalid" }`
	fmt.Fprintln(writer, jsonResponse)
}

func TestCloudControllerGatewayInvalidTokenHandling(t *testing.T) {
	gateway := NewCloudControllerGateway()

	ts := httptest.NewTLSServer(http.HandlerFunc(invalidTokenCloudControllerRequest))
	defer ts.Close()

	request, apiResponse := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
	assert.False(t, apiResponse.IsNotSuccessful())

	apiResponse = gateway.PerformRequest(request)

	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Contains(t, apiResponse.Message, "The token is invalid")
	assert.Contains(t, apiResponse.ErrorCode, INVALID_TOKEN_CODE)
}
