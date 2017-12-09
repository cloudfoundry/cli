package v3_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
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

var _ = Describe("v3-stage Command", func() {
	var (
		cmd             v3.V3StageCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3StageActor
		fakeNOAAClient  *v3actionfakes.FakeNOAAClient

		binaryName  string
		executeErr  error
		app         string
		packageGUID string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3StageActor)
		fakeNOAAClient = new(v3actionfakes.FakeNOAAClient)

		fakeConfig.StagingTimeoutReturns(10 * time.Minute)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"
		packageGUID = "some-package-guid"

		cmd = v3.V3StageCommand{
			RequiredArgs: flag.AppName{AppName: app},
			PackageGUID:  packageGUID,

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			NOAAClient:  fakeNOAAClient,
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
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.HasTargetedOrganizationReturns(true)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.HasTargetedSpaceReturns(true)
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
		})

		Context("when the logging does not error", func() {
			var allLogsWritten chan bool

			BeforeEach(func() {
				allLogsWritten = make(chan bool)
				fakeActor.GetStreamingLogsForApplicationByNameAndSpaceStub = func(appName string, spaceGUID string, client v3action.NOAAClient) (<-chan *v3action.LogMessage, <-chan error, v3action.Warnings, error) {
					logStream := make(chan *v3action.LogMessage)
					errorStream := make(chan error)

					go func() {
						logStream <- v3action.NewLogMessage("Here are some staging logs!", 1, time.Now(), v3action.StagingLog, "sourceInstance")
						logStream <- v3action.NewLogMessage("Here are some other staging logs!", 1, time.Now(), v3action.StagingLog, "sourceInstance")
						allLogsWritten <- true
					}()

					return logStream, errorStream, v3action.Warnings{"steve for all I care"}, nil
				}
			})

			Context("when the staging is successful", func() {
				const dropletCreateTime = "2017-08-14T21:16:42Z"

				BeforeEach(func() {
					fakeActor.StagePackageStub = func(packageGUID string, _ string) (<-chan v3action.Droplet, <-chan v3action.Warnings, <-chan error) {
						dropletStream := make(chan v3action.Droplet)
						warningsStream := make(chan v3action.Warnings)
						errorStream := make(chan error)

						go func() {
							<-allLogsWritten
							defer close(dropletStream)
							defer close(warningsStream)
							defer close(errorStream)
							warningsStream <- v3action.Warnings{"some-warning", "some-other-warning"}
							dropletStream <- v3action.Droplet{
								GUID:      "some-droplet-guid",
								CreatedAt: dropletCreateTime,
								State:     constant.DropletStaged,
							}
						}()

						return dropletStream, warningsStream, errorStream
					}
				})

				It("outputs the droplet GUID", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					createdAtTimeParsed, err := time.Parse(time.RFC3339, dropletCreateTime)
					Expect(err).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Staging package for %s in org some-org / space some-space as steve...", app))
					Expect(testUI.Out).To(Say("\n\n"))
					Expect(testUI.Out).To(Say("Package staged"))
					Expect(testUI.Out).To(Say("droplet guid:\\s+some-droplet-guid"))
					Expect(testUI.Out).To(Say("state:\\s+staged"))
					Expect(testUI.Out).To(Say("created:\\s+%s", testUI.UserFriendlyDate(createdAtTimeParsed)))

					Expect(testUI.Err).To(Say("some-warning"))
					Expect(testUI.Err).To(Say("some-other-warning"))
				})

				It("stages the package", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
					guidArg, _ := fakeActor.StagePackageArgsForCall(0)
					Expect(guidArg).To(Equal(packageGUID))
				})

				It("displays staging logs and their warnings", func() {
					Expect(testUI.Out).To(Say("Here are some staging logs!"))
					Expect(testUI.Out).To(Say("Here are some other staging logs!"))

					Expect(testUI.Err).To(Say("steve for all I care"))

					Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID, noaaClient := fakeActor.GetStreamingLogsForApplicationByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal(app))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(noaaClient).To(Equal(fakeNOAAClient))

					Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
					guidArg, _ := fakeActor.StagePackageArgsForCall(0)
					Expect(guidArg).To(Equal(packageGUID))
				})
			})

			Context("when the staging returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("any gibberish")
					fakeActor.StagePackageStub = func(packageGUID string, _ string) (<-chan v3action.Droplet, <-chan v3action.Warnings, <-chan error) {
						dropletStream := make(chan v3action.Droplet)
						warningsStream := make(chan v3action.Warnings)
						errorStream := make(chan error)

						go func() {
							<-allLogsWritten
							defer close(dropletStream)
							defer close(warningsStream)
							defer close(errorStream)
							warningsStream <- v3action.Warnings{"some-warning", "some-other-warning"}
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

		Context("when the logging stream has errors", func() {
			var (
				expectedErr    error
				allLogsWritten chan bool
			)

			BeforeEach(func() {
				allLogsWritten = make(chan bool)
				expectedErr = errors.New("banana")

				fakeActor.GetStreamingLogsForApplicationByNameAndSpaceStub = func(appName string, spaceGUID string, client v3action.NOAAClient) (<-chan *v3action.LogMessage, <-chan error, v3action.Warnings, error) {
					logStream := make(chan *v3action.LogMessage)
					errorStream := make(chan error)

					go func() {
						defer close(logStream)
						defer close(errorStream)
						logStream <- v3action.NewLogMessage("Here are some staging logs!", 1, time.Now(), v3action.StagingLog, "sourceInstance")
						errorStream <- expectedErr
						allLogsWritten <- true
					}()

					return logStream, errorStream, v3action.Warnings{"steve for all I care"}, nil
				}

				fakeActor.StagePackageStub = func(packageGUID string, _ string) (<-chan v3action.Droplet, <-chan v3action.Warnings, <-chan error) {
					dropletStream := make(chan v3action.Droplet)
					warningsStream := make(chan v3action.Warnings)
					errorStream := make(chan error)

					go func() {
						<-allLogsWritten
						defer close(dropletStream)
						defer close(warningsStream)
						defer close(errorStream)
						warningsStream <- v3action.Warnings{"some-warning", "some-other-warning"}
						dropletStream <- v3action.Droplet{
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

		Context("when the logging returns an error due to an API error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("something is wrong!")
				logStream := make(chan *v3action.LogMessage)
				errorStream := make(chan error)
				fakeActor.GetStreamingLogsForApplicationByNameAndSpaceReturns(logStream, errorStream, v3action.Warnings{"some-warning", "some-other-warning"}, expectedErr)
			})

			It("returns the error and displays warnings", func() {
				Expect(executeErr).To(Equal(expectedErr))

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Err).To(Say("some-other-warning"))
			})
		})
	})
})
