package net_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var failingRoutingApiRequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
	jsonResponse := `{ "name": "some-error", "message": "The host is taken: test1" }`
	fmt.Fprintln(writer, jsonResponse)
}

var invalidTokenRoutingApiRequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusUnauthorized)
	jsonResponse := `{ "name": "UnauthorizedError", "message": "bad token!" }`
	fmt.Fprintln(writer, jsonResponse)
}

var _ = Describe("Routing Api Gateway", func() {
	var gateway Gateway
	var config core_config.Reader

	BeforeEach(func() {
		config = testconfig.NewRepository()
		gateway = NewRoutingApiGateway(config, time.Now, &testterm.FakeUI{})
	})

	It("parses error responses", func() {
		ts := httptest.NewTLSServer(http.HandlerFunc(failingRoutingApiRequest))
		defer ts.Close()
		gateway.SetTrustedCerts(ts.TLS.Certificates)

		request, apiErr := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
		_, apiErr = gateway.PerformRequest(request)

		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Error()).To(ContainSubstring("The host is taken: test1"))
		Expect(apiErr.(errors.HttpError).ErrorCode()).To(ContainSubstring("some-error"))
	})

	It("parses invalid token responses", func() {
		ts := httptest.NewTLSServer(http.HandlerFunc(invalidTokenRoutingApiRequest))
		defer ts.Close()
		gateway.SetTrustedCerts(ts.TLS.Certificates)

		request, apiErr := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
		_, apiErr = gateway.PerformRequest(request)

		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Error()).To(ContainSubstring("bad token"))
		Expect(apiErr.(errors.HttpError)).To(HaveOccurred())
	})

	Context("when the Routing API returns a invalid Json", func() {
		var invalidJsonResponse = func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusUnauthorized)
			jsonResponse := `¯\_(ツ)_/¯`
			fmt.Fprintln(writer, jsonResponse)
		}

		It("returns a 500 http error", func() {
			ts := httptest.NewTLSServer(http.HandlerFunc(invalidJsonResponse))
			defer ts.Close()
			gateway.SetTrustedCerts(ts.TLS.Certificates)

			request, apiErr := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
			_, apiErr = gateway.PerformRequest(request)

			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.(errors.HttpError)).To(HaveOccurred())
			Expect(apiErr.(errors.HttpError).StatusCode()).To(Equal(http.StatusInternalServerError))
		})
	})
})
