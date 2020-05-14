package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("rename-org Command", func() {
	var (
		cmd             RenameOrgCommand
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

		cmd = RenameOrgCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		cmd.RequiredArgs.OldOrgName = "old-org-name"
		cmd.RequiredArgs.NewOrgName = "new-org-name"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
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
				fakeActor.RenameOrganizationReturns(
					resources.Organization{GUID: "old-org-guid", Name: "new-org-name"},
					v7action.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("renames the org", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(fakeActor.RenameOrganizationCallCount()).To(Equal(1))
				oldOrgName, newOrgName := fakeActor.RenameOrganizationArgsForCall(0)
				Expect(oldOrgName).To(Equal("old-org-name"))
				Expect(newOrgName).To(Equal("new-org-name"))
			})

			When("renaming a targeted org", func() {
				BeforeEach(func() {
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: "old-org-guid", Name: "old-org-name"})
				})

				It("renames the org and records it in the config file", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
					newOrgGUID, newOrgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
					Expect(newOrgGUID).To(Equal("old-org-guid"))
					Expect(newOrgName).To(Equal("new-org-name"))
				})
			})
		})
	})
})
