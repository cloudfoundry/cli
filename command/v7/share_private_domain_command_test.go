package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"

	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("share-private-domain command", func() {
	var (
		cmd             SharePrivateDomainCommand
		DomainName      = "some-domain-name"
		OrgName         = "some-org-name"
		fakeActor       *v7fakes.FakeSharePrivateDomainActor
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI
		binaryName      string

		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v7fakes.FakeSharePrivateDomainActor)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		cmd = SharePrivateDomainCommand{
			Actor:       fakeActor,
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
		}
		cmd.RequiredArgs = flag.OrgDomain{
			Organization: OrgName,
			Domain:       DomainName,
		}
		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the user is not logged in", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{})
		})

		It("checks target and returns the error", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkedOrg, checkedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkedOrg).To(Equal(false))
			Expect(checkedSpace).To(Equal(false))

			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{}))
		})
	})

	When("getting the current user fails", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("current-user-error"))
		})

		It("returns an error", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkedOrg, checkedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkedOrg).To(Equal(false))
			Expect(checkedSpace).To(Equal(false))

			Expect(executeErr).To(MatchError(errors.New("current-user-error")))
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "the-user"}, nil)
		})

		It("should print text indicating it is sharing a domain", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Sharing domain %s with org %s as the-user\.\.\.`, DomainName, OrgName))
		})

		When("sharing the domain errors", func() {
			BeforeEach(func() {
				fakeActor.SharePrivateDomainReturns(v7action.Warnings{"warnings-1", "warnings-2"}, errors.New("err-share-domain"))
			})

			It("returns an error and displays warnings", func() {
				Expect(executeErr).To(MatchError("err-share-domain"))
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
			})
		})

		When("sharing the domain is successful", func() {
			BeforeEach(func() {
				fakeActor.SharePrivateDomainReturns(v7action.Warnings{"warnings-1", "warnings-2"}, nil)
			})

			It("prints all warnings and OK", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
				Expect(testUI.Out).To(Say("OK"))
			})

			It("shares the domain", func() {
				Expect(fakeActor.SharePrivateDomainCallCount()).To(Equal(1))
				expectedDomainName, expectedOrgName := fakeActor.SharePrivateDomainArgsForCall(0)
				Expect(expectedDomainName).To(Equal(DomainName))
				Expect(expectedOrgName).To(Equal(OrgName))
			})
		})
	})

})
