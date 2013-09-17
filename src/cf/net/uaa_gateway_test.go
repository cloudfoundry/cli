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
	gateway := NewUAAGateway(nil)

	ts := httptest.NewTLSServer(http.HandlerFunc(failingUAARequest))
	defer ts.Close()

	request, err := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
	assert.NoError(t, err)

	err = gateway.PerformRequest(request)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "The foo is wrong")
	assert.Contains(t, err.ErrorCode, "foo")
}
