package net_test

import (
	"bytes"
	"net/http"
	"strings"

	. "code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/trace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequestDumper", func() {
	Describe("DumpRequest", func() {
		var (
			printer trace.Printer
			buffer  *bytes.Buffer
			dumper  RequestDumper
		)

		BeforeEach(func() {
			buffer = new(bytes.Buffer)
			printer = trace.NewWriterPrinter(buffer, false)
			dumper = NewRequestDumper(printer)
		})

		When("the request body is JSON", func() {
			var (
				request *http.Request
				reqErr  error
			)

			BeforeEach(func() {
				bodyString := `{"password":"verysecret","some-field":"some-value"}`
				request, reqErr = http.NewRequest("GET", "example.com?code=code-from-uri", strings.NewReader(bodyString))
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("Authorization", "bearer: some-secret-token")
				request.Header.Set("Set-Cookie", "some-secret-cookie")
				request.Header.Set("Location", "https://api.cli.fun?code=secret-ssh-code")
				Expect(reqErr).ToNot(HaveOccurred())
			})

			JustBeforeEach(func() {
				dumper.DumpRequest(request)
			})

			It("redacts code=* from all headers", func() {
				Expect(buffer.String()).To(ContainSubstring("?code="))
				Expect(buffer.String()).ToNot(ContainSubstring("secret-ssh-code"))
			})

			It("redacts code=* from the uri", func() {
				Expect(buffer.String()).To(ContainSubstring("?code="))
				Expect(buffer.String()).ToNot(ContainSubstring("code-from-uri"))
			})

			It("redacts values from the key 'password'", func() {
				Expect(buffer.String()).To(ContainSubstring("password"))
				Expect(buffer.String()).ToNot(ContainSubstring("verysecret"))
			})

			It("redacts the authorization header", func() {
				Expect(buffer.String()).To(ContainSubstring("Authorization"))
				Expect(buffer.String()).ToNot(ContainSubstring("some-secret-token"))
			})

			It("redacts Set-Cookie headers", func() {
				Expect(buffer.String()).To(ContainSubstring("Set-Cookie: "))
				Expect(buffer.String()).ToNot(ContainSubstring("some-secret-cookie"))
			})
		})

		When("the request body is x-www-form-urlencoded", func() {
			var (
				request *http.Request
				reqErr  error
			)

			BeforeEach(func() {
				bodyString := `grant_type=password&password=somesecret&scope=&username=admin&refresh_token=secret-refresh-token&access_token=secret-access-token`
				request, reqErr = http.NewRequest("GET", "example.com", strings.NewReader(bodyString))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				request.Header.Set("Authorization", "bearer: some-secret-token")
				Expect(reqErr).ToNot(HaveOccurred())
			})

			JustBeforeEach(func() {
				dumper.DumpRequest(request)
			})

			It("redacts the value from keys called 'password'", func() {
				Expect(buffer.String()).To(ContainSubstring("password"))
				Expect(buffer.String()).ToNot(ContainSubstring("somesecret"))
			})

			It("redacts the authorization header", func() {
				Expect(buffer.String()).To(ContainSubstring("Authorization: "))
				Expect(buffer.String()).ToNot(ContainSubstring("some-secret-token"))
			})

			It("redacts fields containing 'token'", func() {
				Expect(buffer.String()).To(ContainSubstring("refresh_token="))
				Expect(buffer.String()).ToNot(ContainSubstring("secret-refresh-token"))
				Expect(buffer.String()).To(ContainSubstring("access_token="))
				Expect(buffer.String()).ToNot(ContainSubstring("secret-access-token"))
			})
		})
	})
})
