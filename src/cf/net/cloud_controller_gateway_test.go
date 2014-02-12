package net_test

import (
	. "cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var failingCloudControllerRequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
	jsonResponse := `{ "code": 210003, "description": "The host is taken: test1" }`
	fmt.Fprintln(writer, jsonResponse)
}

var invalidTokenCloudControllerRequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
	jsonResponse := `{ "code": 1000, "description": "The token is invalid" }`
	fmt.Fprintln(writer, jsonResponse)
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestCloudControllerGatewayErrorHandling", func() {
		gateway := NewCloudControllerGateway()

		ts := httptest.NewTLSServer(http.HandlerFunc(failingCloudControllerRequest))
		defer ts.Close()

		request, apiResponse := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())

		apiResponse = gateway.PerformRequest(request)

		Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
		Expect(apiResponse.Message).To(ContainSubstring("The host is taken: test1"))
		Expect(apiResponse.ErrorCode).To(ContainSubstring("210003"))
	})
	It("TestCloudControllerGatewayInvalidTokenHandling", func() {

		gateway := NewCloudControllerGateway()

		ts := httptest.NewTLSServer(http.HandlerFunc(invalidTokenCloudControllerRequest))
		defer ts.Close()

		request, apiResponse := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
		Expect(apiResponse.IsNotSuccessful()).To(BeFalse())

		apiResponse = gateway.PerformRequest(request)

		Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
		Expect(apiResponse.Message).To(ContainSubstring("The token is invalid"))
		Expect(apiResponse.ErrorCode).To(ContainSubstring(INVALID_TOKEN_CODE))
	})
})
