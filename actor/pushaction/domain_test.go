package pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domains", func() {
	var (
		actor       *Actor
		fakeV2Actor *pushactionfakes.FakeV2Actor
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		actor = NewActor(fakeV2Actor, nil)
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

		Context("when retrieving the domains is successful", func() {
			BeforeEach(func() {
				fakeV2Actor.GetOrganizationDomainsReturns([]v2action.Domain{
					{
						Name: "private-domain.com",
						GUID: "some-private-domain-guid",
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

			It("returns the first domain and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
				Expect(defaultDomain).To(Equal(v2action.Domain{
					Name: "private-domain.com",
					GUID: "some-private-domain-guid",
				}))

				Expect(fakeV2Actor.GetOrganizationDomainsCallCount()).To(Equal(1))
				Expect(fakeV2Actor.GetOrganizationDomainsArgsForCall(0)).To(Equal(orgGUID))
			})
		})

		Context("no domains exist", func() {
			BeforeEach(func() {
				fakeV2Actor.GetOrganizationDomainsReturns([]v2action.Domain{}, v2action.Warnings{"private-domain-warnings", "shared-domain-warnings"}, nil)
			})

			It("returns the first shared domain and warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.NoDomainsFoundError{OrganizationGUID: orgGUID}))
				Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
			})
		})

		Context("when retrieving the domains errors", func() {
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
