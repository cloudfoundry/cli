package cloudcontrollerv2_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontrollerv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

type DummyResponse struct {
	Val1 string `json:"val1"`
	Val2 int    `json:"val2"`
}

var _ = Describe("Connection", func() {
	var (
		connection *Connection
	)

	BeforeEach(func() {
		connection = NewConnection(server.URL(), true)
	})

	Describe("Send", func() {
		Describe("GETs", func() {
			BeforeEach(func() {
				response := `{
					"val1":"2.59.0",
					"val2":2
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v2/info"),
						RespondWith(http.StatusOK, response),
					),
				)
			})

			It("sends the request to the server", func() {
				request := Request{
					RequestName: InfoRequest,
				}

				var body DummyResponse
				response := Response{
					Result: &body,
				}

				err := connection.Make(request, &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(server.ReceivedRequests()).To(HaveLen(1))

				Expect(body.Val1).To(Equal("2.59.0"))
				Expect(body.Val2).To(Equal(2))
			})
		})

		Describe("X-Cf-Warnings", func() {
			BeforeEach(func() {
				response := `{
					"val1":"2.59.0"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v2/info"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"42, Ed McMann, the 1942 doggers"}}),
					),
				)
			})

			It("returns them in Response", func() {
				request := Request{
					RequestName: InfoRequest,
				}

				var body DummyResponse
				response := Response{
					Result: &body,
				}

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

		Describe("Errors", func() {
			Context("when the server does not exist", func() {
				BeforeEach(func() {
					connection = NewConnection("http://i.hope.this.doesnt.exist.com", false)
				})

				It("returns a RequestError", func() {
					request := Request{
						RequestName: InfoRequest,
					}

					var body DummyResponse
					response := Response{
						Result: &body,
					}

					err := connection.Make(request, &response)
					Expect(err).To(HaveOccurred())

					_, ok := err.(RequestError)
					Expect(ok).To(BeTrue())
				})
			})

			Context("when the server does not have a verified certificate", func() {
				Context("skipSSLValidation is false", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/v2/info"),
								//RespondWith(http.StatusOK, "{}"),
							),
						)

						connection = NewConnection(server.URL(), false)
					})

					It("returns a UnverifiedServerError", func() {
						request := Request{
							RequestName: InfoRequest,
						}

						var body DummyResponse
						response := Response{
							Result: &body,
						}

						err := connection.Make(request, &response)
						Expect(err).To(MatchError(UnverifiedServerError{}))
					})
				})
			})

			Describe("Status Not Found", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("GET", "/v2/info"),
							RespondWith(http.StatusNotFound, ""),
						),
					)
				})

				It("returns a ResourceNotFoundError", func() {
					request := Request{
						RequestName: InfoRequest,
					}

					var body DummyResponse
					response := Response{
						Result: &body,
					}

					err := connection.Make(request, &response)
					Expect(err).To(MatchError(ResourceNotFoundError{}))

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Describe("Status Unauthorized", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("GET", "/v2/info"),
							RespondWith(http.StatusUnauthorized, ""),
						),
					)
				})

				It("returns a UnauthorizedError", func() {
					request := Request{
						RequestName: InfoRequest,
					}

					var body DummyResponse
					response := Response{
						Result: &body,
					}

					err := connection.Make(request, &response)
					Expect(err).To(MatchError(UnauthorizedError{}))

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Describe("Status Forbidden", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("GET", "/v2/info"),
							RespondWith(http.StatusForbidden, ""),
						),
					)
				})

				It("returns a ForbiddenError", func() {
					request := Request{
						RequestName: InfoRequest,
					}

					var body DummyResponse
					response := Response{
						Result: &body,
					}

					err := connection.Make(request, &response)
					Expect(err).To(MatchError(ForbiddenError{}))

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
