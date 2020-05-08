package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "code.cloudfoundry.org/cli/command/v7"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("unset-space-quota Command", func() {
	var (
		cmd             UnsetSpaceQuotaCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		input           *Buffer

		unsetQuotaWarning string
		currentUser       = "some-user"
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		unsetQuotaWarning = RandomString("unset-quota-warning")

		cmd = UnsetSpaceQuotaCommand{
			BaseCommand: command.BaseCommand{
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
			fakeSharedActor.CheckTargetReturns(errors.New("check-target-failure"))
		})

		It("does not require a targeted org or space", func() {
			targetedOrgRequired, targetedSpaceRequired := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(targetedOrgRequired).To(BeTrue())
			Expect(targetedSpaceRequired).To(BeFalse())
		})

		It("should return the error", func() {
			Expect(executeErr).To(MatchError("check-target-failure"))
		})
	})

	When("getting the current user fails", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserNameReturns("", errors.New("current-user-error"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("current-user-error"))
		})
	})

	When("unsetting the quota succeeds", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.UnsetSpaceQuotaArgs{}
			cmd.RequiredArgs.Space = RandomString("space-name")
			cmd.RequiredArgs.SpaceQuota = RandomString("space-quota-name")

			fakeActor.UnsetSpaceQuotaReturns(
				v7action.Warnings{unsetQuotaWarning},
				nil)

			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org-name",
				GUID: "some-org-guid",
			})

			fakeConfig.CurrentUserNameReturns(currentUser, nil)

		})

		It("unsets the quota to the organization", func() {
			Expect(fakeActor.UnsetSpaceQuotaCallCount()).To(Equal(1))
			spaceQuotaName, spaceName, orgGUID := fakeActor.UnsetSpaceQuotaArgsForCall(0)
			Expect(spaceQuotaName).To(Equal(cmd.RequiredArgs.SpaceQuota))
			Expect(spaceName).To(Equal(cmd.RequiredArgs.Space))
			Expect(orgGUID).To(Equal("some-org-guid"))
		})

		It("displays warnings and returns without error", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Err).To(Say(unsetQuotaWarning))
		})

		It("displays flavor text and returns without error", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(testUI.Out).To(Say("Unassigning space quota %s from space %s as %s...", cmd.RequiredArgs.SpaceQuota, cmd.RequiredArgs.Space, currentUser))
			Expect(testUI.Out).To(Say("OK"))
		})
	})

	When("the quota is already dissociated", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.UnsetSpaceQuotaArgs{}
			cmd.RequiredArgs.Space = RandomString("space-name")
			cmd.RequiredArgs.SpaceQuota = RandomString("space-quota-name")

			fakeActor.UnsetSpaceQuotaReturns(
				v7action.Warnings{unsetQuotaWarning},
				nil)

			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org-name",
				GUID: "some-org-guid",
			})

			fakeConfig.CurrentUserNameReturns(currentUser, nil)

		})

		It("unsets the quota to the organization", func() {
			Expect(fakeActor.UnsetSpaceQuotaCallCount()).To(Equal(1))
			spaceQuotaName, spaceName, orgGUID := fakeActor.UnsetSpaceQuotaArgsForCall(0)
			Expect(spaceQuotaName).To(Equal(cmd.RequiredArgs.SpaceQuota))
			Expect(spaceName).To(Equal(cmd.RequiredArgs.Space))
			Expect(orgGUID).To(Equal("some-org-guid"))
		})

		It("displays warnings and returns without error", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Err).To(Say(unsetQuotaWarning))
		})

		It("displays flavor text and returns without error", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(testUI.Out).To(Say("Unassigning space quota %s from space %s as %s...", cmd.RequiredArgs.SpaceQuota, cmd.RequiredArgs.Space, currentUser))
			Expect(testUI.Out).To(Say("OK"))
		})
	})
})
