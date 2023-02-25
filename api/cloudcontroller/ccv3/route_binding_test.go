package ccv3_test

import (
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouteBinding", func() {
	var (
		requester *ccv3fakes.FakeRequester
		client    *Client
	)

	BeforeEach(func() {
		requester = new(ccv3fakes.FakeRequester)
		client, _ = NewFakeRequesterTestClient(requester)
	})

	Describe("CreateRouteBinding", func() {
		When("the request succeeds", func() {
			It("returns warnings and no errors", func() {
				requester.MakeRequestReturns("fake-job-url", Warnings{"fake-warning"}, nil)

				binding := resources.RouteBinding{
					ServiceInstanceGUID: "fake-service-instance-guid",
					RouteGUID:           "fake-route-guid",
				}

				jobURL, warnings, err := client.CreateRouteBinding(binding)

				Expect(jobURL).To(Equal(JobURL("fake-job-url")))
				Expect(warnings).To(ConsistOf("fake-warning"))
				Expect(err).NotTo(HaveOccurred())

				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				Expect(requester.MakeRequestArgsForCall(0)).To(Equal(RequestParams{
					RequestName: internal.PostRouteBindingRequest,
					RequestBody: binding,
				}))
			})
		})

		When("the request fails", func() {
			It("returns errors and warnings", func() {
				requester.MakeRequestReturns("", Warnings{"fake-warning"}, errors.New("bang"))

				jobURL, warnings, err := client.CreateRouteBinding(resources.RouteBinding{})

				Expect(jobURL).To(BeEmpty())
				Expect(warnings).To(ConsistOf("fake-warning"))
				Expect(err).To(MatchError("bang"))
			})
		})
	})

	Describe("GetRouteBindings", func() {
		var (
			query      []Query
			bindings   []resources.RouteBinding
			included   IncludedResources
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
				for i := 1; i <= 3; i++ {
					Expect(requestParams.AppendToList(resources.RouteBinding{
						GUID:                fmt.Sprintf("route-binding-%d-guid", i),
						ServiceInstanceGUID: fmt.Sprintf("si-%d-guid", i),
					})).NotTo(HaveOccurred())
				}
				return IncludedResources{}, Warnings{"warning-1", "warning-2"}, nil
			})

			query = []Query{
				{Key: ServiceInstanceGUIDFilter, Values: []string{"si-1-guid", "si-2-guid", "si-3-guid", "si-4-guid"}},
			}
		})

		JustBeforeEach(func() {
			bindings, included, warnings, executeErr = client.GetRouteBindings(query...)
		})

		It("makes the correct call", func() {
			Expect(requester.MakeListRequestCallCount()).To(Equal(1))
			actualParams := requester.MakeListRequestArgsForCall(0)
			Expect(actualParams.RequestName).To(Equal(internal.GetRouteBindingsRequest))
			Expect(actualParams.Query).To(ConsistOf(Query{Key: ServiceInstanceGUIDFilter, Values: []string{"si-1-guid", "si-2-guid", "si-3-guid", "si-4-guid"}}))
			Expect(actualParams.ResponseBody).To(BeAssignableToTypeOf(resources.RouteBinding{}))
		})

		It("returns a list of route bindings", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

			Expect(bindings).To(ConsistOf(
				resources.RouteBinding{
					GUID:                "route-binding-1-guid",
					ServiceInstanceGUID: "si-1-guid",
				},
				resources.RouteBinding{
					GUID:                "route-binding-2-guid",
					ServiceInstanceGUID: "si-2-guid",
				},
				resources.RouteBinding{
					GUID:                "route-binding-3-guid",
					ServiceInstanceGUID: "si-3-guid",
				},
			))
		})

		When("there are included resources", func() {
			BeforeEach(func() {
				requester.MakeListRequestReturns(
					IncludedResources{
						ServiceInstances: []resources.ServiceInstance{
							{Name: "foo", GUID: "foo-guid"},
							{Name: "bar", GUID: "bar-guid"},
						},
					},
					nil,
					nil,
				)
			})

			It("returns the included resources", func() {
				Expect(included).To(Equal(IncludedResources{
					ServiceInstances: []resources.ServiceInstance{
						{Name: "foo", GUID: "foo-guid"},
						{Name: "bar", GUID: "bar-guid"},
					},
				}))
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

	Describe("DeleteRouteBinding", func() {
		const (
			guid   = "fake-route-binding-guid"
			jobURL = JobURL("fake-job-url")
		)

		It("makes the right request", func() {
			client.DeleteRouteBinding(guid)

			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			Expect(requester.MakeRequestArgsForCall(0)).To(Equal(RequestParams{
				RequestName: internal.DeleteRouteBindingRequest,
				URIParams:   internal.Params{"route_binding_guid": guid},
			}))
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns(jobURL, Warnings{"fake-warning"}, nil)
			})

			It("returns warnings and no errors", func() {
				job, warnings, err := client.DeleteRouteBinding(guid)

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
				jobURL, warnings, err := client.DeleteRouteBinding(guid)

				Expect(jobURL).To(BeEmpty())
				Expect(warnings).To(ConsistOf("fake-warning"))
				Expect(err).To(MatchError("bang"))
			})
		})
	})
})
