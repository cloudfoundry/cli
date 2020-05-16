package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "code.cloudfoundry.org/cli/resources"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Organization Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("GetOrganizations", func() {
		var (
			returnOrganizations []Organization
			organizations       []Organization

			organization1Name string
			organization1GUID string

			organization2Name string
			organization2GUID string

			organization3Name string
			organization3GUID string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			returnOrganizations = []Organization{
				{Name: organization1Name, GUID: organization1GUID},
				{Name: organization2Name, GUID: organization2GUID},
				{Name: organization3Name, GUID: organization3GUID},
			}
		})

		When("the API layer call is successful", func() {

			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					returnOrganizations,
					ccv3.Warnings{"some-organizations-warning"},
					nil,
				)
			})

			JustBeforeEach(func() {
				organizations, warnings, executeErr = actor.GetOrganizations("")
			})

			It("does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("does not pass through the label selector", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				expectedQuery := []ccv3.Query{
					{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
				}
				actualQuery := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(actualQuery).To(Equal(expectedQuery))
			})

			It("returns back the organizations and warnings", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))

				Expect(organizations).To(ConsistOf(
					Organization{Name: organization1Name, GUID: organization1GUID},
					Organization{Name: organization2Name, GUID: organization2GUID},
					Organization{Name: organization3Name, GUID: organization3GUID},
				))
				Expect(warnings).To(ConsistOf("some-organizations-warning"))
			})
		})

		When("a label selector is provided", func() {
			JustBeforeEach(func() {
				organizations, warnings, executeErr = actor.GetOrganizations("some-label-selector")
			})

			It("passes the label selector through", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))

				expectedQuery := []ccv3.Query{
					{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
					{Key: ccv3.LabelSelectorFilter, Values: []string{"some-label-selector"}},
				}
				actualQuery := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(actualQuery).To(Equal(expectedQuery))
			})

		})

		When("when the API layer call returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]Organization{},
					ccv3.Warnings{"some-organizations-warning"},
					errors.New("some-organizations-error"),
				)
			})

			JustBeforeEach(func() {
				organizations, warnings, executeErr = actor.GetOrganizations("")
			})

			It("returns the error and prints warnings", func() {
				Expect(executeErr).To(MatchError("some-organizations-error"))
				Expect(warnings).To(ConsistOf("some-organizations-warning"))
				Expect(organizations).To(ConsistOf([]Organization{}))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
			})

		})
	})

	Describe("GetOrganizationByGUID", func() {
		When("the org exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationReturns(
					Organization{
						Name: "some-org-name",
						GUID: "some-org-guid",
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the organization and warnings", func() {
				org, warnings, err := actor.GetOrganizationByGUID("some-org-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(org).To(Equal(Organization{
					Name: "some-org-name",
					GUID: "some-org-guid",
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(0)).To(Equal(
					"some-org-guid",
				))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetOrganizationReturns(
					Organization{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.GetOrganizationByGUID("some-org-guid")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetOrganizationByName", func() {
		When("the org exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]Organization{
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
				org, warnings, err := actor.GetOrganizationByName("some-org-name")
				Expect(err).ToNot(HaveOccurred())
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
					[]Organization{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.GetOrganizationByName("some-org-name")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetDefaultDomain", func() {
		var (
			domain     Domain
			warnings   Warnings
			executeErr error

			orgGUID = "org-guid"
		)

		JustBeforeEach(func() {
			domain, warnings, executeErr = actor.GetDefaultDomain(orgGUID)
		})

		When("the api call is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDefaultDomainReturns(
					ccv3.Domain{
						Name: "some-domain-name",
						GUID: "some-domain-guid",
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the domain and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(domain).To(Equal(Domain{
					Name: "some-domain-name",
					GUID: "some-domain-guid",
				}))
				Expect(warnings).To(ConsistOf("some-warning"))

				Expect(fakeCloudControllerClient.GetDefaultDomainCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetDefaultDomainArgsForCall(0)).To(Equal(orgGUID))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetDefaultDomainReturns(
					ccv3.Domain{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(executeErr).To(MatchError(expectedError))
			})
		})
	})

	When("the org does not exist", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetOrganizationsReturns(
				[]Organization{},
				ccv3.Warnings{"some-warning"},
				nil,
			)
		})

		It("returns an OrganizationNotFoundError and the warnings", func() {
			_, warnings, err := actor.GetOrganizationByName("some-org-name")
			Expect(warnings).To(ConsistOf("some-warning"))
			Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org-name"}))
		})
	})

	Describe("CreateOrganization", func() {
		var (
			org      Organization
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			org, warnings, err = actor.CreateOrganization("some-org-name")
		})

		When("the org is created successfully", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateOrganizationReturns(
					Organization{Name: "some-org-name", GUID: "some-org-guid"},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the org resource", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(org).To(Equal(Organization{Name: "some-org-name", GUID: "some-org-guid"}))
			})
		})

		When("the request fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateOrganizationReturns(
					Organization{},
					ccv3.Warnings{"warning-1", "warning-2"},
					errors.New("create-org-failed"),
				)
			})

			It("returns warnings and the client error", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(err).To(MatchError("create-org-failed"))
			})
		})
	})

	Describe("DeleteOrganization", func() {
		var (
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			warnings, err = actor.DeleteOrganization("some-org")
		})

		When("the org is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]Organization{},
					ccv3.Warnings{
						"warning-1",
						"warning-2",
					},
					nil,
				)
			})

			It("returns an OrganizationNotFoundError", func() {
				Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the org is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]Organization{{Name: "some-org", GUID: "some-org-guid"}},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			When("the delete returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some delete space error")
					fakeCloudControllerClient.DeleteOrganizationReturns(
						ccv3.JobURL(""),
						ccv3.Warnings{"warning-5", "warning-6"},
						expectedErr,
					)
				})

				It("returns the error", func() {
					Expect(err).To(Equal(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-5", "warning-6"))
				})
			})

			When("the delete returns a job", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteOrganizationReturns(
						ccv3.JobURL("some-url"),
						ccv3.Warnings{"warning-5", "warning-6"},
						nil,
					)
				})

				When("polling errors", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("Never expected, by anyone")
						fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"warning-7", "warning-8"}, expectedErr)
					})

					It("returns the error", func() {
						Expect(err).To(Equal(expectedErr))
						Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-5", "warning-6", "warning-7", "warning-8"))
					})
				})

				When("the job is successful", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"warning-7", "warning-8"}, nil)
					})

					It("returns warnings and no error", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-5", "warning-6", "warning-7", "warning-8"))

						Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(Equal([]ccv3.Query{{
							Key:    ccv3.NameFilter,
							Values: []string{"some-org"},
						}}))

						Expect(fakeCloudControllerClient.DeleteOrganizationCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.DeleteOrganizationArgsForCall(0)).To(Equal("some-org-guid"))

						Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(Equal(ccv3.JobURL("some-url")))
					})
				})
			})
		})
	})

	Describe("RenameOrganization", func() {
		var (
			oldOrgName = "old-and-stale-org-name"
			newOrgName = "fresh-and-new-org-name"

			org        Organization
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			org, warnings, executeErr = actor.RenameOrganization(
				oldOrgName,
				newOrgName,
			)
		})

		It("delegate to the actor to get the org", func() {
			// assert on the underlying client call because we dont have a fake actor
			Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
		})

		When("getting the org fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					nil,
					ccv3.Warnings{"get-org-warning"},
					errors.New("get-org-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("get-org-error"))
				Expect(warnings).To(ConsistOf("get-org-warning"))
			})
		})

		When("getting the org succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]Organization{{Name: oldOrgName, GUID: "org-guid"}},
					ccv3.Warnings{"get-org-warning"},
					nil,
				)
			})

			It("delegates to the client to update the org", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateOrganizationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateOrganizationArgsForCall(0)).To(Equal(Organization{
					GUID: "org-guid",
					Name: newOrgName,
				}))
			})

			When("updating the org fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateOrganizationReturns(
						Organization{},
						ccv3.Warnings{"update-org-warning"},
						errors.New("update-org-error"),
					)
				})

				It("returns an error and all warnings", func() {
					Expect(executeErr).To(MatchError("update-org-error"))
					Expect(warnings).To(ConsistOf("get-org-warning", "update-org-warning"))
				})

			})

			When("updating the org succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateOrganizationReturns(
						Organization{Name: newOrgName, GUID: "org-guid"},
						ccv3.Warnings{"update-org-warning"},
						nil,
					)
				})

				It("returns warnings and no error", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-org-warning", "update-org-warning"))
					Expect(org).To(Equal(Organization{Name: newOrgName, GUID: "org-guid"}))
				})
			})
		})
	})
})
