package shared_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("app stager", func() {
	var (
		appStager          shared.AppStager
		executeErr         error
		testUI             *ui.UI
		fakeConfig         *commandfakes.FakeConfig
		fakeActor          *v7fakes.FakeActor
		fakeLogCacheClient *sharedactionfakes.FakeLogCacheClient

		app      v7action.Application
		pkgGUID  string
		strategy constant.DeploymentStrategy
		noWait   bool

		allLogsWritten   chan bool
		closedTheStreams bool
	)
	const dropletCreateTime = "2017-08-14T21:16:42Z"

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns("some-binary-name")
		fakeActor = new(v7fakes.FakeActor)
		fakeLogCacheClient = new(sharedactionfakes.FakeLogCacheClient)
		allLogsWritten = make(chan bool)

		strategy = constant.DeploymentStrategyDefault

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})
		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
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
				logStream <- *sharedaction.NewLogMessage("Here's an output log!", "OUT", time.Now(), "OUT", "sourceInstance-1")
				logStream <- *sharedaction.NewLogMessage("Here's a staging log!", sharedaction.StagingLog, time.Now(), sharedaction.StagingLog, "sourceInstance-2")
				errorStream <- errors.New("something bad happened while trying to get staging logs")
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
		fakeActor.GetDetailedAppSummaryReturns(v7action.DetailedApplicationSummary{}, v7action.Warnings{"application-summary-warning-1", "application-summary-warning-2"}, nil)
	})

	JustBeforeEach(func() {
		appStager = shared.NewAppStager(fakeActor, testUI, fakeConfig, fakeLogCacheClient)
		executeErr = appStager.StageAndStart(
			app,
			pkgGUID,
			strategy,
			noWait,
		)
	})

	When("getting logs for the app fails because getting the app fails", func() {
		BeforeEach(func() {
			fakeActor.GetStreamingLogsForApplicationByNameAndSpaceReturns(
				nil, nil, nil, v7action.Warnings{"get-log-streaming-warning"}, errors.New("get-log-streaming-error"))
		})

		It("displays all warnings and returns an error", func() {
			Expect(executeErr).To(MatchError("get-log-streaming-error"))
			Expect(testUI.Err).To(Say("get-log-streaming-warning"))
		})
	})

	It("displays that it's staging the app and tracing the logs", func() {
		Expect(testUI.Out).To(Say("Staging app and tracing logs..."))
	})

	When("log streams are opened", func() {
		JustAfterEach(func() {
			Expect(closedTheStreams).To(BeTrue())
		})

		When("staging the package (StagePackage + PollStage) fails", func() {
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

		When("the deployment strategy is rolling", func() {
			BeforeEach(func() {
				strategy = constant.DeploymentStrategyRolling
				fakeActor.CreateDeploymentReturns(
					"some-deployment-guid",
					v7action.Warnings{"create-deployment-warning"},
					nil,
				)
				fakeActor.PollStartForRollingReturns(
					v7action.Warnings{"poll-start-warning"},
					nil,
				)

			})
			It("displays that it's creating the deployment", func() {
				Expect(testUI.Out).To(Say("Creating deployment for app %s...", app.Name))
			})
			It("creates the deployment", func() {
				Expect(fakeActor.CreateDeploymentCallCount()).To(Equal(1))
				appGUID, dropletGUID := fakeActor.CreateDeploymentArgsForCall(0)
				Expect(appGUID).To(Equal(app.GUID))
				Expect(dropletGUID).To(Equal("some-droplet-guid"))
			})
			It("displays the warnings", func() {
				Expect(testUI.Err).To(Say("create-deployment-warning"))
			})

			When("creating a deployment fails", func() {
				BeforeEach(func() {
					fakeActor.CreateDeploymentReturns(
						"",
						v7action.Warnings{"create-deployment-warning"},
						errors.New("create-deployment-error"),
					)
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("create-deployment-error"))
				})
			})
			It("displays that it's waiting for the app to deploy", func() {
				Expect(testUI.Out).To(Say("Waiting for app to deploy..."))
			})

			It("displays all warnings", func() {
				Expect(testUI.Err).To(Say("poll-start-warning"))
			})

			When("polling fails for a rolling restage", func() {
				BeforeEach(func() {
					fakeActor.PollStartForRollingReturns(
						v7action.Warnings{"poll-start-warning"},
						errors.New("poll-start-error"),
					)
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("poll-start-error"))
				})
			})
		})

		When("the deployment strategy is NOT rolling", func() {
			BeforeEach(func() {
				fakeActor.StopApplicationReturns(
					v7action.Warnings{"stop-app-warning"}, nil)
				fakeActor.SetApplicationDropletReturns(
					v7action.Warnings{"set-droplet-warning"}, nil)
				fakeActor.StartApplicationReturns(
					v7action.Warnings{"start-app-warning"}, nil)
				fakeActor.PollStartReturns(
					v7action.Warnings{"poll-app-warning"}, nil)

			})

			It("displays all warnings", func() {
				Expect(testUI.Err).To(Say("stop-app-warning"))
			})

			When("stopping the app fails", func() {
				BeforeEach(func() {
					fakeActor.StopApplicationReturns(
						v7action.Warnings{"stop-app-warning"}, errors.New("stop-app-error"))
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("stop-app-error"))
				})
			})

			It("displays all warnings", func() {
				Expect(testUI.Err).To(Say("set-droplet-warning"))
			})

			When("setting the droplet fails", func() {
				BeforeEach(func() {
					fakeActor.SetApplicationDropletReturns(
						v7action.Warnings{"set-droplet-warning"}, errors.New("set-droplet-error"))
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("set-droplet-error"))
				})
			})

			It("displays that it's waiting for the app to start", func() {
				Expect(testUI.Out).To(Say("Waiting for app to start..."))
			})

			It("displays all warnings", func() {
				Expect(testUI.Err).To(Say("start-app-warning"))
			})

			When("starting the application fails", func() {
				BeforeEach(func() {
					fakeActor.StartApplicationReturns(
						v7action.Warnings{"start-app-warning"}, errors.New("start-app-error"))
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("start-app-error"))
				})
			})

			It("displays all warnings", func() {
				Expect(testUI.Err).To(Say("poll-app-warning"))
			})

			When("polling the application fails", func() {
				BeforeEach(func() {
					fakeActor.PollStartReturns(
						v7action.Warnings{"poll-app-warning"}, errors.New("poll-app-error"))
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("poll-app-error"))
				})
			})
		})
	})

	It("Gets the detailed app summary", func() {
		Expect(fakeActor.GetDetailedAppSummaryCallCount()).To(Equal(1))
		appName, targetedSpaceGUID, obfuscatedValues := fakeActor.GetDetailedAppSummaryArgsForCall(0)
		Expect(appName).To(Equal(app.Name))
		Expect(targetedSpaceGUID).To(Equal("some-space-guid"))
		Expect(obfuscatedValues).To(BeFalse())
	})

	It("displays the warnings", func() {
		Expect(testUI.Err).To(Say("application-summary-warning-1"))
		Expect(testUI.Err).To(Say("application-summary-warning-2"))

	})

	When("getting the app summary fails", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = actionerror.ApplicationNotFoundError{Name: app.Name}
			fakeActor.GetDetailedAppSummaryReturns(v7action.DetailedApplicationSummary{}, v7action.Warnings{"application-summary-warning-1", "application-summary-warning-2"}, expectedErr)
		})

		It("displays all warnings and returns an error", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app.Name}))
		})
	})

	It("succeeds", func() {
		Expect(executeErr).To(Not(HaveOccurred()))
	})
})
