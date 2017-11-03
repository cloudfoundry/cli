package v3_test

import (
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

var _ = Describe("set-space-isolation-segment Command", func() {
	var (
		cmd              v3.SetSpaceIsolationSegmentCommand
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v3fakes.FakeSetSpaceIsolationSegmentActor
		fakeActorV2      *v3fakes.FakeSetSpaceIsolationSegmentActorV2
		binaryName       string
		executeErr       error
		isolationSegment string
		space            string
		org              string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeSetSpaceIsolationSegmentActor)
		fakeActorV2 = new(v3fakes.FakeSetSpaceIsolationSegmentActorV2)

		cmd = v3.SetSpaceIsolationSegmentCommand{
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
		isolationSegment = "segment1"

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
			cmd.RequiredArgs.IsolationSegmentName = isolationSegment
		})

		Context("when the space lookup is unsuccessful", func() {
			BeforeEach(func() {
				fakeActorV2.GetSpaceByOrganizationAndNameReturns(v2action.Space{}, v2action.Warnings{"I am a warning", "I am also a warning"}, actionerror.SpaceNotFoundError{Name: space})
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: space}))
				Expect(testUI.Err).To(Say("I am a warning"))
				Expect(testUI.Err).To(Say("I am also a warning"))
			})
		})

		Context("when the space lookup is successful", func() {
			BeforeEach(func() {
				fakeActorV2.GetSpaceByOrganizationAndNameReturns(v2action.Space{
					Name: space,
					GUID: "some-space-guid",
				}, v2action.Warnings{"I am a warning", "I am also a warning"}, nil)
			})

			Context("when the entitlement is successful", func() {
				BeforeEach(func() {
					fakeActor.AssignIsolationSegmentToSpaceByNameAndSpaceReturns(v3action.Warnings{"entitlement-warning", "banana"}, nil)
				})

				It("Displays the header and okay", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Updating isolation segment of space %s in org %s as banana...", space, org))
					Expect(testUI.Out).To(Say("OK"))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
					Expect(testUI.Err).To(Say("entitlement-warning"))
					Expect(testUI.Err).To(Say("banana"))

					Expect(testUI.Out).To(Say("In order to move running applications to this isolation segment, they must be restarted."))

					Expect(fakeActor.AssignIsolationSegmentToSpaceByNameAndSpaceCallCount()).To(Equal(1))
					isolationSegmentName, spaceGUID := fakeActor.AssignIsolationSegmentToSpaceByNameAndSpaceArgsForCall(0)
					Expect(isolationSegmentName).To(Equal(isolationSegment))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when the entitlement errors", func() {
				BeforeEach(func() {
					fakeActor.AssignIsolationSegmentToSpaceByNameAndSpaceReturns(v3action.Warnings{"entitlement-warning", "banana"}, actionerror.IsolationSegmentNotFoundError{Name: "segment1"})
				})

				It("returns the warnings and error", func() {
					Expect(testUI.Out).To(Say("Updating isolation segment of space %s in org %s as banana...", space, org))
					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
					Expect(testUI.Err).To(Say("entitlement-warning"))
					Expect(testUI.Err).To(Say("banana"))
					Expect(executeErr).To(MatchError(actionerror.IsolationSegmentNotFoundError{Name: "segment1"}))
				})
			})
		})
	})
})
