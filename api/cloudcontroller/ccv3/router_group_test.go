package ccv3_test

import (
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouterGroup", func() {
	var (
		client    *Client
		requester *ccv3fakes.FakeRequester
	)

	BeforeEach(func() {
		requester = new(ccv3fakes.FakeRequester)
		client, _ = NewFakeRequesterTestClient(requester)
	})

	Describe("GetRouterGroups", func() {
		var (
			routerGroups []RouterGroup
			warnings     Warnings
			executeErr   error
		)

		BeforeEach(func() {
			client.Info.Links.Routing.HREF = "some-routing-url"

			requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
				var toReturn []RouterGroup
				toReturn = append(toReturn, RouterGroup{Name: "rg1"})
				toReturn = append(toReturn, RouterGroup{Name: "rg2"})

				responseBody := requestParams.ResponseBody.(*[]RouterGroup)
				*responseBody = toReturn

				return "", Warnings{"some-warning"}, errors.New("some-error")
			})
		})

		JustBeforeEach(func() {
			routerGroups, warnings, executeErr = client.GetRouterGroups()
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			params := requester.MakeRequestArgsForCall(0)

			Expect(params.URL).To(Equal("some-routing-url/v1/router_groups"))
			Expect(params.ResponseBody).To(matchers.HaveTypeOf(&[]RouterGroup{}))
		})

		It("returns the resources and all warnings", func() {
			Expect(warnings).To(ConsistOf("some-warning"))
			Expect(executeErr).To(MatchError("some-error"))
			Expect(routerGroups).To(Equal([]RouterGroup{
				{Name: "rg1"},
				{Name: "rg2"},
			}))
		})
	})
})
