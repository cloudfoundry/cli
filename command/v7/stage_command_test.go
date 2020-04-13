package v7_test

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
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("stage Command", func() {
	const dropletCreateTime = "2017-08-14T21:16:42Z"

	var (
		cmd                v7.StageCommand
		testUI             *ui.UI
		fakeConfig         *commandfakes.FakeConfig
		fakeSharedActor    *commandfakes.FakeSharedActor
		fakeActor          *v7fakes.FakeActor
		fakeLogCacheClient *sharedactionfakes.FakeLogCacheClient

		binaryName  string
		executeErr  error
		appName     string
		packageGUID string
		spaceGUID   string
		app         v7action.Application

		allLogsWritten   chan bool
		closedTheStreams bool
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		fakeLogCacheClient = new(sharedactionfakes.FakeLogCacheClient)

		fakeConfig.StagingTimeoutReturns(10 * time.Minute)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		appName = "some-app"
		packageGUID = "some-package-guid"
		spaceGUID = "some-space-guid"

		cmd = v7.StageCommand{
			RequiredArgs: flag.AppName{AppName: appName},
			PackageGUID:  packageGUID,
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			LogCacheClient: fakeLogCacheClient,
		}

		fakeConfig.HasTargetedOrganizationReturns(true)
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			GUID: "some-org-guid",
			Name: "some-org",
		})
		fakeConfig.HasTargetedSpaceReturns(true)
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			GUID: spaceGUID,
			Name: "some-space",
		})
		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)

		allLogsWritten = make(chan bool)
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
				logStream <- *sharedaction.NewLogMessage("Here are some other staging logs!", "OUT", time.Now(), sharedaction.StagingLog, "sourceInstance")
				errorStream <- errors.New("problem getting more staging logs")
				allLogsWritten <- true
			}()

			return logStream, errorStream, cancelFunc, v7action.Warnings{"steve for all I care"}, nil
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
				warningsStream <- v7action.Warnings{"some-warning", "some-other-warning"}
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

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("displays the experimental warning", func() {
			Expect(testUI.Err).NotTo(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the package's GUID is not passed in", func() {
		var (
			newestPackageGUID string
		)

		BeforeEach(func() {
			cmd.PackageGUID = ""
			app = v7action.Application{GUID: "some-app-guid", Name: "some-name"}
			newestPackageGUID = "newest-package-guid"

			fakeActor.GetApplicationByNameAndSpaceReturns(
				app,
				v7action.Warnings{"app-by-name-warning"},
				nil)

			fakeActor.GetNewestReadyPackageForApplicationReturns(
				v7action.Package{GUID: newestPackageGUID},
				v7action.Warnings{"newest-pkg-warning"},
				nil)
		})

		It("grabs the most recent version", func() {
			Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
			appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
			Expect(appNameArg).To(Equal(cmd.RequiredArgs.AppName))
			Expect(spaceGUIDArg).To(Equal(spaceGUID))
			Expect(testUI.Err).To(Say("app-by-name-warning"))

			Expect(fakeActor.GetNewestReadyPackageForApplicationCallCount()).To(Equal(1))
			appArg := fakeActor.GetNewestReadyPackageForApplicationArgsForCall(0)
			Expect(appArg).To(Equal(app))
			Expect(testUI.Err).To(Say("newest-pkg-warning"))

			Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
			guidArg, appNameArg, spaceGUIDArg := fakeActor.StagePackageArgsForCall(0)
			Expect(guidArg).To(Equal(newestPackageGUID))
			Expect(appNameArg).To(Equal(appName))
			Expect(spaceGUIDArg).To(Equal(spaceGUID))
		})
		When("It can't get the application's information", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v7action.Application{},
					v7action.Warnings{"app-warning"},
					errors.New("cant-get-app-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(testUI.Err).To(Say("app-warning"))
				Expect(executeErr).To(MatchError("cant-get-app-error"))
			})
		})
	})

	When("the logging stream has errors", func() {
		var (
			expectedErr      error
			allLogsWritten   chan bool
			closedTheStreams bool
		)

		BeforeEach(func() {
			allLogsWritten = make(chan bool)
			expectedErr = errors.New("banana")

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
					logStream <- *sharedaction.NewLogMessage("Here are some staging logs!", "err", time.Now(), sharedaction.StagingLog, "sourceInstance")
					errorStream <- expectedErr
					allLogsWritten <- true
				}()

				return logStream, errorStream, cancelFunc, v7action.Warnings{"steve for all I care"}, nil
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

		JustAfterEach(func() {
			Expect(closedTheStreams).To(BeTrue())
		})

		It("displays the errors and continues staging", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Err).To(Say("banana"))
			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Err).To(Say("some-other-warning"))
		})
	})

	When("the logging returns an error due to an API error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("something is wrong!")
			logStream := make(chan sharedaction.LogMessage)
			errorStream := make(chan error)
			cancelFunc := func() {
				close(logStream)
				close(errorStream)
			}
			fakeActor.GetStreamingLogsForApplicationByNameAndSpaceReturns(logStream, errorStream, cancelFunc, v7action.Warnings{"some-warning", "some-other-warning"}, expectedErr)
		})

		It("returns the error and displays warnings", func() {
			Expect(executeErr).To(Equal(expectedErr))

			Expect(testUI.Err).To(Say("some-warning"))
			Expect(testUI.Err).To(Say("some-other-warning"))
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
			Expect(closedTheStreams).To(BeTrue())
		})
	})

	It("outputs the droplet GUID", func() {
		Expect(executeErr).ToNot(HaveOccurred())

		createdAtTimeParsed, err := time.Parse(time.RFC3339, dropletCreateTime)
		Expect(err).ToNot(HaveOccurred())

		Expect(testUI.Out).To(Say("Staging package for %s in org some-org / space some-space as steve...", appName))
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
		Expect(appNameArg).To(Equal(appName))
		Expect(spaceGUIDArg).To(Equal("some-space-guid"))
	})

	It("displays staging logs and their warnings", func() {
		Expect(testUI.Out).To(Say("Here are some staging logs!"))
		Expect(testUI.Out).To(Say("Here are some other staging logs!"))

		Expect(testUI.Err).To(Say("steve for all I care"))
		Eventually(testUI.Err).Should(Say("Failed to retrieve logs from Log Cache: problem getting more staging logs"))

		Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(1))
		appNameArg, spaceGUID, logCacheClient := fakeActor.GetStreamingLogsForApplicationByNameAndSpaceArgsForCall(0)
		Expect(appNameArg).To(Equal(appName))
		Expect(spaceGUID).To(Equal("some-space-guid"))
		Expect(logCacheClient).To(Equal(fakeLogCacheClient))

		Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
		guidArg, appNameArg, spaceGUIDArg := fakeActor.StagePackageArgsForCall(0)
		Expect(guidArg).To(Equal(packageGUID))
		Expect(appNameArg).To(Equal(appName))
		Expect(spaceGUIDArg).To(Equal(spaceGUID))

		Expect(closedTheStreams).To(BeTrue())
	})
})
