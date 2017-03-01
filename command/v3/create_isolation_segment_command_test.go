package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-isolation-segment Command", func() {
	var (
		cmd             v3.CreateIsolationSegmentCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeCreateIsolationSegmentActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeCreateIsolationSegmentActor)

		cmd = v3.CreateIsolationSegmentCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		// fakeActor.CloudControllerAPIVersionReturns("3.0.0")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	// Context("when the API version is below the minimum", func() {
	// 	BeforeEach(func() {
	// 		fakeActor.CloudControllerAPIVersionReturns("0.0.0")
	// 	})

	// 	It("returns a MinimumAPIVersionNotMetError", func() {
	// 		Expect(executeErr).To(MatchError(command.MinimumAPIVersionNotMetError{
	// 			CurrentVersion: "0.0.0",
	// 			MinimumVersion: "3.0.0",
	// 		}))
	// 	})
	// })

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(command.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		Context("when the tag placement_tags exists", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
				fakeActor.CreateIsolationSegmentReturns(v3action.Warnings{"I am a warning", "I am also a warning"}, nil)

				cmd.RequiredArgs.IsolationSegmentName = "segment1"
			})

			It("displays the header and ok", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Creating isolation segment segment1 as banana..."))
				Expect(testUI.Out).To(Say("OK"))

				Expect(testUI.Err).To(Say("I am a warning"))
				Expect(testUI.Err).To(Say("I am also a warning"))

				Expect(fakeActor.CreateIsolationSegmentCallCount()).To(Equal(1))
				Expect(fakeActor.CreateIsolationSegmentArgsForCall(0)).To(Equal("segment1"))
			})
		})

		Context("when the tag placement_tags not exist", func() {
			var expectedErr error

			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
				expectedErr = errors.New("I am an error")
				fakeActor.CreateIsolationSegmentReturns(v3action.Warnings{"I am a warning", "I am also a warning"}, expectedErr)

				cmd.RequiredArgs.IsolationSegmentName = "segment1"
			})

			It("displays the header and error", func() {
				Expect(executeErr).To(MatchError(expectedErr))

				Expect(testUI.Out).To(Say("Creating isolation segment segment1 as banana..."))

				Expect(testUI.Err).To(Say("I am a warning"))
				Expect(testUI.Err).To(Say("I am also a warning"))
			})
		})
	})

})
