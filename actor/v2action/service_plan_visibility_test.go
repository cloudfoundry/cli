package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Plan Visibility Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetServicePlanVisibilities", func() {
		var (
			visibilities []ServicePlanVisibility
			warnings     Warnings
			err          error
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServicePlanVisibilitiesReturns([]ccv2.ServicePlanVisibility{
				{GUID: "guid-1", ServicePlanGUID: "plan-guid-1", OrganizationGUID: "org-guid-1"},
				{GUID: "guid-2", ServicePlanGUID: "plan-guid-2", OrganizationGUID: "org-guid-2"},
			}, ccv2.Warnings{"cc-warning"}, nil)
		})

		JustBeforeEach(func() {
			visibilities, warnings, err = actor.GetServicePlanVisibilities("some-service-plan-guid")
		})

		It("doesn't error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("fetches visibilities from the cloud controller client, filtering by plan GUID", func() {
			Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).NotTo(BeZero())
			expectedFilters := []ccv2.Filter{
				{
					Type:     constant.ServicePlanGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-service-plan-guid"},
				},
			}
			Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(0)).To(Equal(expectedFilters))
		})

		It("returns visibilities fetched from the cloud controller client", func() {
			expectedVisibilities := []ServicePlanVisibility{
				{GUID: "guid-1", ServicePlanGUID: "plan-guid-1", OrganizationGUID: "org-guid-1"},
				{GUID: "guid-2", ServicePlanGUID: "plan-guid-2", OrganizationGUID: "org-guid-2"},
			}
			Expect(visibilities).To(Equal(expectedVisibilities))
		})

		It("returns any warnings from the cloud controller client", func() {
			Expect(warnings).To(ConsistOf("cc-warning"))
		})

		When("fetching visibilities from the cloud controller client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlanVisibilitiesReturns(nil, ccv2.Warnings{"cc-warning"}, errors.New("boom"))
			})

			It("propagates the error", func() {
				Expect(err).To(MatchError("boom"))
			})
		})

		It("returns any warnings from the cloud controller client", func() {
			Expect(warnings).To(ConsistOf("cc-warning"))
		})
	})
})
