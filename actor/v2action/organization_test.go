package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Org Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("DeleteOrganization", func() {
		var (
			warnings     Warnings
			deleteOrgErr error
		)

		JustBeforeEach(func() {
			warnings, deleteOrgErr = actor.DeleteOrganization("some-org")
		})

		Context("when there are no errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv2.Organization{
					{GUID: "some-org-guid"},
				}, ccv2.Warnings{"get-org-warning"}, nil)

				fakeCloudControllerClient.DeleteOrganizationReturns(
					ccv2.Warnings{"delete-org-warning"}, nil)
			})

			It("returns warnings and deletes the org", func() {
				Expect(warnings).To(ConsistOf("get-org-warning", "delete-org-warning"))
				Expect(deleteOrgErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(query).To(Equal(
					[]ccv2.Query{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Value:    "some-org",
					}}))

				Expect(fakeCloudControllerClient.DeleteOrganizationCallCount()).To(Equal(1))
				orgGuid := fakeCloudControllerClient.DeleteOrganizationArgsForCall(0)
				Expect(orgGuid).To(Equal("some-org-guid"))
			})
		})

		Context("when the org is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{
						"get-org-warning",
					},
					nil,
				)
			})

			It("returns an error and all warnings", func() {
				Expect(warnings).To(ConsistOf("get-org-warning"))
				Expect(deleteOrgErr).To(MatchError(OrganizationNotFoundError{
					Name: "some-org",
				}))
			})
		})

		Context("when more than one org is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{
						{GUID: "org-1-guid"},
						{GUID: "org-2-guid"},
					},
					ccv2.Warnings{
						"get-org-warning",
					},
					nil,
				)
			})

			It("returns an error and all warnings", func() {
				Expect(warnings).To(ConsistOf("get-org-warning"))
				Expect(deleteOrgErr).To(MatchError(
					"Organization name 'some-org' matches multiple GUIDs: org-1-guid, org-2-guid"))
			})
		})

		Context("when getting the org returns an error", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("get-orgs-error")
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{
						"get-org-warning",
					},
					returnedErr,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(deleteOrgErr).To(MatchError(returnedErr))
				Expect(warnings).To(ConsistOf("get-org-warning"))
			})
		})

		Context("when the delete returns an error", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("delete-org-error")

				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{GUID: "org-1-guid"}},
					ccv2.Warnings{
						"get-org-warning",
					},
					nil,
				)

				fakeCloudControllerClient.DeleteOrganizationReturns(
					ccv2.Warnings{"delete-org-warning"},
					returnedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(deleteOrgErr).To(MatchError(returnedErr))
				Expect(warnings).To(ConsistOf("get-org-warning", "delete-org-warning"))
			})
		})
	})
})
