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

var _ = Describe("Cloud Controller Connection", func() {
	var (
		connection *CloudControllerConnection
	)

	BeforeEach(func() {
		connection = NewConnection(server.URL(), true)
	})

	Describe("Make", func() {
		Describe("URL Generation", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v2/apps", "q=a:b&q=c:d"),
						RespondWith(http.StatusOK, "{}"),
					),
				)
			})

			Context("when passing a RequestName", func() {
				It("sends the request to the server", func() {
					request := Request{
						RequestName: AppsRequest,
						Query: FormatQueryParameters([]Query{
							{
								Filter:   "a",
								Operator: EqualOperator,
								Value:    "b",
							},
							{
								Filter:   "c",
								Operator: EqualOperator,
								Value:    "d",
							},
						}),
					}

					err := connection.Make(request, &Response{})
					Expect(err).NotTo(HaveOccurred())

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("when passing a URI", func() {
				It("sends the request to the server", func() {
					request := Request{
						URI:    "/v2/apps?q=a:b&q=c:d",
						Method: "GET",
					}

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
						VerifyRequest("GET", "/v2/apps", ""),
						VerifyHeaderKV("foo", "bar"),
						VerifyHeaderKV("accept", "application/json"),
						VerifyHeaderKV("content-type", "application/json"),
						RespondWith(http.StatusOK, "{}"),
					),
				)
			})

			Context("when passed a response with a result set", func() {
				It("unmarshals the data into a struct", func() {
					request := Request{
						URI:    "/v2/apps",
						Method: "GET",
						Header: http.Header{
							"foo": {"bar"},
						},
					}

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
						VerifyRequest("GET", "/v2/apps", ""),
						RespondWith(http.StatusOK, response),
					),
				)

				request = Request{
					URI:    "/v2/apps",
					Method: "GET",
				}
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
							VerifyRequest("GET", "/v2/info"),
							RespondWith(http.StatusOK, "{}", http.Header{"X-Cf-Warnings": {"42, Ed McMann, the 1942 doggers"}}),
						),
					)
				})

				It("returns them in Response", func() {
					request := Request{
						RequestName: InfoRequest,
					}

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
					connection = NewConnection("http://i.hope.this.doesnt.exist.com", false)
				})

				It("returns a RequestError", func() {
					request := Request{
						RequestName: InfoRequest,
					}

					var response Response
					err := connection.Make(request, &response)
					Expect(err).To(HaveOccurred())

					requestErr, ok := err.(RequestError)
					Expect(ok).To(BeTrue())
					Expect(requestErr.Error()).To(MatchRegexp(".*http://i.hope.this.doesnt.exist.com/v2/info.*[nN]o such host"))
				})
			})

			Context("when the server does not have a verified certificate", func() {
				Context("skipSSLValidation is false", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/v2/info"),
							),
						)

						connection = NewConnection(server.URL(), false)
					})

					It("returns a UnverifiedServerError", func() {
						request := Request{
							RequestName: InfoRequest,
						}

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
							VerifyRequest("GET", "/v2/info"),
							RespondWith(http.StatusNotFound, ccResponse),
						),
					)
				})

				It("returns a CCRawResponse", func() {
					request := Request{
						RequestName: InfoRequest,
					}

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
