package v2_test

import (
	. "code.cloudfoundry.org/cli/commands/v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors", func() {
	translateFunc := func(s string, vars ...interface{}) string {
		return "translated " + s
	}

	Describe("APIRequestError", func() {
		Describe("Error", func() {
			It("returns the error template", func() {
				e := APIRequestError{}
				Expect(e).To(MatchError("Request error: {{.Error}}\nTIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."))
			})
		})

		Describe("Translate", func() {
			It("returns the translated error", func() {
				e := APIRequestError{}
				Expect(e.Translate(translateFunc)).To(Equal("translated Request error: {{.Error}}\nTIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."))
			})
		})
	})

	Describe("InvalidSSLCertError", func() {
		Describe("Error", func() {
			It("returns the error template", func() {
				e := InvalidSSLCertError{}
				Expect(e).To(MatchError("Invalid SSL Cert for {{.API}}\nTIP: Use 'cf api --skip-ssl-validation' to continue with an insecure API endpoint"))
			})
		})

		Describe("Translate", func() {
			It("returns the translated error", func() {
				e := InvalidSSLCertError{}
				Expect(e.Translate(translateFunc)).To(Equal("translated Invalid SSL Cert for {{.API}}\nTIP: Use 'cf api --skip-ssl-validation' to continue with an insecure API endpoint"))
			})
		})
	})
})
