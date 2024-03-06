package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
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
					[]resources.Organization{
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
					[]resources.Organization{},
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
					[]resources.Organization{},
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
					[]resources.Organization{
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
				fakeCloudControllerClient.CloudControllerAPIVersionReturns("3.65.0")
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
						[]resources.Organization{},
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

	Describe("GetOrganizations", func() {
		var (
			executeErr error
			warnings   Warnings
			orgs       []Organization
		)

		JustBeforeEach(func() {
			orgs, warnings, executeErr = actor.GetOrganizations()
		})

		It("fetches all the organizations", func() {
			Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
		})

		When("no error occurs", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]resources.Organization{
						resources.Organization{
							Name: "some-org-1",
							GUID: "some-org-guid-1",
						},
						resources.Organization{
							Name: "some-org-2",
							GUID: "some-org-guid-2",
						},
					},
					ccv3.Warnings{
						"some-warning-1",
						"some-warning-2",
					},
					nil,
				)
			})

			It("returns the organizations", func() {
				Expect(orgs).To(Equal(
					[]Organization{
						Organization{
							Name: "some-org-1",
							GUID: "some-org-guid-1",
						},
						Organization{
							Name: "some-org-2",
							GUID: "some-org-guid-2",
						},
					},
				))
			})

			It("returns all the warnings", func() {
				Expect(warnings).To(Equal(Warnings{"some-warning-1", "some-warning-2"}))
			})
		})

		When("an error occurs", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]resources.Organization{},
					ccv3.Warnings{
						"some-warning-1",
						"some-warning-2",
					},
					errors.New("get-organization-error"),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("get-organization-error"))
			})

			It("returns the warnings", func() {
				Expect(warnings).To(Equal(Warnings{"some-warning-1", "some-warning-2"}))
			})
		})
	})

})
