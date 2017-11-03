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

var _ = Describe("delete-isolation-segment Command", func() {
	var (
		cmd              v3.DeleteIsolationSegmentCommand
		input            *Buffer
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v3fakes.FakeDeleteIsolationSegmentActor
		binaryName       string
		executeErr       error
		isolationSegment string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeDeleteIsolationSegmentActor)

		cmd = v3.DeleteIsolationSegmentCommand{
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

		Context("when the -f flag is provided", func() {
			BeforeEach(func() {
				cmd.Force = true
			})

			Context("when the iso segment exists", func() {
				Context("when the delete is successful", func() {
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

				Context("when the delete is unsuccessful", func() {
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

			Context("when the iso segment does not exist", func() {
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

		Context("when the -f flag is not provided", func() {
			Context("when the user chooses the default", func() {
				BeforeEach(func() {
					input.Write([]byte("\n"))
				})

				It("cancels the deletion", func() {
					Expect(testUI.Out).To(Say("Really delete the isolation segment %s?", isolationSegment))
					Expect(testUI.Out).To(Say("Delete cancelled"))
					Expect(fakeActor.DeleteIsolationSegmentByNameCallCount()).To(Equal(0))
				})
			})

			Context("when the user inputs yes", func() {
				BeforeEach(func() {
					input.Write([]byte("yes\n"))
				})

				It("deletes the isolation segment", func() {
					Expect(testUI.Out).To(Say("Really delete the isolation segment %s?", isolationSegment))
					Expect(testUI.Out).To(Say("OK"))
					Expect(fakeActor.DeleteIsolationSegmentByNameCallCount()).To(Equal(1))
				})
			})

			Context("when the user inputs no", func() {
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
