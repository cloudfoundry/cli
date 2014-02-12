package net_test

import (
	. "cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

var _ = Describe("Testing with ginkgo", func() {
	It("TestUAAGatewayErrorHandling", func() {

		gateway := NewUAAGateway()

		ts := httptest.NewTLSServer(http.HandlerFunc(failingUAARequest))
		defer ts.Close()

		request, apiResponse := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())

		apiResponse = gateway.PerformRequest(request)

		Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
		assert.Contains(mr.T(), apiResponse.Message, "The foo is wrong")
		assert.Contains(mr.T(), apiResponse.ErrorCode, "foo")
	})
})
