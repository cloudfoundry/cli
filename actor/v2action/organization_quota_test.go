package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OrganizationQuota Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetOrganizationQuota", func() {
		When("the org quota exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotaReturns(
					ccv2.OrganizationQuota{
						GUID: "some-org-quota-guid",
						Name: "some-org-quota",
					},
					ccv2.Warnings{"warning-1"},
					nil,
				)
			})

			It("returns the org quota and warnings", func() {
				orgQuota, warnings, err := actor.GetOrganizationQuota("some-org-quota-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(orgQuota).To(Equal(OrganizationQuota{
					GUID: "some-org-quota-guid",
					Name: "some-org-quota",
				}))
				Expect(warnings).To(ConsistOf("warning-1"))

				Expect(fakeCloudControllerClient.GetOrganizationQuotaCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationQuotaArgsForCall(0)).To(Equal(
					"some-org-quota-guid"))
			})
		})

		When("the org quota does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotaReturns(ccv2.OrganizationQuota{}, nil, ccerror.ResourceNotFoundError{})
			})

			It("returns an OrganizationQuotaNotFoundError", func() {
				_, _, err := actor.GetOrganizationQuota("some-org-quota-guid")
				Expect(err).To(MatchError(actionerror.OrganizationQuotaNotFoundError{GUID: "some-org-quota-guid"}))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some org quota error")
				fakeCloudControllerClient.GetOrganizationQuotaReturns(ccv2.OrganizationQuota{}, nil, expectedErr)
			})

			It("returns the error", func() {
				_, _, err := actor.GetOrganizationQuota("some-org-quota-guid")
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})

	Describe("GetOrganizationQuotaByName", func() {
		var (
			quotaName string

			quota      OrganizationQuota
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			quotaName = "some-quota"
			quota, warnings, executeErr = actor.GetOrganizationQuotaByName(quotaName)
		})

		When("fetching quotas succeeds", func() {
			When("a single quota is returned", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationQuotasReturns(
						[]ccv2.OrganizationQuota{
							{
								GUID: "some-quota-definition-guid",
								Name: "some-quota",
							},
						},
						ccv2.Warnings{"quota-warning-1", "quota-warning-2"},
						nil)
				})

				It("returns the found quota and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(quota).To(Equal(OrganizationQuota{
						GUID: "some-quota-definition-guid",
						Name: "some-quota",
					},
					))
					Expect(warnings).To(ConsistOf("quota-warning-1", "quota-warning-2"))

					Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetOrganizationQuotasArgsForCall(0)).To(Equal([]ccv2.Filter{
						{
							Type:     constant.NameFilter,
							Operator: constant.EqualOperator,
							Values:   []string{"some-quota"},
						},
					}))
				})
			})

			When("more than one quota is returned", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationQuotasReturns(
						[]ccv2.OrganizationQuota{
							{
								GUID: "some-quota-definition-guid-1",
								Name: "some-quota-1",
							},
							{
								GUID: "some-quota-definition-guid-2",
								Name: "some-quota-2",
							},
						},
						ccv2.Warnings{"quota-warning-1", "quota-warning-2"},
						nil)
				})

				It("returns an error that multiple quotas were found and does not try to create the org", func() {
					Expect(executeErr).To(MatchError(actionerror.MultipleOrganizationQuotasFoundForNameError{
						Name: quotaName,
						GUIDs: []string{
							"some-quota-definition-guid-1",
							"some-quota-definition-guid-2",
						},
					}))
					Expect(warnings).To(ConsistOf("quota-warning-1", "quota-warning-2"))

					Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.CreateOrganizationCallCount()).To(Equal(0))
				})
			})

			When("no quotas are returned", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationQuotasReturns(
						[]ccv2.OrganizationQuota{},
						ccv2.Warnings{"quota-warning-1", "quota-warning-2"},
						nil)
				})

				It("returns an error that no quotas were found and does not try to create the org", func() {
					Expect(executeErr).To(MatchError(actionerror.QuotaNotFoundForNameError{Name: quotaName}))
					Expect(warnings).To(ConsistOf("quota-warning-1", "quota-warning-2"))

					Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.CreateOrganizationCallCount()).To(Equal(0))
				})
			})

		})

		When("fetching the quota fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotasReturns(
					nil,
					ccv2.Warnings{"quota-warning-1", "quota-warning-2"},
					errors.New("no quota found"))
			})

			It("returns warnings and the error, and does not try to create the org", func() {
				Expect(executeErr).To(MatchError("no quota found"))
				Expect(warnings).To(ConsistOf("quota-warning-1", "quota-warning-2"))

				Expect(fakeCloudControllerClient.GetOrganizationQuotasCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateOrganizationCallCount()).To(Equal(0))
			})
		})
	})
})
