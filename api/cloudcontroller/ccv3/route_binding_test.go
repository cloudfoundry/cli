package ccv3_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
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
})
