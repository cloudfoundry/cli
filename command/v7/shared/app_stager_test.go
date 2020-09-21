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
	"code.cloudfoundry.org/cli/resources"
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

		app          resources.Application
		space        configv3.Space
		organization configv3.Organization
		pkgGUID      string
		strategy     constant.DeploymentStrategy
		noWait       bool
		appAction    constant.ApplicationAction

		allLogsWritten   chan bool
		closedTheStreams bool
	)

	const dropletCreateTime = "2017-08-14T21:16:42Z"

	Context("StageAndStart", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeConfig = new(commandfakes.FakeConfig)
			fakeConfig.BinaryNameReturns("some-binary-name")
			fakeActor = new(v7fakes.FakeActor)
			fakeLogCacheClient = new(sharedactionfakes.FakeLogCacheClient)
			allLogsWritten = make(chan bool)

			pkgGUID = "package-guid"
			app = resources.Application{GUID: "app-guid", Name: "app-name"}
			space = configv3.Space{Name: "some-space", GUID: "some-space-guid"}
			organization = configv3.Organization{Name: "some-org"}
			strategy = constant.DeploymentStrategyDefault
			appAction = constant.ApplicationRestarting

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
			fakeActor.StagePackageStub = func(packageGUID, appName, spaceGUID string) (<-chan resources.Droplet, <-chan v7action.Warnings, <-chan error) {
				dropletStream := make(chan resources.Droplet)
				warningsStream := make(chan v7action.Warnings)
				errorStream := make(chan error)

				go func() {
					<-allLogsWritten
					defer close(dropletStream)
					defer close(warningsStream)
					defer close(errorStream)
					warningsStream <- v7action.Warnings{"stage-package-warning"}
					dropletStream <- resources.Droplet{
						GUID:      "some-droplet-guid",
						CreatedAt: dropletCreateTime,
						State:     constant.DropletStaged,
					}
				}()

				return dropletStream, warningsStream, errorStream
			}
		})

		JustBeforeEach(func() {
			appStager = shared.NewAppStager(fakeActor, testUI, fakeConfig, fakeLogCacheClient)
			executeErr = appStager.StageAndStart(
				app,
				space,
				organization,
				pkgGUID,
				strategy,
				noWait,
				appAction,
			)
		})

		It("stages and starts the app", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(testUI.Out).To(Say("Staging app and tracing logs..."))

			user, err := fakeConfig.CurrentUser()
			Expect(err).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, app.Name, organization.Name, space.Name, user.Name))
			Expect(testUI.Out).To(Say("Waiting for app to start..."))
		})

		When("staging fails", func() {
			JustAfterEach(func() {
				Expect(closedTheStreams).To(BeTrue())
			})
			BeforeEach(func() {
				expectedErr := errors.New("package-staging-error")
				fakeActor.StagePackageStub = func(packageGUID, _, _ string) (<-chan resources.Droplet, <-chan v7action.Warnings, <-chan error) {

					dropletStream := make(chan resources.Droplet)
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

		When("starting fails", func() {
			BeforeEach(func() {
				fakeActor.StartApplicationReturns(
					v7action.Warnings{"start-app-warning"}, errors.New("start-app-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("start-app-error"))
			})
		})
	})

	Context("StageApp", func() {
		var (
			droplet resources.Droplet
		)

		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeConfig = new(commandfakes.FakeConfig)
			fakeConfig.BinaryNameReturns("some-binary-name")
			fakeActor = new(v7fakes.FakeActor)
			fakeLogCacheClient = new(sharedactionfakes.FakeLogCacheClient)
			allLogsWritten = make(chan bool)

			pkgGUID = "package-guid"
			app = resources.Application{GUID: "app-guid", Name: "app-name"}
			space = configv3.Space{Name: "some-space", GUID: "some-space-guid"}
			organization = configv3.Organization{Name: "some-org"}
			strategy = constant.DeploymentStrategyDefault

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
			fakeActor.StagePackageStub = func(packageGUID, appName, spaceGUID string) (<-chan resources.Droplet, <-chan v7action.Warnings, <-chan error) {
				dropletStream := make(chan resources.Droplet)
				warningsStream := make(chan v7action.Warnings)
				errorStream := make(chan error)

				go func() {
					<-allLogsWritten
					defer close(dropletStream)
					defer close(warningsStream)
					defer close(errorStream)
					warningsStream <- v7action.Warnings{"stage-package-warning"}
					dropletStream <- resources.Droplet{
						GUID:      "some-droplet-guid",
						CreatedAt: dropletCreateTime,
						State:     constant.DropletStaged,
					}
				}()

				return dropletStream, warningsStream, errorStream
			}
		})

		JustBeforeEach(func() {
			appStager = shared.NewAppStager(fakeActor, testUI, fakeConfig, fakeLogCacheClient)
			droplet, executeErr = appStager.StageApp(
				app,
				pkgGUID,
				space,
			)
		})

		It("stages the app and polls until it is complete", func() {
			Expect(testUI.Out).To(Say("Staging app and tracing logs..."))
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(droplet).To(Equal(
				resources.Droplet{
					GUID:      "some-droplet-guid",
					CreatedAt: dropletCreateTime,
					State:     constant.DropletStaged,
				},
			))
			Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
			inputPackageGUID, inputAppName, inputSpaceGUID := fakeActor.StagePackageArgsForCall(0)
			Expect(inputPackageGUID).To(Equal("package-guid"))
			Expect(inputAppName).To(Equal("app-name"))
			Expect(inputSpaceGUID).To(Equal("some-space-guid"))
		})

		When("getting logs for the app fails because getting the app fails", func() {
			BeforeEach(func() {
				fakeActor.GetStreamingLogsForApplicationByNameAndSpaceReturns(
					nil, nil, nil, v7action.Warnings{"get-log-streaming-warning"}, errors.New("get-log-streaming-error"))
			})

			It("displays all warnings and returns an error", func() {
				Expect(executeErr).To(MatchError("get-log-streaming-error"))
				Expect(droplet).To(Equal(resources.Droplet{}))
				Expect(testUI.Err).To(Say("get-log-streaming-warning"))
			})
		})

		When("staging the package fails", func() {
			JustAfterEach(func() {
				Expect(closedTheStreams).To(BeTrue())
			})
			BeforeEach(func() {
				expectedErr := errors.New("package-staging-error")
				fakeActor.StagePackageStub = func(packageGUID, _, _ string) (<-chan resources.Droplet, <-chan v7action.Warnings, <-chan error) {

					dropletStream := make(chan resources.Droplet)
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
	})

	Context("StartApp", func() {
		var (
			resourceGUID string
		)

		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeConfig = new(commandfakes.FakeConfig)
			fakeConfig.BinaryNameReturns("some-binary-name")
			fakeActor = new(v7fakes.FakeActor)
			fakeLogCacheClient = new(sharedactionfakes.FakeLogCacheClient)
			allLogsWritten = make(chan bool)

			strategy = constant.DeploymentStrategyDefault
			noWait = true
			appAction = constant.ApplicationRestarting

			app = resources.Application{GUID: "app-guid", Name: "app-name", State: constant.ApplicationStarted}
			space = configv3.Space{Name: "some-space", GUID: "some-space-guid"}
			organization = configv3.Organization{Name: "some-org"}
			resourceGUID = "droplet-guid"

			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			fakeActor.GetDetailedAppSummaryReturns(v7action.DetailedApplicationSummary{}, v7action.Warnings{"application-summary-warning-1", "application-summary-warning-2"}, nil)
		})

		JustBeforeEach(func() {
			appStager = shared.NewAppStager(fakeActor, testUI, fakeConfig, fakeLogCacheClient)
			executeErr = appStager.StartApp(
				app,
				resourceGUID,
				strategy,
				noWait,
				space,
				organization,
				appAction,
			)
		})

		When("the deployment strategy is rolling", func() {
			BeforeEach(func() {
				strategy = constant.DeploymentStrategyRolling
				fakeActor.CreateDeploymentByApplicationAndDropletReturns(
					"some-deployment-guid",
					v7action.Warnings{"create-deployment-warning"},
					nil,
				)

				fakeActor.PollStartForRollingReturns(
					v7action.Warnings{"poll-start-warning"},
					nil,
				)
			})

			When("the appAction is rolling back", func() {
				BeforeEach(func() {
					appAction = constant.ApplicationRollingBack
					resourceGUID = "revision-guid"
					fakeActor.CreateDeploymentByApplicationAndRevisionReturns(
						"some-deployment-guid",
						v7action.Warnings{"create-deployment-warning"},
						nil,
					)
				})

				It("displays output for each step of rolling back", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(testUI.Out).To(Say("Creating deployment for app %s...", app.Name))
					Expect(fakeActor.CreateDeploymentByApplicationAndRevisionCallCount()).To(Equal(1), "CreateDeployment...")
					appGUID, revisionGUID := fakeActor.CreateDeploymentByApplicationAndRevisionArgsForCall(0)
					Expect(appGUID).To(Equal(app.GUID))
					Expect(revisionGUID).To(Equal("revision-guid"))
					Expect(testUI.Err).To(Say("create-deployment-warning"))

					Expect(testUI.Out).To(Say("Waiting for app to deploy..."))
					Expect(fakeActor.PollStartForRollingCallCount()).To(Equal(1))
					Expect(testUI.Err).To(Say("poll-start-warning"))
				})
			})

			When("the app starts successfully", func() {
				It("displays output for each step of deploying", func() {
					Expect(executeErr).To(BeNil())

					Expect(testUI.Out).To(Say("Creating deployment for app %s...", app.Name))
					Expect(fakeActor.CreateDeploymentByApplicationAndDropletCallCount()).To(Equal(1))
					appGUID, dropletGUID := fakeActor.CreateDeploymentByApplicationAndDropletArgsForCall(0)
					Expect(appGUID).To(Equal(app.GUID))
					Expect(dropletGUID).To(Equal("droplet-guid"))
					Expect(testUI.Err).To(Say("create-deployment-warning"))

					Expect(testUI.Out).To(Say("Waiting for app to deploy..."))
					Expect(fakeActor.PollStartForRollingCallCount()).To(Equal(1))
					Expect(testUI.Err).To(Say("poll-start-warning"))
				})
			})

			When("creating a deployment fails", func() {
				BeforeEach(func() {
					fakeActor.CreateDeploymentByApplicationAndDropletReturns(
						"",
						v7action.Warnings{"create-deployment-warning"},
						errors.New("create-deployment-error"),
					)
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("create-deployment-error"))
				})
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

			When("the app action is starting", func() {
				BeforeEach(func() {
					appAction = constant.ApplicationStarting
					app = resources.Application{GUID: "app-guid", Name: "app-name", State: constant.ApplicationStopped}
				})

				It("displays output for each step of starting", func() {
					Expect(executeErr).To(BeNil())

					user, err := fakeConfig.CurrentUser()
					Expect(err).NotTo(HaveOccurred())
					Expect(testUI.Out).To(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, app.Name, organization.Name, space.Name, user.Name))

					Expect(testUI.Out).NotTo(Say("Stopping app..."))
					Expect(fakeActor.StopApplicationCallCount()).To(Equal(0))

					Expect(fakeActor.SetApplicationDropletCallCount()).To(Equal(1))
					Expect(testUI.Err).To(Say("set-droplet-warning"))

					Expect(testUI.Out).To(Say("Waiting for app to start..."))
					Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
					Expect(testUI.Err).To(Say("start-app-warning"))

					Expect(fakeActor.PollStartCallCount()).To(Equal(1))
					Expect(testUI.Err).To(Say("poll-app-warning"))
				})

				When("the app is already started", func() {
					BeforeEach(func() {
						app = resources.Application{GUID: "app-guid", Name: "app-name", State: constant.ApplicationStarted}
					})
					It("does not attempt to start the application", func() {
						Expect(executeErr).To(BeNil())
						Expect(testUI.Out).To(Say(`App '%s' is already started.`, app.Name))
						Expect(fakeActor.StartApplicationCallCount()).To(Equal(0))
					})
				})
			})

			When("the app action is restarting", func() {
				It("displays output for each step of restarting", func() {
					Expect(executeErr).To(BeNil())

					user, err := fakeConfig.CurrentUser()
					Expect(err).NotTo(HaveOccurred())
					Expect(testUI.Out).To(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, app.Name, organization.Name, space.Name, user.Name))

					Expect(testUI.Out).To(Say("Stopping app..."))
					Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
					Expect(testUI.Err).To(Say("stop-app-warning"))

					Expect(fakeActor.SetApplicationDropletCallCount()).To(Equal(1))
					Expect(testUI.Err).To(Say("set-droplet-warning"))

					Expect(testUI.Out).To(Say("Waiting for app to start..."))
					Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
					Expect(testUI.Err).To(Say("start-app-warning"))

					Expect(fakeActor.PollStartCallCount()).To(Equal(1))
					Expect(testUI.Err).To(Say("poll-app-warning"))
				})

				When("the app is already stopped", func() {
					BeforeEach(func() {
						app = resources.Application{GUID: "app-guid", Name: "app-name", State: constant.ApplicationStopped}
					})

					It("does not stop the application", func() {
						Expect(executeErr).To(BeNil())
						Expect(testUI.Out).NotTo(Say("Stopping app..."))
						Expect(fakeActor.StopApplicationCallCount()).To(Equal(0))
					})
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
			})

			When("a droplet guid is not provided", func() {
				BeforeEach(func() {
					resourceGUID = ""
				})

				It("does not try to set the application droplet", func() {
					Expect(executeErr).To(BeNil())
					Expect(fakeActor.SetApplicationDropletCallCount()).To(Equal(0))
				})
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

			When("starting the application fails", func() {
				BeforeEach(func() {
					fakeActor.StartApplicationReturns(
						v7action.Warnings{"start-app-warning"}, errors.New("start-app-error"))
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("start-app-error"))
				})
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
})
