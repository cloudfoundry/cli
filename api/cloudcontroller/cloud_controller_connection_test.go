package cloudcontroller_test

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	. "code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

type DummyResponse struct {
	Val1 string      `json:"val1"`
	Val2 int         `json:"val2"`
	Val3 interface{} `json:"val3,omitempty"`
}

var _ = Describe("Cloud Controller Connection", func() {
	var connection *CloudControllerConnection

	BeforeEach(func() {
		connection = NewConnection(Config{SkipSSLValidation: true})
	})

	Describe("Make", func() {
		Describe("Data Unmarshalling", func() {
			var request *Request

			BeforeEach(func() {
				response := `{
					"val1":"2.59.0",
					"val2":2,
					"val3":1111111111111111111
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/foo", ""),
						RespondWith(http.StatusOK, response),
					),
				)

				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
				Expect(err).ToNot(HaveOccurred())
				request = &Request{Request: req}
			})

			When("passed a response with a result set", func() {
				It("unmarshals the data into a struct", func() {
					var body DummyResponse
					response := Response{
						DecodeJSONResponseInto: &body,
					}

					err := connection.Make(request, &response)
					Expect(err).NotTo(HaveOccurred())

					Expect(body.Val1).To(Equal("2.59.0"))
					Expect(body.Val2).To(Equal(2))
				})

				It("keeps numbers unmarshalled to interfaces as interfaces", func() {
					var body DummyResponse
					response := Response{
						DecodeJSONResponseInto: &body,
					}

					err := connection.Make(request, &response)
					Expect(err).NotTo(HaveOccurred())
					Expect(fmt.Sprint(body.Val3)).To(Equal("1111111111111111111"))
				})
			})

			When("passed an empty response", func() {
				It("skips the unmarshalling step", func() {
					var response Response
					err := connection.Make(request, &response)
					Expect(err).NotTo(HaveOccurred())
					Expect(response.DecodeJSONResponseInto).To(BeNil())
				})
			})
		})

		Describe("HTTP Response", func() {
			var request *Request

			BeforeEach(func() {
				response := `{}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/foo", ""),
						RespondWith(http.StatusOK, response),
					),
				)

				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
				Expect(err).ToNot(HaveOccurred())
				request = &Request{Request: req}
			})

			It("returns the status", func() {
				response := Response{}

				err := connection.Make(request, &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.HTTPResponse.Status).To(Equal("200 OK"))
			})
		})

		Describe("Response Headers", func() {
			Describe("Location", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/foo"),
							RespondWith(http.StatusAccepted, "{}", http.Header{"Location": {"/v2/some-location"}}),
						),
					)
				})

				It("returns the location in the ResourceLocationURL", func() {
					req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
					Expect(err).ToNot(HaveOccurred())
					request := &Request{Request: req}

					var response Response
					err = connection.Make(request, &response)
					Expect(err).NotTo(HaveOccurred())

					Expect(server.ReceivedRequests()).To(HaveLen(1))
					Expect(response.ResourceLocationURL).To(Equal("/v2/some-location"))
				})
			})
			Describe("X-Cf-Warnings", func() {
				When("there are warnings", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v2/foo"),
								RespondWith(http.StatusOK, "{}", http.Header{"X-Cf-Warnings": {"42,+Ed+McMann,+the+1942+doggers,a%2Cb"}}),
							),
						)
					})

					It("returns them in Response", func() {
						req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
						Expect(err).ToNot(HaveOccurred())
						request := &Request{Request: req}

						var response Response
						err = connection.Make(request, &response)
						Expect(err).NotTo(HaveOccurred())

						Expect(server.ReceivedRequests()).To(HaveLen(1))

						warnings := response.Warnings
						Expect(warnings).ToNot(BeNil())
						Expect(warnings).To(HaveLen(4))
						Expect(warnings).To(ContainElement("42"))
						Expect(warnings).To(ContainElement("Ed McMann"))
						Expect(warnings).To(ContainElement("the 1942 doggers"))
						Expect(warnings).To(ContainElement("a,b"))
					})
				})

				When("there are warnings using multi-value header", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v2/foo"),
								RespondWith(http.StatusOK, "{}", http.Header{"X-Cf-Warnings": {
									"42,+Ed+McMann,+the+1942+doggers,a%2Cb",
									"something,simpler",
								}}),
							),
						)
					})

					It("returns them in Response", func() {
						req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
						Expect(err).ToNot(HaveOccurred())
						request := &Request{Request: req}

						var response Response
						err = connection.Make(request, &response)
						Expect(err).NotTo(HaveOccurred())

						Expect(server.ReceivedRequests()).To(HaveLen(1))

						warnings := response.Warnings
						Expect(warnings).ToNot(BeNil())
						Expect(warnings).To(HaveLen(6))
						Expect(warnings).To(ContainElement("42"))
						Expect(warnings).To(ContainElement("Ed McMann"))
						Expect(warnings).To(ContainElement("the 1942 doggers"))
						Expect(warnings).To(ContainElement("a,b"))
						Expect(warnings).To(ContainElement("something"))
						Expect(warnings).To(ContainElement("simpler"))
					})
				})

				When("there are no warnings", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v2/foo"),
								RespondWith(http.StatusOK, "{}"),
							),
						)
					})

					It("returns them in Response", func() {
						req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
						Expect(err).ToNot(HaveOccurred())
						request := &Request{Request: req}

						var response Response
						err = connection.Make(request, &response)
						Expect(err).NotTo(HaveOccurred())

						Expect(response.Warnings).To(BeEmpty())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})
			})
		})

		Describe("Errors", func() {
			When("the server does not exist", func() {
				BeforeEach(func() {
					connection = NewConnection(Config{})
				})

				It("returns a RequestError", func() {
					req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", "http://garbledyguk.com"), nil)
					Expect(err).ToNot(HaveOccurred())
					request := &Request{Request: req}

					var response Response
					err = connection.Make(request, &response)
					Expect(err).To(HaveOccurred())

					requestErr, ok := err.(ccerror.RequestError)
					Expect(ok).To(BeTrue())
					Expect(requestErr.Error()).To(MatchRegexp(".*http://garbledyguk.com/v2/foo.*[nN]o such host"))
				})
			})

			When("the server does not have a verified certificate", func() {
				Context("skipSSLValidation is false", func() {
					BeforeEach(func() {
						if runtime.GOOS == "darwin" {
							Skip("ssl verification is different on darwin")
						}
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v2/foo"),
							),
						)

						connection = NewConnection(Config{})
					})

					It("returns a UnverifiedServerError", func() {
						req, err := http.NewRequest(http.MethodGet, server.URL(), nil)
						Expect(err).ToNot(HaveOccurred())
						request := &Request{Request: req}

						var response Response
						err = connection.Make(request, &response)
						Expect(err).To(MatchError(ccerror.UnverifiedServerError{URL: server.URL()}))
					})
				})
			})

			When("the server's certificate does not match the hostname", func() {
				Context("skipSSLValidation is false", func() {
					BeforeEach(func() {
						if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
							Skip("ssl validation has a different order on windows/darwin, will not be returned properly")
						}
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/"),
							),
						)

						connection = NewConnection(Config{})
					})

					// loopback.cli.fun is a custom DNS record setup to point to 127.0.0.1
					It("returns a SSLValidationHostnameError", func() {
						altHostURL := strings.Replace(server.URL(), "127.0.0.1", "loopback.cli.fun", -1)
						req, err := http.NewRequest(http.MethodGet, altHostURL, nil)
						Expect(err).ToNot(HaveOccurred())
						request := &Request{Request: req}

						var response Response
						err = connection.Make(request, &response)
						Expect(err).To(MatchError(ccerror.SSLValidationHostnameError{
							Message: "x509: certificate is valid for example.com, not loopback.cli.fun",
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
							RespondWith(http.StatusNotFound, ccResponse, http.Header{"X-Vcap-Request-Id": {"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95", "6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f"}}),
						),
					)
				})

				It("returns a CCRawResponse", func() {
					req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
					Expect(err).ToNot(HaveOccurred())
					request := &Request{Request: req}

					var response Response
					err = connection.Make(request, &response)
					Expect(err).To(MatchError(ccerror.RawHTTPStatusError{
						StatusCode:  http.StatusNotFound,
						RawResponse: []byte(ccResponse),
						RequestIDs:  []string{"6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95", "6e0b4379-f5f7-4b2b-56b0-9ab7e96eed95::7445d9db-c31e-410d-8dc5-9f79ec3fc26f"},
					}))

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
