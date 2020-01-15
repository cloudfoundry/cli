package v7action_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"errors"
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
	})

	Describe("GetOrganizationQuotaByName", func() {
		var (
			quotaName  string
			quota      OrganizationQuota
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			quotaName = "quota-name"
		})

		JustBeforeEach(func() {
			quota, warnings, executeErr = actor.GetOrganizationQuotaByName(quotaName)
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

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(executeErr).To(MatchError("list-error"))
				Expect(quota).To(Equal(OrganizationQuota{}))
			})
		})

		When("when the org quota could not be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]ccv3.OrgQuota{},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(executeErr).To(MatchError(actionerror.OrganizationQuotaNotFoundForNameError{Name: quotaName}))
				Expect(quota).To(Equal(OrganizationQuota{}))
			})
		})

		When("getting a single quota by name", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					[]ccv3.OrgQuota{
						ccv3.OrgQuota{
							GUID: "quota-guid",
							Name: quotaName,
						},
					},
					ccv3.Warnings{"some-quota-warning"},
					nil,
				)
			})

			It("queries the API and returns the matching organization quota", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetOrganizationQuotasArgsForCall(0)
				Expect(query).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{quotaName}},
				))

				Expect(warnings).To(ConsistOf("some-quota-warning"))
				Expect(quota).To(Equal(OrganizationQuota{GUID: "quota-guid", Name: quotaName}))
			})
		})
	})
})
