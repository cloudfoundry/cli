package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
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

var _ = Describe("create-isolation-segment Command", func() {
	var (
		cmd              v3.CreateIsolationSegmentCommand
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v3fakes.FakeCreateIsolationSegmentActor
		binaryName       string
		executeErr       error
		isolationSegment string
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
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)

			cmd.RequiredArgs.IsolationSegmentName = isolationSegment
		})

		Context("when the create is successful", func() {
			BeforeEach(func() {
				fakeActor.CreateIsolationSegmentByNameReturns(v3action.Warnings{"I am a warning", "I am also a warning"}, nil)
			})

			It("displays the header and ok", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Creating isolation segment segment1 as banana..."))
				Expect(testUI.Out).To(Say("OK"))

				Expect(testUI.Err).To(Say("I am a warning"))
				Expect(testUI.Err).To(Say("I am also a warning"))

				Expect(fakeActor.CreateIsolationSegmentByNameCallCount()).To(Equal(1))
				Expect(fakeActor.CreateIsolationSegmentByNameArgsForCall(0)).To(Equal(v3action.IsolationSegment{Name: isolationSegment}))
			})
		})

		Context("when the create is unsuccessful", func() {
			Context("due to an unexpected error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am an error")
					fakeActor.CreateIsolationSegmentByNameReturns(v3action.Warnings{"I am a warning", "I am also a warning"}, expectedErr)
				})

				It("displays the header and error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Out).To(Say("Creating isolation segment segment1 as banana..."))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
				})
			})

			Context("due to an IsolationSegmentAlreadyExistsError", func() {
				BeforeEach(func() {
					fakeActor.CreateIsolationSegmentByNameReturns(v3action.Warnings{"I am a warning", "I am also a warning"}, actionerror.IsolationSegmentAlreadyExistsError{})
				})

				It("displays the header and ok", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Creating isolation segment segment1 as banana..."))
					Expect(testUI.Out).To(Say("OK"))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
					Expect(testUI.Err).To(Say("Isolation segment %s already exists.", isolationSegment))
				})
			})
		})
	})
})
