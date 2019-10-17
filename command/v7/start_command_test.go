package v7_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"

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

var _ = Describe("start Command", func() {
	var (
		cmd             v7.StartCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeStartActor
		fakeNOAAClient  *v7actionfakes.FakeNOAAClient

		binaryName  string
		executeErr  error
		app         string
		packageGUID string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeStartActor)
		fakeNOAAClient = new(v7actionfakes.FakeNOAAClient)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"
		packageGUID = "some-package-guid"

		cmd = v7.StartCommand{
			RequiredArgs: flag.AppName{AppName: app},

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			NOAAClient:  fakeNOAAClient,
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

	When("the actor does not return an error", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
				GUID: "some-space-guid",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{GUID: "some-app-guid", State: constant.ApplicationStopped}, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			fakeActor.StartApplicationReturns(v7action.Warnings{"start-warning-1", "start-warning-2"}, nil)
			fakeActor.PollStartReturns(v7action.Warnings{"poll-warning-1", "poll-warning-2"}, nil)
		})

		It("says that the app was started and outputs warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Starting app some-app in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).To(Say(`Waiting for app to start\.\.\.`))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("start-warning-1"))
			Expect(testUI.Err).To(Say("start-warning-2"))
			Expect(testUI.Err).To(Say("poll-warning-1"))
			Expect(testUI.Err).To(Say("poll-warning-2"))

			Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
			appGUID := fakeActor.StartApplicationArgsForCall(0)
			Expect(appGUID).To(Equal("some-app-guid"))

			Expect(fakeActor.PollStartCallCount()).To(Equal(1))
			appGUID, noWait := fakeActor.PollStartArgsForCall(0)
			Expect(appGUID).To(Equal("some-app-guid"))
			Expect(noWait).To(Equal(false))
		})
	})

	When("the get app call returns a ApplicationNotFoundError", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			expectedErr = actionerror.ApplicationNotFoundError{Name: app}
			fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{State: constant.ApplicationStopped}, v7action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
		})

		It("says that the app failed to start", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))
			Expect(testUI.Out).ToNot(Say("Starting"))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
		})
	})

	When("the start app call returns a ApplicationNotFoundError (someone else deleted app after we fetched summary)", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{State: constant.ApplicationStopped}, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			expectedErr = actionerror.ApplicationNotFoundError{Name: app}
			fakeActor.StartApplicationReturns(v7action.Warnings{"start-warning-1", "start-warning-2"}, expectedErr)
		})

		It("says that the app failed to start", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))
			Expect(testUI.Out).To(Say(`Starting app some-app in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).NotTo(Say(`Waiting for app to start\.\.\.`))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("start-warning-1"))
			Expect(testUI.Err).To(Say("start-warning-2"))
		})
	})

	When("there is an error when polling the app", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{State: constant.ApplicationStopped}, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			expectedErr = actionerror.ApplicationNotStartedError{Name: app}
			fakeActor.StartApplicationReturns(v7action.Warnings{"start-warning-1", "start-warning-2"}, nil)
			fakeActor.PollStartReturns(v7action.Warnings{"poll-warning"}, expectedErr)
		})

		It("says that the app failed to start", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotStartedError{Name: app}))
			Expect(testUI.Out).To(Say(`Starting app some-app in org some-org / space some-space as steve\.\.\.`))
			Expect(testUI.Out).To(Say(`Waiting for app to start\.\.\.`))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("start-warning-1"))
			Expect(testUI.Err).To(Say("start-warning-2"))
			Expect(testUI.Err).To(Say("poll-warning"))
		})
	})

	When("the app is already started", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{State: constant.ApplicationStarted, Name: app}, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
		})

		It("says that the app failed to start", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(testUI.Out).ToNot(Say("Starting"))
			Expect(testUI.Out).To(Say(`App 'some-app' is already started\.`))
			Expect(testUI.Out).To(Say("OK"))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))

			Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
		})
	})

	When("the app needs staging", func() {
		BeforeEach(func() {
			app = "some-app"
			fakeActor.GetUnstagedNewestPackageGUIDReturns(packageGUID, v7action.Warnings{"needs-stage-warnings"}, nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
				GUID: "some-space-guid",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{State: constant.ApplicationStopped, LifecycleType: constant.AppLifecycleTypeBuildpack, GUID: "some-app-guid"}, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
		})

		When("the logging does not error", func() {
			var allLogsWritten chan bool

			BeforeEach(func() {
				allLogsWritten = make(chan bool)
				fakeActor.GetStreamingLogsForApplicationByNameAndSpaceStub = func(appName string, spaceGUID string, client v7action.NOAAClient) (<-chan *v7action.LogMessage, <-chan error, v7action.Warnings, error) {
					logStream := make(chan *v7action.LogMessage)
					errorStream := make(chan error)

					go func() {
						logStream <- v7action.NewLogMessage("Here are some staging logs!", 1, time.Now(), v7action.StagingLog, "sourceInstance")
						logStream <- v7action.NewLogMessage("Here are some other staging logs!", 1, time.Now(), v7action.StagingLog, "sourceInstance")
						allLogsWritten <- true
					}()

					return logStream, errorStream, v7action.Warnings{"steve for all I care"}, nil
				}
			})

			When("the staging is successful", func() {
				const dropletCreateTime = "2017-08-14T21:16:42Z"
				BeforeEach(func() {
					fakeActor.StagePackageStub = func(packageGUID, appName, spaceGUID string) (<-chan v7action.Droplet, <-chan v7action.Warnings, <-chan error) {
						dropletStream := make(chan v7action.Droplet)
						warningsStream := make(chan v7action.Warnings)
						errorStream := make(chan error)

						go func() {
							<-allLogsWritten
							defer close(dropletStream)
							defer close(warningsStream)
							defer close(errorStream)
							warningsStream <- v7action.Warnings{"some-warning", "some-other-warning"}
							dropletStream <- v7action.Droplet{
								GUID:      "some-droplet-guid",
								CreatedAt: dropletCreateTime,
								State:     constant.DropletStaged,
							}
						}()

						return dropletStream, warningsStream, errorStream
					}
					fakeActor.SetApplicationDropletReturns(v7action.Warnings{"some-set-droplet-warning"}, nil)
				})

				It("outputs the droplet GUID", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					createdAtTimeParsed, err := time.Parse(time.RFC3339, dropletCreateTime)
					Expect(err).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Staging package for %s in org some-org / space some-space as steve...", app))
					Expect(testUI.Out).To(Say("\n\n"))
					Expect(testUI.Out).To(Say("Package staged"))
					Expect(testUI.Out).To(Say(`droplet guid:\s+some-droplet-guid`))
					Expect(testUI.Out).To(Say(`state:\s+staged`))
					Expect(testUI.Out).To(Say(`created:\s+%s`, testUI.UserFriendlyDate(createdAtTimeParsed)))

					Expect(testUI.Err).To(Say("some-warning"))
					Expect(testUI.Err).To(Say("some-other-warning"))
				})

				It("stages the package", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
					guidArg, appNameArg, spaceGUIDArg := fakeActor.StagePackageArgsForCall(0)
					Expect(guidArg).To(Equal(packageGUID))
					Expect(appNameArg).To(Equal(app))
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))
				})

				It("displays staging logs and their warnings", func() {
					Expect(testUI.Out).To(Say("Here are some staging logs!"))
					Expect(testUI.Out).To(Say("Here are some other staging logs!"))

					Expect(testUI.Err).To(Say("steve for all I care"))

					Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appNameArg, spaceGUID, noaaClient := fakeActor.GetStreamingLogsForApplicationByNameAndSpaceArgsForCall(0)
					Expect(appNameArg).To(Equal(app))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(noaaClient).To(Equal(fakeNOAAClient))

					Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
					guidArg, appNameArg, spaceGUIDArg := fakeActor.StagePackageArgsForCall(0)
					Expect(guidArg).To(Equal(packageGUID))
					Expect(appNameArg).To(Equal(app))
					Expect(spaceGUIDArg).To(Equal(spaceGUID))
				})

				When("Assigning the droplet is successful", func() {
					BeforeEach(func() {
						fakeActor.SetApplicationDropletReturns(v7action.Warnings{"some-set-droplet-warning"}, nil)
					})
					It("displays that the droplet was assigned", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Setting app some-app to droplet some-droplet-guid in org some-org / space some-space as steve..."))

						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
						Expect(testUI.Out).To(Say("OK"))

						Expect(fakeActor.SetApplicationDropletCallCount()).To(Equal(1))
						appGuid, dropletGUID := fakeActor.SetApplicationDropletArgsForCall(0)
						Expect(appGuid).To(Equal("some-app-guid"))
						Expect(dropletGUID).To(Equal("some-droplet-guid"))

					})
				})

				When("Assigning the droplet is not successful", func() {
					var expectedErr error
					BeforeEach(func() {
						expectedErr = errors.New("some-error")
						fakeActor.SetApplicationDropletReturns(v7action.Warnings{"some-set-droplet-warning"}, expectedErr)
					})
					It("errors and displays warnings", func() {
						Expect(testUI.Out).To(Say("Setting app some-app to droplet some-droplet-guid in org some-org / space some-space as steve..."))
						Expect(executeErr).To(HaveOccurred())
						Expect(testUI.Err).To(Say("some-set-droplet-warning"))
					})
				})
			})

			When("the staging returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("any gibberish")
					fakeActor.StagePackageStub = func(packageGUID, _, _ string) (<-chan v7action.Droplet, <-chan v7action.Warnings, <-chan error) {
						dropletStream := make(chan v7action.Droplet)
						warningsStream := make(chan v7action.Warnings)
						errorStream := make(chan error)

						go func() {
							<-allLogsWritten
							defer close(dropletStream)
							defer close(warningsStream)
							defer close(errorStream)
							warningsStream <- v7action.Warnings{"some-warning", "some-other-warning"}
							errorStream <- expectedErr
						}()

						return dropletStream, warningsStream, errorStream
					}
				})

				It("returns the error and displays warnings", func() {
					Expect(executeErr).To(Equal(expectedErr))

					Expect(testUI.Err).To(Say("some-warning"))
					Expect(testUI.Err).To(Say("some-other-warning"))
				})
			})
		})

		When("the logging stream has errors", func() {
			var (
				expectedErr    error
				allLogsWritten chan bool
			)

			BeforeEach(func() {
				allLogsWritten = make(chan bool)
				expectedErr = errors.New("banana")

				fakeActor.GetStreamingLogsForApplicationByNameAndSpaceStub = func(appName string, spaceGUID string, client v7action.NOAAClient) (<-chan *v7action.LogMessage, <-chan error, v7action.Warnings, error) {
					logStream := make(chan *v7action.LogMessage)
					errorStream := make(chan error)

					go func() {
						defer close(logStream)
						defer close(errorStream)
						logStream <- v7action.NewLogMessage("Here are some staging logs!", 1, time.Now(), v7action.StagingLog, "sourceInstance")
						errorStream <- expectedErr
						allLogsWritten <- true
					}()

					return logStream, errorStream, v7action.Warnings{"steve for all I care"}, nil
				}

				fakeActor.StagePackageStub = func(packageGUID, _, _ string) (<-chan v7action.Droplet, <-chan v7action.Warnings, <-chan error) {
					dropletStream := make(chan v7action.Droplet)
					warningsStream := make(chan v7action.Warnings)
					errorStream := make(chan error)

					go func() {
						<-allLogsWritten
						defer close(dropletStream)
						defer close(warningsStream)
						defer close(errorStream)
						warningsStream <- v7action.Warnings{"some-warning", "some-other-warning"}
						dropletStream <- v7action.Droplet{
							GUID:      "some-droplet-guid",
							CreatedAt: "2017-08-14T21:16:42Z",
							State:     constant.DropletStaged,
						}
					}()

					return dropletStream, warningsStream, errorStream
				}
			})

			It("displays the errors and continues staging", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("banana"))
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Err).To(Say("some-other-warning"))
			})
		})
	})

	When("the get application returns an unknown error", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			expectedErr = errors.New("some-error")
			fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{State: constant.ApplicationStopped}, v7action.Warnings{"get-warning-1", "get-warning-2"}, expectedErr)
		})

		It("says that the app failed to start", func() {
			Expect(executeErr).To(Equal(expectedErr))
			Expect(testUI.Out).ToNot(Say("Starting"))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))

			Expect(fakeActor.StartApplicationCallCount()).To(BeZero(), "Expected StartApplication to not be called")
		})
	})

	When("the start application returns an unknown error", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetApplicationByNameAndSpaceReturns(v7action.Application{State: constant.ApplicationStopped}, v7action.Warnings{"get-warning-1", "get-warning-2"}, nil)
			expectedErr = errors.New("some-error")
			fakeActor.StartApplicationReturns(v7action.Warnings{"start-warning-1", "start-warning-2"}, expectedErr)
		})

		It("says that the app failed to start", func() {
			Expect(executeErr).To(Equal(expectedErr))
			Expect(testUI.Out).To(Say(`Starting app some-app in org some-org / space some-space as steve\.\.\.`))

			Expect(testUI.Err).To(Say("get-warning-1"))
			Expect(testUI.Err).To(Say("get-warning-2"))
			Expect(testUI.Err).To(Say("start-warning-1"))
			Expect(testUI.Err).To(Say("start-warning-2"))
		})
	})

	When("getting the application summary returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = actionerror.ApplicationNotFoundError{Name: app}
			fakeActor.GetDetailedAppSummaryReturns(v7action.DetailedApplicationSummary{}, v7action.Warnings{"warning-1", "warning-2"}, expectedErr)
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
			summary := v7action.DetailedApplicationSummary{
				ApplicationSummary: v7action.ApplicationSummary{
					Application: v7action.Application{
						Name:  "some-app",
						State: constant.ApplicationStarted,
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
			}
			fakeActor.GetDetailedAppSummaryReturns(summary, v7action.Warnings{"warning-1", "warning-2"}, nil)
		})

		It("prints the application summary and outputs warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Waiting for app to start\.\.\.`))
			Expect(testUI.Out).To(Say(`name:\s+some-app`))
			Expect(testUI.Out).To(Say(`requested state:\s+started`))
			Expect(testUI.Out).ToNot(Say("start command:"))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))

			Expect(fakeActor.GetDetailedAppSummaryCallCount()).To(Equal(1))
			appName, spaceGUID, withObfuscatedValues := fakeActor.GetDetailedAppSummaryArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
			Expect(withObfuscatedValues).To(BeFalse())
		})
	})
})
