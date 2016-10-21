package cloudcontroller_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/rata"
)

type DummyResponse struct {
	Val1 string `json:"val1"`
	Val2 int    `json:"val2"`
}

var _ = Describe("Cloud Controller Connection", func() {
	var (
		connection *CloudControllerConnection
		FooRequest string
		routes     rata.Routes
	)

	BeforeEach(func() {
		FooRequest = "Foo"
		routes = rata.Routes{
			{Path: "/v2/foo", Method: "GET", Name: FooRequest},
		}
		connection = NewConnection(server.URL(), routes, true)
	})

	Describe("Make", func() {
		Describe("HTTP request generation", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v2/foo", "q=a:b&q=c:d"),
						RespondWith(http.StatusOK, "{}"),
					),
				)
			})

			Context("when generating the request via a RequestName", func() {
				It("sends the request to the server", func() {
					request := NewRequest(
						FooRequest,
						nil,
						nil,
						map[string][]string{
							"q": {"a:b", "c:d"},
						},
					)

					err := connection.Make(request, &Response{})
					Expect(err).NotTo(HaveOccurred())

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				Context("when an error is encountered", func() {
					It("returns the error", func() {
						request := NewRequest(
							"some-invalid-request-name",
							nil,
							nil,
							nil,
						)

						err := connection.Make(request, &Response{})
						Expect(err).To(MatchError("No route exists with the name some-invalid-request-name"))
					})
				})
			})

			Context("when generating the request via an URI", func() {
				It("sends the request to the server", func() {
					request := NewRequestFromURI(
						"/v2/foo?q=a:b&q=c:d",
						"GET",
						nil,
					)

					err := connection.Make(request, &Response{})
					Expect(err).NotTo(HaveOccurred())

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				Context("when an error is encountered", func() {
					It("returns the error", func() {
						request := NewRequestFromURI(
							"/v2/foo?q=a:b&q=c:d",
							"INVALID:METHOD",
							nil,
						)

						err := connection.Make(request, &Response{})
						Expect(err).To(MatchError("net/http: invalid method \"INVALID:METHOD\""))
					})
				})
			})
		})

		Describe("Request Headers", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v2/foo", ""),
						VerifyHeaderKV("foo", "bar"),
						VerifyHeaderKV("accept", "application/json"),
						VerifyHeaderKV("content-type", "application/json"),
						RespondWith(http.StatusOK, "{}"),
					),
				)
			})

			Context("when passed headers", func() {
				It("passes headers to the server", func() {
					request := NewRequestFromURI(
						"/v2/foo",
						"GET",
						http.Header{
							"foo": {"bar"},
						},
					)

					err := connection.Make(request, &Response{})
					Expect(err).NotTo(HaveOccurred())

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})

		Describe("Data Unmarshalling", func() {
			var request Request

			BeforeEach(func() {
				response := `{
					"val1":"2.59.0",
					"val2":2
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v2/foo", ""),
						RespondWith(http.StatusOK, response),
					),
				)

				request = NewRequestFromURI(
					"/v2/foo",
					"GET",
					nil,
				)
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

		Describe("Response Headers", func() {
			Describe("X-Cf-Warnings", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("GET", "/v2/foo"),
							RespondWith(http.StatusOK, "{}", http.Header{"X-Cf-Warnings": {"42, Ed McMann, the 1942 doggers"}}),
						),
					)
				})

				It("returns them in Response", func() {
					request := NewRequest(
						FooRequest,
						nil,
						nil,
						nil,
					)

					var response Response
					err := connection.Make(request, &response)
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
					connection = NewConnection("http://i.hope.this.doesnt.exist.com", routes, false)
				})

				It("returns a RequestError", func() {
					request := NewRequest(
						FooRequest,
						nil,
						nil,
						nil,
					)

					var response Response
					err := connection.Make(request, &response)
					Expect(err).To(HaveOccurred())

					requestErr, ok := err.(RequestError)
					Expect(ok).To(BeTrue())
					Expect(requestErr.Error()).To(MatchRegexp(".*http://i.hope.this.doesnt.exist.com/v2/foo.*[nN]o such host"))
				})
			})

			Context("when the server does not have a verified certificate", func() {
				Context("skipSSLValidation is false", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/v2/foo"),
							),
						)

						connection = NewConnection(server.URL(), routes, false)
					})

					It("returns a UnverifiedServerError", func() {
						request := NewRequest(
							FooRequest,
							nil,
							nil,
							nil,
						)

						var response Response
						err := connection.Make(request, &response)
						Expect(err).To(MatchError(UnverifiedServerError{URL: server.URL()}))
					})
				})
			})

			Describe("RawCCError", func() {
				var ccResponse string
				BeforeEach(func() {
					ccResponse = `{
						"code": 90004,
						"description": "The service binding could not be found: some-guid",
						"error_code": "CF-ServiceBindingNotFound"
					}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("GET", "/v2/foo"),
							RespondWith(http.StatusNotFound, ccResponse),
						),
					)
				})

				It("returns a CCRawResponse", func() {
					request := NewRequest(
						FooRequest,
						nil,
						nil,
						nil,
					)

					var response Response
					err := connection.Make(request, &response)
					Expect(err).To(MatchError(RawCCError{
						StatusCode:  http.StatusNotFound,
						RawResponse: []byte(ccResponse),
					}))

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
