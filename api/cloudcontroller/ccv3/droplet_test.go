package ccv3_test

import (
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
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
			droplet    resources.Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			droplet, warnings, executeErr = client.CreateDroplet("app-guid")
		})

		BeforeEach(func() {
			requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
				requestParams.ResponseBody.(*resources.Droplet).GUID = "some-guid"
				return "", Warnings{"some-warning"}, errors.New("some-error")
			})
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			actualParams := requester.MakeRequestArgsForCall(0)
			Expect(actualParams.RequestName).To(Equal(internal.PostDropletRequest))
			Expect(actualParams.RequestBody).To(Equal(DropletCreateRequest{
				Relationships: resources.Relationships{
					constant.RelationshipTypeApplication: resources.Relationship{GUID: "app-guid"},
				},
			}))
			_, ok := actualParams.ResponseBody.(*resources.Droplet)
			Expect(ok).To(BeTrue())
		})

		It("returns the given droplet and all warnings", func() {
			Expect(droplet).To(Equal(resources.Droplet{GUID: "some-guid"}))
			Expect(warnings).To(ConsistOf("some-warning"))
			Expect(executeErr).To(MatchError("some-error"))
		})
	})

	Describe("GetApplicationDropletCurrent", func() {
		var (
			droplet    resources.Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			droplet, warnings, executeErr = client.GetApplicationDropletCurrent("some-app-guid")
		})

		BeforeEach(func() {
			requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
				requestParams.ResponseBody.(*resources.Droplet).GUID = "some-guid"
				return "", Warnings{"some-warning"}, errors.New("some-error")
			})
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			actualParams := requester.MakeRequestArgsForCall(0)
			Expect(actualParams.RequestName).To(Equal(internal.GetApplicationDropletCurrentRequest))
			Expect(actualParams.URIParams).To(Equal(internal.Params{"app_guid": "some-app-guid"}))
			_, ok := actualParams.ResponseBody.(*resources.Droplet)
			Expect(ok).To(BeTrue())
		})

		It("returns the given droplet and all warnings", func() {
			Expect(droplet).To(Equal(resources.Droplet{GUID: "some-guid"}))
			Expect(warnings).To(ConsistOf("some-warning"))
			Expect(executeErr).To(MatchError("some-error"))
		})
	})

	Describe("GetPackageDroplets", func() {
		var (
			droplets   []resources.Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			droplets, warnings, executeErr = client.GetPackageDroplets(
				"package-guid",
				Query{Key: PerPage, Values: []string{"2"}},
			)
		})

		BeforeEach(func() {
			requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
				err := requestParams.AppendToList(resources.Droplet{GUID: "some-droplet-guid"})
				Expect(err).NotTo(HaveOccurred())
				return IncludedResources{}, Warnings{"some-warning"}, errors.New("some-error")
			})
		})

		It("makes the correct request", func() {
			Expect(requester.MakeListRequestCallCount()).To(Equal(1))
			actualParams := requester.MakeListRequestArgsForCall(0)
			Expect(actualParams.RequestName).To(Equal(internal.GetPackageDropletsRequest))
			Expect(actualParams.URIParams).To(Equal(internal.Params{"package_guid": "package-guid"}))
			_, ok := actualParams.ResponseBody.(resources.Droplet)
			Expect(ok).To(BeTrue())
		})

		It("returns the given droplet and all warnings", func() {
			Expect(droplets).To(Equal([]resources.Droplet{{GUID: "some-droplet-guid"}}))
			Expect(warnings).To(ConsistOf("some-warning"))
			Expect(executeErr).To(MatchError("some-error"))
		})
	})

	Describe("GetDroplet", func() {
		var (
			droplet    resources.Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			droplet, warnings, executeErr = client.GetDroplet("some-guid")
		})

		BeforeEach(func() {
			requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
				requestParams.ResponseBody.(*resources.Droplet).GUID = "some-droplet-guid"
				return "", Warnings{"some-warning"}, errors.New("some-error")
			})
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			actualParams := requester.MakeRequestArgsForCall(0)
			Expect(actualParams.RequestName).To(Equal(internal.GetDropletRequest))
			Expect(actualParams.URIParams).To(Equal(internal.Params{"droplet_guid": "some-guid"}))
			_, ok := actualParams.ResponseBody.(*resources.Droplet)
			Expect(ok).To(BeTrue())
		})

		It("returns the given droplet and all warnings", func() {
			Expect(droplet).To(Equal(resources.Droplet{GUID: "some-droplet-guid"}))
			Expect(warnings).To(ConsistOf("some-warning"))
			Expect(executeErr).To(MatchError("some-error"))
		})
	})

	Describe("GetDroplets", func() {
		var (
			droplets   []resources.Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			droplets, warnings, executeErr = client.GetDroplets(
				Query{Key: AppGUIDFilter, Values: []string{"some-app-guid"}},
				Query{Key: PerPage, Values: []string{"2"}},
			)
		})

		BeforeEach(func() {
			requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
				err := requestParams.AppendToList(resources.Droplet{GUID: "some-droplet-guid"})
				Expect(err).NotTo(HaveOccurred())
				return IncludedResources{}, Warnings{"some-warning"}, errors.New("some-error")
			})
		})

		It("makes the correct request", func() {
			Expect(requester.MakeListRequestCallCount()).To(Equal(1))
			actualParams := requester.MakeListRequestArgsForCall(0)
			Expect(actualParams.RequestName).To(Equal(internal.GetDropletsRequest))
			Expect(actualParams.Query).To(Equal([]Query{
				{Key: AppGUIDFilter, Values: []string{"some-app-guid"}},
				{Key: PerPage, Values: []string{"2"}},
			}))
			_, ok := actualParams.ResponseBody.(resources.Droplet)
			Expect(ok).To(BeTrue())
		})

		It("returns the given droplet and all warnings", func() {
			Expect(droplets).To(Equal([]resources.Droplet{{GUID: "some-droplet-guid"}}))
			Expect(warnings).To(ConsistOf("some-warning"))
			Expect(executeErr).To(MatchError("some-error"))
		})
	})

	Describe("UploadDropletBits", func() {
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

			client, _ = NewTestClient()
		})

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.UploadDropletBits(dropletGUID, dropletFilePath, dropletFile, int64(len(dropletContent)))
		})

		When("the upload is successful", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-droplet-guid",
					"state": "PROCESSING_UPLOAD"
				}`

				verifyHeaderAndBody := func(_ http.ResponseWriter, req *http.Request) {
					contentType := req.Header.Get("Content-Type")
					Expect(contentType).To(MatchRegexp("multipart/form-data; boundary=[\\w\\d]+"))

					defer req.Body.Close()
					requestReader := multipart.NewReader(req.Body, contentType[30:])

					dropletPart, err := requestReader.NextPart()
					Expect(err).NotTo(HaveOccurred())

					Expect(dropletPart.FormName()).To(Equal("bits"))
					Expect(dropletPart.FileName()).To(Equal("fake-droplet.tgz"))

					defer dropletPart.Close()
					partContents, err := ioutil.ReadAll(dropletPart)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(partContents)).To(Equal(dropletContent))
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/droplets/some-droplet-guid/upload"),
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
				expectedErr = errors.New("droplet read error")
				fakeReader = new(ccv3fakes.FakeReader)
				fakeReader.ReadReturns(0, expectedErr)
				dropletFile = fakeReader

				server.AppendHandlers(
					VerifyRequest(http.MethodPost, "/v3/droplets/some-droplet-guid/upload"),
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
                        "detail": "The droplet could not be found: some-droplet-guid",
                        "title": "CF-ResourceNotFound",
                        "code": 10010
                    }]
                }`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/droplets/some-droplet-guid/upload"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
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
						VerifyRequest(http.MethodPost, "/v3/droplets/some-guid/upload"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
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

	Describe("DownloadDroplet", func() {
		var (
			dropletBytes []byte
			warnings     Warnings
			executeErr   error
		)

		JustBeforeEach(func() {
			dropletBytes, warnings, executeErr = client.DownloadDroplet("some-droplet-guid")
		})

		BeforeEach(func() {
			requester.MakeRequestReceiveRawCalls(func(string, internal.Params, string) ([]byte, ccv3.Warnings, error) {
				return []byte{'d', 'r', 'o', 'p'}, Warnings{"some-warning"}, errors.New("some-error")
			})
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestReceiveRawCallCount()).To(Equal(1))
			requestType, requestParams, responseType := requester.MakeRequestReceiveRawArgsForCall(0)
			Expect(requestType).To(Equal(internal.GetDropletBitsRequest))
			Expect(requestParams).To(Equal(internal.Params{"droplet_guid": "some-droplet-guid"}))
			Expect(responseType).To(Equal("application/json"))
		})

		It("returns the given droplet and all warnings", func() {
			Expect(dropletBytes).To(Equal([]byte{'d', 'r', 'o', 'p'}))
			Expect(warnings).To(ConsistOf("some-warning"))
			Expect(executeErr).To(MatchError("some-error"))
		})
	})
})
