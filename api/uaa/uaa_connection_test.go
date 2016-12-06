package uaa_test

import (
	"fmt"
	"net/http"

	. "code.cloudfoundry.org/cli/api/uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

type DummyResponse struct {
	Val1 string `json:"val1"`
	Val2 int    `json:"val2"`
}

var _ = Describe("UAA Connection", func() {
	var (
		connection *UAAConnection
		request    *http.Request
	)

	BeforeEach(func() {
		connection = NewConnection(true, 0)
	})

	Describe("Make", func() {
		Describe("Data Unmarshalling", func() {
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

		Describe("Errors", func() {
			Context("when the server does not exist", func() {
				BeforeEach(func() {
					connection = NewConnection(false, 0)
				})

				It("returns a RequestError", func() {
					request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", "http://i.hope.this.doesnt.exist.com"), nil)
					Expect(err).ToNot(HaveOccurred())

					var response Response
					err = connection.Make(request, &response)
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

						connection = NewConnection(false, 0)
					})

					It("returns a UnverifiedServerError", func() {
						request, err := http.NewRequest(http.MethodGet, server.URL(), nil)
						Expect(err).ToNot(HaveOccurred())

						var response Response
						err = connection.Make(request, &response)
						Expect(err).To(MatchError(UnverifiedServerError{URL: server.URL()}))
					})
				})
			})

			Describe("RawHTTPStatusError", func() {
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

				It("returns a RawHTTPStatusError", func() {
					request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/foo", server.URL()), nil)
					Expect(err).ToNot(HaveOccurred())

					var response Response
					err = connection.Make(request, &response)
					Expect(err).To(MatchError(RawHTTPStatusError{
						StatusCode:  http.StatusUnauthorized,
						RawResponse: []byte(uaaResponse),
					}))

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
