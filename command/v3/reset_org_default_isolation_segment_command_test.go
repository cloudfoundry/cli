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

var _ = Describe("reset-org-default-isolation-segment Command", func() {
	var (
		cmd             v3.ResetOrgDefaultIsolationSegmentCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeResetOrgDefaultIsolationSegmentActor
		fakeActorV2     *v3fakes.FakeResetOrgDefaultIsolationSegmentActorV2
		binaryName      string
		executeErr      error
		orgName         string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeResetOrgDefaultIsolationSegmentActor)
		fakeActorV2 = new(v3fakes.FakeResetOrgDefaultIsolationSegmentActorV2)

		cmd = v3.ResetOrgDefaultIsolationSegmentCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			ActorV2:     fakeActorV2,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		orgName = "some-org"

		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionIsolationSegmentV3)
		cmd.RequiredArgs.OrgName = orgName
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

	Context("when checking file succeeds", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: orgName,
				GUID: "some-org-guid",
			})
		})

		Context("when the user is not logged in", func() {
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

		Context("when the user is logged in", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			})

			Context("when the org lookup is unsuccessful", func() {
				BeforeEach(func() {
					fakeActorV2.GetOrganizationByNameReturns(v2action.Organization{}, v2action.Warnings{"warning-1", "warning-2"}, actionerror.OrganizationNotFoundError{Name: orgName})
				})

				It("returns the warnings and error", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: orgName}))
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			Context("when the org lookup is successful", func() {
				BeforeEach(func() {
					fakeActorV2.GetOrganizationByNameReturns(v2action.Organization{
						Name: orgName,
						GUID: "some-org-guid",
					}, v2action.Warnings{"warning-1", "warning-2"}, nil)
				})

				Context("when the reset succeeds", func() {
					BeforeEach(func() {
						fakeActor.ResetOrganizationDefaultIsolationSegmentReturns(v3action.Warnings{"warning-3", "warning-4"}, nil)
					})

					It("displays the header and okay", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Resetting default isolation segment of org %s as banana...", orgName))

						Expect(testUI.Out).To(Say("OK\n\n"))

						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
						Expect(testUI.Err).To(Say("warning-3"))
						Expect(testUI.Err).To(Say("warning-4"))

						Expect(testUI.Out).To(Say("Applications in spaces of this org that have no isolation segment assigned will be placed in the platform default isolation segment."))
						Expect(testUI.Out).To(Say("Running applications need a restart to be moved there."))

						Expect(fakeActor.ResetOrganizationDefaultIsolationSegmentCallCount()).To(Equal(1))
						orgGUID := fakeActor.ResetOrganizationDefaultIsolationSegmentArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))
					})
				})

				Context("when the reset errors", func() {
					var expectedErr error
					BeforeEach(func() {
						expectedErr = errors.New("some error")
						fakeActor.ResetOrganizationDefaultIsolationSegmentReturns(v3action.Warnings{"warning-3", "warning-4"}, expectedErr)
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
