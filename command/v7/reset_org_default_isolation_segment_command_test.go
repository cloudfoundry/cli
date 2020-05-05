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

var _ = Describe("reset-org-default-isolation-segment Command", func() {
	var (
		cmd             ResetOrgDefaultIsolationSegmentCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		orgName         string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = ResetOrgDefaultIsolationSegmentCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		orgName = "some-org"

		cmd.RequiredArgs.OrgName = orgName
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("checking file succeeds", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: orgName,
				GUID: "some-org-guid",
			})
		})

		When("the user is not logged in", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some current user error")
				fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
			})

			It("return an error", func() {
				Expect(executeErr).To(Equal(expectedErr))

				Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
			})
		})

		When("the user is logged in", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			})

			When("the org lookup is unsuccessful", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationByNameReturns(v7action.Organization{}, v7action.Warnings{"warning-1", "warning-2"}, actionerror.OrganizationNotFoundError{Name: orgName})
				})

				It("returns the warnings and error", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: orgName}))
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			When("the org lookup is successful", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationByNameReturns(v7action.Organization{
						Name: orgName,
						GUID: "some-org-guid",
					}, v7action.Warnings{"warning-1", "warning-2"}, nil)
				})

				When("the reset succeeds", func() {
					BeforeEach(func() {
						fakeActor.ResetOrganizationDefaultIsolationSegmentReturns(v7action.Warnings{"warning-3", "warning-4"}, nil)
					})

					It("displays the header and okay", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Resetting default isolation segment of org %s as banana...", orgName))

						Expect(testUI.Out).To(Say("OK\n\n"))

						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
						Expect(testUI.Err).To(Say("warning-3"))
						Expect(testUI.Err).To(Say("warning-4"))

						Expect(testUI.Out).To(Say("TIP: Restart applications in spaces without assigned isolation segments to move them to the platform default."))

						Expect(fakeActor.ResetOrganizationDefaultIsolationSegmentCallCount()).To(Equal(1))
						orgGUID := fakeActor.ResetOrganizationDefaultIsolationSegmentArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))
					})
				})

				When("the reset errors", func() {
					var expectedErr error
					BeforeEach(func() {
						expectedErr = errors.New("some error")
						fakeActor.ResetOrganizationDefaultIsolationSegmentReturns(v7action.Warnings{"warning-3", "warning-4"}, expectedErr)
					})

					It("returns the warnings and error", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Out).To(Say("Resetting default isolation segment of org %s as banana...", orgName))
						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
						Expect(testUI.Err).To(Say("warning-3"))
						Expect(testUI.Err).To(Say("warning-4"))
					})
				})
			})
		})
	})
})
