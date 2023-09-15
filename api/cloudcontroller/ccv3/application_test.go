package ccv3_test

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application", func() {
	var (
		client    *Client
		requester *ccv3fakes.FakeRequester
	)

	BeforeEach(func() {
		requester = new(ccv3fakes.FakeRequester)
		client, _ = NewFakeRequesterTestClient(requester)
	})

	Describe("Application", func() {
		Describe("MarshalJSON", func() {
			var (
				app      resources.Application
				appBytes []byte
				err      error
			)

			BeforeEach(func() {
				app = resources.Application{}
			})

			JustBeforeEach(func() {
				appBytes, err = app.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
			})

			When("no lifecycle is provided", func() {
				BeforeEach(func() {
					app = resources.Application{}
				})

				It("omits the lifecycle from the JSON", func() {
					Expect(string(appBytes)).To(Equal("{}"))
				})
			})

			When("lifecycle type docker is provided", func() {
				BeforeEach(func() {
					app = resources.Application{
						LifecycleType: constant.AppLifecycleTypeDocker,
					}
				})

				It("sets lifecycle type to docker with empty data", func() {
					Expect(string(appBytes)).To(MatchJSON(`{"lifecycle":{"type":"docker","data":{}}}`))
				})
			})

			When("lifecycle type buildpack is provided", func() {
				BeforeEach(func() {
					app.LifecycleType = constant.AppLifecycleTypeBuildpack
				})

				When("no buildpacks are provided", func() {
					It("omits the lifecycle from the JSON", func() {
						Expect(string(appBytes)).To(MatchJSON("{}"))
					})

					When("but you do specify a stack", func() {
						BeforeEach(func() {
							app.StackName = "cflinuxfs9000"
						})

						It("does, in fact, send the stack in the json", func() {
							Expect(string(appBytes)).To(MatchJSON(`{"lifecycle":{"data":{"stack":"cflinuxfs9000"},"type":"buildpack"}}`))
						})
					})
				})

				When("default buildpack is provided", func() {
					BeforeEach(func() {
						app.LifecycleBuildpacks = []string{"default"}
					})

					It("sets the lifecycle buildpack to be empty in the JSON", func() {
						Expect(string(appBytes)).To(MatchJSON(`{"lifecycle":{"data":{"buildpacks":null},"type":"buildpack"}}`))
					})
				})

				When("null buildpack is provided", func() {
					BeforeEach(func() {
						app.LifecycleBuildpacks = []string{"null"}
					})

					It("sets the Lifecycle buildpack to be empty in the JSON", func() {
						Expect(string(appBytes)).To(MatchJSON(`{"lifecycle":{"data":{"buildpacks":null},"type":"buildpack"}}`))
					})
				})

				When("other buildpacks are provided", func() {
					BeforeEach(func() {
						app.LifecycleBuildpacks = []string{"some-buildpack"}
					})

					It("sets them in the JSON", func() {
						Expect(string(appBytes)).To(MatchJSON(`{"lifecycle":{"data":{"buildpacks":["some-buildpack"]},"type":"buildpack"}}`))
					})
				})
			})

			When("metadata is provided", func() {
				BeforeEach(func() {
					app = resources.Application{
						Metadata: &resources.Metadata{
							Labels: map[string]types.NullString{
								"some-key":  types.NewNullString("some-value"),
								"other-key": types.NewNullString("other-value\nwith a newline & a \" quote")},
						},
					}
				})

				It("should include the labels in the JSON", func() {
					Expect(string(appBytes)).To(MatchJSON(`{
						"metadata": {
							"labels": {
								"some-key":"some-value",
								"other-key":"other-value\nwith a newline & a \" quote"
							}
						}
					}`))
				})

				When("labels need to be removed", func() {
					BeforeEach(func() {
						app = resources.Application{
							Metadata: &resources.Metadata{
								Labels: map[string]types.NullString{
									"some-key":      types.NewNullString("some-value"),
									"other-key":     types.NewNullString("other-value\nwith a newline & a \" quote"),
									"key-to-delete": types.NewNullString(),
								},
							},
						}
					})

					It("should send nulls for those labels", func() {
						Expect(string(appBytes)).To(MatchJSON(`{
						"metadata": {
							"labels": {
								"some-key":"some-value",
								"other-key":"other-value\nwith a newline & a \" quote",
								"key-to-delete":null
							}
						}
					}`))
					})
				})
			})
		})

		Describe("UnmarshalJSON", func() {
			var (
				app      resources.Application
				appBytes []byte
				err      error
			)

			BeforeEach(func() {
				appBytes = []byte("{}")
				app = resources.Application{}
			})

			JustBeforeEach(func() {
				err = json.Unmarshal(appBytes, &app)
				Expect(err).ToNot(HaveOccurred())
			})

			When("no lifecycle is provided", func() {
				BeforeEach(func() {
					appBytes = []byte("{}")
				})

				It("omits the lifecycle from the JSON", func() {
					Expect(app).To(Equal(resources.Application{}))
				})
			})

			When("lifecycle type docker is provided", func() {
				BeforeEach(func() {
					appBytes = []byte(`{"lifecycle":{"type":"docker","data":{}}}`)
				})
				It("sets the lifecycle type to docker with empty data", func() {
					Expect(app).To(Equal(resources.Application{
						LifecycleType: constant.AppLifecycleTypeDocker,
					}))
				})
			})

			When("lifecycle type buildpack is provided", func() {

				When("other buildpacks are provided", func() {
					BeforeEach(func() {
						appBytes = []byte(`{"lifecycle":{"data":{"buildpacks":["some-buildpack"]},"type":"buildpack"}}`)
					})

					It("sets them in the JSON", func() {
						Expect(app).To(Equal(resources.Application{
							LifecycleType:       constant.AppLifecycleTypeBuildpack,
							LifecycleBuildpacks: []string{"some-buildpack"},
						}))
					})
				})
			})

			When("labels are provided", func() {
				BeforeEach(func() {
					appBytes = []byte(`{"metadata":{"labels":{"some-key":"some-value"}}}`)
				})

				It("sets the labels", func() {
					Expect(app).To(Equal(resources.Application{
						Metadata: &resources.Metadata{
							Labels: map[string]types.NullString{
								"some-key": types.NewNullString("some-value"),
							},
						},
					}))
				})
			})
		})
	})

	Describe("CreateApplication", func() {
		var (
			appToCreate resources.Application

			createdApp resources.Application
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			createdApp, warnings, executeErr = client.CreateApplication(appToCreate)
		})

		When("the application successfully is created", func() {
			BeforeEach(func() {
				requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
					requestParams.ResponseBody.(*resources.Application).GUID = "some-app-guid"
					requestParams.ResponseBody.(*resources.Application).Name = requestParams.RequestBody.(resources.Application).Name
					return "", Warnings{"this is a warning"}, nil
				})
				appToCreate = resources.Application{
					Name:      "some-app-name",
					SpaceGUID: "some-space-guid",
				}
			})

			It("makes the correct request", func() {
				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.PostApplicationRequest))
				Expect(actualParams.RequestBody).To(Equal(appToCreate))
				_, ok := actualParams.ResponseBody.(*resources.Application)
				Expect(ok).To(BeTrue())
			})

			It("returns the created app and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(createdApp).To(Equal(resources.Application{
					Name: "some-app-name",
					GUID: "some-app-guid",
				}))
			})
		})

		When("cc returns back an error or warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					},
					{
						Code:   10010,
						Detail: "App not found",
						Title:  "CF-ResourceNotFound",
					},
				}

				requester.MakeRequestReturns(
					"",
					Warnings{"this is a warning"},
					ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors},
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
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetApplicationByNameAndSpace", func() {
		var (
			appName   = "some-app-name"
			spaceGUID = "some-space-guid"

			app        resources.Application
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			app, warnings, executeErr = client.GetApplicationByNameAndSpace(appName, spaceGUID)
		})

		When("the application exists", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					err := requestParams.AppendToList(resources.Application{GUID: "app-guid-2"})
					Expect(err).NotTo(HaveOccurred())
					return IncludedResources{}, Warnings{"this is a warning"}, nil
				})
			})

			It("makes the correct request", func() {
				Expect(requester.MakeListRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeListRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.GetApplicationsRequest))
				Expect(actualParams.Query).To(Equal([]Query{
					{Key: NameFilter, Values: []string{"some-app-name"}},
					{Key: SpaceGUIDFilter, Values: []string{"some-space-guid"}},
				}))
				_, ok := actualParams.ResponseBody.(resources.Application)
				Expect(ok).To(BeTrue())
			})

			It("returns the queried application and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(app).To(Equal(resources.Application{GUID: "app-guid-2"}))
			})
		})

		When("the application does not exist", func() {
			BeforeEach(func() {
				requester.MakeListRequestReturns(IncludedResources{}, Warnings{"this is a warning"}, nil)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ApplicationNotFoundError{Name: appName}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					},
					{
						Code:   10010,
						Detail: "App not found",
						Title:  "CF-ResourceNotFound",
					},
				}

				requester.MakeListRequestReturns(
					IncludedResources{},
					Warnings{"this is a warning"},
					ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors},
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
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetApplications", func() {
		var (
			filters []Query

			apps       []resources.Application
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			apps, warnings, executeErr = client.GetApplications(filters...)
		})

		When("applications exist", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					err := requestParams.AppendToList(resources.Application{GUID: "app-guid-1"})
					Expect(err).NotTo(HaveOccurred())
					return IncludedResources{}, Warnings{"this is a warning", "this is another warning"}, nil
				})

				filters = []Query{
					{Key: SpaceGUIDFilter, Values: []string{"some-space-guid"}},
					{Key: NameFilter, Values: []string{"some-app-name"}},
				}
			})

			It("makes the correct request", func() {
				Expect(requester.MakeListRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeListRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.GetApplicationsRequest))
				Expect(actualParams.Query).To(Equal(filters))
				_, ok := actualParams.ResponseBody.(resources.Application)
				Expect(ok).To(BeTrue())
			})

			It("returns the queried applications and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))

				Expect(apps).To(ConsistOf(resources.Application{GUID: "app-guid-1"}))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					},
					{
						Code:   10010,
						Detail: "App not found",
						Title:  "CF-ResourceNotFound",
					},
				}

				requester.MakeListRequestReturns(
					IncludedResources{},
					Warnings{"this is a warning"},
					ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors},
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
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateApplication", func() {
		var (
			appToUpdate resources.Application

			updatedApp resources.Application
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			updatedApp, warnings, executeErr = client.UpdateApplication(appToUpdate)
		})

		When("the application successfully is updated", func() {
			BeforeEach(func() {
				requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
					requestParams.ResponseBody.(*resources.Application).GUID = "some-app-guid"
					requestParams.ResponseBody.(*resources.Application).Name = requestParams.RequestBody.(resources.Application).Name
					requestParams.ResponseBody.(*resources.Application).StackName = requestParams.RequestBody.(resources.Application).StackName
					requestParams.ResponseBody.(*resources.Application).LifecycleType = requestParams.RequestBody.(resources.Application).LifecycleType
					requestParams.ResponseBody.(*resources.Application).LifecycleBuildpacks = requestParams.RequestBody.(resources.Application).LifecycleBuildpacks
					requestParams.ResponseBody.(*resources.Application).SpaceGUID = requestParams.RequestBody.(resources.Application).SpaceGUID
					return "", Warnings{"this is a warning"}, nil
				})

				appToUpdate = resources.Application{
					GUID:                "some-app-guid",
					Name:                "some-app-name",
					StackName:           "some-stack-name",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"some-buildpack"},
					SpaceGUID:           "some-space-guid",
				}
			})

			It("makes the correct request", func() {
				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.PatchApplicationRequest))
				Expect(actualParams.URIParams).To(Equal(internal.Params{"app_guid": "some-app-guid"}))
				Expect(actualParams.RequestBody).To(Equal(appToUpdate))
				_, ok := actualParams.ResponseBody.(*resources.Application)
				Expect(ok).To(BeTrue())
			})

			It("returns the updated app and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(updatedApp).To(Equal(resources.Application{
					GUID:                "some-app-guid",
					StackName:           "some-stack-name",
					LifecycleBuildpacks: []string{"some-buildpack"},
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					Name:                "some-app-name",
					SpaceGUID:           "some-space-guid",
				}))
			})
		})

		When("cc returns back an error or warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					},
					{
						Code:   10010,
						Detail: "App not found",
						Title:  "CF-ResourceNotFound",
					},
				}

				requester.MakeRequestReturns(
					"",
					Warnings{"this is a warning"},
					ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors},
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
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateApplicationName", func() {
		var (
			newAppName string
			appGUID    string

			updatedApp resources.Application
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			newAppName = "some-new-app-name"
			appGUID = "some-app-guid"

			updatedApp, warnings, executeErr = client.UpdateApplicationName(newAppName, appGUID)
		})

		When("the application successfully is updated", func() {
			BeforeEach(func() {
				requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
					requestParams.ResponseBody.(*resources.Application).GUID = appGUID
					requestParams.ResponseBody.(*resources.Application).Name = requestParams.RequestBody.(resources.ApplicationNameOnly).Name
					requestParams.ResponseBody.(*resources.Application).StackName = "some-stack-name"
					requestParams.ResponseBody.(*resources.Application).LifecycleType = constant.AppLifecycleTypeBuildpack
					requestParams.ResponseBody.(*resources.Application).LifecycleBuildpacks = []string{"some-buildpack"}
					requestParams.ResponseBody.(*resources.Application).SpaceGUID = "some-space-guid"
					return "", Warnings{"this is a warning"}, nil
				})
			})

			It("makes the correct request", func() {
				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.PatchApplicationRequest))
				Expect(actualParams.URIParams).To(Equal(internal.Params{"app_guid": "some-app-guid"}))
				Expect(actualParams.RequestBody).To(Equal(resources.ApplicationNameOnly{Name: newAppName}))
				_, ok := actualParams.ResponseBody.(*resources.Application)
				Expect(ok).To(BeTrue())
			})

			It("returns the updated app and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(updatedApp).To(Equal(resources.Application{
					GUID:                "some-app-guid",
					StackName:           "some-stack-name",
					LifecycleBuildpacks: []string{"some-buildpack"},
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					Name:                "some-new-app-name",
					SpaceGUID:           "some-space-guid",
				}))
			})
		})

		When("cc returns back an error or warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					},
					{
						Code:   10010,
						Detail: "App not found",
						Title:  "CF-ResourceNotFound",
					},
				}

				requester.MakeRequestReturns(
					"",
					Warnings{"this is a warning"},
					ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors},
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
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateApplicationStop", func() {
		var (
			responseApp resources.Application
			warnings    Warnings
			executeErr  error
		)

		JustBeforeEach(func() {
			responseApp, warnings, executeErr = client.UpdateApplicationStop("some-app-guid")
		})

		When("the response succeeds", func() {
			BeforeEach(func() {
				requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
					requestParams.ResponseBody.(*resources.Application).GUID = "some-app-guid"
					requestParams.ResponseBody.(*resources.Application).Name = "some-app"
					requestParams.ResponseBody.(*resources.Application).State = constant.ApplicationStopped
					return "", Warnings{"this is a warning"}, nil
				})
			})

			It("makes the correct request", func() {
				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.PostApplicationActionStopRequest))
				Expect(actualParams.URIParams).To(Equal(internal.Params{"app_guid": "some-app-guid"}))
				_, ok := actualParams.ResponseBody.(*resources.Application)
				Expect(ok).To(BeTrue())
			})

			It("returns the application, warnings, and no error", func() {
				Expect(responseApp).To(Equal(resources.Application{
					GUID:  "some-app-guid",
					Name:  "some-app",
					State: constant.ApplicationStopped,
				}))
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the CC returns an error", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					},
					{
						Code:   10010,
						Detail: "App not found",
						Title:  "CF-ResourceNotFound",
					},
				}

				requester.MakeRequestReturns(
					"",
					Warnings{"this is a warning"},
					ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors},
				)
			})

			It("returns no app, the error and all warnings", func() {
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
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateApplicationStart", func() {
		var (
			responseApp resources.Application
			warnings    Warnings
			executeErr  error
		)

		JustBeforeEach(func() {
			responseApp, warnings, executeErr = client.UpdateApplicationStart("some-app-guid")
		})

		When("the response succeeds", func() {
			BeforeEach(func() {
				requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
					requestParams.ResponseBody.(*resources.Application).GUID = "some-app-guid"
					requestParams.ResponseBody.(*resources.Application).Name = "some-app"
					return "", Warnings{"this is a warning"}, nil
				})
			})

			It("makes the correct request", func() {
				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.PostApplicationActionStartRequest))
				Expect(actualParams.URIParams).To(Equal(internal.Params{"app_guid": "some-app-guid"}))
				_, ok := actualParams.ResponseBody.(*resources.Application)
				Expect(ok).To(BeTrue())
			})

			It("returns warnings and no error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(responseApp).To(Equal(resources.Application{
					GUID: "some-app-guid",
					Name: "some-app",
				}))
			})
		})

		When("cc returns back an error or warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					},
					{
						Code:   10010,
						Detail: "App not found",
						Title:  "CF-ResourceNotFound",
					},
				}

				requester.MakeRequestReturns(
					"",
					Warnings{"this is a warning"},
					ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors},
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
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateApplicationRestart", func() {
		var (
			responseApp resources.Application
			warnings    Warnings
			executeErr  error
		)

		JustBeforeEach(func() {
			responseApp, warnings, executeErr = client.UpdateApplicationRestart("some-app-guid")
		})

		When("the response succeeds", func() {
			BeforeEach(func() {
				requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
					requestParams.ResponseBody.(*resources.Application).GUID = "some-app-guid"
					requestParams.ResponseBody.(*resources.Application).Name = "some-app"
					requestParams.ResponseBody.(*resources.Application).State = constant.ApplicationStarted
					return "", Warnings{"this is a warning"}, nil
				})
			})

			It("makes the correct request", func() {
				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.PostApplicationActionRestartRequest))
				Expect(actualParams.URIParams).To(Equal(internal.Params{"app_guid": "some-app-guid"}))
				_, ok := actualParams.ResponseBody.(*resources.Application)
				Expect(ok).To(BeTrue())
			})

			It("returns the application, warnings, and no error", func() {
				Expect(responseApp).To(Equal(resources.Application{
					GUID:  "some-app-guid",
					Name:  "some-app",
					State: constant.ApplicationStarted,
				}))
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the CC returns an error", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					},
					{
						Code:   10010,
						Detail: "App not found",
						Title:  "CF-ResourceNotFound",
					},
				}

				requester.MakeRequestReturns(
					"",
					Warnings{"this is a warning"},
					ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors},
				)
			})

			It("returns no app, the error and all warnings", func() {
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
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
