package net_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	. "code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var failingUAARequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
	jsonResponse := `{ "error": "foo", "error_description": "The foo is wrong..." }`
	fmt.Fprintln(writer, jsonResponse)
}

var _ = Describe("UAA Gateway", func() {
	var gateway Gateway
	var config coreconfig.Reader
	var timeout string

	BeforeEach(func() {
		config = testconfig.NewRepository()
		timeout = "1"
	})

	JustBeforeEach(func() {
		gateway = NewUAAGateway(config, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), timeout)
	})

	It("parses error responses", func() {
		ts := httptest.NewTLSServer(http.HandlerFunc(failingUAARequest))
		defer ts.Close()
		gateway.SetTrustedCerts(ts.TLS.Certificates)

		request, apiErr := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
		_, apiErr = gateway.PerformRequest(request)

		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Error()).To(ContainSubstring("The foo is wrong"))
		Expect(apiErr.(errors.HTTPError).ErrorCode()).To(ContainSubstring("foo"))
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
