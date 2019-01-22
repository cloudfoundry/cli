package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-isolation-segment Command", func() {
	var (
		cmd              DeleteIsolationSegmentCommand
		input            *Buffer
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v6fakes.FakeDeleteIsolationSegmentActor
		binaryName       string
		executeErr       error
		isolationSegment string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeDeleteIsolationSegmentActor)

		cmd = DeleteIsolationSegmentCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		isolationSegment = "segment1"
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
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			cmd.RequiredArgs.IsolationSegmentName = isolationSegment
		})

		When("the -f flag is provided", func() {
			BeforeEach(func() {
				cmd.Force = true
			})

			When("the iso segment exists", func() {
				When("the delete is successful", func() {
					BeforeEach(func() {
						fakeActor.DeleteIsolationSegmentByNameReturns(v3action.Warnings{"I am a warning", "I am also a warning"}, nil)
					})

					It("displays the header and ok", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Deleting isolation segment segment1 as banana..."))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Err).To(Say("I am a warning"))
						Expect(testUI.Err).To(Say("I am also a warning"))

						Expect(fakeActor.DeleteIsolationSegmentByNameCallCount()).To(Equal(1))
						Expect(fakeActor.DeleteIsolationSegmentByNameArgsForCall(0)).To(Equal("segment1"))
					})
				})

				When("the delete is unsuccessful", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("I am an error")
						fakeActor.DeleteIsolationSegmentByNameReturns(v3action.Warnings{"I am a warning", "I am also a warning"}, expectedErr)
					})

					It("displays the header and error", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Out).To(Say("Deleting isolation segment segment1 as banana..."))

						Expect(testUI.Err).To(Say("I am a warning"))
						Expect(testUI.Err).To(Say("I am also a warning"))
					})
				})
			})

			When("the iso segment does not exist", func() {
				BeforeEach(func() {
					fakeActor.DeleteIsolationSegmentByNameReturns(v3action.Warnings{"I am a warning", "I am also a warning"}, actionerror.IsolationSegmentNotFoundError{})
				})

				It("displays does not exist warning", func() {
					Expect(testUI.Out).To(Say("OK"))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("Isolation segment %s does not exist.", isolationSegment))
				})
			})
		})

		When("the -f flag is not provided", func() {
			When("the user chooses the default", func() {
				BeforeEach(func() {
					input.Write([]byte("\n"))
				})

				It("cancels the deletion", func() {
					Expect(testUI.Out).To(Say("Really delete the isolation segment %s?", isolationSegment))
					Expect(testUI.Out).To(Say("Delete cancelled"))
					Expect(fakeActor.DeleteIsolationSegmentByNameCallCount()).To(Equal(0))
				})
			})

			When("the user inputs yes", func() {
				BeforeEach(func() {
					input.Write([]byte("yes\n"))
				})

				It("deletes the isolation segment", func() {
					Expect(testUI.Out).To(Say("Really delete the isolation segment %s?", isolationSegment))
					Expect(testUI.Out).To(Say("OK"))
					Expect(fakeActor.DeleteIsolationSegmentByNameCallCount()).To(Equal(1))
				})
			})

			When("the user inputs no", func() {
				BeforeEach(func() {
					input.Write([]byte("no\n"))
				})

				It("cancels the deletion", func() {
					Expect(testUI.Out).To(Say("Really delete the isolation segment %s?", isolationSegment))
					Expect(testUI.Out).To(Say("Delete cancelled"))
					Expect(fakeActor.DeleteIsolationSegmentByNameCallCount()).To(Equal(0))
				})
			})
		})
	})
})
