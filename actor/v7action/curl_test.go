package v7action_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Curl Actions", func() {
	Describe("MakeCurlRequest", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			fakeConfig                *v7actionfakes.FakeConfig

			method          string
			path            string
			customHeaders   []string
			data            string
			failOnHTTPError bool

			mockResponseBody []byte
			mockHTTPResponse *http.Response
			mockErr          error

			responseBody []byte
			httpResponse *http.Response
			executeErr   error
		)

		BeforeEach(func() {
			actor, fakeCloudControllerClient, fakeConfig, _, _, _, _ = NewTestActor()

			fakeConfig.TargetReturns("api.com")

			mockResponseBody = []byte(`{"response":"yep"}`)
			mockHTTPResponse = &http.Response{}
			mockErr = nil

			method = ""
			path = "/v3/is/great"
			customHeaders = []string{}
			data = ""
			failOnHTTPError = false
		})

		JustBeforeEach(func() {
			fakeCloudControllerClient.MakeRequestSendReceiveRawReturns(
				mockResponseBody,
				mockHTTPResponse,
				mockErr,
			)

			responseBody, httpResponse, executeErr = actor.MakeCurlRequest(method, path, customHeaders, data, failOnHTTPError)
		})

		When("no method is given", func() {
			It("makes the request", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeCloudControllerClient.MakeRequestSendReceiveRawCallCount()).To(Equal(1))

				givenMethod, givenURL, givenHeaders, givenBody := fakeCloudControllerClient.MakeRequestSendReceiveRawArgsForCall(0)
				Expect(givenMethod).To(Equal(""))
				Expect(givenURL).To(Equal("api.com/v3/is/great"))
				Expect(givenHeaders).To(Equal(http.Header{}))
				Expect(givenBody).To(Equal([]byte("")))

				Expect(responseBody).To(Equal(mockResponseBody))
				Expect(httpResponse).To(Equal(mockHTTPResponse))
			})
		})

		When("method and data are given", func() {
			BeforeEach(func() {
				method = "PATCH"
				data = `{"name": "cool"}`
			})

			It("uses the given method", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeCloudControllerClient.MakeRequestSendReceiveRawCallCount()).To(Equal(1))

				givenMethod, givenURL, givenHeaders, givenBody := fakeCloudControllerClient.MakeRequestSendReceiveRawArgsForCall(0)
				Expect(givenMethod).To(Equal("PATCH"))
				Expect(givenURL).To(Equal("api.com/v3/is/great"))
				Expect(givenHeaders).To(Equal(http.Header{}))
				Expect(givenBody).To(Equal([]byte(`{"name": "cool"}`)))

				Expect(responseBody).To(Equal(mockResponseBody))
				Expect(httpResponse).To(Equal(mockHTTPResponse))
			})
		})

		When("data is given, but no method", func() {
			BeforeEach(func() {
				data = `{"name": "cool"}`
			})

			It("uses method POST", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeCloudControllerClient.MakeRequestSendReceiveRawCallCount()).To(Equal(1))

				givenMethod, givenURL, givenHeaders, givenBody := fakeCloudControllerClient.MakeRequestSendReceiveRawArgsForCall(0)
				Expect(givenMethod).To(Equal("POST"))
				Expect(givenURL).To(Equal("api.com/v3/is/great"))
				Expect(givenHeaders).To(Equal(http.Header{}))
				Expect(givenBody).To(Equal([]byte(`{"name": "cool"}`)))

				Expect(responseBody).To(Equal(mockResponseBody))
				Expect(httpResponse).To(Equal(mockHTTPResponse))
			})
		})

		When("data is a path to a file", func() {
			When("the file exists", func() {
				var tempFile string

				BeforeEach(func() {
					file, err := ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())
					tempFile = file.Name()

					_, err = file.WriteString(`{"file": "wow"}`)
					Expect(err).NotTo(HaveOccurred())

					data = "@" + tempFile
				})

				AfterEach(func() {
					os.RemoveAll(tempFile)
				})

				It("reads the file and uses the contents as the request body", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(fakeCloudControllerClient.MakeRequestSendReceiveRawCallCount()).To(Equal(1))

					givenMethod, givenURL, givenHeaders, givenBody := fakeCloudControllerClient.MakeRequestSendReceiveRawArgsForCall(0)
					Expect(givenMethod).To(Equal("POST"))
					Expect(givenURL).To(Equal("api.com/v3/is/great"))
					Expect(givenHeaders).To(Equal(http.Header{}))
					Expect(givenBody).To(Equal([]byte(`{"file": "wow"}`)))

					Expect(responseBody).To(Equal(mockResponseBody))
					Expect(httpResponse).To(Equal(mockHTTPResponse))
				})
			})

			When("the file exists, but the given path has extra quotes", func() {
				var tempFile string

				BeforeEach(func() {
					file, err := ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())
					tempFile = file.Name()

					_, err = file.WriteString(`{"file": "wow"}`)
					Expect(err).NotTo(HaveOccurred())

					data = `'@"` + tempFile + `"'`
				})

				AfterEach(func() {
					os.RemoveAll(tempFile)
				})

				It("reads the file and uses the contents as the request body", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(fakeCloudControllerClient.MakeRequestSendReceiveRawCallCount()).To(Equal(1))

					givenMethod, givenURL, givenHeaders, givenBody := fakeCloudControllerClient.MakeRequestSendReceiveRawArgsForCall(0)
					Expect(givenMethod).To(Equal("POST"))
					Expect(givenURL).To(Equal("api.com/v3/is/great"))
					Expect(givenHeaders).To(Equal(http.Header{}))
					Expect(givenBody).To(Equal([]byte(`{"file": "wow"}`)))

					Expect(responseBody).To(Equal(mockResponseBody))
					Expect(httpResponse).To(Equal(mockHTTPResponse))
				})
			})

			When("the file doesn't exist", func() {
				BeforeEach(func() {
					data = "@/not/real"
				})

				It("returns a helpful error", func() {
					Expect(executeErr.Error()).To(MatchRegexp("Error creating request"))
				})
			})
		})

		When("valid headers are given", func() {
			BeforeEach(func() {
				customHeaders = append(customHeaders, "X-Wow: Amazing")
				customHeaders = append(customHeaders, "X-Wow: Cool")
				customHeaders = append(customHeaders, "X-Great: Yeah")
			})

			It("uses those headers for the request", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeCloudControllerClient.MakeRequestSendReceiveRawCallCount()).To(Equal(1))

				givenMethod, givenURL, givenHeaders, givenBody := fakeCloudControllerClient.MakeRequestSendReceiveRawArgsForCall(0)
				Expect(givenMethod).To(Equal(""))
				Expect(givenURL).To(Equal("api.com/v3/is/great"))
				Expect(givenHeaders).To(Equal(http.Header{
					"X-Wow":   {"Amazing", "Cool"},
					"X-Great": {"Yeah"},
				}))
				Expect(givenBody).To(Equal([]byte("")))
			})
		})

		When("invalid headers are given", func() {
			BeforeEach(func() {
				customHeaders = append(customHeaders, "notformattedcorrectly")
			})

			It("returns a helpful error", func() {
				Expect(executeErr.Error()).To(MatchRegexp("Error creating request"))
			})
		})

		When("the request fails", func() {
			BeforeEach(func() {
				mockErr = errors.New("uh oh")
				mockHTTPResponse.StatusCode = 500
			})

			It("does not return the error by default", func() {
				Expect(executeErr).NotTo(HaveOccurred())
			})

			When("the fail-on-http-errors flag is set", func() {
				BeforeEach(func() {
					failOnHTTPError = true
				})

				It("returns an error containing the status code", func() {
					Expect(executeErr).To(MatchError(translatableerror.CurlExit22Error{StatusCode: 500}))
				})
			})
		})
	})
})
