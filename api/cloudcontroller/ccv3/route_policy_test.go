package ccv3_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v8/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoutePolicy", func() {
	var (
		requester *ccv3fakes.FakeRequester
		client    *Client
	)

	BeforeEach(func() {
		requester = new(ccv3fakes.FakeRequester)
		client, _ = NewFakeRequesterTestClient(requester)
	})

	Describe("CreateRoutePolicy", func() {
		var (
			policy     resources.RoutePolicy
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			policy, warnings, executeErr = client.CreateRoutePolicy(resources.RoutePolicy{
				Source:    "cf:any",
				RouteGUID: "route-guid",
			})
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns("", Warnings{"create-warning"}, nil)
			})

			It("makes the correct request and returns warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("create-warning"))
				Expect(policy).To(Equal(resources.RoutePolicy{}))

				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.PostRoutePolicyRequest))
				Expect(actualParams.RequestBody).To(Equal(resources.RoutePolicy{
					Source:    "cf:any",
					RouteGUID: "route-guid",
				}))
			})
		})

		When("the request fails", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns("", Warnings{"warning"}, errors.New("bang"))
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("bang"))
				Expect(warnings).To(ConsistOf("warning"))
			})
		})
	})

	Describe("GetRoutePolicies", func() {
		var (
			policies   []resources.RoutePolicy
			included   IncludedResources
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			policies, included, warnings, executeErr = client.GetRoutePolicies()
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					Expect(requestParams.AppendToList(resources.RoutePolicy{
						GUID:      "policy-guid",
						Source:    "cf:any",
						RouteGUID: "route-guid",
					})).NotTo(HaveOccurred())
					return IncludedResources{}, Warnings{"list-warning"}, nil
				})
			})

			It("makes the correct request and returns policies", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("list-warning"))
				Expect(policies).To(HaveLen(1))
				Expect(policies[0].GUID).To(Equal("policy-guid"))
				Expect(policies[0].Source).To(Equal("cf:any"))
				Expect(included).To(Equal(IncludedResources{}))

				Expect(requester.MakeListRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeListRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.GetRoutePoliciesRequest))
				Expect(actualParams.ResponseBody).To(BeAssignableToTypeOf(resources.RoutePolicy{}))
			})
		})

		When("the request fails", func() {
			BeforeEach(func() {
				requester.MakeListRequestReturns(IncludedResources{}, Warnings{"warning"}, errors.New("bang"))
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("bang"))
				Expect(warnings).To(ConsistOf("warning"))
			})
		})
	})

	Describe("DeleteRoutePolicy", func() {
		var (
			jobURL     JobURL
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.DeleteRoutePolicy("policy-guid")
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns("", Warnings{"delete-warning"}, nil)
			})

			It("makes the correct request and returns warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("delete-warning"))
				Expect(jobURL).To(Equal(JobURL("")))

				Expect(requester.MakeRequestCallCount()).To(Equal(1))
				actualParams := requester.MakeRequestArgsForCall(0)
				Expect(actualParams.RequestName).To(Equal(internal.DeleteRoutePolicyRequest))
				Expect(actualParams.URIParams).To(Equal(internal.Params{"route_policy_guid": "policy-guid"}))
			})
		})

		When("the request fails", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns("", Warnings{"warning"}, errors.New("bang"))
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("bang"))
				Expect(warnings).To(ConsistOf("warning"))
			})
		})
	})
})
