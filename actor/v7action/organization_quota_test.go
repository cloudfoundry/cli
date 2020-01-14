package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Organization Quota Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _ = NewTestActor()
	})

	Describe("GetOrganizationQuotas", func() {
		var (
			quotas     []OrganizationQuota
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			quotas, warnings, executeErr = actor.GetOrganizationQuotas()
		})

		When("getting organization quotas", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]ccv3.OrgQuota{
						ccv3.OrgQuota{
							GUID: "quota-guid",
							Name: "kiwi",
						},
						ccv3.OrgQuota{
							GUID: "quota-2-guid",
							Name: "strawberry",
						},
					},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("queries the API and returns organization quotas", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))

				Expect(quotas).To(ConsistOf(
					OrganizationQuota{
						GUID: "quota-guid",
						Name: "kiwi",
					},
					OrganizationQuota{
						GUID: "quota-2-guid",
						Name: "strawberry",
					},
				))
				Expect(warnings).To(ConsistOf("some-quota-warning"))
			})
		})

		When("when the API layer call returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]ccv3.OrgQuota{},
					ccv3.Warnings{"some-quota-warning"},
					errors.New("list-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))

				Expect(executeErr).To(MatchError("list-error"))
				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(quotas).To(BeEmpty())
			})
		})
	})
})
