package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Organization Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("GetOrganizationByName", func() {
		var (
			executeErr error
			warnings   Warnings
			org        Organization
		)
		JustBeforeEach(func() {
			org, warnings, executeErr = actor.GetOrganizationByName("some-org-name")
		})

		When("the org exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv3.Organization{
						{
							Name: "some-org-name",
							GUID: "some-org-guid",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the organization and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(org).To(Equal(Organization{
					Name: "some-org-name",
					GUID: "some-org-guid",
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-org-name"}},
				))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv3.Organization{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(executeErr).To(MatchError(expectedError))
			})
		})

		When("the org does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv3.Organization{},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns an OrganizationNotFoundError and the warnings", func() {
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org-name"}))
			})
		})
	})

	Describe("GetOrganizationsByGUIDs", func() {
		When("the orgs exists", func() {
			var (
				executeErr error
				warnings   Warnings
				orgs       []Organization
			)
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv3.Organization{
						{
							Name: "some-org-name",
							GUID: "some-org-guid",
						},
						{
							Name: "another-org-name",
							GUID: "another-org-guid",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			JustBeforeEach(func() {
				orgs, warnings, executeErr = actor.GetOrganizationsByGUIDs("some-org-guid", "another-org-guid")
			})

			It("returns the organizations and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(orgs).To(ConsistOf(
					Organization{
						Name: "some-org-name",
						GUID: "some-org-guid",
					},
					Organization{
						Name: "another-org-name",
						GUID: "another-org-guid",
					},
				))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{"some-org-guid", "another-org-guid"}},
				))
			})

			When("the cloud controller client returns an error", func() {
				var expectedError error

				BeforeEach(func() {
					expectedError = errors.New("I am a CloudControllerClient Error")
					fakeCloudControllerClient.GetOrganizationsReturns(
						[]ccv3.Organization{},
						ccv3.Warnings{"some-warning"},
						expectedError)
				})

				It("returns the warnings and the error", func() {
					Expect(warnings).To(ConsistOf("some-warning"))
					Expect(executeErr).To(MatchError(expectedError))
				})
			})
		})
	})

})
