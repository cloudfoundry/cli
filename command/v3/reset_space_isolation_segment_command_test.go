package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("reset-space-isolation-segment Command", func() {
	var (
		cmd             v3.ResetSpaceIsolationSegmentCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeResetSpaceIsolationSegmentActor
		fakeActorV2     *v3fakes.FakeResetSpaceIsolationSegmentActorV2
		binaryName      string
		executeErr      error
		space           string
		org             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeResetSpaceIsolationSegmentActor)
		fakeActorV2 = new(v3fakes.FakeResetSpaceIsolationSegmentActorV2)

		cmd = v3.ResetSpaceIsolationSegmentCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			ActorV2:     fakeActorV2,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		space = "some-space"
		org = "some-org"

		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionIsolationSegmentV3)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the API version is below the minimum", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns("0.0.0")
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumAPIVersionNotMetError{
				CurrentVersion: "0.0.0",
				MinimumVersion: ccversion.MinVersionIsolationSegmentV3,
			}))
		})
	})

	Context("when checking target fails", func() {
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

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: org,
				GUID: "some-org-guid",
			})

			cmd.RequiredArgs.SpaceName = space
		})

		Context("when the space lookup is unsuccessful", func() {
			BeforeEach(func() {
				fakeActorV2.GetSpaceByOrganizationAndNameReturns(v2action.Space{}, v2action.Warnings{"warning-1", "warning-2"}, actionerror.SpaceNotFoundError{Name: space})
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: space}))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
			})
		})

		Context("when the space lookup is successful", func() {
			BeforeEach(func() {
				fakeActorV2.GetSpaceByOrganizationAndNameReturns(v2action.Space{
					Name: space,
					GUID: "some-space-guid",
				}, v2action.Warnings{"warning-1", "warning-2"}, nil)
			})

			Context("when the reset changes the isolation segment to platform default", func() {
				BeforeEach(func() {
					fakeActor.ResetSpaceIsolationSegmentReturns("", v3action.Warnings{"warning-3", "warning-4"}, nil)
				})

				It("Displays the header and okay", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Resetting isolation segment assignment of space %s in org %s as banana...", space, org))

					Expect(testUI.Out).To(Say("OK\n\n"))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("warning-3"))
					Expect(testUI.Err).To(Say("warning-4"))

					Expect(testUI.Out).To(Say("Applications in this space will be placed in the platform default isolation segment."))
					Expect(testUI.Out).To(Say("Running applications need a restart to be moved there."))

					Expect(fakeActor.ResetSpaceIsolationSegmentCallCount()).To(Equal(1))
					orgGUID, spaceGUID := fakeActor.ResetSpaceIsolationSegmentArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when the reset changes the isolation segment to the org's default", func() {
				BeforeEach(func() {
					fakeActor.ResetSpaceIsolationSegmentReturns("some-org-iso-seg-name", v3action.Warnings{"warning-3", "warning-4"}, nil)
				})

				It("Displays the header and okay", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Resetting isolation segment assignment of space %s in org %s as banana...", space, org))

					Expect(testUI.Out).To(Say("OK\n\n"))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("warning-3"))
					Expect(testUI.Err).To(Say("warning-4"))

					Expect(testUI.Out).To(Say("Applications in this space will be placed in isolation segment some-org-iso-seg-name."))
					Expect(testUI.Out).To(Say("Running applications need a restart to be moved there."))

					Expect(fakeActor.ResetSpaceIsolationSegmentCallCount()).To(Equal(1))
					orgGUID, spaceGUID := fakeActor.ResetSpaceIsolationSegmentArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when the reset errors", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("some error")
					fakeActor.ResetSpaceIsolationSegmentReturns("some-org-iso-seg", v3action.Warnings{"warning-3", "warning-4"}, expectedErr)
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
