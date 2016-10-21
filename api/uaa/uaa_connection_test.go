package uaa_test

import (
	"net/http"
	"net/url"
	"strings"

	. "code.cloudfoundry.org/cli/api/uaa"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/rata"
)

type DummyResponse struct {
	Val1 string `json:"val1"`
	Val2 int    `json:"val2"`
}

var _ = Describe("UAA Connection", func() {
	var (
		connection *UAAConnection
		FooRequest string
		routes     rata.Routes
	)

	BeforeEach(func() {
		FooRequest = "Foo"
		routes = rata.Routes{
			{Path: "/v2/foo", Method: http.MethodGet, Name: FooRequest},
		}
		connection = NewConnection(server.URL(), routes, true)
	})

	Describe("Make", func() {
		Describe("URL Generation", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/foo", "q=a:b&q=c:d"),
						RespondWith(http.StatusOK, "{}"),
					),
				)
			})

			Context("when passing a RequestName", func() {
				It("sends the request to the server", func() {
					request := NewRequest(FooRequest, nil, nil, url.Values{"q": {"a:b", "c:d"}})

					err := connection.Make(request, &Response{})
					Expect(err).NotTo(HaveOccurred())

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})

		Describe("Request Headers", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/foo"),
						VerifyHeaderKV("foo", "bar"),
						VerifyHeaderKV("accept", "application/json"),
						RespondWith(http.StatusOK, "{}"),
					),
				)
			})

			Context("when passed request headers", func() {
				It("merges the request headers and passes them to the server", func() {
					request := NewRequest(FooRequest, nil, http.Header{"foo": {"bar"}}, nil)

					err := connection.Make(request, &Response{})
					Expect(err).NotTo(HaveOccurred())

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})

		Describe("Request body", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/foo"),
						VerifyBody([]byte("some-body-parameters")),
						RespondWith(http.StatusOK, "{}"),
					),
				)
			})

			Context("when passing a Request body", func() {
				It("sends the request body to the server", func() {
					request := NewRequest(FooRequest, nil, nil, nil, strings.NewReader("some-body-parameters"))

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
						VerifyRequest(http.MethodGet, "/v2/foo", ""),
						RespondWith(http.StatusOK, response),
					),
				)

				request = NewRequest(FooRequest, nil, nil, nil)
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

		Describe("Errors", func() {
			Context("when the server does not exist", func() {
				BeforeEach(func() {
					connection = NewConnection("http://i.hope.this.doesnt.exist.com", routes, false)
				})

				It("returns a RequestError", func() {
					request := NewRequest(FooRequest, nil, nil, nil)

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
								VerifyRequest(http.MethodGet, "/v2/foo"),
							),
						)

						connection = NewConnection(server.URL(), routes, false)
					})

					It("returns a UnverifiedServerError", func() {
						request := NewRequest(FooRequest, nil, nil, nil)

						var response Response
						err := connection.Make(request, &response)
						Expect(err).To(MatchError(UnverifiedServerError{URL: server.URL()}))
					})
				})
			})

			Describe("UAAError", func() {
				var uaaResponse string
				BeforeEach(func() {
					uaaResponse = `{
						"error":"unauthorized",
						"error_description":"Bad credentials"
					}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/foo"),
							RespondWith(http.StatusUnauthorized, uaaResponse),
						),
					)
				})

				It("returns a UAAError", func() {
					request := NewRequest(FooRequest, nil, nil, nil)

					var response Response
					err := connection.Make(request, &response)
					Expect(err).To(MatchError(Error{
						Type:        "unauthorized",
						Description: "Bad credentials",
					}))

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
