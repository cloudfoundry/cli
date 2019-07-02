package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("restart Command", func() {
	var (
		cmd             v7.RestartCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeRestartActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeRestartActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = v7.RestartCommand{
			RequiredArgs: flag.AppName{AppName: app},

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("displays the experimental warning", func() {
		Expect(testUI.Err).NotTo(Say("This command is in EXPERIMENTAL stage and may change without notice"))
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

		When("stop app does not return an error", func() {
			BeforeEach(func() {
				fakeActor.StopApplicationReturns(v7action.Warnings{"stop-warning-1", "stop-warning-2"}, nil)
			})

			When("start app does not return an error", func() {
				BeforeEach(func() {
					fakeActor.StartApplicationReturns(v7action.Warnings{"start-warning-1", "start-warning-2"}, nil)
				})

				When("polling the app does not return an error", func() {
					BeforeEach(func() {
						fakeActor.PollStartReturns(v7action.Warnings{"poll-warning-1", "poll-warning-2"}, nil)
					})

					When("get app does not return an error", func() {
						Context("if the app was already started", func() {
							BeforeEach(func() {
								fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{GUID: "some-app-guid", State: constant.ApplicationStarted}, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
							})

							It("says that the app was stopped, then started, and outputs warnings", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Err).To(Say("get-warning-1"))
								Expect(testUI.Err).To(Say("get-warning-2"))

								Expect(testUI.Out).To(Say(`Restarting app some-app in org some-org / space some-space as steve\.\.\.`))
								Expect(testUI.Err).To(Say("stop-warning-1"))
								Expect(testUI.Err).To(Say("stop-warning-2"))
								Expect(testUI.Out).To(Say(`Stopping app\.\.\.`))

								Expect(testUI.Err).To(Say("start-warning-1"))
								Expect(testUI.Err).To(Say("start-warning-2"))

								Expect(testUI.Out).To(Say(`Waiting for app to start\.\.\.`))

								Expect(testUI.Err).To(Say("poll-warning-1"))
								Expect(testUI.Err).To(Say("poll-warning-2"))

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

								Expect(fakeActor.PollStartCallCount()).To(Equal(1))
								appGUID, noWait := fakeActor.PollStartArgsForCall(0)
								Expect(appGUID).To(Equal("some-app-guid"))
								Expect(noWait).To(Equal(false))
							})
						})

						Context("if the app was not already started", func() {
							BeforeEach(func() {
								fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{GUID: "some-app-guid", State: constant.ApplicationStopped}, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
							})

							It("says that the app was stopped, then started, and outputs warnings", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Err).To(Say("get-warning-1"))
								Expect(testUI.Err).To(Say("get-warning-2"))

								Expect(testUI.Out).To(Say(`Restarting app some-app in org some-org / space some-space as steve\.\.\.`))
								Expect(testUI.Out).ToNot(Say("Stopping"))
								Expect(testUI.Err).ToNot(Say("stop-warning"))

								Expect(testUI.Err).To(Say("start-warning-1"))
								Expect(testUI.Err).To(Say("start-warning-2"))

								Expect(testUI.Out).To(Say(`Waiting for app to start\.\.\.`))
								Expect(testUI.Err).To(Say("poll-warning-1"))
								Expect(testUI.Err).To(Say("poll-warning-2"))

								Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
								appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
								Expect(appName).To(Equal("some-app"))
								Expect(spaceGUID).To(Equal("some-space-guid"))

								Expect(fakeActor.StopApplicationCallCount()).To(BeZero(), "Expected StopApplication to not be called")

								Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
								appGUID := fakeActor.StartApplicationArgsForCall(0)
								Expect(appGUID).To(Equal("some-app-guid"))

								Expect(fakeActor.PollStartCallCount()).To(Equal(1))
								appGUID, noWait := fakeActor.PollStartArgsForCall(0)
								Expect(appGUID).To(Equal("some-app-guid"))
								Expect(noWait).To(Equal(false))
							})
						})
					})

					When("the get app call returns an error", func() {
						Context("which is an ApplicationNotFoundError", func() {
							BeforeEach(func() {
								fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{}, v7action.Warnings{"get-warning-1", "get-warning-2"}, actionerror.ApplicationNotFoundError{Name: app})
							})

							It("says that the app wasn't found", func() {
								Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))
								Expect(testUI.Out).ToNot(Say("Stopping"))
								Expect(testUI.Out).ToNot(Say("Waiting for app to start"))

								Expect(testUI.Err).To(Say("get-warning-1"))
								Expect(testUI.Err).To(Say("get-warning-2"))

								Expect(fakeActor.StopApplicationCallCount()).To(BeZero(), "Expected StopApplication to not be called")
								Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
							})

							When("it is an unknown error", func() {
								var expectedErr error

								BeforeEach(func() {
									expectedErr = errors.New("some get app error")
									fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{State: constant.ApplicationStopped}, v7action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
								})

								It("says that the app failed to start", func() {
									Expect(executeErr).To(Equal(expectedErr))
									Expect(testUI.Out).ToNot(Say("Stopping"))
									Expect(testUI.Out).ToNot(Say("Waiting for app to start"))

									Expect(testUI.Err).To(Say("get-warning-1"))
									Expect(testUI.Err).To(Say("get-warning-2"))

									Expect(fakeActor.StopApplicationCallCount()).To(BeZero(), "Expected StopApplication to not be called")
									Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
								})
							})
						})
					})
				})
			})

			When("the start app call returns an error", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{GUID: "some-app-guid", State: constant.ApplicationStarted}, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
				})

				Context("and the error is some random error", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("some start error")
						fakeActor.StartApplicationReturns(v7action.Warnings{"start-warning-1", "start-warning-2"}, expectedErr)
					})

					It("says that the app failed to start", func() {
						Expect(executeErr).To(Equal(expectedErr))
						Expect(testUI.Out).To(Say(`Restarting app some-app in org some-org / space some-space as steve\.\.\.`))
						Expect(testUI.Out).NotTo(Say(`Waiting for app to start\.\.\.`))

						Expect(testUI.Err).To(Say("get-warning-1"))
						Expect(testUI.Err).To(Say("get-warning-2"))
						Expect(testUI.Err).To(Say("start-warning-1"))
						Expect(testUI.Err).To(Say("start-warning-2"))

						Expect(fakeActor.PollStartCallCount()).To(BeZero(), "Expected PollStart to not be called")
					})
				})

				When("the start app call returns an ApplicationNotFoundError (someone else deleted app after we fetched app)", func() {
					BeforeEach(func() {
						fakeActor.StartApplicationReturns(v7action.Warnings{"start-warning-1", "start-warning-2"}, actionerror.ApplicationNotFoundError{Name: app})
					})

					It("says that the app failed to start", func() {
						Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))
						Expect(testUI.Out).To(Say(`Restarting app some-app in org some-org / space some-space as steve\.\.\.`))
						Expect(testUI.Out).NotTo(Say(`Waiting for app to start\.\.\.`))

						Expect(testUI.Err).To(Say("get-warning-1"))
						Expect(testUI.Err).To(Say("get-warning-2"))
						Expect(testUI.Err).To(Say("start-warning-1"))
						Expect(testUI.Err).To(Say("start-warning-2"))

						Expect(fakeActor.PollStartCallCount()).To(BeZero(), "Expected PollStart to not be called")
					})
				})
			})
		})

		When("the stop app call returns an error", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{GUID: "some-app-guid", State: constant.ApplicationStarted}, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			})

			Context("and the error is some random error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some stop error")
					fakeActor.StopApplicationReturns(v7action.Warnings{"stop-warning-1", "stop-warning-2"}, expectedErr)
				})

				It("says that the app failed to start", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(testUI.Out).To(Say(`Restarting app some-app in org some-org / space some-space as steve\.\.\.`))
					Expect(testUI.Out).To(Say(`Stopping app\.\.\.`))
					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))
					Expect(testUI.Err).To(Say("stop-warning-1"))
					Expect(testUI.Err).To(Say("stop-warning-2"))

					Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
				})
			})

			When("the stop app call returns a ApplicationNotFoundError (someone else deleted app after we fetched summary)", func() {
				BeforeEach(func() {
					fakeActor.StopApplicationReturns(v7action.Warnings{"stop-warning-1", "stop-warning-2"}, actionerror.ApplicationNotFoundError{Name: app})
				})

				It("says that the app failed to start", func() {
					Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))
					Expect(testUI.Out).To(Say(`Restarting app some-app in org some-org / space some-space as steve\.\.\.`))

					Expect(testUI.Err).To(Say("get-warning-1"))
					Expect(testUI.Err).To(Say("get-warning-2"))
					Expect(testUI.Err).To(Say("stop-warning-1"))
					Expect(testUI.Err).To(Say("stop-warning-2"))

					Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
				})
			})
		})
	})

	When("getting the application summary returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = actionerror.ApplicationNotFoundError{Name: app}
			fakeActor.GetApplicationSummaryByNameAndSpaceReturns(v7action.ApplicationSummary{}, v7action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))

			Expect(testUI.Out).To(Say(`Waiting for app to start\.\.\.`))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	When("getting the application summary is successful", func() {
		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
				GUID: "some-space-guid",
			})
			summary := v7action.ApplicationSummary{
				Application: v7action.Application{
					Name:  "some-app",
					State: constant.ApplicationStarted,
				},
				CurrentDroplet: v7action.Droplet{
					Stack: "cflinuxfs2",
					Buildpacks: []v7action.DropletBuildpack{
						{
							Name:         "ruby_buildpack",
							DetectOutput: "some-detect-output",
						},
						{
							Name:         "some-buildpack",
							DetectOutput: "",
						},
					},
				},
				ProcessSummaries: v7action.ProcessSummaries{
					{
						Process: v7action.Process{
							Type:    constant.ProcessTypeWeb,
							Command: *types.NewFilteredString("some-command-1"),
						},
					},
					{
						Process: v7action.Process{
							Type:    "console",
							Command: *types.NewFilteredString("some-command-2"),
						},
					},
				},
			}
			fakeActor.GetApplicationSummaryByNameAndSpaceReturns(summary, v7action.Warnings{"warning-1", "warning-2"}, nil)
		})

		It("prints the application summary and outputs warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Waiting for app to start\.\.\.`))
			Expect(testUI.Out).To(Say(`name:\s+some-app`))
			Expect(testUI.Out).To(Say(`requested state:\s+started`))
			Expect(testUI.Out).ToNot(Say("start command:"))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))

			Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID, withObfuscatedValues, _ := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
			Expect(withObfuscatedValues).To(BeFalse())
		})
	})
})
