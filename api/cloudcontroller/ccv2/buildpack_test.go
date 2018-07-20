package ccv2_test

import (
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/ccv2fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Buildpack", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("CreateBuildpack", func() {
		var (
			buildpack  Buildpack
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			buildpack, warnings, executeErr = client.CreateBuildpack(Buildpack{
				Name:     "potato",
				Position: 1,
				Enabled:  true,
			})
		})

		Context("when the creation is successful", func() {
			BeforeEach(func() {
				response := `
				{
					"metadata": {
						"guid": "some-guid"
					},
					"entity": {
						"name": "potato",
						"stack": "null",
						"position": 1,
						"enabled": true
					}
				}`
				requestBody := map[string]interface{}{
					"name":     "potato",
					"position": 1,
					"enabled":  true,
				}
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/buildpacks"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns a buildpack and any warnings", func() {
				Expect(server.ReceivedRequests()).To(HaveLen(2))

				Expect(executeErr).ToNot(HaveOccurred())
				Expect(buildpack).To(Equal(Buildpack{
					GUID:     "some-guid",
					Name:     "potato",
					Enabled:  true,
					Position: 1,
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when the create returns an error", func() {
			BeforeEach(func() {
				response := `
					{
						"description": "Request invalid due to parse error: Field: name, Error: Missing field name",
						"error_code": "CF-MessageParseError",
						"code": 1001
					}
				`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/buildpacks"),
						RespondWith(http.StatusBadRequest, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.BadRequestError{Message: "Request invalid due to parse error: Field: name, Error: Missing field name"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("UploadBuildpack", func() {
		var (
			warnings   Warnings
			executeErr error
			bpFile     io.Reader
			bpFilePath string
			bpContent  string
		)

		BeforeEach(func() {
			bpContent = "some-content"
			bpFile = strings.NewReader(bpContent)
			bpFilePath = "some/fake-buildpack.zip"
		})

		JustBeforeEach(func() {
			warnings, executeErr = client.UploadBuildpack("some-buildpack-guid", bpFilePath, bpFile, int64(len(bpContent)))
		})

		Context("when the upload is successful", func() {
			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "some-buildpack-guid",
						"url": "/v2/buildpacks/buildpack-guid/bits"
					},
					"entity": {
						"guid": "some-buildpack-guid",
						"status": "queued"
					}
				}`

				verifyHeaderAndBody := func(_ http.ResponseWriter, req *http.Request) {
					contentType := req.Header.Get("Content-Type")
					Expect(contentType).To(MatchRegexp("multipart/form-data; boundary=[\\w\\d]+"))

					defer req.Body.Close()
					requestReader := multipart.NewReader(req.Body, contentType[30:])

					buildpackPart, err := requestReader.NextPart()
					Expect(err).NotTo(HaveOccurred())

					Expect(buildpackPart.FormName()).To(Equal("buildpack"))
					Expect(buildpackPart.FileName()).To(Equal("fake-buildpack.zip"))

					defer buildpackPart.Close()
					partContents, err := ioutil.ReadAll(buildpackPart)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(partContents)).To(Equal(bpContent))
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/buildpacks/some-buildpack-guid/bits"),
						verifyHeaderAndBody,
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns warnings", func() {
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(executeErr).ToNot(HaveOccurred())
			})
		})

		Context("when there is an error reading the buildpack", func() {
			var (
				fakeReader  *ccv2fakes.FakeReader
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("some read error")
				fakeReader = new(ccv2fakes.FakeReader)
				fakeReader.ReadReturns(0, expectedErr)
				bpFile = fakeReader

				server.AppendHandlers(
					VerifyRequest(http.MethodPut, "/v2/buildpacks/some-buildpack-guid/bits"),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		Context("when the upload returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 30003,
					"description": "The buildpack could not be found: some-buildpack-guid",
					"error_code": "CF-Banana"
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/buildpacks/some-buildpack-guid/bits"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{Message: "The buildpack could not be found: some-buildpack-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when a retryable error occurs", func() {
			BeforeEach(func() {
				wrapper := &wrapper.CustomWrapper{
					CustomMake: func(connection cloudcontroller.Connection, request *cloudcontroller.Request, response *cloudcontroller.Response) error {
						defer GinkgoRecover() // Since this will be running in a thread

						if strings.HasSuffix(request.URL.String(), "/v2/buildpacks/some-buildpack-guid/bits") {
							_, err := ioutil.ReadAll(request.Body)
							Expect(err).ToNot(HaveOccurred())
							Expect(request.Body.Close()).ToNot(HaveOccurred())
							return request.ResetBody()
						}
						return connection.Make(request, response)
					},
				}

				client = NewTestClient(Config{Wrappers: []ConnectionWrapper{wrapper}})
			})

			It("returns the PipeSeekError", func() {
				Expect(executeErr).To(MatchError(ccerror.PipeSeekError{}))
			})
		})

		Context("when an http error occurs mid-transfer", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some read error")

				wrapper := &wrapper.CustomWrapper{
					CustomMake: func(connection cloudcontroller.Connection, request *cloudcontroller.Request, response *cloudcontroller.Response) error {
						defer GinkgoRecover() // Since this will be running in a thread

						if strings.HasSuffix(request.URL.String(), "/v2/buildpacks/some-buildpack-guid/bits") {
							defer request.Body.Close()
							readBytes, err := ioutil.ReadAll(request.Body)
							Expect(err).ToNot(HaveOccurred())
							Expect(len(readBytes)).To(BeNumerically(">", len(bpContent)))
							return expectedErr
						}
						return connection.Make(request, response)
					},
				}

				client = NewTestClient(Config{Wrappers: []ConnectionWrapper{wrapper}})
			})

			It("returns the http error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})
	})

})
