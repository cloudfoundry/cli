package net_test

import (
	. "cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
)

var failingUAARequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
	jsonResponse := `{ "error": "foo", "error_description": "The foo is wrong..." }`
	fmt.Fprintln(writer, jsonResponse)
}

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestUAAGatewayErrorHandling", func() {

			gateway := NewUAAGateway()

			ts := httptest.NewTLSServer(http.HandlerFunc(failingUAARequest))
			defer ts.Close()

			request, apiResponse := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
			assert.False(mr.T(), apiResponse.IsNotSuccessful())

			apiResponse = gateway.PerformRequest(request)

			assert.True(mr.T(), apiResponse.IsNotSuccessful())
			assert.Contains(mr.T(), apiResponse.Message, "The foo is wrong")
			assert.Contains(mr.T(), apiResponse.ErrorCode, "foo")
		})
	})
}
