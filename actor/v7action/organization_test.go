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
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
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
})
