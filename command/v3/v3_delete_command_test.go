package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-delete Command", func() {
	var (
		cmd             v3.V3DeleteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3DeleteActor
		input           *Buffer
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3DeleteActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = v3.V3DeleteCommand{
			RequiredArgs: flag.AppName{AppName: app},

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})

		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionV3)
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
				MinimumVersion: ccversion.MinVersionV3,
			}))
		})

		It("displays the experimental warning", func() {
			Expect(testUI.Out).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		})
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
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
		})
	})

	Context("when the -f flag is NOT provided", func() {
		BeforeEach(func() {
			cmd.Force = false
		})

		Context("when the user inputs yes", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("y\n"))
				Expect(err).ToNot(HaveOccurred())

				fakeActor.DeleteApplicationByNameAndSpaceReturns(v3action.Warnings{"some-warning"}, nil)
			})

			It("deletes the app", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say("Deleting app some-app in org some-org / space some-space as steve\\.\\.\\."))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).NotTo(Say("App some-app does not exist"))
			})
		})

		Context("when the user inputs no", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("cancels the delete", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Delete cancelled"))
				Expect(fakeActor.DeleteApplicationByNameAndSpaceCallCount()).To(Equal(0))
			})
		})

		Context("when the user chooses the default", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("cancels the delete", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Delete cancelled"))
				Expect(fakeActor.DeleteApplicationByNameAndSpaceCallCount()).To(Equal(0))
			})
		})

		Context("when the user input is invalid", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("e\n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("asks the user again", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say("Really delete the app some-app\\? \\[yN\\]"))
				Expect(testUI.Out).To(Say("invalid input \\(not y, n, yes, or no\\)"))
				Expect(testUI.Out).To(Say("Really delete the app some-app\\? \\[yN\\]"))

				Expect(fakeActor.DeleteApplicationByNameAndSpaceCallCount()).To(Equal(0))
			})
		})
	})

	Context("when the -f flag is provided", func() {
		BeforeEach(func() {
			cmd.Force = true
		})

		Context("when deleting the app errors", func() {
			Context("generic error", func() {
				BeforeEach(func() {
					fakeActor.DeleteApplicationByNameAndSpaceReturns(v3action.Warnings{"some-warning"}, errors.New("some-error"))
				})

				It("displays all warnings, and returns the erorr", func() {
					Expect(testUI.Err).To(Say("some-warning"))
					Expect(testUI.Out).To(Say("Deleting app some-app in org some-org / space some-space as steve\\.\\.\\."))
					Expect(testUI.Out).ToNot(Say("OK"))
					Expect(executeErr).To(MatchError("some-error"))
				})
			})
		})

		Context("when the app doesn't exist", func() {
			BeforeEach(func() {
				fakeActor.DeleteApplicationByNameAndSpaceReturns(v3action.Warnings{"some-warning"}, actionerror.ApplicationNotFoundError{Name: "some-app"})
			})

			It("displays all warnings, that the app wasn't found, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say("Deleting app some-app in org some-org / space some-space as steve\\.\\.\\."))
				Expect(testUI.Out).To(Say("App some-app does not exist"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				fakeActor.DeleteApplicationByNameAndSpaceReturns(v3action.Warnings{"some-warning"}, nil)
			})

			It("displays all warnings, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say("Deleting app some-app in org some-org / space some-space as steve\\.\\.\\."))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).NotTo(Say("App some-app does not exist"))
			})
		})
	})
})
