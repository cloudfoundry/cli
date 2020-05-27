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

var _ = Describe("create-private-domain Command", func() {
	var (
		cmd             CreatePrivateDomainCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor

		executeErr error

		binaryName string
		domainName string
		orgName    string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		domainName = "example.com"
		orgName = "my-org"
		cmd = CreatePrivateDomainCommand{
			RequiredArgs: flag.OrgDomain{
				Organization: orgName,
				Domain:       domainName,
			},
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the environment is not set up correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "the-user"}, nil)
		})

		It("should print text indicating it is creating a domain", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Creating private domain %s for org %s as the-user\.\.\.`, domainName, orgName))
		})

		When("creating the domain errors", func() {
			BeforeEach(func() {
				fakeActor.CreatePrivateDomainReturns(v7action.Warnings{"warnings-1", "warnings-2"}, errors.New("err-create-domain"))
			})

			It("returns an error and displays warnings", func() {
				Expect(executeErr).To(MatchError("err-create-domain"))
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
			})
		})

		When("creating the domain is successful", func() {
			BeforeEach(func() {
				fakeActor.CreatePrivateDomainReturns(v7action.Warnings{"warnings-1", "warnings-2"}, nil)
			})

			It("prints all warnings, ok and then a tip", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).To(Say("TIP: Domain '%s' is a private domain. Run 'cf share-private-domain' to share this domain with a different org.", domainName))
			})

			It("creates the domain", func() {
				Expect(fakeActor.CreatePrivateDomainCallCount()).To(Equal(1))
				expectedDomainName, expectedOrgName := fakeActor.CreatePrivateDomainArgsForCall(0)
				Expect(expectedDomainName).To(Equal(domainName))
				Expect(expectedOrgName).To(Equal(orgName))
			})
		})
	})
})
