package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Organization Summary Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		orgSummary                OrganizationSummary
		warnings                  Warnings
		err                       error
		expectedErr               error
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	JustBeforeEach(func() {
		orgSummary, warnings, err = actor.GetOrganizationSummaryByName("some-org")
	})

	Describe("GetOrganizationSummaryByName", func() {
		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{
						{
							GUID:                "some-org-guid",
							Name:                "some-org",
							QuotaDefinitionGUID: "some-quota-definition-guid",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)

				fakeCloudControllerClient.GetSharedDomainsReturns(
					[]ccv2.Domain{
						{
							GUID: "shared-domain-guid-2",
							Name: "shared-domain-2",
						},
						{
							GUID: "shared-domain-guid-1",
							Name: "shared-domain-1",
						},
					},
					ccv2.Warnings{"warning-3", "warning-4"},
					nil)

				fakeCloudControllerClient.GetOrganizationPrivateDomainsReturns(
					[]ccv2.Domain{
						{
							GUID: "private-domain-guid-2",
							Name: "private-domain-2",
						},
						{
							GUID: "private-domain-guid-1",
							Name: "private-domain-1",
						},
					},
					ccv2.Warnings{"warning-5", "warning-6"},
					nil)

				fakeCloudControllerClient.GetOrganizationQuotaReturns(
					ccv2.OrganizationQuota{
						GUID: "some-org-quota-guid",
						Name: "some-org-quota",
					},
					ccv2.Warnings{"warning-7", "warning-8"},
					nil)

				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{
						{
							GUID:     "space-2-guid",
							Name:     "space-2",
							AllowSSH: false,
						},
						{
							GUID:     "space-1-guid",
							Name:     "space-1",
							AllowSSH: true,
						},
					},
					ccv2.Warnings{"warning-9", "warning-10"},
					nil)
			})

			It("returns the organization summary and all warnings", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)[0].Values).To(ConsistOf("some-org"))
				Expect(fakeCloudControllerClient.GetSharedDomainsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsArgsForCall(0)).To(Equal("some-org-guid"))
				Expect(fakeCloudControllerClient.GetOrganizationQuotaCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationQuotaArgsForCall(0)).To(Equal("some-quota-definition-guid"))
				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)[0].Values).To(ConsistOf("some-org-guid"))

				Expect(orgSummary).To(Equal(OrganizationSummary{
					Organization: Organization{
						Name:                "some-org",
						GUID:                "some-org-guid",
						QuotaDefinitionGUID: "some-quota-definition-guid",
					},
					QuotaName:   "some-org-quota",
					DomainNames: []string{"private-domain-1", "private-domain-2", "shared-domain-1", "shared-domain-2"},
					SpaceNames:  []string{"space-1", "space-2"},
				}))
				Expect(warnings).To(ConsistOf([]string{"warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6", "warning-7", "warning-8", "warning-9", "warning-10"}))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when an error is encountered getting the organization", func() {
			BeforeEach(func() {
				expectedErr = errors.New("get-orgs-error")
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{
						"warning-1",
						"warning-2",
					},
					expectedErr,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("when the organization exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{
						{GUID: "some-org-guid"},
					},
					ccv2.Warnings{
						"warning-1",
						"warning-2",
					},
					nil,
				)
			})

			Context("when an error is encountered getting the organization domains", func() {
				BeforeEach(func() {
					expectedErr = errors.New("shared domains error")
					fakeCloudControllerClient.GetSharedDomainsReturns([]ccv2.Domain{}, ccv2.Warnings{"shared domains warning"}, expectedErr)
				})

				It("returns that error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "shared domains warning"))
				})
			})

			Context("when an error is encountered getting the organization quota", func() {
				BeforeEach(func() {
					expectedErr = errors.New("some org quota error")
					fakeCloudControllerClient.GetOrganizationQuotaReturns(ccv2.OrganizationQuota{}, ccv2.Warnings{"quota warning"}, expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "quota warning"))
				})
			})

			Context("when an error is encountered getting the organization spaces", func() {
				BeforeEach(func() {
					expectedErr = errors.New("cc-get-spaces-error")
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{},
						ccv2.Warnings{"spaces warning"},
						expectedErr,
					)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "spaces warning"))
				})
			})
		})
	})
})
