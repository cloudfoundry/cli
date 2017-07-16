package plugin_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	. "code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	"code.cloudfoundry.org/cli/api/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

type DummyResponse struct {
	Val1 string      `json:"val1"`
	Val2 int         `json:"val2"`
	Val3 interface{} `json:"val3,omitempty"`
}

var _ = Describe("Plugin Connection", func() {
	var (
		connection      *PluginConnection
		fakeProxyReader *pluginfakes.FakeProxyReader
	)

	BeforeEach(func() {
		connection = NewConnection(true, 0)
		fakeProxyReader = new(pluginfakes.FakeProxyReader)

		fakeProxyReader.WrapStub = func(reader io.Reader) io.ReadCloser {
			return ioutil.NopCloser(reader)
		}
	})

	Describe("Make", func() {
		Describe("Data Unmarshalling", func() {
			var (
				request      *http.Request
				responseBody string
			)

			BeforeEach(func() {
				responseBody = `{
					"val1":"2.59.0",
					"val2":2,
					"val3":1111111111111111111
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/list", ""),
						RespondWith(http.StatusOK, responseBody),
					),
				)

				var err error
				request, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/list", server.URL()), nil)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when passed a response with a result set", func() {
				It("unmarshals the data into a struct", func() {
					var body DummyResponse
					response := Response{
						Result: &body,
					}

					err := connection.Make(request, &response, fakeProxyReader)
					Expect(err).NotTo(HaveOccurred())

					Expect(body.Val1).To(Equal("2.59.0"))
					Expect(body.Val2).To(Equal(2))

					Expect(fakeProxyReader.StartCallCount()).To(Equal(1))
					Expect(fakeProxyReader.StartArgsForCall(0)).To(BeEquivalentTo(len(responseBody)))

					Expect(fakeProxyReader.WrapCallCount()).To(Equal(1))

					Expect(fakeProxyReader.FinishCallCount()).To(Equal(1))
				})

				It("keeps numbers unmarshalled to interfaces as interfaces", func() {
					var body DummyResponse
					response := Response{
						Result: &body,
					}

					err := connection.Make(request, &response, nil)
					Expect(err).NotTo(HaveOccurred())
					Expect(fmt.Sprint(body.Val3)).To(Equal("1111111111111111111"))
				})
			})

			Context("when passed an empty response", func() {
				It("skips the unmarshalling step", func() {
					var response Response
					err := connection.Make(request, &response, nil)
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
						VerifyRequest(http.MethodGet, "/list", ""),
						RespondWith(http.StatusOK, response),
					),
				)

				var err error
				request, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/list", server.URL()), nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the status", func() {
				response := Response{}

				err := connection.Make(request, &response, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.HTTPResponse.Status).To(Equal("200 OK"))
			})
		})

		Describe("Request errors", func() {
			Context("when the server does not exist", func() {
				BeforeEach(func() {
					connection = NewConnection(false, 0)
				})

				It("returns a RequestError", func() {
					request, err := http.NewRequest(http.MethodGet, "http://i.hope.this.doesnt.exist.com/list", nil)
					Expect(err).ToNot(HaveOccurred())

					var response Response
					err = connection.Make(request, &response, nil)
					Expect(err).To(HaveOccurred())

					requestErr, ok := err.(pluginerror.RequestError)
					Expect(ok).To(BeTrue())
					Expect(requestErr.Error()).To(MatchRegexp(".*http://i.hope.this.doesnt.exist.com/list.*"))
				})
			})

			Context("when the server does not have a verified certificate", func() {
				Context("skipSSLValidation is false", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/list"),
							),
						)

						connection = NewConnection(false, 0)
					})

					It("returns a UnverifiedServerError", func() {
						request, err := http.NewRequest(http.MethodGet, server.URL(), nil)
						Expect(err).ToNot(HaveOccurred())

						var response Response
						err = connection.Make(request, &response, nil)
						Expect(err).To(MatchError(pluginerror.UnverifiedServerError{URL: server.URL()}))
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

						connection = NewConnection(false, 0)
					})

					// loopback.cli.ci.cf-app.com is a custom DNS record setup to point to 127.0.0.1
					It("returns a SSLValidationHostnameError", func() {
						altHostURL := strings.Replace(server.URL(), "127.0.0.1", "loopback.cli.ci.cf-app.com", -1)
						request, err := http.NewRequest(http.MethodGet, altHostURL, nil)
						Expect(err).ToNot(HaveOccurred())

						var response Response
						err = connection.Make(request, &response, nil)
						Expect(err).To(MatchError(pluginerror.SSLValidationHostnameError{
							Message: "x509: certificate is valid for example.com, not loopback.cli.ci.cf-app.com",
						}))
					})
				})
			})
		})

		Describe("4xx and 5xx response codes", func() {
			Context("when any 4xx or 5xx response codes are encountered", func() {
				var rawResponse string

				BeforeEach(func() {
					rawResponse = `{
						"error":"some error"
						"description": "some error description",
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/list"),
							RespondWith(http.StatusTeapot, rawResponse),
						),
					)
				})

				It("returns a RawHTTPStatusError", func() {
					request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/list", server.URL()), nil)
					Expect(err).ToNot(HaveOccurred())

					var response Response
					err = connection.Make(request, &response, nil)
					Expect(err).To(MatchError(pluginerror.RawHTTPStatusError{
						Status:      "418 I'm a teapot",
						RawResponse: []byte(rawResponse),
					}))

					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
