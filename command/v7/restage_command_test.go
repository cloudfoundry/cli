package v7_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
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
)

var _ = Describe("restage Command", func() {
	var (
		cmd                v7.RestageCommand
		testUI             *ui.UI
		fakeConfig         *commandfakes.FakeConfig
		fakeSharedActor    *commandfakes.FakeSharedActor
		fakeActor          *v7fakes.FakeRestageActor
		fakeLogCacheClient *v7actionfakes.FakeLogCacheClient

		executeErr       error
		app              string
		allLogsWritten   chan bool
		closedTheStreams bool
	)
	const dropletCreateTime = "2017-08-14T21:16:42Z"

	BeforeEach(func() {
		app = "some-app"

		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeRestageActor)
		fakeLogCacheClient = new(v7actionfakes.FakeLogCacheClient)
		cmd = v7.RestageCommand{
			RequiredArgs: flag.AppName{AppName: app},

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
		fakeActor.GetStreamingLogsForApplicationByNameAndSpaceStub = func(appName string, spaceGUID string, client v7action.LogCacheClient) (<-chan v7action.LogMessage, <-chan error, context.CancelFunc, v7action.Warnings, error) {
			logStream := make(chan v7action.LogMessage)
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
				logStream <- *v7action.NewLogMessage("Here are some staging logs!", "OUT", time.Now(), v7action.StagingLog, "sourceInstance")
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
				expectedErr = actionerror.ApplicationNotFoundError{Name: app}
				fakeActor.GetDetailedAppSummaryReturns(v7action.DetailedApplicationSummary{}, v7action.Warnings{"application-summary-warning-1", "application-summary-warning-2"}, expectedErr)
			})

			It("displays all warnings and returns an error", func() {
				Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))

				Expect(testUI.Err).To(Say("application-summary-warning-1"))
				Expect(testUI.Err).To(Say("application-summary-warning-2"))
			})
		})
	})
})
