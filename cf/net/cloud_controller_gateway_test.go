package net_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	. "code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"

	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	var config coreconfig.Reader
	var timeout string

	BeforeEach(func() {
		timeout = "1"
	})

	JustBeforeEach(func() {
		config = testconfig.NewRepository()
		gateway = NewCloudControllerGateway(config, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), timeout)
	})

	It("parses error responses", func() {
		ts := httptest.NewTLSServer(http.HandlerFunc(failingCloudControllerRequest))
		ts.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
		defer ts.Close()
		gateway.SetTrustedCerts(ts.TLS.Certificates)

		request, apiErr := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
		_, apiErr = gateway.PerformRequest(request)

		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Error()).To(ContainSubstring("The host is taken: test1"))
		Expect(apiErr.(errors.HTTPError).ErrorCode()).To(ContainSubstring("210003"))
	})

	It("parses invalid token responses", func() {
		ts := httptest.NewTLSServer(http.HandlerFunc(invalidTokenCloudControllerRequest))
		ts.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
		defer ts.Close()
		gateway.SetTrustedCerts(ts.TLS.Certificates)

		request, apiErr := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
		_, apiErr = gateway.PerformRequest(request)

		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Error()).To(ContainSubstring("The token is invalid"))
		Expect(apiErr.(*errors.InvalidTokenError)).To(HaveOccurred())
	})

	It("uses the set dial timeout", func() {
		Expect(gateway.DialTimeout).To(Equal(1 * time.Second))
	})

	Context("with an invalid timeout", func() {
		BeforeEach(func() {
			timeout = ""
		})

		It("uses the default dial timeout", func() {
			Expect(gateway.DialTimeout).To(Equal(5 * time.Second))
		})
	})
})
