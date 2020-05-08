package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("reset-space-isolation-segment Command", func() {
	var (
		cmd             v7.ResetSpaceIsolationSegmentCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
		space           string
		org             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = v7.ResetSpaceIsolationSegmentCommand{
			BaseCommand: command.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		space = "some-space"
		org = "some-org"
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

	When("the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: org,
				GUID: "some-org-guid",
			})

			cmd.RequiredArgs.SpaceName = space
		})

		When("the space lookup is unsuccessful", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceByNameAndOrganizationReturns(v7action.Space{}, v7action.Warnings{"warning-1", "warning-2"}, actionerror.SpaceNotFoundError{Name: space})
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: space}))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
			})
		})

		When("the space lookup is successful", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceByNameAndOrganizationReturns(v7action.Space{
					Name: space,
					GUID: "some-space-guid",
				}, v7action.Warnings{"warning-1", "warning-2"}, nil)
			})

			When("the reset changes the isolation segment to platform default", func() {
				BeforeEach(func() {
					fakeActor.ResetSpaceIsolationSegmentReturns("", v7action.Warnings{"warning-3", "warning-4"}, nil)
				})

				It("Displays the header and okay", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Resetting isolation segment assignment of space %s in org %s as banana...", space, org))

					Expect(testUI.Out).To(Say("OK\n\n"))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("warning-3"))
					Expect(testUI.Err).To(Say("warning-4"))

					Expect(testUI.Out).To(Say("TIP: Restart applications in this space to relocate them to the platform default."))

					Expect(fakeActor.ResetSpaceIsolationSegmentCallCount()).To(Equal(1))
					orgGUID, spaceGUID := fakeActor.ResetSpaceIsolationSegmentArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			When("the reset changes the isolation segment to the org's default", func() {
				BeforeEach(func() {
					fakeActor.ResetSpaceIsolationSegmentReturns("some-org-iso-seg-name", v7action.Warnings{"warning-3", "warning-4"}, nil)
				})

				It("Displays the header and okay", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Resetting isolation segment assignment of space %s in org %s as banana...", space, org))

					Expect(testUI.Out).To(Say("OK\n\n"))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("warning-3"))
					Expect(testUI.Err).To(Say("warning-4"))

					Expect(testUI.Out).To(Say("TIP: Restart applications in this space to relocate them to this organization's default isolation segment."))

					Expect(fakeActor.ResetSpaceIsolationSegmentCallCount()).To(Equal(1))
					orgGUID, spaceGUID := fakeActor.ResetSpaceIsolationSegmentArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			When("the reset errors", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("some error")
					fakeActor.ResetSpaceIsolationSegmentReturns("some-org-iso-seg", v7action.Warnings{"warning-3", "warning-4"}, expectedErr)
				})

				It("returns the warnings and error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Out).To(Say("Resetting isolation segment assignment of space %s in org %s as banana...", space, org))
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("warning-3"))
					Expect(testUI.Err).To(Say("warning-4"))
				})
			})
		})
	})
})
