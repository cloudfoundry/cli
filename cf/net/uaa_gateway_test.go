package net_test

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var failingUAARequest = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusBadRequest)
	jsonResponse := `{ "error": "foo", "error_description": "The foo is wrong..." }`
	fmt.Fprintln(writer, jsonResponse)
}

var _ = Describe("UAA Gateway", func() {
	var gateway Gateway
	var config configuration.Reader

	BeforeEach(func() {
		config = testconfig.NewRepository()
		gateway = NewUAAGateway(config)
	})

	It("parses error responses", func() {
		ts := httptest.NewTLSServer(http.HandlerFunc(failingUAARequest))
		defer ts.Close()
		gateway.SetTrustedCerts(ts.TLS.Certificates)

		request, apiErr := gateway.NewRequest("GET", ts.URL, "TOKEN", nil)
		_, apiErr = gateway.PerformRequest(request)

		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Error()).To(ContainSubstring("The foo is wrong"))
		Expect(apiErr.(errors.HttpError).ErrorCode()).To(ContainSubstring("foo"))
	})
})
