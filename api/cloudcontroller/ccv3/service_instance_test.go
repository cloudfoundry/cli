package ccv3_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Instance", func() {
	var (
		requester *ccv3fakes.FakeRequester
		client    *Client
	)

	BeforeEach(func() {
		requester = new(ccv3fakes.FakeRequester)
		client, _ = NewFakeRequesterTestClient(requester)
	})

	Describe("GetServiceInstances", func() {
		var (
			query      Query
			instances  []resources.ServiceInstance
			included   IncludedResources
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			instances, included, warnings, executeErr = client.GetServiceInstances(query)
		})

		When("service instances exist", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					for i := 1; i <= 3; i++ {
						Expect(requestParams.AppendToList(resources.ServiceInstance{
							GUID: fmt.Sprintf("service-instance-%d-guid", i),
							Name: fmt.Sprintf("service-instance-%d-name", i),
						})).NotTo(HaveOccurred())
					}
					return IncludedResources{ServiceOfferings: []resources.ServiceOffering{{GUID: "fake-service-offering"}}}, Warnings{"warning-1", "warning-2"}, nil
				})

				query = Query{
					Key:    NameFilter,
					Values: []string{"some-service-instance-name"},
				}
			})

			It("returns a list of service instances with warnings and included resources", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(instances).To(ConsistOf(
					resources.ServiceInstance{
						GUID: "service-instance-1-guid",
						Name: "service-instance-1-name",
					},
					resources.ServiceInstance{
						GUID: "service-instance-2-guid",
						Name: "service-instance-2-name",
					},
					resources.ServiceInstance{
						GUID: "service-instance-3-guid",
						Name: "service-instance-3-name",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(included).To(Equal(IncludedResources{ServiceOfferings: []resources.ServiceOffering{{GUID: "fake-service-offering"}}}))

				Expect(requester.MakeListRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeListRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.GetServiceInstancesRequest))
				Expect(actualParams.Query).To(ConsistOf(query))
				Expect(actualParams.ResponseBody).To(BeAssignableToTypeOf(resources.ServiceInstance{}))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   42424,
						Detail: "Some detailed error message",
						Title:  "CF-SomeErrorTitle",
					},
					{
						Code:   11111,
						Detail: "Some other detailed error message",
						Title:  "CF-SomeOtherErrorTitle",
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
							Code:   42424,
							Detail: "Some detailed error message",
							Title:  "CF-SomeErrorTitle",
						},
						{
							Code:   11111,
							Detail: "Some other detailed error message",
							Title:  "CF-SomeOtherErrorTitle",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetServiceInstanceByNameAndSpace", func() {
		const (
			name      = "fake-service-instance-name"
			spaceGUID = "fake-space-guid"
		)
		var (
			instance   resources.ServiceInstance
			included   IncludedResources
			warnings   Warnings
			executeErr error
			query      []Query
		)

		BeforeEach(func() {
			query = []Query{{Key: Include, Values: []string{"unicorns"}}}
		})

		JustBeforeEach(func() {
			instance, included, warnings, executeErr = client.GetServiceInstanceByNameAndSpace(name, spaceGUID, query...)
		})

		It("makes the correct API request", func() {
			Expect(requester.MakeListRequestCallCount()).To(Equal(1))
			actualParams := requester.MakeListRequestArgsForCall(0)
			Expect(actualParams.RequestName).To(Equal(internal.GetServiceInstancesRequest))
			Expect(actualParams.Query).To(ConsistOf(
				Query{
					Key:    NameFilter,
					Values: []string{name},
				},
				Query{
					Key:    SpaceGUIDFilter,
					Values: []string{spaceGUID},
				},
				Query{
					Key:    Include,
					Values: []string{"unicorns"},
				},
			))
			Expect(actualParams.ResponseBody).To(BeAssignableToTypeOf(resources.ServiceInstance{}))
		})

		When("there are no matches", func() {
			BeforeEach(func() {
				requester.MakeListRequestReturns(
					IncludedResources{},
					Warnings{"this is a warning"},
					nil,
				)
			})

			It("returns an error and warnings", func() {
				Expect(instance).To(Equal(resources.ServiceInstance{}))
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(executeErr).To(MatchError(ccerror.ServiceInstanceNotFoundError{
					Name:      name,
					SpaceGUID: spaceGUID,
				}))
			})
		})

		When("there is a single match", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					Expect(requestParams.AppendToList(resources.ServiceInstance{
						Name: name,
						GUID: "service-instance-guid",
					})).NotTo(HaveOccurred())

					return IncludedResources{ServiceOfferings: []resources.ServiceOffering{{GUID: "fake-offering-guid"}}},
						Warnings{"warning-1", "warning-2"},
						nil
				})
			})

			It("returns the resource, included resources, and warnings", func() {
				Expect(instance).To(Equal(resources.ServiceInstance{
					Name: name,
					GUID: "service-instance-guid",
				}))
				Expect(included).To(Equal(IncludedResources{ServiceOfferings: []resources.ServiceOffering{{GUID: "fake-offering-guid"}}}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})

		When("there are multiple matches", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					for i := 1; i <= 3; i++ {
						Expect(requestParams.AppendToList(resources.ServiceInstance{
							GUID: fmt.Sprintf("service-instance-%d-guid", i),
							Name: fmt.Sprintf("service-instance-%d-name", i),
						})).NotTo(HaveOccurred())
					}
					return IncludedResources{}, Warnings{"warning-1", "warning-2"}, nil
				})
			})

			It("returns the first resource and warnings", func() {
				Expect(instance).To(Equal(resources.ServiceInstance{
					Name: "service-instance-1-name",
					GUID: "service-instance-1-guid",
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   42424,
						Detail: "Some detailed error message",
						Title:  "CF-SomeErrorTitle",
					},
					{
						Code:   11111,
						Detail: "Some other detailed error message",
						Title:  "CF-SomeOtherErrorTitle",
					},
				}

				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					Expect(requestParams.AppendToList(resources.ServiceInstance{
						GUID: "service-instance-guid",
						Name: "service-instance-name",
					})).NotTo(HaveOccurred())

					return IncludedResources{},
						Warnings{"warning-1", "warning-2"},
						ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors}
				})
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   42424,
							Detail: "Some detailed error message",
							Title:  "CF-SomeErrorTitle",
						},
						{
							Code:   11111,
							Detail: "Some other detailed error message",
							Title:  "CF-SomeOtherErrorTitle",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetServiceInstanceParameters", func() {
		const guid = "fake-service-instance-guid"

		BeforeEach(func() {
			requester.MakeRequestCalls(func(params RequestParams) (JobURL, Warnings, error) {
				json.Unmarshal([]byte(`{"foo":"bar"}`), params.ResponseBody)
				return "", Warnings{"one", "two"}, nil
			})
		})

		It("makes the correct API request", func() {
			client.GetServiceInstanceParameters(guid)

			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			actualRequest := requester.MakeRequestArgsForCall(0)
			Expect(actualRequest.RequestName).To(Equal(internal.GetServiceInstanceParametersRequest))
			Expect(actualRequest.URIParams).To(Equal(internal.Params{"service_instance_guid": guid}))
		})

		It("returns the parameters", func() {
			params, warnings, err := client.GetServiceInstanceParameters(guid)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("one", "two"))
			Expect(params).To(Equal(types.NewOptionalObject(map[string]interface{}{"foo": "bar"})))
		})

		When("there are no parameters", func() {
			BeforeEach(func() {
				requester.MakeRequestCalls(func(params RequestParams) (JobURL, Warnings, error) {
					json.Unmarshal([]byte(``), params.ResponseBody)
					return "", nil, nil
				})
			})

			It("returns a set empty empty object", func() {
				params, _, _ := client.GetServiceInstanceParameters(guid)
				Expect(params.Value).To(BeEmpty())
				Expect(params.IsSet).To(BeTrue())
			})
		})

		When("there is an error getting the parameters", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns("", Warnings{"one", "two"}, errors.New("boom"))
			})

			It("returns warnings and an error", func() {
				params, warnings, err := client.GetServiceInstanceParameters(guid)
				Expect(err).To(MatchError("boom"))
				Expect(warnings).To(ConsistOf("one", "two"))
				Expect(params.Value).To(BeEmpty())
				Expect(params.IsSet).To(BeFalse())
			})
		})
	})

	Describe("CreateServiceInstance", func() {
		Context("synchronous response", func() {
			When("the request succeeds", func() {
				It("returns warnings and no errors", func() {
					requester.MakeRequestReturns("", Warnings{"fake-warning"}, nil)

					si := resources.ServiceInstance{
						Type:            resources.UserProvidedServiceInstance,
						Name:            "fake-user-provided-service-instance",
						SpaceGUID:       "fake-space-guid",
						Tags:            types.NewOptionalStringSlice("foo", "bar"),
						RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
						SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
						Credentials: types.NewOptionalObject(map[string]interface{}{
							"foo": "bar",
							"baz": 42,
						}),
					}

					jobURL, warnings, err := client.CreateServiceInstance(si)

					Expect(jobURL).To(BeEmpty())
					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).NotTo(HaveOccurred())

					Expect(requester.MakeRequestCallCount()).To(Equal(1))
					Expect(requester.MakeRequestArgsForCall(0)).To(Equal(RequestParams{
						RequestName: internal.PostServiceInstanceRequest,
						RequestBody: si,
					}))
				})
			})

			When("the request fails", func() {
				It("returns errors and warnings", func() {
					requester.MakeRequestReturns("", Warnings{"fake-warning"}, errors.New("bang"))

					si := resources.ServiceInstance{
						Type:            resources.UserProvidedServiceInstance,
						Name:            "fake-user-provided-service-instance",
						SpaceGUID:       "fake-space-guid",
						Tags:            types.NewOptionalStringSlice("foo", "bar"),
						RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
						SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
						Credentials: types.NewOptionalObject(map[string]interface{}{
							"foo": "bar",
							"baz": 42,
						}),
					}

					jobURL, warnings, err := client.CreateServiceInstance(si)

					Expect(jobURL).To(BeEmpty())
					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).To(MatchError("bang"))
				})
			})
		})
	})

	Describe("UpdateServiceInstance", func() {
		const (
			guid   = "fake-service-instance-guid"
			jobURL = JobURL("fake-job-url")
		)

		var serviceInstance resources.ServiceInstance

		BeforeEach(func() {
			serviceInstance = resources.ServiceInstance{
				Name:            "fake-new-user-provided-service-instance",
				Tags:            types.NewOptionalStringSlice("foo", "bar"),
				RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
				SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
				Credentials: types.NewOptionalObject(map[string]interface{}{
					"foo": "bar",
					"baz": 42,
				}),
				MaintenanceInfoVersion: "9.1.2",
			}
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns(jobURL, Warnings{"fake-warning"}, nil)
			})

			It("returns warnings and no errors", func() {
				job, warnings, err := client.UpdateServiceInstance(guid, serviceInstance)

				Expect(job).To(Equal(jobURL))
				Expect(warnings).To(ConsistOf("fake-warning"))
				Expect(err).NotTo(HaveOccurred())

				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				Expect(requester.MakeRequestArgsForCall(0)).To(Equal(RequestParams{
					RequestName: internal.PatchServiceInstanceRequest,
					URIParams:   internal.Params{"service_instance_guid": guid},
					RequestBody: serviceInstance,
				}))
			})
		})

		When("the request fails", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns("", Warnings{"fake-warning"}, errors.New("bang"))
			})

			It("returns errors and warnings", func() {
				jobURL, warnings, err := client.UpdateServiceInstance(guid, serviceInstance)

				Expect(jobURL).To(BeEmpty())
				Expect(warnings).To(ConsistOf("fake-warning"))
				Expect(err).To(MatchError("bang"))
			})
		})
	})

	Describe("DeleteServiceInstance", func() {
		const (
			guid   = "fake-service-instance-guid"
			jobURL = JobURL("fake-job-url")
		)

		When("the request succeeds", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns(jobURL, Warnings{"fake-warning"}, nil)
			})

			It("returns warnings and no errors", func() {
				job, warnings, err := client.DeleteServiceInstance(guid)

				Expect(job).To(Equal(jobURL))
				Expect(warnings).To(ConsistOf("fake-warning"))
				Expect(err).NotTo(HaveOccurred())

				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				Expect(requester.MakeRequestArgsForCall(0)).To(Equal(RequestParams{
					RequestName: internal.DeleteServiceInstanceRequest,
					URIParams:   internal.Params{"service_instance_guid": guid},
				}))
			})
		})

		When("the request fails", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns("", Warnings{"fake-warning"}, errors.New("bang"))
			})

			It("returns errors and warnings", func() {
				jobURL, warnings, err := client.DeleteServiceInstance(guid)

				Expect(jobURL).To(BeEmpty())
				Expect(warnings).To(ConsistOf("fake-warning"))
				Expect(err).To(MatchError("bang"))
			})
		})
	})

	Describe("shared service instances", func() {
		BeforeEach(func() {
			client, _ = NewTestClient()
		})

		Describe("ShareServiceInstanceToSpaces", func() {
			var (
				serviceInstanceGUID string
				spaceGUIDs          []string

				relationshipList resources.RelationshipList
				warnings         Warnings
				executeErr       error
			)

			BeforeEach(func() {
				serviceInstanceGUID = "some-service-instance-guid"
				spaceGUIDs = []string{"some-space-guid", "some-other-space-guid"}
			})

			JustBeforeEach(func() {
				relationshipList, warnings, executeErr = client.ShareServiceInstanceToSpaces(serviceInstanceGUID, spaceGUIDs)
			})

			When("no errors are encountered", func() {
				BeforeEach(func() {
					response := `{
					"data": [
						{
							"guid": "some-space-guid"
						},
						{
							"guid": "some-other-space-guid"
						}
					]
				}`

					requestBody := map[string][]map[string]string{
						"data": {{"guid": "some-space-guid"}, {"guid": "some-other-space-guid"}},
					}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/service_instances/some-service-instance-guid/relationships/shared_spaces"),
							VerifyJSONRepresenting(requestBody),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns all relationships and warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(relationshipList).To(Equal(resources.RelationshipList{
						GUIDs: []string{"some-space-guid", "some-other-space-guid"},
					}))
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
						}
					]
				}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/service_instances/some-service-instance-guid/relationships/shared_spaces"),
							RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{
						ResponseCode: http.StatusTeapot,
						V3ErrorResponse: ccerror.V3ErrorResponse{
							Errors: []ccerror.V3Error{
								{
									Code:   10008,
									Detail: "The request is semantically invalid: command presence",
									Title:  "CF-UnprocessableEntity",
								},
							},
						},
					}))
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})

		Describe("GetServiceInstanceSharedSpaces", func() {
			const (
				serviceInstanceGUID string = "some-service-instance-guid"
			)

			var (
				spaces     []resources.Space
				warnings   Warnings
				executeErr error
			)

			JustBeforeEach(func() {
				spaces, warnings, executeErr = client.GetServiceInstanceSharedSpaces(serviceInstanceGUID)
			})

			When("service instance has been shared with spaces", func() {
				BeforeEach(func() {
					response := `{
					"data": [
						{
							"guid": "some-space-guid"
						},
						{
							"guid": "some-other-space-guid"
						}
					]
				}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/service_instances/some-service-instance-guid/relationships/shared_spaces"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns a list of shared spaces with warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(spaces).To(ConsistOf(
						resources.Space{
							GUID: "some-space-guid",
						},
						resources.Space{
							GUID: "some-other-space-guid",
						},
					))

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
						}
					]
				}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/service_instances/some-service-instance-guid/relationships/shared_spaces"),
							RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{
						ResponseCode: http.StatusTeapot,
						V3ErrorResponse: ccerror.V3ErrorResponse{
							Errors: []ccerror.V3Error{
								{
									Code:   10008,
									Detail: "The request is semantically invalid: command presence",
									Title:  "CF-UnprocessableEntity",
								},
							},
						},
					}))
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})
	})
})
