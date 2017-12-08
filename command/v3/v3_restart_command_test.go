package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
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

var _ = Describe("v3-restart Command", func() {
	var (
		cmd             v3.V3RestartCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3RestartActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3RestartActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = v3.V3RestartCommand{
			RequiredArgs: flag.AppName{AppName: app},

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

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

	Context("when the user is logged in", func() {
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

		Context("when stop app does not return an error", func() {
			BeforeEach(func() {
				fakeActor.StopApplicationReturns(v3action.Warnings{"stop-warning-1", "stop-warning-2"}, nil)
			})

			Context("when start app does not return an error", func() {
				BeforeEach(func() {
					fakeActor.StartApplicationReturns(v3action.Application{}, v3action.Warnings{"start-warning-1", "start-warning-2"}, nil)
				})

				Context("when get app does not return an error", func() {
					Context("if the app was already started", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid", State: constant.ApplicationStarted}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
						})

						It("says that the app was stopped, then started, and outputs warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Err).To(Say("get-warning-1"))
							Expect(testUI.Err).To(Say("get-warning-2"))

							Expect(testUI.Out).To(Say("Stopping app some-app in org some-org / space some-space as steve\\.\\.\\."))
							Expect(testUI.Err).To(Say("stop-warning-1"))
							Expect(testUI.Err).To(Say("stop-warning-2"))
							Expect(testUI.Out).To(Say("OK"))

							Expect(testUI.Out).To(Say("Starting app some-app in org some-org / space some-space as steve\\.\\.\\."))
							Expect(testUI.Err).To(Say("start-warning-1"))
							Expect(testUI.Err).To(Say("start-warning-2"))
							Expect(testUI.Out).To(Say("OK"))

							Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
							appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
							Expect(appName).To(Equal("some-app"))
							Expect(spaceGUID).To(Equal("some-space-guid"))

							Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
							appGUID := fakeActor.StopApplicationArgsForCall(0)
							Expect(appGUID).To(Equal("some-app-guid"))

							Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
							appGUID = fakeActor.StartApplicationArgsForCall(0)
							Expect(appGUID).To(Equal("some-app-guid"))
						})
					})

					Context("if the app was not already started", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid", State: constant.ApplicationStopped}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
						})

						It("says that the app was stopped, then started, and outputs warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Err).To(Say("get-warning-1"))
							Expect(testUI.Err).To(Say("get-warning-2"))

							Expect(testUI.Out).ToNot(Say("Stopping"))
							Expect(testUI.Err).ToNot(Say("stop-warning"))

							Expect(testUI.Out).To(Say("Starting app some-app in org some-org / space some-space as steve\\.\\.\\."))
							Expect(testUI.Err).To(Say("start-warning-1"))
							Expect(testUI.Err).To(Say("start-warning-2"))
							Expect(testUI.Out).To(Say("OK"))

							Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
							appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
							Expect(appName).To(Equal("some-app"))
							Expect(spaceGUID).To(Equal("some-space-guid"))

							Expect(fakeActor.StopApplicationCallCount()).To(BeZero(), "Expected StopApplication to not be called")

							Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
							appGUID := fakeActor.StartApplicationArgsForCall(0)
							Expect(appGUID).To(Equal("some-app-guid"))
						})
					})
				})

				Context("when the get app call returns an error", func() {
					Context("which is an ApplicationNotFoundError", func() {
						BeforeEach(func() {
							fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"get-warning-1", "get-warning-2"}, actionerror.ApplicationNotFoundError{Name: app})
						})

						It("says that the app wasn't found", func() {
							Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))
							Expect(testUI.Out).ToNot(Say("Stopping"))
							Expect(testUI.Out).ToNot(Say("Starting"))

							Expect(testUI.Err).To(Say("get-warning-1"))
							Expect(testUI.Err).To(Say("get-warning-2"))

							Expect(fakeActor.StopApplicationCallCount()).To(BeZero(), "Expected StopApplication to not be called")
							Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
						})

						Context("when it is an unknown error", func() {
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

								Expect(fakeActor.StopApplicationCallCount()).To(BeZero(), "Expected StopApplication to not be called")
								Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
							})
						})
					})
				})
			})

			Context("when the start app call returns an error", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid", State: constant.ApplicationStarted}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
				})

				Context("and the error is some random error", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("some start error")
						fakeActor.StartApplicationReturns(v3action.Application{}, v3action.Warnings{"start-warning-1", "start-warning-2"}, expectedErr)
					})

					It("says that the app failed to start", func() {
						Expect(executeErr).To(Equal(expectedErr))
						Expect(testUI.Out).To(Say("Starting app some-app in org some-org / space some-space as steve\\.\\.\\."))

						Expect(testUI.Err).To(Say("get-warning-1"))
						Expect(testUI.Err).To(Say("get-warning-2"))
						Expect(testUI.Err).To(Say("start-warning-1"))
						Expect(testUI.Err).To(Say("start-warning-2"))
					})
				})

				Context("when the start app call returns an ApplicationNotFoundError (someone else deleted app after we fetched app)", func() {
					BeforeEach(func() {
						fakeActor.StartApplicationReturns(v3action.Application{}, v3action.Warnings{"start-warning-1", "start-warning-2"}, actionerror.ApplicationNotFoundError{Name: app})
					})

					It("says that the app failed to start", func() {
						Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))
						Expect(testUI.Out).To(Say("Starting app some-app in org some-org / space some-space as steve\\.\\.\\."))

						Expect(testUI.Err).To(Say("get-warning-1"))
						Expect(testUI.Err).To(Say("get-warning-2"))
						Expect(testUI.Err).To(Say("start-warning-1"))
						Expect(testUI.Err).To(Say("start-warning-2"))
					})
				})
			})
		})

		Context("when the stop app call returns an error", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "some-app-guid", State: constant.ApplicationStarted}, v3action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			})

			Context("and the error is some random error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some stop error")
					fakeActor.StopApplicationReturns(v3action.Warnings{"stop-warning-1", "stop-warning-2"}, expectedErr)
				})

				It("says that the app failed to start", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(testUI.Out).To(Say("Stopping app some-app in org some-org / space some-space as steve\\.\\.\\."))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))
					Expect(testUI.Err).To(Say("stop-warning-1"))
					Expect(testUI.Err).To(Say("stop-warning-2"))

					Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
				})
			})

			Context("when the stop app call returns a ApplicationNotFoundError (someone else deleted app after we fetched summary)", func() {
				BeforeEach(func() {
					fakeActor.StopApplicationReturns(v3action.Warnings{"stop-warning-1", "stop-warning-2"}, actionerror.ApplicationNotFoundError{Name: app})
				})

				It("says that the app failed to start", func() {
					Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))
					Expect(testUI.Out).To(Say("Stopping app some-app in org some-org / space some-space as steve\\.\\.\\."))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))
					Expect(testUI.Err).To(Say("stop-warning-1"))
					Expect(testUI.Err).To(Say("stop-warning-2"))

					Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
				})
			})
		})
	})
})
