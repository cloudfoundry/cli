package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/clock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Organization Summary Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		orgSummary                OrganizationSummary
		warnings                  Warnings
		err                       error
		expectedErr               error
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, clock.NewClock())
	})

	JustBeforeEach(func() {
		orgSummary, warnings, err = actor.GetOrganizationSummaryByName("some-org")
	})

	Describe("GetOrganizationSummaryByName", func() {
		When("no errors are encountered", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]resources.Organization{
						{
							GUID:      "some-org-guid",
							Name:      "some-org",
							QuotaGUID: "org-quota-guid",
						},
					},
					ccv3.Warnings{"get-org-warning-1", "get-org-warning-2"},
					nil)

				fakeCloudControllerClient.GetOrganizationDomainsReturns(
					[]resources.Domain{
						{
							GUID: "shared-domain-guid-2",
							Name: "shared-domain-2",
						},
						{
							GUID: "shared-domain-guid-1",
							Name: "shared-domain-1",
						},
					},
					ccv3.Warnings{"domain-warning-1", "domain-warning-2"},
					nil)

				fakeCloudControllerClient.GetOrganizationQuotaReturns(
					resources.OrganizationQuota{Quota: resources.Quota{Name: "my-quota", GUID: "quota-guid"}},
					ccv3.Warnings{"get-quota-warning-1"}, nil)

				fakeCloudControllerClient.GetSpacesReturns(
					[]resources.Space{
						{
							GUID: "space-2-guid",
							Name: "space-2",
						},
						{
							GUID: "space-1-guid",
							Name: "space-1",
						},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"space-warning-1", "space-warning-2"},
					nil)

				fakeCloudControllerClient.GetOrganizationDefaultIsolationSegmentReturns(
					resources.Relationship{
						GUID: "default-iso-seg-guid",
					},
					ccv3.Warnings{"iso-seg-warning-1", "iso-seg-warning-2"},
					nil)
			})

			It("returns the organization summary and all warnings", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)[0].Values).To(ConsistOf("some-org"))

				Expect(fakeCloudControllerClient.GetOrganizationDomainsCallCount()).To(Equal(1))
				orgGuid, labelSelector := fakeCloudControllerClient.GetOrganizationDomainsArgsForCall(0)
				Expect(orgGuid).To(Equal("some-org-guid"))
				Expect(labelSelector).To(Equal([]ccv3.Query{}))

				Expect(fakeCloudControllerClient.GetOrganizationQuotaCallCount()).To(Equal(1))
				givenQuotaGUID := fakeCloudControllerClient.GetOrganizationQuotaArgsForCall(0)
				Expect(givenQuotaGUID).To(Equal("org-quota-guid"))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)[0].Values).To(ConsistOf("some-org-guid"))

				Expect(orgSummary).To(Equal(OrganizationSummary{
					Organization: resources.Organization{
						Name:      "some-org",
						GUID:      "some-org-guid",
						QuotaGUID: "org-quota-guid",
					},
					DomainNames:                 []string{"shared-domain-1", "shared-domain-2"},
					QuotaName:                   "my-quota",
					SpaceNames:                  []string{"space-1", "space-2"},
					DefaultIsolationSegmentGUID: "default-iso-seg-guid",
				}))
				Expect(warnings).To(ConsistOf(
					"get-org-warning-1",
					"get-org-warning-2",
					"domain-warning-1",
					"domain-warning-2",
					"get-quota-warning-1",
					"space-warning-1",
					"space-warning-2",
					"iso-seg-warning-1",
					"iso-seg-warning-2",
				))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("an error is encountered getting the organization", func() {
			BeforeEach(func() {
				expectedErr = errors.New("get-orgs-error")
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]resources.Organization{},
					ccv3.Warnings{
						"get-org-warning-1",
						"get-org-warning-2",
					},
					expectedErr,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get-org-warning-1", "get-org-warning-2"))
			})
		})

		When("the organization exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]resources.Organization{
						{GUID: "some-org-guid"},
					},
					ccv3.Warnings{
						"get-org-warning-1",
						"get-org-warning-2",
					},
					nil,
				)
			})

			When("an error is encountered getting the organization domains", func() {
				BeforeEach(func() {
					expectedErr = errors.New("domains error")
					fakeCloudControllerClient.GetOrganizationDomainsReturns([]resources.Domain{}, ccv3.Warnings{"domains warning"}, expectedErr)
				})

				It("returns that error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("get-org-warning-1", "get-org-warning-2", "domains warning"))
				})
			})

			When("an error is encountered getting the organization quota", func() {
				BeforeEach(func() {
					expectedErr = errors.New("quota error")
					fakeCloudControllerClient.GetOrganizationQuotaReturns(resources.OrganizationQuota{}, ccv3.Warnings{"Quota warning"}, expectedErr)
				})

				It("returns that error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("get-org-warning-1", "get-org-warning-2", "Quota warning"))
				})
			})

			When("an error is encountered getting the organization spaces", func() {
				BeforeEach(func() {
					expectedErr = errors.New("cc-get-spaces-error")
					fakeCloudControllerClient.GetSpacesReturns(
						[]resources.Space{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"spaces warning"},
						expectedErr,
					)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("get-org-warning-1", "get-org-warning-2", "spaces warning"))
				})
			})
		})
	})
})
