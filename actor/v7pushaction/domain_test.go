package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domains", func() {
	var (
		actor       *Actor
		fakeV2Actor *v7pushactionfakes.FakeV2Actor
	)

	BeforeEach(func() {
		actor, fakeV2Actor, _, _ = getTestPushActor()
	})

	Describe("DefaultDomain", func() {
		var (
			orgGUID       string
			defaultDomain v2action.Domain
			warnings      Warnings
			executeErr    error
		)

		BeforeEach(func() {
			orgGUID = "some-org-guid"
		})

		JustBeforeEach(func() {
			defaultDomain, warnings, executeErr = actor.DefaultDomain(orgGUID)
		})

		When("retrieving the domains is successful", func() {
			When("there is an external domain", func() {
				BeforeEach(func() {
					fakeV2Actor.GetOrganizationDomainsReturns([]v2action.Domain{
						{
							Name:     "private-domain.com",
							GUID:     "some-private-domain-guid",
							Internal: true,
						},
						{
							Name: "private-domain-2.com",
							GUID: "some-private-domain-guid-2",
						},
						{
							Name: "shared-domain.com",
							GUID: "some-shared-domain-guid",
						},
					},
						v2action.Warnings{"private-domain-warnings", "shared-domain-warnings"},
						nil,
					)
				})

				It("returns the first external domain and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
					Expect(defaultDomain).To(Equal(v2action.Domain{
						Name: "private-domain-2.com",
						GUID: "some-private-domain-guid-2",
					}))

					Expect(fakeV2Actor.GetOrganizationDomainsCallCount()).To(Equal(1))
					Expect(fakeV2Actor.GetOrganizationDomainsArgsForCall(0)).To(Equal(orgGUID))
				})
			})

			When("all the domains are internal", func() {
				BeforeEach(func() {
					fakeV2Actor.GetOrganizationDomainsReturns([]v2action.Domain{
						{
							Name:     "internal-domain-1.com",
							GUID:     "some-internal-domain-guid",
							Internal: true,
						},
						{
							Name:     "internal-domain-2.com",
							GUID:     "some-internal-domain-guid",
							Internal: true,
						},
					},
						v2action.Warnings{"warnings-1", "warnings-2"},
						nil,
					)
				})

				It("returns an error and warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.NoDomainsFoundError{OrganizationGUID: orgGUID}))
					Expect(warnings).To(ConsistOf("warnings-1", "warnings-2"))
				})

			})
		})

		Context("no domains exist", func() {
			BeforeEach(func() {
				fakeV2Actor.GetOrganizationDomainsReturns([]v2action.Domain{}, v2action.Warnings{"private-domain-warnings", "shared-domain-warnings"}, nil)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.NoDomainsFoundError{OrganizationGUID: orgGUID}))
				Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
			})
		})

		When("retrieving the domains errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("whoops")
				fakeV2Actor.GetOrganizationDomainsReturns([]v2action.Domain{}, v2action.Warnings{"private-domain-warnings", "shared-domain-warnings"}, expectedErr)
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
			})
		})
	})
})
