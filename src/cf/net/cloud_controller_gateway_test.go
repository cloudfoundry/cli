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

var _ = Describe("Cloud Controller Gateway", func() {
	var gateway Gateway

	BeforeEach(func() {
		gateway = NewCloudControllerGateway()
	})

	It("parses error responses", func() {
		ts := httptest.NewTLSServer(http.HandlerFunc(failingCloudControllerRequest))
		defer ts.Close()
		gateway.AddTrustedCerts(ts.TLS.Certificates)

		request, apiResponse := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
		apiResponse = gateway.PerformRequest(request)

		Expect(apiResponse).NotTo(BeNil())
		Expect(apiResponse.Error()).To(ContainSubstring("The host is taken: test1"))
		Expect(apiResponse.ErrorCode()).To(ContainSubstring("210003"))
	})

	It("parses invalid token responses", func() {
		ts := httptest.NewTLSServer(http.HandlerFunc(invalidTokenCloudControllerRequest))
		defer ts.Close()
		gateway.AddTrustedCerts(ts.TLS.Certificates)

		request, apiResponse := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
		apiResponse = gateway.PerformRequest(request)

		Expect(apiResponse).NotTo(BeNil())
		Expect(apiResponse.Error()).To(ContainSubstring("The token is invalid"))
		Expect(apiResponse.ErrorCode()).To(ContainSubstring(INVALID_TOKEN_CODE))
	})
})
