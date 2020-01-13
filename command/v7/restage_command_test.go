package v7_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/cli/command/translatableerror"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
)

var _ = Describe("restage Command", func() {
	var (
		cmd                v7.RestageCommand
		testUI             *ui.UI
		fakeConfig         *commandfakes.FakeConfig
		fakeSharedActor    *commandfakes.FakeSharedActor
		fakeActor          *v7fakes.FakeRestageActor
		fakeLogCacheClient *sharedactionfakes.FakeLogCacheClient

		executeErr       error
		appName          string
		allLogsWritten   chan bool
		closedTheStreams bool
	)
	const dropletCreateTime = "2017-08-14T21:16:42Z"

	BeforeEach(func() {
		appName = "some-app"

		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns("some-binary-name")
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeRestageActor)
		fakeLogCacheClient = new(sharedactionfakes.FakeLogCacheClient)
		cmd = v7.RestageCommand{
			RequiredArgs: flag.AppName{AppName: appName},

			UI:             testUI,
			Config:         fakeConfig,
			SharedActor:    fakeSharedActor,
			Actor:          fakeActor,
			LogCacheClient: fakeLogCacheClient,
		}
		expectedErr := errors.New("banana")
		allLogsWritten = make(chan bool)

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})
		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
		fakeSharedActor.CheckTargetReturns(nil)
		fakeActor.GetApplicationByNameAndSpaceReturns(
			v7action.Application{GUID: "app-guid"},
			v7action.Warnings{"get-app-warning"},
			nil,
		)
		fakeActor.GetNewestReadyPackageForApplicationReturns(
			v7action.Package{GUID: "earliest-package-guid"},
			v7action.Warnings{"get-package-warning"},
			nil,
		)
		fakeActor.GetStreamingLogsForApplicationByNameAndSpaceStub = func(appName string, spaceGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc, v7action.Warnings, error) {
			logStream := make(chan sharedaction.LogMessage)
			errorStream := make(chan error)
			closedTheStreams = false

			cancelFunc := func() {
				if closedTheStreams {
					return
				}
				closedTheStreams = true
				close(logStream)
				close(errorStream)
			}
			go func() {
				logStream <- *sharedaction.NewLogMessage("Here are some staging logs!", "OUT", time.Now(), sharedaction.StagingLog, "sourceInstance")
				errorStream <- expectedErr
				allLogsWritten <- true
			}()

			return logStream, errorStream, cancelFunc, v7action.Warnings{"get-logs-warning"}, nil
		}
		fakeActor.StagePackageStub = func(packageGUID, appName, spaceGUID string) (<-chan v7action.Droplet, <-chan v7action.Warnings, <-chan error) {
			dropletStream := make(chan v7action.Droplet)
			warningsStream := make(chan v7action.Warnings)
			errorStream := make(chan error)

			go func() {
				<-allLogsWritten
				defer close(dropletStream)
				defer close(warningsStream)
				defer close(errorStream)
				warningsStream <- v7action.Warnings{"stage-package-warning"}
				dropletStream <- v7action.Droplet{
					GUID:      "some-droplet-guid",
					CreatedAt: dropletCreateTime,
					State:     constant.DropletStaged,
				}
			}()

			return dropletStream, warningsStream, errorStream
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("The app is successfully restaged", func() {
		It("stages the latest package for the requested app", func() {
			Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))

			Expect(fakeActor.GetNewestReadyPackageForApplicationCallCount()).To(Equal(1))
			appGUID := fakeActor.GetNewestReadyPackageForApplicationArgsForCall(0)
			Expect(appGUID).To(Equal("app-guid"))

			Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(1))

			Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
			pkgGUID, appName, spaceGUID := fakeActor.StagePackageArgsForCall(0)
			Expect(pkgGUID).To(Equal("earliest-package-guid"))
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
		})

		It("stops the app", func() {
			Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
			appGUID := fakeActor.StopApplicationArgsForCall(0)
			Expect(appGUID).To(Equal("app-guid"))
		})

		It("assigns the droplet and starts the application", func() {
			Expect(fakeActor.SetApplicationDropletCallCount()).To(Equal(1))
			appGUID, dropletGUID := fakeActor.SetApplicationDropletArgsForCall(0)
			Expect(appGUID).To(Equal("app-guid"))
			Expect(dropletGUID).To(Equal("some-droplet-guid"))

			Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
			appGUID = fakeActor.StartApplicationArgsForCall(0)
			Expect(appGUID).To(Equal("app-guid"))

			Expect(fakeActor.PollStartCallCount()).To(Equal(1))
			appGUID, _ = fakeActor.PollStartArgsForCall(0)
			Expect(appGUID).To(Equal("app-guid"))
		})

		It("prints the app summary", func() {
			Expect(testUI.Out).To(Say(`Restaging app some-app in org some-org / space some-space as steve...`))

			Expect(testUI.Out).To(Say("Staging app and tracing logs..."))
			Expect(testUI.Err).To(Say("get-app-warning"))
			Expect(testUI.Err).To(Say("get-package-warning"))
			Expect(testUI.Err).To(Say("get-logs-warning"))
			Expect(testUI.Err).To(Say("stage-package-warning"))

			Expect(fakeActor.GetDetailedAppSummaryCallCount()).To(Equal(1))
			appName, spaceGUID, withObfuscatedValues := fakeActor.GetDetailedAppSummaryArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))
			Expect(withObfuscatedValues).To(BeFalse())
		})

		When("a rolling strategy is specified", func() {
			BeforeEach(func() {
				cmd.Strategy.Name = constant.DeploymentStrategyRolling

				fakeActor.CreateDeploymentReturns(
					"deployment-guid",
					v7action.Warnings{"create-deployment-warning"},
					nil,
				)
				fakeActor.PollStartForRollingReturns(
					v7action.Warnings{"poll-deployment-warning"},
					nil,
				)

			})
			It("creates a deployment", func() {
				Expect(fakeActor.CreateDeploymentCallCount()).To(Equal(1))
				appGUID, dropletGUID := fakeActor.CreateDeploymentArgsForCall(0)
				Expect(appGUID).To(Equal("app-guid"))
				Expect(dropletGUID).To(Equal("some-droplet-guid"))

				Expect(fakeActor.PollStartForRollingCallCount()).To(Equal(1))
				appGUID, deploymentGUID, noWait := fakeActor.PollStartForRollingArgsForCall(0)
				Expect(appGUID).To(Equal("app-guid"))
				Expect(deploymentGUID).To(Equal("deployment-guid"))
				Expect(noWait).To(BeFalse())
			})

			It("print deployment output and the app summary", func() {
				Expect(testUI.Out).To(Say(`Restaging app some-app in org some-org / space some-space as steve...`))
				Expect(testUI.Out).To(Say("Waiting for app to deploy..."))
				Expect(testUI.Err).To(Say("create-deployment-warning"))
				Expect(testUI.Err).To(Say("poll-deployment-warning"))

				Expect(fakeActor.GetDetailedAppSummaryCallCount()).To(Equal(1))
				appName, spaceGUID, withObfuscatedValues := fakeActor.GetDetailedAppSummaryArgsForCall(0)
				Expect(appName).To(Equal("some-app"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(withObfuscatedValues).To(BeFalse())
			})
		})

		When("--no-wait is specified", func() {
			BeforeEach(func() {
				cmd.NoWait = true
			})

			It("respects the no-wait value and polls appropriately", func() {
				// deployment strategy is null
				Expect(fakeActor.PollStartCallCount()).To(Equal(1))
				_, noWait := fakeActor.PollStartArgsForCall(0)
				Expect(noWait).To(BeTrue())

				// deployment strategy is rolling
				cmd.Strategy.Name = constant.DeploymentStrategyRolling
				executeErr = cmd.Execute(nil)

				Expect(fakeActor.PollStartForRollingCallCount()).To(Equal(1))
				_, _, noWait = fakeActor.PollStartForRollingArgsForCall(0)
				Expect(noWait).To(BeTrue())
			})
		})
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: "binary"})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: "binary"}))

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

	When("getting the application fails", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationByNameAndSpaceReturns(
				v7action.Application{},
				v7action.Warnings{"get-app-warning"},
				errors.New("get-app-error"),
			)
		})

		It("displays all warnings and returns an error", func() {
			Expect(executeErr).To(MatchError("get-app-error"))
			Expect(testUI.Err).To(Say("get-app-warning"))
		})
	})

	When("getting the package fails", func() {
		BeforeEach(func() {
			fakeActor.GetNewestReadyPackageForApplicationReturns(
				v7action.Package{},
				v7action.Warnings{"get-package-warning"},
				errors.New("get-package-error"),
			)
		})

		It("displays all warnings and returns an error", func() {
			Expect(executeErr).To(MatchError("get-package-error"))
			Expect(testUI.Err).To(Say("get-package-warning"))
		})
	})

	When("there are no packages available to stage", func() {
		BeforeEach(func() {
			fakeActor.GetNewestReadyPackageForApplicationReturns(
				v7action.Package{},
				v7action.Warnings{"get-package-warning"},
				actionerror.PackageNotFoundInAppError{},
			)
		})

		It("displays all warnings and returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.PackageNotFoundInAppError{
				AppName: appName, BinaryName: "some-binary-name"}))
			Expect(testUI.Err).To(Say("get-package-warning"))
		})

	})

	When("getting logs for the app fails", func() {
		BeforeEach(func() {
			fakeActor.GetStreamingLogsForApplicationByNameAndSpaceReturns(
				nil, nil, nil, v7action.Warnings{"get-log-streaming-warning"}, errors.New("get-log-streaming-error"))
		})

		It("displays all warnings and returns an error", func() {
			Expect(executeErr).To(MatchError("get-log-streaming-error"))
			Expect(testUI.Err).To(Say("get-log-streaming-warning"))
		})
	})

	When("log streams are opened", func() {
		JustAfterEach(func() {
			Expect(closedTheStreams).To(BeTrue())
		})
		When("staging the package fails", func() {
			BeforeEach(func() {
				expectedErr := errors.New("package-staging-error")
				fakeActor.StagePackageStub = func(packageGUID, _, _ string) (<-chan v7action.Droplet, <-chan v7action.Warnings, <-chan error) {

					dropletStream := make(chan v7action.Droplet)
					warningsStream := make(chan v7action.Warnings)
					errorStream := make(chan error)

					go func() {
						<-allLogsWritten
						defer close(dropletStream)
						defer close(warningsStream)
						defer close(errorStream)
						warningsStream <- v7action.Warnings{"some-package-warning", "some-other-package-warning"}
						errorStream <- expectedErr
					}()

					return dropletStream, warningsStream, errorStream
				}
			})

			It("displays all warnings and returns an error", func() {
				Expect(executeErr).To(MatchError("package-staging-error"))

				Expect(testUI.Err).To(Say("some-package-warning"))
				Expect(testUI.Err).To(Say("some-other-package-warning"))
			})
		})

		When("stopping the app fails", func() {
			BeforeEach(func() {
				fakeActor.StopApplicationReturns(
					v7action.Warnings{"stop-app-warning"}, errors.New("stop-app-error"))
			})

			It("displays all warnings and returns an error", func() {
				Expect(executeErr).To(MatchError("stop-app-error"))
				Expect(testUI.Err).To(Say("stop-app-warning"))
			})
		})

		When("setting the droplet fails", func() {
			BeforeEach(func() {
				fakeActor.SetApplicationDropletReturns(
					v7action.Warnings{"set-droplet-warning"}, errors.New("set-droplet-error"))
			})

			It("displays all warnings and returns an error", func() {
				Expect(executeErr).To(MatchError("set-droplet-error"))
				Expect(testUI.Err).To(Say("set-droplet-warning"))
			})
		})

		When("starting the application fails", func() {
			BeforeEach(func() {
				fakeActor.PollStartReturns(
					v7action.Warnings{"start-app-warning"}, errors.New("start-app-error"))
			})

			It("displays all warnings and returns an error", func() {
				Expect(executeErr).To(MatchError("start-app-error"))
				Expect(testUI.Err).To(Say("start-app-warning"))
			})
		})

		When("getting the app summary fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = actionerror.ApplicationNotFoundError{Name: appName}
				fakeActor.GetDetailedAppSummaryReturns(v7action.DetailedApplicationSummary{}, v7action.Warnings{"application-summary-warning-1", "application-summary-warning-2"}, expectedErr)
			})

			It("displays all warnings and returns an error", func() {
				Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: appName}))

				Expect(testUI.Err).To(Say("application-summary-warning-1"))
				Expect(testUI.Err).To(Say("application-summary-warning-2"))
			})
		})
	})

	When("creating a deployment fails for a rolling restage", func() {
		BeforeEach(func() {
			cmd.Strategy.Name = constant.DeploymentStrategyRolling
			fakeActor.CreateDeploymentReturns(
				"",
				v7action.Warnings{"create-deployment-warning"},
				errors.New("create-deployment-error"),
			)
		})

		It("displays all warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("create-deployment-warning"))
			Expect(executeErr).To(MatchError("create-deployment-error"))
		})
	})

	When("polling fails for a rolling restage", func() {
		BeforeEach(func() {
			cmd.Strategy.Name = constant.DeploymentStrategyRolling
			fakeActor.CreateDeploymentReturns(
				"some-deployment",
				v7action.Warnings{},
				nil,
			)
			fakeActor.PollStartForRollingReturns(
				v7action.Warnings{"poll-start-warning"},
				errors.New("poll-start-error"),
			)
		})

		It("displays all warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("poll-start-warning"))
			Expect(executeErr).To(MatchError("poll-start-error"))
		})
	})

	When("starting the application fails", func() {
		BeforeEach(func() {
			fakeActor.StartApplicationReturns(
				v7action.Warnings{"start-app-warning"},
				errors.New("start-app-error"),
			)
		})

		It("displays all warnings and returns an error", func() {
			Expect(testUI.Err).To(Say("start-app-warning"))
			Expect(executeErr).To(MatchError("start-app-error"))
		})
	})

})
