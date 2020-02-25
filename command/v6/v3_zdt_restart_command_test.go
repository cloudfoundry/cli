package v6_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-zdt-restart Command", func() {
	var (
		cmd             V3ZeroDowntimeRestartCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeV3ZeroDowntimeRestartActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeV3ZeroDowntimeRestartActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = V3ZeroDowntimeRestartCommand{
			RequiredArgs: flag.AppName{AppName: app},

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}
		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinSupportedV3ClientVersion)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the API version is below the minimum", func() {
		olderCurrentVersion := "3.0.1"

		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns(olderCurrentVersion)
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
				CurrentVersion: olderCurrentVersion,
				MinimumVersion: ccversion.MinSupportedV3ClientVersion,
			}))
		})
	})

	It("displays the experimental warning", func() {
		Expect(testUI.Err).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
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
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
				GUID: "some-space-guid",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
		})

		When("the app exists", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v3action.Application{
						GUID:  "some-app-guid",
						State: constant.ApplicationStarted,
					}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
				fakeActor.CreateDeploymentReturns("", v3action.Warnings{"deploy-warning-1", "deploy-warning-2"}, nil)
			})

			It("makes a deployment for the app", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.CreateDeploymentCallCount()).To(Equal(1))

				Expect(testUI.Err).To(Say("deploy-warning-1"))
				Expect(testUI.Err).To(Say("deploy-warning-2"))
				Expect(testUI.Out).To(Say("Starting deployment for app some-app in org some-org / space some-space as steve..."))
				Expect(testUI.Out).To(Say("Waiting for app to start..."))
				Expect(testUI.Out).To(Say("OK"))

				Expect(fakeActor.ZeroDowntimePollStartCallCount()).To(Equal(1))
			})

			When("the app is stopped", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v3action.Application{
							GUID:  "some-app-guid",
							State: constant.ApplicationStopped,
						}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)

					fakeActor.StartApplicationReturns(v3action.Warnings{"start-warning-1", "start-warning-2"}, nil)
				})

				It("starts the app", func() {
					Expect(testUI.Out).To(Say(`Starting app some-app in org some-org / space some-space as steve\.\.\.`))
					Expect(testUI.Err).To(Say("start-warning-1"))
					Expect(testUI.Err).To(Say("start-warning-2"))
					Expect(testUI.Out).To(Say("OK"))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))

					Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
					appGUID := fakeActor.StartApplicationArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))
				})

				When("it fails to start the app", func() {
					BeforeEach(func() {
						fakeActor.StartApplicationReturns(nil, errors.New("lol error"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("lol error"))
					})
				})
			})

			When("the app fails to start", func() {
				BeforeEach(func() {
					fakeActor.ZeroDowntimePollStartReturns(errors.New("lol error"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("lol error"))
				})
			})
		})

		When("it fails to get the app", func() {
			When("the app isn't found", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"get-warning-1", "get-warning-2"}, actionerror.ApplicationNotFoundError{Name: app})
				})

				It("says that the app wasn't found", func() {
					Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))
					Expect(testUI.Out).ToNot(Say("Stopping"))
					Expect(testUI.Out).ToNot(Say("Starting"))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))

					Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
					Expect(fakeActor.CreateDeploymentCallCount()).To(BeZero(), "Expected CreateDeployment to not be called")
				})
			})

			When("it is an unknown error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some get app error")
					fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{State: constant.ApplicationStopped}, v3action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
				})

				It("says that the app failed to start", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(testUI.Out).ToNot(Say("Stopping"))
					Expect(testUI.Out).ToNot(Say("Starting"))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))

					Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
					Expect(fakeActor.CreateDeploymentCallCount()).To(BeZero(), "Expected CreateDeployment to not be called")
				})
			})
		})
	})
})
