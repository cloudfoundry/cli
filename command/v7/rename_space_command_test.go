package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
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

var _ = Describe("rename-space Command", func() {
	var (
		cmd             RenameSpaceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		input           *Buffer
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = RenameSpaceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		cmd.RequiredArgs.OldSpaceName = "old-space-name"
		cmd.RequiredArgs.NewSpaceName = "new-space-name"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target errors", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns the NotLoggedInError", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeFalse())
		})

	})

	When("checking target succeeds", func() {
		When("getting the current user returns an error", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("some error")
				fakeConfig.CurrentUserReturns(configv3.User{}, returnedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(returnedErr))
			})
		})

		When("when the command succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: "username"}, nil)
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-targeted-org", GUID: "org-guid"})
				fakeActor.RenameSpaceByNameAndOrganizationGUIDReturns(
					v7action.Space{GUID: "old-space-guid", Name: "new-space-name"},
					v7action.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("renames the space in the targeted org", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(fakeActor.RenameSpaceByNameAndOrganizationGUIDCallCount()).To(Equal(1))
				oldSpaceName, newSpaceName, orgArg := fakeActor.RenameSpaceByNameAndOrganizationGUIDArgsForCall(0)
				Expect(oldSpaceName).To(Equal("old-space-name"))
				Expect(newSpaceName).To(Equal("new-space-name"))
				Expect(orgArg).To(Equal("org-guid"))
			})

			When("renaming a targeted space", func() {
				BeforeEach(func() {
					fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: "old-space-guid", Name: "old-space-name"})
				})

				It("targets the renamed space", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					Expect(fakeConfig.V7SetSpaceInformationCallCount()).To(Equal(1))
					newSpaceGUID, newSpaceName := fakeConfig.V7SetSpaceInformationArgsForCall(0)
					Expect(newSpaceGUID).To(Equal("old-space-guid"))
					Expect(newSpaceName).To(Equal("new-space-name"))
				})
			})
		})
	})
})
