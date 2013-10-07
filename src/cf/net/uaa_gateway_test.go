package net_test

import (
	. "cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var failingUAARequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
	jsonResponse := `{ "error": "foo", "error_description": "The foo is wrong..." }`
	fmt.Fprintln(writer, jsonResponse)
}

func TestUAAGatewayErrorHandling(t *testing.T) {
	gateway := NewUAAGateway()

	ts := httptest.NewTLSServer(http.HandlerFunc(failingUAARequest))
	defer ts.Close()

	request, apiResponse := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
	assert.False(t, apiResponse.IsNotSuccessful())

	apiResponse = gateway.PerformRequest(request)

	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Contains(t, apiResponse.Message, "The foo is wrong")
	assert.Contains(t, apiResponse.ErrorCode, "foo")
}
