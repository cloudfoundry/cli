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
	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			Expect(params).To(Equal(types.JSONObject{"foo": "bar"}))
		})

		When("there is an error getting the parameters", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns("", Warnings{"one", "two"}, errors.New("boom"))
			})

			It("returns warnings and an error", func() {
				params, warnings, err := client.GetServiceInstanceParameters(guid)
				Expect(err).To(MatchError("boom"))
				Expect(warnings).To(ConsistOf("one", "two"))
				Expect(params).To(BeEmpty())
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

		Context("user provided", func() {
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
		})

		Context("managed", func() {
			BeforeEach(func() {
				serviceInstance = resources.ServiceInstance{
					Name:            "fake-new-user-provided-service-instance",
					Tags:            types.NewOptionalStringSlice("foo", "bar"),
					ServicePlanGUID: guid,
					Parameters:      types.NewOptionalObject(map[string]interface{}{"some-param": "some-value"}),
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

		It("makes the right request", func() {
			client.DeleteServiceInstance(guid)

			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			Expect(requester.MakeRequestArgsForCall(0)).To(Equal(RequestParams{
				RequestName: internal.DeleteServiceInstanceRequest,
				URIParams:   internal.Params{"service_instance_guid": guid},
			}))
		})

		When("there are query parameters", func() {
			It("passes them through", func() {
				client.DeleteServiceInstance(guid, Query{Key: NameFilter, Values: []string{"foo"}})

				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				Expect(requester.MakeRequestArgsForCall(0).Query).To(ConsistOf(Query{Key: NameFilter, Values: []string{"foo"}}))
			})
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns(jobURL, Warnings{"fake-warning"}, nil)
			})

			It("returns warnings and no errors", func() {
				job, warnings, err := client.DeleteServiceInstance(guid)

				Expect(job).To(Equal(jobURL))
				Expect(warnings).To(ConsistOf("fake-warning"))
				Expect(err).NotTo(HaveOccurred())
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
		Describe("ShareServiceInstanceToSpaces", func() {
			var (
				serviceInstanceGUID string
				spaceGUIDs          []string
			)

			BeforeEach(func() {
				serviceInstanceGUID = "some-service-instance-guid"
				spaceGUIDs = []string{"some-space-guid", "some-other-space-guid"}
			})

			It("makes the right request", func() {
				client.ShareServiceInstanceToSpaces(serviceInstanceGUID, spaceGUIDs)

				Expect(requester.MakeRequestCallCount()).To(Equal(1))

				actualRequest := requester.MakeRequestArgsForCall(0)
				Expect(actualRequest.RequestName).To(Equal(internal.PostServiceInstanceRelationshipsSharedSpacesRequest))
				Expect(actualRequest.URIParams).To(Equal(internal.Params{"service_instance_guid": serviceInstanceGUID}))
				Expect(actualRequest.RequestBody).To(Equal(resources.RelationshipList{
					GUIDs: spaceGUIDs,
				}))
			})

			When("the request succeeds", func() {
				BeforeEach(func() {
					requester.MakeRequestCalls(func(params RequestParams) (JobURL, Warnings, error) {
						json.Unmarshal([]byte(`{"data":[{"guid":"some-space-guid"}, {"guid":"some-other-space-guid"}]}`), params.ResponseBody)
						return "", Warnings{"fake-warning"}, nil
					})
				})

				It("returns warnings and no errors", func() {
					relationships, warnings, err := client.ShareServiceInstanceToSpaces(serviceInstanceGUID, spaceGUIDs)

					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).NotTo(HaveOccurred())
					Expect(relationships).To(Equal(resources.RelationshipList{GUIDs: spaceGUIDs}))
				})
			})

			When("the request fails", func() {
				BeforeEach(func() {
					requester.MakeRequestReturns("", Warnings{"fake-warning"}, errors.New("bang"))
				})

				It("returns errors and warnings", func() {
					_, warnings, err := client.ShareServiceInstanceToSpaces(serviceInstanceGUID, spaceGUIDs)

					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).To(MatchError("bang"))
				})
			})
		})

		Describe("UnshareServiceInstanceFromSpace", func() {
			var (
				serviceInstanceGUID string
				spaceGUID           string
			)

			BeforeEach(func() {
				serviceInstanceGUID = "some-service-instance-guid"
				spaceGUID = "some-space-guid"
			})

			It("makes the right request", func() {
				client.UnshareServiceInstanceFromSpace(serviceInstanceGUID, spaceGUID)

				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				Expect(requester.MakeRequestArgsForCall(0)).To(Equal(RequestParams{
					RequestName: internal.DeleteServiceInstanceRelationshipsSharedSpaceRequest,
					URIParams: internal.Params{
						"service_instance_guid": serviceInstanceGUID,
						"space_guid":            spaceGUID},
				}))
			})

			When("the request succeeds", func() {
				BeforeEach(func() {
					requester.MakeRequestReturns("", Warnings{"fake-warning"}, nil)
				})

				It("returns warnings and no errors", func() {
					warnings, err := client.UnshareServiceInstanceFromSpace(serviceInstanceGUID, spaceGUID)

					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("the request fails", func() {
				BeforeEach(func() {
					requester.MakeRequestReturns("", Warnings{"fake-warning"}, errors.New("bang"))
				})

				It("returns errors and warnings", func() {
					warnings, err := client.UnshareServiceInstanceFromSpace(serviceInstanceGUID, spaceGUID)

					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).To(MatchError("bang"))
				})
			})
		})

		Describe("GetServiceInstanceSharedSpaces", func() {
			var (
				serviceInstanceGUID string
			)

			BeforeEach(func() {
				serviceInstanceGUID = "some-service-instance-guid"
			})

			It("makes the right request", func() {
				client.GetServiceInstanceSharedSpaces(serviceInstanceGUID)

				Expect(requester.MakeRequestCallCount()).To(Equal(1))

				actualRequest := requester.MakeRequestArgsForCall(0)
				Expect(actualRequest.RequestName).To(Equal(internal.GetServiceInstanceRelationshipsSharedSpacesRequest))
				Expect(actualRequest.URIParams).To(Equal(internal.Params{"service_instance_guid": serviceInstanceGUID}))
				Expect(actualRequest.Query).To(ConsistOf(
					Query{
						Key:    FieldsSpace,
						Values: []string{"guid", "name", "relationships.organization"},
					},
					Query{
						Key:    FieldsSpaceOrganization,
						Values: []string{"guid", "name"},
					},
				))
			})

			When("the request succeeds", func() {
				BeforeEach(func() {
					requester.MakeRequestCalls(func(params RequestParams) (JobURL, Warnings, error) {
						jsonry.Unmarshal([]byte(`{
								   "data": [{"guid":"some-space-guid"},{"guid":"some-other-space-guid"}],
								   "links": {
									  "self": {
										 "href": "https://some-url/v3/service_instances/7915bc51-8203-4758-b0e2-f77bfcdc38cb/relationships/shared_spaces"
									  }
								   },
								   "included": {
									  "spaces": [
										 {
											"name": "some-space-name",
											"guid": "some-space-guid",
											"relationships": {
											   "organization": {
												  "data": {
													 "guid": "some-org-guid"
												  }
											   }
											}
										 },
										{
											"name": "some-other-space-name",
											"guid": "some-other-space-guid",
											"relationships": {
											   "organization": {
												  "data": {
													 "guid": "some-org-guid"
												  }
											   }
											}
										 }
									  ],
									  "organizations": [
										 {
											"name": "some-org-name",
											"guid": "some-org-guid"
										 }
									  ]
								   }
								}`), params.ResponseBody)
						return "", Warnings{"fake-warning"}, nil
					})
				})

				It("returns warnings and no errors", func() {
					spaces, warnings, err := client.GetServiceInstanceSharedSpaces(serviceInstanceGUID)

					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).NotTo(HaveOccurred())
					Expect(spaces).To(Equal([]SpaceWithOrganization{
						{
							SpaceGUID:        "some-space-guid",
							SpaceName:        "some-space-name",
							OrganizationName: "some-org-name",
						},
						{
							SpaceGUID:        "some-other-space-guid",
							SpaceName:        "some-other-space-name",
							OrganizationName: "some-org-name",
						},
					}))
				})
			})

			When("the request fails", func() {
				BeforeEach(func() {
					requester.MakeRequestReturns("", Warnings{"fake-warning"}, errors.New("bang"))
				})

				It("returns errors and warnings", func() {
					_, warnings, err := client.GetServiceInstanceSharedSpaces(serviceInstanceGUID)

					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).To(MatchError("bang"))
				})
			})
		})

		Describe("GetServiceInstanceUsageSummary", func() {
			var (
				serviceInstanceGUID string
				spaceGUIDs          []string
			)

			BeforeEach(func() {
				serviceInstanceGUID = "some-service-instance-guid"
				spaceGUIDs = []string{"some-space-guid", "some-other-space-guid"}
			})

			It("makes the right request", func() {
				client.GetServiceInstanceUsageSummary(serviceInstanceGUID)

				Expect(requester.MakeRequestCallCount()).To(Equal(1))

				actualRequest := requester.MakeRequestArgsForCall(0)
				Expect(actualRequest.RequestName).To(Equal(internal.GetServiceInstanceSharedSpacesUsageSummaryRequest))
				Expect(actualRequest.URIParams).To(Equal(internal.Params{"service_instance_guid": serviceInstanceGUID}))
			})

			When("the request succeeds", func() {
				BeforeEach(func() {
					requester.MakeRequestCalls(func(params RequestParams) (JobURL, Warnings, error) {
						jsonry.Unmarshal([]byte(`{
							"usage_summary": [
								{
									"space": {
										"guid": "some-space-guid"
									},
									"bound_app_count": 2
								},
								{
									"space": {
										"guid": "some-other-space-guid"
									},
									"bound_app_count": 1
								}
							],
							"links": {
								"self": {
									"href": "https://api.example.org/v3/service_instances/some_instance_guid/relationships/shared_spaces/usage_summary"
								},
								"shared_spaces": {
									"href": "https://api.example.org/v3/service_instances/some_instance_guid/relationships/shared_spaces"
								},
								"service_instance": {
									"href": "https://api.example.org/v3/service_instances/some_instance_guid"
								}
							}
						}`), params.ResponseBody)
						return "", Warnings{"fake-warning"}, nil
					})
				})

				It("returns warnings and no errors", func() {
					usageSummary, warnings, err := client.GetServiceInstanceUsageSummary(serviceInstanceGUID)

					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).NotTo(HaveOccurred())
					Expect(usageSummary).To(Equal([]resources.ServiceInstanceUsageSummary{{SpaceGUID: spaceGUIDs[0], BoundAppCount: 2}, {SpaceGUID: spaceGUIDs[1], BoundAppCount: 1}}))
				})
			})

			When("the request fails", func() {
				BeforeEach(func() {
					requester.MakeRequestReturns("", Warnings{"fake-warning"}, errors.New("bang"))
				})

				It("returns errors and warnings", func() {
					_, warnings, err := client.GetServiceInstanceUsageSummary(serviceInstanceGUID)

					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).To(MatchError("bang"))
				})
			})
		})
	})
})
