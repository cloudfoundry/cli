package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-org-quota Command", func() {

	var (
		cmd             DeleteOrgQuotaCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		input           *Buffer
		binaryName      string
		quotaName       string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		fakeActor = new(v7fakes.FakeActor)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)

		cmd = DeleteOrgQuotaCommand{
			BaseCommand: command.BaseCommand{
				Actor:       fakeActor,
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
			},
		}

		quotaName = "some-quota"
		cmd.RequiredArgs.Quota = quotaName
		cmd.Force = true
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error if the check fails", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			shouldCheckTargetedOrg, shouldCheckTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(shouldCheckTargetedOrg).To(BeFalse())
			Expect(shouldCheckTargetedSpace).To(BeFalse())
		})
	})

	When("the deletion completes successfully", func() {
		BeforeEach(func() {
			fakeActor.DeleteOrganizationQuotaReturns(v7action.Warnings{"some-warning"}, nil)
		})

		When("--force is specified", func() {
			BeforeEach(func() {
				cmd.Force = true
			})

			It("calls the actor method correctly", func() {
				Expect(fakeActor.DeleteOrganizationQuotaCallCount()).To(Equal(1))

				givenQuotaName := fakeActor.DeleteOrganizationQuotaArgsForCall(0)
				Expect(givenQuotaName).To(Equal(quotaName))
			})

			It("prints warnings and appropriate output", func() {
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say("Deleting org quota some-quota as some-user..."))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("--force is not specified", func() {
			BeforeEach(func() {
				cmd.Force = false
			})

			When("the user inputs yes", func() {
				BeforeEach(func() {
					_, err := input.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("prompted the user for confirmation", func() {
					Expect(testUI.Out).To(Say("Really delete the org quota some-quota?"))
					Expect(testUI.Out).To(Say("Deleting org quota some-quota as some-user..."))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("the user inputs no", func() {
				BeforeEach(func() {
					_, err := input.Write([]byte("n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("cancels the delete", func() {
					Expect(testUI.Out).To(Say("Really delete the org quota some-quota?"))
					Expect(testUI.Out).To(Say("'some-quota' has not been deleted."))
					Expect(testUI.Out).NotTo(Say("Deleting org quota some-quota as some-user..."))
				})
			})
		})
	})

	When("the deletion request returns an error", func() {
		BeforeEach(func() {
			fakeActor.DeleteOrganizationQuotaReturns(
				v7action.Warnings{"a-warning"},
				errors.New("uh oh"),
			)
		})

		It("prints warnings and returns error", func() {
			Expect(executeErr).To(MatchError("uh oh"))
			Expect(testUI.Err).To(Say("a-warning"))
		})
	})

	When("the quota does not exist", func() {
		BeforeEach(func() {
			fakeActor.DeleteOrganizationQuotaReturns(
				v7action.Warnings{"a-warning"},
				actionerror.OrganizationQuotaNotFoundForNameError{Name: quotaName},
			)
		})

		It("prints warnings and helpful message, but exits with OK", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Err).To(Say("a-warning"))
			Expect(testUI.Err).To(Say(`Organization quota with name 'some-quota' not found\.`))
			Expect(testUI.Out).To(Say(`OK`))
		})
	})
})
