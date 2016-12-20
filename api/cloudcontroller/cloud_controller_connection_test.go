package cloudcontroller_test

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	. "code.cloudfoundry.org/cli/api/cloudcontroller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

type DummyResponse struct {
	Val1 string `json:"val1"`
	Val2 int    `json:"val2"`
}

var _ = Describe("Cloud Controller Connection", func() {
	var connection *CloudControllerConnection

	BeforeEach(func() {
		connection = NewConnection(Config{SkipSSLValidation: true})
	})

	Describe("Make", func() {
		Describe("Data Unmarshalling", func() {
			var request *http.Request

			BeforeEach(func() {
				response := `{
					"val1":"2.59.0",
					"val2":2
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/foo", ""),
						RespondWith(http.StatusOK, response),
					),
				)

				var err error
				request, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when passed a response with a result set", func() {
				It("unmarshals the data into a struct", func() {
					var body DummyResponse
					response := Response{
						Result: &body,
					}

					err := connection.Make(request, &response)
					Expect(err).NotTo(HaveOccurred())

					Expect(body.Val1).To(Equal("2.59.0"))
					Expect(body.Val2).To(Equal(2))
				})
			})

			Context("when passed an empty response", func() {
				It("skips the unmarshalling step", func() {
					var response Response
					err := connection.Make(request, &response)
					Expect(err).NotTo(HaveOccurred())
					Expect(response.Result).To(BeNil())
				})
			})
		})

		Describe("HTTP Response", func() {
			var request *http.Request

			BeforeEach(func() {
				response := `{}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/foo", ""),
						RespondWith(http.StatusOK, response),
					),
				)

				var err error
				request, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the status", func() {
				response := Response{}

				err := connection.Make(request, &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.HTTPResponse.Status).To(Equal("200 OK"))
			})
		})

		Describe("Response Headers", func() {
			Describe("X-Cf-Warnings", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/foo"),
							RespondWith(http.StatusOK, "{}", http.Header{"X-Cf-Warnings": {"42, Ed McMann, the 1942 doggers"}}),
						),
					)
				})

				It("returns them in Response", func() {
					request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
					Expect(err).ToNot(HaveOccurred())

					var response Response
					err = connection.Make(request, &response)
					Expect(err).NotTo(HaveOccurred())

					Expect(server.ReceivedRequests()).To(HaveLen(1))

					warnings := response.Warnings
					Expect(warnings).ToNot(BeNil())
					Expect(warnings).To(HaveLen(3))
					Expect(warnings).To(ContainElement("42"))
					Expect(warnings).To(ContainElement("Ed McMann"))
					Expect(warnings).To(ContainElement("the 1942 doggers"))
				})
			})
		})

		Describe("Errors", func() {
			Context("when the server does not exist", func() {
				BeforeEach(func() {
					connection = NewConnection(Config{})
				})

				It("returns a RequestError", func() {
					request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", "http://garbledyguk.com"), nil)
					Expect(err).ToNot(HaveOccurred())

					var response Response
					err = connection.Make(request, &response)
					Expect(err).To(HaveOccurred())

					requestErr, ok := err.(RequestError)
					Expect(ok).To(BeTrue())
					Expect(requestErr.Error()).To(MatchRegexp(".*http://garbledyguk.com/v2/foo.*[nN]o such host"))
				})
			})

			Context("when the server does not have a verified certificate", func() {
				Context("skipSSLValidation is false", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v2/foo"),
							),
						)

						connection = NewConnection(Config{})
					})

					It("returns a UnverifiedServerError", func() {
						request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s", server.URL()), nil)
						Expect(err).ToNot(HaveOccurred())

						var response Response
						err = connection.Make(request, &response)
						Expect(err).To(MatchError(UnverifiedServerError{URL: server.URL()}))
					})
				})
			})

			Context("when the server's certificate does not match the hostname", func() {
				Context("skipSSLValidation is false", func() {
					BeforeEach(func() {
						if runtime.GOOS == "windows" {
							Skip("ssl validation has a different order on windows, will not be returned properly")
						}
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/"),
							),
						)

						connection = NewConnection(Config{})
					})

					// loopback.cli.ci.cf-app.com is a custom DNS record setup to point to 127.0.0.1
					It("returns a SSLValidationHostnameError", func() {
						altHostURL := strings.Replace(server.URL(), "127.0.0.1", "loopback.cli.ci.cf-app.com", -1)
						request, err := http.NewRequest(http.MethodGet, altHostURL, nil)
						Expect(err).ToNot(HaveOccurred())

						var response Response
						err = connection.Make(request, &response)
						Expect(err).To(MatchError(SSLValidationHostnameError{
							Message: "x509: certificate is valid for example.com, not loopback.cli.ci.cf-app.com",
						}))
					})
				})
			})

			Describe("RawHTTPStatusError", func() {
				var ccResponse string
				BeforeEach(func() {
					ccResponse = `{
						"code": 90004,
						"description": "The service binding could not be found: some-guid",
						"error_code": "CF-ServiceBindingNotFound"
					}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/foo"),
							RespondWith(http.StatusNotFound, ccResponse),
						),
					)
				})

				It("returns a CCRawResponse", func() {
					request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
					Expect(err).ToNot(HaveOccurred())

					var response Response
					err = connection.Make(request, &response)
					Expect(err).To(MatchError(RawHTTPStatusError{
						StatusCode:  http.StatusNotFound,
						RawResponse: []byte(ccResponse),
					}))

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
