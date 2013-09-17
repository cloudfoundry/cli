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
	gateway := NewCloudControllerGateway(nil)

	ts := httptest.NewTLSServer(http.HandlerFunc(failingCloudControllerRequest))
	defer ts.Close()

	request, err := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
	assert.NoError(t, err)

	err = gateway.PerformRequest(request)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "The host is taken: test1")
	assert.Contains(t, err.ErrorCode, "210003")
}

var invalidTokenCloudControllerRequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
	jsonResponse := `{ "code": 1000, "description": "The token is invalid" }`
	fmt.Fprintln(writer, jsonResponse)
}

func TestCloudControllerGatewayInvalidTokenHandling(t *testing.T) {
	gateway := NewCloudControllerGateway(nil)

	ts := httptest.NewTLSServer(http.HandlerFunc(invalidTokenCloudControllerRequest))
	defer ts.Close()

	request, err := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
	assert.NoError(t, err)

	err = gateway.PerformRequest(request)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "The token is invalid")
	assert.Contains(t, err.ErrorCode, INVALID_TOKEN_CODE)
}
