package ccv3_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/types"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Buildpacks", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetBuildpacks", func() {
		var (
			query Query

			buildpacks []Buildpack
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			buildpacks, warnings, executeErr = client.GetBuildpacks(query)
		})

		When("buildpacks exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
					"pagination": {
						"next": {
							"href": "%s/v3/buildpacks?names=some-buildpack-name&page=2&per_page=2"
						}
					},
					"resources": [
						{
							"guid": "guid1",
							"name": "ruby_buildpack",
							"state": "AWAITING_UPLOAD",
							"stack": "windows64",
							"position": 1,
							"enabled": true,
							"locked": false
						},
						{
							"guid": "guid2",
							"name": "staticfile_buildpack",
							"state": "AWAITING_UPLOAD",
							"stack": "cflinuxfs3",
							"position": 2,
							"enabled": false,
							"locked": true
						}
					]
				}`, server.URL())
				response2 := `{
					"pagination": {
						"next": null
					},
					"resources": [
						{
							"guid": "guid3",
							"name": "go_buildpack",
							"state": "AWAITING_UPLOAD",
							"stack": "cflinuxfs2",
							"position": 3,
							"enabled": true,
							"locked": false
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/buildpacks", "names=some-buildpack-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/buildpacks", "names=some-buildpack-name&page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)

				query = Query{
					Key:    NameFilter,
					Values: []string{"some-buildpack-name"},
				}
			})

			It("returns the queried buildpacks and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(buildpacks).To(ConsistOf(
					Buildpack{
						Name:     "ruby_buildpack",
						GUID:     "guid1",
						Position: types.NullInt{Value: 1, IsSet: true},
						Enabled:  types.NullBool{Value: true, IsSet: true},
						Locked:   types.NullBool{Value: false, IsSet: true},
						Stack:    "windows64",
						State:    "AWAITING_UPLOAD",
					},
					Buildpack{
						Name:     "staticfile_buildpack",
						GUID:     "guid2",
						Position: types.NullInt{Value: 2, IsSet: true},
						Enabled:  types.NullBool{Value: false, IsSet: true},
						Locked:   types.NullBool{Value: true, IsSet: true},
						Stack:    "cflinuxfs3",
						State:    "AWAITING_UPLOAD",
					},
					Buildpack{
						Name:     "go_buildpack",
						GUID:     "guid3",
						Position: types.NullInt{Value: 3, IsSet: true},
						Enabled:  types.NullBool{Value: true, IsSet: true},
						Locked:   types.NullBool{Value: false, IsSet: true},
						Stack:    "cflinuxfs2",
						State:    "AWAITING_UPLOAD",
					},
				))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
    {
      "code": 10010,
      "detail": "buildpack not found",
      "title": "CF-buildpackNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/buildpacks"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "buildpack not found",
							Title:  "CF-buildpackNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("CreateBuildpack", func() {
		var (
			inputBuildpack Buildpack

			bp         Buildpack
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			bp, warnings, executeErr = client.CreateBuildpack(inputBuildpack)
		})

		When("the buildpack is successfully created", func() {
			BeforeEach(func() {
				inputBuildpack = Buildpack{
					Name:  "some-buildpack",
					Stack: "some-stack",
				}
				response := `{
    				"guid": "some-bp-guid",
    				"created_at": "2016-03-18T23:26:46Z",
    				"updated_at": "2016-10-17T20:00:42Z",
    				"name": "some-buildpack",
    				"state": "AWAITING_UPLOAD",
    				"filename": null,
    				"stack": "some-stack",
    				"position": 42,
    				"enabled": true,
    				"locked": false,
    				"links": {
    				  "self": {
    				    "href": "/v3/buildpacks/some-bp-guid"
    				  },
						"upload": {
							"href": "/v3/buildpacks/some-bp-guid/upload",
							"method": "POST"
						}
    				}
				}`

				expectedBody := map[string]interface{}{
					"name":  "some-buildpack",
					"stack": "some-stack",
				}
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/buildpacks"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the created buildpack and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				expectedBuildpack := Buildpack{
					GUID:     "some-bp-guid",
					Name:     "some-buildpack",
					Stack:    "some-stack",
					Enabled:  types.NullBool{Value: true, IsSet: true},
					Filename: "",
					Locked:   types.NullBool{Value: false, IsSet: true},
					State:    constant.BuildpackAwaitingUpload,
					Position: types.NullInt{Value: 42, IsSet: true},
					Links: APILinks{
						"upload": APILink{
							Method: "POST",
							HREF:   "/v3/buildpacks/some-bp-guid/upload",
						},
						"self": APILink{
							HREF: "/v3/buildpacks/some-bp-guid",
						},
					},
				}
				Expect(bp).To(Equal(expectedBuildpack))
			})
		})

		When("cc returns back an error or warnings", func() {
			BeforeEach(func() {
				inputBuildpack = Buildpack{}
				response := ` {
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
    {
      "code": 10010,
      "detail": "Buildpack not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/buildpacks"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "Buildpack not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UploadBuildpack", func() {
		var (
			jobURL     JobURL
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
			jobURL, warnings, executeErr = client.UploadBuildpack("some-buildpack-guid", bpFilePath, bpFile, int64(len(bpContent)))
		})

		When("the upload is successful", func() {
			BeforeEach(func() {
				response := `{
										"metadata": {
											"guid": "some-buildpack-guid",
											"url": "/v3/buildpacks/buildpack-guid/upload"
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

					Expect(buildpackPart.FormName()).To(Equal("bits"))
					Expect(buildpackPart.FileName()).To(Equal("fake-buildpack.zip"))

					defer buildpackPart.Close()
					partContents, err := ioutil.ReadAll(buildpackPart)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(partContents)).To(Equal(bpContent))
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/buildpacks/some-buildpack-guid/upload"),
						verifyHeaderAndBody,
						RespondWith(
							http.StatusAccepted,
							response,
							http.Header{
								"X-Cf-Warnings": {"this is a warning"},
								"Location":      {"http://example.com/job-guid"},
							},
						),
					),
				)
			})

			It("returns the processing job URL and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(jobURL).To(Equal(JobURL("http://example.com/job-guid")))
			})
		})

		When("there is an error reading the buildpack", func() {
			var (
				fakeReader  *ccv3fakes.FakeReader
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("some read error")
				fakeReader = new(ccv3fakes.FakeReader)
				fakeReader.ReadReturns(0, expectedErr)
				bpFile = fakeReader

				server.AppendHandlers(
					VerifyRequest(http.MethodPost, "/v3/buildpacks/some-buildpack-guid/upload"),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		When("the upload returns an error", func() {
			BeforeEach(func() {
				response := `{
					"errors": [{
                        "detail": "The buildpack could not be found: some-buildpack-guid",
                        "title": "CF-ResourceNotFound",
                        "code": 10010
                    }]
                }`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/buildpacks/some-buildpack-guid/upload"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(
					ccerror.ResourceNotFoundError{
						Message: "The buildpack could not be found: some-buildpack-guid",
					},
				))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		When("a retryable error occurs", func() {
			BeforeEach(func() {
				wrapper := &wrapper.CustomWrapper{
					CustomMake: func(connection cloudcontroller.Connection, request *cloudcontroller.Request, response *cloudcontroller.Response) error {
						defer GinkgoRecover() // Since this will be running in a thread

						if strings.HasSuffix(request.URL.String(), "/v3/buildpacks/some-buildpack-guid/upload") {
							_, err := ioutil.ReadAll(request.Body)
							Expect(err).ToNot(HaveOccurred())
							Expect(request.Body.Close()).ToNot(HaveOccurred())
							return request.ResetBody()
						}
						return connection.Make(request, response)
					},
				}

				client, _ = NewTestClient(Config{Wrappers: []ConnectionWrapper{wrapper}})
			})

			It("returns the PipeSeekError", func() {
				Expect(executeErr).To(MatchError(ccerror.PipeSeekError{}))
			})
		})

		When("an http error occurs mid-transfer", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some read error")

				wrapper := &wrapper.CustomWrapper{
					CustomMake: func(connection cloudcontroller.Connection, request *cloudcontroller.Request, response *cloudcontroller.Response) error {
						defer GinkgoRecover() // Since this will be running in a thread

						if strings.HasSuffix(request.URL.String(), "/v3/buildpacks/some-buildpack-guid/upload") {
							defer request.Body.Close()
							readBytes, err := ioutil.ReadAll(request.Body)
							Expect(err).ToNot(HaveOccurred())
							Expect(len(readBytes)).To(BeNumerically(">", len(bpContent)))
							return expectedErr
						}
						return connection.Make(request, response)
					},
				}

				client, _ = NewTestClient(Config{Wrappers: []ConnectionWrapper{wrapper}})
			})

			It("returns the http error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})
	})

	Describe("DeleteBuildpacks", func() {
		var (
			buildpackGUID = "some-guid"

			jobURL     JobURL
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.DeleteBuildpack(buildpackGUID)
		})

		When("buildpacks exist", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/buildpacks/"+buildpackGUID),
						RespondWith(http.StatusAccepted, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}, "Location": {"some-job-url"}}),
					),
				)
			})

			It("returns the delete job URL and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(jobURL).To(Equal(JobURL("some-job-url")))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
    {
      "code": 10010,
      "detail": "buildpack not found",
      "title": "CF-buildpackNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/buildpacks/"+buildpackGUID),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "buildpack not found",
							Title:  "CF-buildpackNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
