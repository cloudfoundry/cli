package ccv3_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/cloudcontroller/uploads"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
					Image: "docker/some-image",
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
				requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
					requestParams.ResponseBody.(*Droplet).GUID = "some-guid"
					requestParams.ResponseBody.(*Droplet).State = "STAGED"
					requestParams.ResponseBody.(*Droplet).Image = "docker/some-image"
					requestParams.ResponseBody.(*Droplet).Stack = "some-stack"
					requestParams.ResponseBody.(*Droplet).Buildpacks = []DropletBuildpack{{Name: "some-buildpack", DetectOutput: "detected-buildpack"}}
					return "", Warnings{"warning-1"}, nil
				})
			})

			It("makes the correct request", func() {
				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.GetApplicationDropletCurrentRequest))
				Expect(actualParams.URIParams).To(Equal(internal.Params{"app_guid": "some-guid"}))

				_, ok := actualParams.ResponseBody.(*Droplet)
				Expect(ok).To(BeTrue())
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
					Image: "docker/some-image",
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

			It("returns the error and all given warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.DropletNotFoundError{}))
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
				requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
					requestParams.ResponseBody.(*Droplet).GUID = "some-guid"
					requestParams.ResponseBody.(*Droplet).State = constant.DropletStaged
					requestParams.ResponseBody.(*Droplet).Image = "docker/some-image"
					requestParams.ResponseBody.(*Droplet).Stack = "some-stack"
					requestParams.ResponseBody.(*Droplet).Buildpacks = []DropletBuildpack{{Name: "some-buildpack", DetectOutput: "detected-buildpack"}}
					return "", Warnings{"warning-1"}, nil
				})

			})

			It("makes the correct request", func() {
				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.GetDropletRequest))
				Expect(actualParams.URIParams).To(Equal(internal.Params{"droplet_guid": "some-guid"}))

				_, ok := actualParams.ResponseBody.(*Droplet)
				Expect(ok).To(BeTrue())
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
					Image: "docker/some-image",
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

	Describe("GetDroplets", func() {
		var (
			query []Query

			droplets   []Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			query = []Query{
				{Key: AppGUIDFilter, Values: []string{"some-app-guid"}},
				{Key: PerPage, Values: []string{"2"}},
			}
			droplets, warnings, executeErr = client.GetDroplets(query...)
		})

		When("the CC returns back droplets", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					requestParams.AppendToList(Droplet{
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
					})

					requestParams.AppendToList(Droplet{
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
					})

					requestParams.AppendToList(Droplet{
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
					})

					return IncludedResources{}, Warnings{"warning-1", "warning-2"}, nil
				})
			})

			It("makes the correct request", func() {
				Expect(requester.MakeListRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeListRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.GetDropletsRequest))
				Expect(actualParams.Query).To(Equal(query))
				_, ok := actualParams.ResponseBody.(Droplet)
				Expect(ok).To(BeTrue())
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
				requester.MakeListRequestReturns(
					IncludedResources{},
					Warnings{"warning-1"},
					ccerror.ApplicationNotFoundError{},
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
			dropletLength	int64
			jobURL          JobURL
			warnings        Warnings
			executeErr      error
		)

		BeforeEach(func() {
			dropletGUID = "some-droplet-guid"
			dropletContent = "some-content"
			dropletLength = int64(len(dropletContent))
			dropletFile = strings.NewReader(dropletContent)
			dropletFilePath = "some/fake-droplet.tgz"
		})

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.UploadDropletBits(dropletGUID, dropletFilePath, dropletFile, dropletLength)
		})

		FWhen("the upload is successful", func() {
			BeforeEach(func() {
				requester.MakeRequestUploadAsyncReturns("http://example.com/job-guid", Warnings{"this is a warning"}, nil)
			})

			It("makes the correct request", func() {
				contentLength, _ := uploads.CalculateRequestSize(dropletLength, dropletFilePath, "bits")

				contentType, _, _ := uploads.CreateMultipartBodyAndHeader(dropletFile, dropletFilePath, "bits")

				Expect(requester.MakeRequestUploadAsyncCallCount()).To(Equal(1))
				requestName, requestParams, contentType, bodyParam, contentLen, _, _ := requester.MakeRequestUploadAsyncArgsForCall(0)
				Expect(requestName).To(Equal(internal.PostDropletBitsRequest))
				Expect(requestParams).To(Equal(internal.Params{"droplet_guid": dropletGUID}))
				Expect(contentType).To(MatchRegexp("multipart/form-data; boundary=[\\w\\d]+"))

				requestReader := multipart.NewReader(bodyParam, contentType[30:])

				dropletPart, err := requestReader.NextPart()
				Expect(err).NotTo(HaveOccurred())

				Expect(dropletPart.FormName()).To(Equal("bits"))
				Expect(dropletPart.FileName()).To(Equal("fake-droplet.tgz"))

				partContents, err := ioutil.ReadAll(dropletPart)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(partContents)).To(Equal(dropletContent))
				dropletPart.Close()

				Expect(contentLen).To(Equal(contentLength))
			})

			It("returns the processing job URL and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(jobURL).To(Equal(JobURL("http://example.com/job-guid")))
			})
		})

		FWhen("the upload returns an error", func() {
			BeforeEach(func() {
				requester.MakeRequestUploadAsyncReturns(
					"",
					Warnings{"this is a warning"},
					ccerror.DropletNotFoundError{},
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(
					ccerror.DropletNotFoundError{},
				))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		// TODO should the pended tests be moved to requester spec
		XWhen("a retryable error occurs", func() {
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

		//TODO also this one
		XWhen("an http error occurs mid-transfer", func() {
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
