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
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/types"
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
			inputBuildpack Buildpack

			resultBuildpack Buildpack
			warnings        Warnings
			executeErr      error
		)

		JustBeforeEach(func() {
			resultBuildpack, warnings, executeErr = client.CreateBuildpack(inputBuildpack)
		})

		Context("when the creation is successful", func() {
			Context("when all the properties are passed", func() {
				BeforeEach(func() {
					inputBuildpack = Buildpack{
						Name:     "potato",
						Position: types.NullInt{IsSet: true, Value: 1},
						Enabled:  types.NullBool{IsSet: true, Value: true},
						Stack:    "foobar",
					}

					response := `
				{
					"metadata": {
						"guid": "some-guid"
					},
					"entity": {
						"name": "potato",
						"stack": "foobar",
						"position": 1,
						"enabled": true
					}
				}`
					requestBody := map[string]interface{}{
						"name":     "potato",
						"position": 1,
						"enabled":  true,
						"stack":    "foobar",
					}
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/buildpacks"),
							VerifyJSONRepresenting(requestBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("creates a buildpack and returns it with any warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					validateV2InfoPlusNumberOfRequests(1)

					Expect(resultBuildpack).To(Equal(Buildpack{
						GUID:     "some-guid",
						Name:     "potato",
						Enabled:  types.NullBool{IsSet: true, Value: true},
						Position: types.NullInt{IsSet: true, Value: 1},
						Stack:    "foobar",
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})

			Context("when the minimum properties are being passed", func() {
				BeforeEach(func() {
					inputBuildpack = Buildpack{
						Name: "potato",
					}

					response := `
				{
					"metadata": {
						"guid": "some-guid"
					},
					"entity": {
						"name": "potato",
						"stack": null,
						"position": 10000,
						"enabled": true
					}
				}`
					requestBody := map[string]interface{}{
						"name": "potato",
					}
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/buildpacks"),
							VerifyJSONRepresenting(requestBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("creates a buildpack and returns it with any warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					validateV2InfoPlusNumberOfRequests(1)

					Expect(resultBuildpack).To(Equal(Buildpack{
						GUID:     "some-guid",
						Name:     "potato",
						Enabled:  types.NullBool{IsSet: true, Value: true},
						Position: types.NullInt{IsSet: true, Value: 10000},
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
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

	Describe("GetBuildpacks", func() {
		var (
			buildpacks []Buildpack
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			bpName := Filter{
				Type:     constant.NameFilter,
				Operator: constant.EqualOperator,
				Values:   []string{"some-bp-name"},
			}

			buildpacks, warnings, executeErr = client.GetBuildpacks(bpName)
		})

		Context("when buildpacks are found", func() {
			BeforeEach(func() {
				response1 := `{
										"next_url": "/v2/buildpacks?q=name:some-bp-name",
										"resources": [
											{
												"metadata": {
												"guid": "some-bp-guid1"
												},
												"entity": {
													"name": "some-bp-name1",
													"stack": null,
													"position": 2,
													"enabled": true
												}
											},
											{
												"metadata": {
													"guid": "some-bp-guid2"
												},
												"entity": {
													"name": "some-bp-name2",
													"stack": null,
													"position": 3,
													"enabled": false
												}
											}
										]
									}`
				response2 := `{
										"next_url": null,
										"resources": [
											{
												"metadata": {
													"guid": "some-bp-guid3"
												},
												"entity": {
													"name": "some-bp-name3",
													"stack": "cflinuxfs2",
													"position": 4,
													"enabled": true
												}
											}
										]
									}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/buildpacks", "q=name:some-bp-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"first warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/buildpacks", "q=name:some-bp-name"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"second warning"}}),
					),
				)
			})

			It("returns the buildpacks", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(buildpacks).To(Equal([]Buildpack{
					{
						Name:     "some-bp-name1",
						GUID:     "some-bp-guid1",
						Enabled:  types.NullBool{IsSet: true, Value: true},
						Position: types.NullInt{IsSet: true, Value: 2},
						Stack:    "",
					},
					{
						Name:     "some-bp-name2",
						GUID:     "some-bp-guid2",
						Enabled:  types.NullBool{IsSet: true, Value: false},
						Position: types.NullInt{IsSet: true, Value: 3},
						Stack:    "",
					},
					{
						Name:     "some-bp-name3",
						GUID:     "some-bp-guid3",
						Enabled:  types.NullBool{IsSet: true, Value: true},
						Position: types.NullInt{IsSet: true, Value: 4},
						Stack:    "cflinuxfs2",
					},
				}))

				Expect(warnings).To(ConsistOf(Warnings{"first warning", "second warning"}))
			})
		})

		Context("when no buildpacks are found", func() {
			BeforeEach(func() {
				response := `{
										"resources": []
									}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/buildpacks", "q=name:some-bp-name"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an empty list", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(buildpacks).To(HaveLen(0))
			})
		})

		Context("when the API responds with an error", func() {
			BeforeEach(func() {
				response := `{
										"code": 10001,
										"description": "Whoops",
										"error_code": "CF-SomeError"
									}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/buildpacks", "q=name:some-bp-name"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns warnings and the error", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Whoops",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("UpdateBuildpack", func() {
		var (
			buildpack        Buildpack
			updatedBuildpack Buildpack
			warnings         Warnings
			executeErr       error
		)

		JustBeforeEach(func() {
			updatedBuildpack, warnings, executeErr = client.UpdateBuildpack(buildpack)
		})

		Context("when the buildpack exists", func() {
			Context("when all the properties are provided", func() {
				Context("when the provided properties are golang non-zero values", func() {
					BeforeEach(func() {
						buildpack = Buildpack{
							Name:     "some-bp-name",
							Position: types.NullInt{IsSet: true, Value: 10},
							Enabled:  types.NullBool{IsSet: true, Value: true},
							GUID:     "some-bp-guid",
						}

						response := `
										{
											"metadata": {
											     "guid": "some-bp-guid"
											},
											"entity": {
												"name": "some-bp-name",
												"stack": null,
												"position": 10,
												"enabled": true
											}
										}
									`

						requestBody := map[string]interface{}{
							"name":     "some-bp-name",
							"position": 10,
							"enabled":  true,
						}

						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodPut, "/v2/buildpacks/some-bp-guid"),
								VerifyJSONRepresenting(requestBody),
								RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
							),
						)
					})

					It("updates and returns the updated buildpack", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						validateV2InfoPlusNumberOfRequests(1)
						Expect(warnings).To(ConsistOf("this is a warning"))
						Expect(updatedBuildpack).To(Equal(buildpack))
					})
				})

				Context("when the provided properties are golang zero values", func() {
					BeforeEach(func() {
						buildpack = Buildpack{
							Name:     "some-bp-name",
							GUID:     "some-bp-guid",
							Position: types.NullInt{IsSet: true, Value: 0},
							Enabled:  types.NullBool{IsSet: true, Value: false},
						}

						response := `
										{
											"metadata": {
											"guid": "some-bp-guid"
											},
											"entity": {
												"name": "some-bp-name",
												"stack": null,
												"position": 0,
												"enabled": false
											}
										}
									`
						requestBody := map[string]interface{}{
							"name":     "some-bp-name",
							"position": 0,
							"enabled":  false,
						}

						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodPut, "/v2/buildpacks/some-bp-guid"),
								VerifyJSONRepresenting(requestBody),
								RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
							),
						)
					})

					It("updates and returns the updated buildpack", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						validateV2InfoPlusNumberOfRequests(1)
						Expect(warnings).To(ConsistOf("this is a warning"))
						Expect(updatedBuildpack).To(Equal(buildpack))
					})
				})
			})
		})

		Context("when the buildpack does not exist", func() {
			BeforeEach(func() {
				response := `{
										"description": "buildpack not found",
										"error_code": "CF-NotFound",
										"code": 10000
									}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/buildpacks/some-bp-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

				buildpack = Buildpack{
					GUID: "some-bp-guid",
				}

			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "buildpack not found",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when the API errors", func() {
			BeforeEach(func() {
				response := `{
										"code": 10001,
										"description": "Some Error",
										"error_code": "CF-SomeError"
									}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/buildpacks/some-bp-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

				buildpack = Buildpack{
					GUID: "some-bp-guid",
				}
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
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
