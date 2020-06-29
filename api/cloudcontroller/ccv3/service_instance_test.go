package ccv3_test

import (
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
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
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			instances, warnings, executeErr = client.GetServiceInstances(query)
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
					return IncludedResources{}, Warnings{"warning-1", "warning-2"}, nil
				})

				query = Query{
					Key:    NameFilter,
					Values: []string{"some-service-instance-name"},
				}
			})

			It("returns a list of service instances with their associated warnings", func() {
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
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			instance, warnings, executeErr = client.GetServiceInstanceByNameAndSpace(name, spaceGUID)
		})

		It("makes the correct API request", func() {
			Expect(requester.MakeListRequestCallCount()).To(Equal(1))
			actualParams := requester.MakeListRequestArgsForCall(0)
			Expect(actualParams.RequestName).To(Equal(internal.GetServiceInstancesRequest))
			Expect(actualParams.Query).To(ConsistOf(
				ccv3.Query{
					Key:    ccv3.NameFilter,
					Values: []string{name},
				},
				ccv3.Query{
					Key:    ccv3.SpaceGUIDFilter,
					Values: []string{spaceGUID},
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

					return IncludedResources{},
						Warnings{"warning-1", "warning-2"},
						nil
				})
			})

			It("returns the resource and warnings", func() {
				Expect(instance).To(Equal(resources.ServiceInstance{
					Name: name,
					GUID: "service-instance-guid",
				}))
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

	Describe("CreateServiceInstance", func() {
		Context("synchronous response", func() {
			When("the request succeeds", func() {
				It("returns warnings and no errors", func() {
					requester.MakeRequestReturns("", ccv3.Warnings{"fake-warning"}, nil)

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
					requester.MakeRequestReturns("", ccv3.Warnings{"fake-warning"}, errors.New("bang"))

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
		Context("synchronous response", func() {
			const guid = "fake-user-provided-service-instance-guid"

			When("the request succeeds", func() {
				It("returns warnings and no errors", func() {
					requester.MakeRequestReturns("", ccv3.Warnings{"fake-warning"}, nil)

					si := resources.ServiceInstance{
						Name:            "fake-new-user-provided-service-instance",
						Tags:            types.NewOptionalStringSlice("foo", "bar"),
						RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
						SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
						Credentials: types.NewOptionalObject(map[string]interface{}{
							"foo": "bar",
							"baz": 42,
						}),
					}

					jobURL, warnings, err := client.UpdateServiceInstance(guid, si)

					Expect(jobURL).To(BeEmpty())
					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).NotTo(HaveOccurred())

					Expect(requester.MakeRequestCallCount()).To(Equal(1))
					Expect(requester.MakeRequestArgsForCall(0)).To(Equal(RequestParams{
						RequestName: internal.PatchServiceInstanceRequest,
						URIParams:   internal.Params{"service_instance_guid": guid},
						RequestBody: si,
					}))
				})
			})

			When("the request fails", func() {
				It("returns errors and warnings", func() {
					requester.MakeRequestReturns("", ccv3.Warnings{"fake-warning"}, errors.New("bang"))

					si := resources.ServiceInstance{
						Name:            "fake-new-user-provided-service-instance",
						SpaceGUID:       "fake-space-guid",
						Tags:            types.NewOptionalStringSlice("foo", "bar"),
						RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
						SyslogDrainURL:  types.NewOptionalString("https://fake-sylogg.com"),
						Credentials: types.NewOptionalObject(map[string]interface{}{
							"foo": "bar",
							"baz": 42,
						}),
					}

					jobURL, warnings, err := client.UpdateServiceInstance(guid, si)

					Expect(jobURL).To(BeEmpty())
					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).To(MatchError("bang"))
				})
			})
		})
	})
})
