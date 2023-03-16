package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete Command", func() {
	var (
		cmd             DeleteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
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
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = DeleteCommand{
			RequiredArgs: flag.AppName{AppName: app},

			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})

		fakeActor.GetCurrentUserReturns(configv3.User{Name: "steve"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the -r flag is provided", func() {
		BeforeEach(func() {
			cmd.DeleteMappedRoutes = true
			cmd.Force = false
			_, err := input.Write([]byte("y\n"))
			Expect(err).ToNot(HaveOccurred())

			fakeActor.DeleteApplicationByNameAndSpaceReturns(v7action.Warnings{"some-warning"}, nil)
		})

		It("asks for a special prompt about deleting associated routes", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Out).To(Say(`Deleting the app and associated routes will make apps with this route, in any org, unreachable\.`))
			Expect(testUI.Out).To(Say(`Deleting app some-app in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).To(Say("OK"))
			Expect(testUI.Out).NotTo(Say(`App 'some-app' does not exist\.`))
		})

		It("deletes application and mapped routes", func() {
			actualName, actualSpace, deleteMappedRoutes := fakeActor.DeleteApplicationByNameAndSpaceArgsForCall(0)
			Expect(actualName).To(Equal(app))
			Expect(actualSpace).To(Equal(fakeConfig.TargetedSpace().GUID))
			Expect(deleteMappedRoutes).To(BeTrue())
		})

		When("the route is mapped to a different app", func() {
			BeforeEach(func() {
				fakeActor.DeleteApplicationByNameAndSpaceReturns(v7action.Warnings{"some-warning"}, actionerror.RouteBoundToMultipleAppsError{})
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(actionerror.RouteBoundToMultipleAppsError{}))
			})

			It("defers showing a tip", func() {
				Expect(testUI.Out).NotTo(Say("TIP"))
				testUI.FlushDeferred()
				Expect(testUI.Out).To(Say(`\n\nTIP: Run 'cf delete some-app' to delete the app and 'cf delete-route' to delete the route\.`))
			})
		})
	})

	When("checking target fails", func() {
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

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeActor.GetCurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("the -f flag is NOT provided", func() {
		BeforeEach(func() {
			cmd.Force = false
		})

		When("the user inputs yes", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("y\n"))
				Expect(err).ToNot(HaveOccurred())

				fakeActor.DeleteApplicationByNameAndSpaceReturns(v7action.Warnings{"some-warning"}, nil)
			})

			It("delegates to the Actor", func() {
				actualName, actualSpace, deleteMappedRoutes := fakeActor.DeleteApplicationByNameAndSpaceArgsForCall(0)
				Expect(actualName).To(Equal(app))
				Expect(actualSpace).To(Equal(fakeConfig.TargetedSpace().GUID))
				Expect(deleteMappedRoutes).To(BeFalse())
			})

			It("deletes the app", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting app some-app in org some-org / space some-space as steve\.\.\.`))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).NotTo(Say(`App 'some-app' does not exist\.`))
			})
		})

		When("the user inputs no", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("cancels the delete", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say(`App 'some-app' has not been deleted\.`))
				Expect(fakeActor.DeleteApplicationByNameAndSpaceCallCount()).To(Equal(0))
			})
		})

		When("the user chooses the default", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("cancels the delete", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say(`App 'some-app' has not been deleted\.`))
				Expect(fakeActor.DeleteApplicationByNameAndSpaceCallCount()).To(Equal(0))
			})
		})

		When("the user input is invalid", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("e\n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("asks the user again", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say(`Really delete the app some-app\? \[yN\]`))
				Expect(testUI.Out).To(Say(`invalid input \(not y, n, yes, or no\)`))
				Expect(testUI.Out).To(Say(`Really delete the app some-app\? \[yN\]`))

				Expect(fakeActor.DeleteApplicationByNameAndSpaceCallCount()).To(Equal(0))
			})
		})
	})

	When("the -f flag is provided", func() {
		BeforeEach(func() {
			cmd.Force = true
		})

		When("deleting the app errors", func() {
			Context("generic error", func() {
				BeforeEach(func() {
					fakeActor.DeleteApplicationByNameAndSpaceReturns(v7action.Warnings{"some-warning"}, errors.New("some-error"))
				})

				It("displays all warnings, and returns the error", func() {
					Expect(testUI.Err).To(Say("some-warning"))
					Expect(testUI.Out).To(Say(`Deleting app some-app in org some-org / space some-space as steve\.\.\.`))
					Expect(testUI.Out).ToNot(Say("OK"))
					Expect(executeErr).To(MatchError("some-error"))
				})
			})
		})

		When("the app doesn't exist", func() {
			BeforeEach(func() {
				fakeActor.DeleteApplicationByNameAndSpaceReturns(v7action.Warnings{"some-warning"}, actionerror.ApplicationNotFoundError{Name: "some-app"})
			})

			It("displays all warnings, that the app wasn't found, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting app some-app in org some-org / space some-space as steve\.\.\.`))
				Expect(testUI.Err).To(Say(`App 'some-app' does not exist\.`))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				fakeActor.DeleteApplicationByNameAndSpaceReturns(v7action.Warnings{"some-warning"}, nil)
			})

			It("displays all warnings, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting app some-app in org some-org / space some-space as steve\.\.\.`))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Err).NotTo(Say(`App 'some-app' does not exist\.`))
			})
		})
	})
})
