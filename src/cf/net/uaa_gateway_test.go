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

	request, apiStatus := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
	assert.False(t, apiStatus.NotSuccessful())

	apiStatus = gateway.PerformRequest(request)

	assert.True(t, apiStatus.NotSuccessful())
	assert.Contains(t, apiStatus.Message, "The foo is wrong")
	assert.Contains(t, apiStatus.ErrorCode, "foo")
}
