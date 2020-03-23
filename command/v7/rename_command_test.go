package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("rename Command", func() {
	var (
		input           *Buffer
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		cmd             RenameCommand
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = RenameCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		cmd.RequiredArgs.OldAppName = "old-app-name"
		cmd.RequiredArgs.NewAppName = "new-app-name"

		fakeConfig.CurrentUserReturns(configv3.User{Name: "username"}, nil)
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "targeted-org"})
		fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "targeted-space"})
		fakeActor.RenameApplicationByNameAndSpaceGUIDReturns(v7action.Application{Name: "new-app-name"}, v7action.Warnings{"rename-app-warning"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target succeeds", func() {

		It("renames the app", func() {
			Expect(testUI.Out).To(Say("Renaming app old-app-name to new-app-name in org targeted-org / space targeted-space as username..."))
			Expect(testUI.Err).To(Say("rename-app-warning"))
			Expect(testUI.Out).To(Say("OK"))
		})
	})

	When("app and space are not targeted", func() {
		var (
			returnedError error
		)
		BeforeEach(func() {
			returnedError = errors.New("no org found")
			fakeSharedActor.CheckTargetReturns(returnedError)
		})
		It("fails", func() {
			targetOrg, targetSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(targetOrg).To(BeTrue())
			Expect(targetSpace).To(BeTrue())
			Expect(executeErr).To(MatchError(returnedError))
		})
	})

	When("there is no CurrentUser", func() {
		var (
			returnedError error
		)
		BeforeEach(func() {
			returnedError = errors.New("current user not found")
			fakeConfig.CurrentUserReturns(configv3.User{}, returnedError)
		})
		It("returns the CurrentUser error", func() {
			Expect(executeErr).To(MatchError(returnedError))
		})
	})

	When("the actor returns an error", func() {
		var (
			returnedError error
		)
		BeforeEach(func() {
			returnedError = errors.New("app rename failed!")
			fakeActor.RenameApplicationByNameAndSpaceGUIDReturns(v7action.Application{Name: "new-app-name"}, v7action.Warnings{"rename-app-warning"}, returnedError)
		})

		It("returns the error", func() {
			Expect(testUI.Out).To(Say("Renaming app old-app-name to new-app-name in org targeted-org / space targeted-space as username..."))
			Expect(testUI.Err).To(Say("rename-app-warning"))
			Expect(executeErr).To(MatchError(returnedError))
		})
	})
})
