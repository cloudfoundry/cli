package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

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
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)
	})

	Describe("GetOrganizationByGUID", func() {
		When("the org exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationReturns(
					ccv3.Organization{
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
					ccv3.Organization{},
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
					[]ccv3.Organization{},
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
				[]ccv3.Organization{},
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
					[]ccv3.Organization{},
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
					[]ccv3.Organization{{Name: "some-org", GUID: "some-org-guid"}},
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
})
