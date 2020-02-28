package ccv3_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/cloudcontroller/uploads"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Droplet", func() {
	var (
		client    *Client
		requester *ccv3fakes.FakeRequester
	)

	BeforeEach(func() {
		requester = new(ccv3fakes.FakeRequester)
		client, _ = NewFakeRequesterTestClient(requester)
	})

	Describe("CreateDroplet", func() {
		var (
			droplet    Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			droplet, warnings, executeErr = client.CreateDroplet("app-guid")
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
					requestParams.ResponseBody.(*Droplet).GUID = "some-guid"
					requestParams.ResponseBody.(*Droplet).State = "AWAITING_UPLOAD"
					requestParams.ResponseBody.(*Droplet).Image = "docker/some-image"
					requestParams.ResponseBody.(*Droplet).Stack = "some-stack"
					requestParams.ResponseBody.(*Droplet).Buildpacks = []DropletBuildpack{{Name: "some-buildpack", DetectOutput: "detected-buildpack"}}
					return "", Warnings{"warning-1"}, nil
				})
			})

			It("makes the correct request", func() {
				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.PostDropletRequest))
				Expect(actualParams.RequestBody).To(Equal(DropletCreateRequest{
					Relationships: Relationships{
						constant.RelationshipTypeApplication: Relationship{GUID: "app-guid"},
					},
				}))
				_, ok := actualParams.ResponseBody.(*Droplet)
				Expect(ok).To(BeTrue())
			})

			It("returns the given droplet and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(droplet).To(Equal(Droplet{
					GUID:  "some-guid",
					Stack: "some-stack",
					State: constant.DropletAwaitingUpload,
					Buildpacks: []DropletBuildpack{
						{
							Name:         "some-buildpack",
							DetectOutput: "detected-buildpack",
						},
					},
					Image:     "docker/some-image",
				}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		When("cloud controller returns an error", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns(
					"",
					Warnings{"warning-1"},
					ccerror.DropletNotFoundError{},
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(ccerror.DropletNotFoundError{}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})

	Describe("GetApplicationDropletCurrent", func() {
		var (
			droplet    Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			droplet, warnings, executeErr = client.GetApplicationDropletCurrent("some-guid")
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-guid",
					"state": "STAGED",
					"error": null,
					"lifecycle": {
						"type": "buildpack",
						"data": {}
					},
					"buildpacks": [
						{
							"name": "some-buildpack",
							"detect_output": "detected-buildpack"
						}
					],
					"image": "docker/some-image",
					"stack": "some-stack",
					"created_at": "2016-03-28T23:39:34Z",
					"updated_at": "2016-03-28T23:39:47Z"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-guid/droplets/current"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the given droplet and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(droplet).To(Equal(Droplet{
					GUID:  "some-guid",
					Stack: "some-stack",
					State: constant.DropletStaged,
					Buildpacks: []DropletBuildpack{
						{
							Name:         "some-buildpack",
							DetectOutput: "detected-buildpack",
						},
					},
					Image:     "docker/some-image",
					CreatedAt: "2016-03-28T23:39:34Z",
				}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		When("cloud controller returns an error", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10010,
							"detail": "Droplet not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-guid/droplets/current"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the error and all given warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.DropletNotFoundError{}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})

	Describe("GetPackageDroplets", func() {
		var (
			droplets   []Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			droplets, warnings, executeErr = client.GetPackageDroplets(
				"package-guid",
				Query{Key: PerPage, Values: []string{"2"}},
			)
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
					"pagination": {
						"next": {
							"href": "%s/v3/packages/package-guid/droplets?per_page=2&page=2"
						}
					},
					"resources": [
						{
							"guid": "some-guid-1",
							"stack": "some-stack-1",
							"buildpacks": [{
								"name": "some-buildpack-1",
								"detect_output": "detected-buildpack-1"
							}],
							"state": "STAGED",
							"created_at": "2017-08-16T00:18:24Z",
							"links": {
								"package": "https://api.com/v3/packages/package-guid"
							}
						},
						{
							"guid": "some-guid-2",
							"stack": "some-stack-2",
							"buildpacks": [{
								"name": "some-buildpack-2",
								"detect_output": "detected-buildpack-2"
							}],
							"state": "COPYING",
							"created_at": "2017-08-16T00:19:05Z"
						}
					]
				}`, server.URL())
				response2 := `{
					"pagination": {
						"next": null
					},
					"resources": [
						{
							"guid": "some-guid-3",
							"stack": "some-stack-3",
							"buildpacks": [{
								"name": "some-buildpack-3",
								"detect_output": "detected-buildpack-3"
							}],
							"state": "FAILED",
							"created_at": "2017-08-22T17:55:02Z"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/packages/package-guid/droplets", "per_page=2"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/packages/package-guid/droplets", "per_page=2&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)
			})

			It("returns the droplets", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(droplets).To(HaveLen(3))

				Expect(droplets[0]).To(Equal(Droplet{
					GUID:  "some-guid-1",
					Stack: "some-stack-1",
					State: constant.DropletStaged,
					Buildpacks: []DropletBuildpack{
						{
							Name:         "some-buildpack-1",
							DetectOutput: "detected-buildpack-1",
						},
					},
					CreatedAt: "2017-08-16T00:18:24Z",
				}))
				Expect(droplets[1]).To(Equal(Droplet{
					GUID:  "some-guid-2",
					Stack: "some-stack-2",
					State: constant.DropletCopying,
					Buildpacks: []DropletBuildpack{
						{
							Name:         "some-buildpack-2",
							DetectOutput: "detected-buildpack-2",
						},
					},
					CreatedAt: "2017-08-16T00:19:05Z",
				}))
				Expect(droplets[2]).To(Equal(Droplet{
					GUID:  "some-guid-3",
					Stack: "some-stack-3",
					State: constant.DropletFailed,
					Buildpacks: []DropletBuildpack{
						{
							Name:         "some-buildpack-3",
							DetectOutput: "detected-buildpack-3",
						},
					},
					CreatedAt: "2017-08-22T17:55:02Z",
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the cloud controller returns an error", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10010,
							"detail": "Package not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/packages/package-guid/droplets"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "Package not found",
				}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})

	Describe("GetDroplet", func() {
		var (
			droplet    Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			droplet, warnings, executeErr = client.GetDroplet("some-guid")
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-guid",
					"state": "STAGED",
					"error": null,
					"lifecycle": {
						"type": "buildpack",
						"data": {}
					},
					"buildpacks": [
						{
							"name": "some-buildpack",
							"detect_output": "detected-buildpack"
						}
					],
					"image": "docker/some-image",
					"stack": "some-stack",
					"created_at": "2016-03-28T23:39:34Z",
					"updated_at": "2016-03-28T23:39:47Z"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/droplets/some-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the given droplet and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(droplet).To(Equal(Droplet{
					GUID:  "some-guid",
					Stack: "some-stack",
					State: constant.DropletStaged,
					Buildpacks: []DropletBuildpack{
						{
							Name:         "some-buildpack",
							DetectOutput: "detected-buildpack",
						},
					},
					Image:     "docker/some-image",
					CreatedAt: "2016-03-28T23:39:34Z",
				}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		When("cloud controller returns an error", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10010,
							"detail": "Droplet not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/droplets/some-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(ccerror.DropletNotFoundError{}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})

	Describe("GetDroplets", func() {
		var (
			droplets   []Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			droplets, warnings, executeErr = client.GetDroplets(
				Query{Key: AppGUIDFilter, Values: []string{"some-app-guid"}},
				Query{Key: PerPage, Values: []string{"2"}},
			)
		})

		When("the CC returns back droplets", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
					"pagination": {
						"next": {
							"href": "%s/v3/droplets?app_guids=some-app-guid&per_page=2&page=2"
						}
					},
					"resources": [
						{
							"guid": "some-guid-1",
							"stack": "some-stack-1",
							"buildpacks": [{
								"name": "some-buildpack-1",
								"detect_output": "detected-buildpack-1"
							}],
							"state": "STAGED",
							"created_at": "2017-08-16T00:18:24Z",
							"links": {
								"package": "https://api.com/v3/packages/some-package-guid"
							}
						},
						{
							"guid": "some-guid-2",
							"stack": "some-stack-2",
							"buildpacks": [{
								"name": "some-buildpack-2",
								"detect_output": "detected-buildpack-2"
							}],
							"state": "COPYING",
							"created_at": "2017-08-16T00:19:05Z"
						}
					]
				}`, server.URL())
				response2 := `{
					"pagination": {
						"next": null
					},
					"resources": [
						{
							"guid": "some-guid-3",
							"stack": "some-stack-3",
							"buildpacks": [{
								"name": "some-buildpack-3",
								"detect_output": "detected-buildpack-3"
							}],
							"state": "FAILED",
							"created_at": "2017-08-22T17:55:02Z"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/droplets", "app_guids=some-app-guid&per_page=2"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/droplets", "app_guids=some-app-guid&per_page=2&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)
			})

			It("returns the droplets and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(droplets).To(HaveLen(3))

				Expect(droplets[0]).To(Equal(Droplet{
					GUID:  "some-guid-1",
					Stack: "some-stack-1",
					State: constant.DropletStaged,
					Buildpacks: []DropletBuildpack{
						{
							Name:         "some-buildpack-1",
							DetectOutput: "detected-buildpack-1",
						},
					},
					CreatedAt: "2017-08-16T00:18:24Z",
				}))
				Expect(droplets[1]).To(Equal(Droplet{
					GUID:  "some-guid-2",
					Stack: "some-stack-2",
					State: constant.DropletCopying,
					Buildpacks: []DropletBuildpack{
						{
							Name:         "some-buildpack-2",
							DetectOutput: "detected-buildpack-2",
						},
					},
					CreatedAt: "2017-08-16T00:19:05Z",
				}))
				Expect(droplets[2]).To(Equal(Droplet{
					GUID:  "some-guid-3",
					Stack: "some-stack-3",
					State: constant.DropletFailed,
					Buildpacks: []DropletBuildpack{
						{
							Name:         "some-buildpack-3",
							DetectOutput: "detected-buildpack-3",
						},
					},
					CreatedAt: "2017-08-22T17:55:02Z",
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("cloud controller returns an error", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10010,
							"detail": "App not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/droplets", "app_guids=some-app-guid&per_page=2"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(ccerror.ApplicationNotFoundError{}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})

	FDescribe("UploadDropletBits", func() {
		var (
			dropletGUID     string
			dropletFile     io.Reader
			dropletFilePath string
			dropletContent  string
			jobURL          JobURL
			warnings        Warnings
			executeErr      error
		)

		BeforeEach(func() {
			dropletGUID = "some-droplet-guid"
			dropletContent = "some-content"
			dropletFile = strings.NewReader(dropletContent)
			dropletFilePath = "some/fake-droplet.tgz"
		})

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.UploadDropletBits(dropletGUID, dropletFilePath, dropletFile, int64(len(dropletContent)))
		})

		FWhen("the upload is successful", func() {
			BeforeEach(func() {
				requester.MakeRequestUploadAsyncReturns("http://example.com/job-guid", Warnings{"this is a warning"}, nil)
			})

			It("makes the correct request", func() {
				contentLength, _ := uploads.CalculateRequestSize(int64(len(dropletContent)), dropletFilePath, "bits")
				//_, _, writeErrors := uploads.CreateMultipartBodyAndHeader(dropletFile, dropletFilePath, "bits")

				Expect(requester.MakeRequestUploadAsyncCallCount()).To(Equal(1))
				requestName, requestParams, contentType, _, contentLen, responseBody, _ := requester.MakeRequestUploadAsyncArgsForCall(0)
				Expect(requestName).To(Equal(internal.PostDropletBitsRequest))
				Expect(requestParams).To(Equal(internal.Params{"droplet_guid": dropletGUID}))
				Expect(contentType).To(MatchRegexp("multipart/form-data; boundary="))
				// TODO: check pieces of body that match
				//Expect(Body).To(Equal(body))
				Expect(contentLen).To(Equal(contentLength))
				Expect(responseBody).To(BeNil())
				//TODO: consider how to test this
				//Expect(writeErrs).To(Equal(writeErrors))
			})

			It("returns the processing job URL and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(jobURL).To(Equal(JobURL("http://example.com/job-guid")))
			})
		})

		FWhen("there is an error reading the buildpack", func() {
			var (
				fakeReader  *ccv3fakes.FakeReader
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("droplet read error")
				fakeReader = new(ccv3fakes.FakeReader)
				fakeReader.ReadReturns(0, expectedErr)
				dropletFile = fakeReader

				//server.AppendHandlers(
				//	VerifyRequest(http.MethodPost, "/v3/droplets/some-droplet-guid/upload"),
				//)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		When("the upload returns an error", func() {
			BeforeEach(func() {
				requester.MakeRequestUploadAsyncReturns(
					"",
					Warnings{"this is a warning"},
					ccerror.ResourceNotFoundError{
						Message: "The droplet could not be found: some-droplet-guid",
					},
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(
					ccerror.ResourceNotFoundError{
						Message: "The droplet could not be found: some-droplet-guid",
					},
				))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		When("cloud controller returns an error", func() {
			BeforeEach(func() {
				dropletGUID = "some-guid"

				requester.MakeRequestUploadAsyncReturns(
					"",
					Warnings{"warning-1"},
					ccerror.DropletNotFoundError{},
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(ccerror.DropletNotFoundError{}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		When("a retryable error occurs", func() {
			BeforeEach(func() {
				wrapper := &wrapper.CustomWrapper{
					CustomMake: func(connection cloudcontroller.Connection, request *cloudcontroller.Request, response *cloudcontroller.Response) error {
						defer GinkgoRecover() // Since this will be running in a thread

						if strings.HasSuffix(request.URL.String(), "/v3/droplets/some-droplet-guid/upload") {
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

						if strings.HasSuffix(request.URL.String(), "/v3/droplets/some-droplet-guid/upload") {
							defer request.Body.Close()
							readBytes, err := ioutil.ReadAll(request.Body)
							Expect(err).ToNot(HaveOccurred())
							Expect(len(readBytes)).To(BeNumerically(">", len(dropletContent)))
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
})
