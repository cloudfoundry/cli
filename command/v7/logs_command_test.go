package v7_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("logs command", func() {
	var (
		cmd             LogsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeLogsActor
		logCacheClient  *sharedactionfakes.FakeLogCacheClient
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeLogsActor)
		logCacheClient = new(sharedactionfakes.FakeLogCacheClient)

		cmd = LogsCommand{
			UI:             testUI,
			Config:         fakeConfig,
			SharedActor:    fakeSharedActor,
			Actor:          fakeActor,
			LogCacheClient: logCacheClient,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		cmd.RequiredArgs.AppName = "some-app"
		fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the checkTarget fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(
				actionerror.NotLoggedInError{BinaryName: binaryName})
		})
		It("returns an error", func() {
			orgRequired, spaceRequired := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(orgRequired).To(BeTrue())
			Expect(spaceRequired).To(BeTrue())

			Expect(executeErr).To(MatchError(
				actionerror.NotLoggedInError{BinaryName: binaryName}))
		})
	})

	When("checkTarget succeeds", func() {
		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space-name",
				GUID: "some-space-guid",
			})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org-name",
			})
		})

		When("the --recent flag is provided", func() {
			BeforeEach(func() {
				cmd.Recent = true
			})

			It("displays flavor text", func() {
				Expect(testUI.Out).To(Say("Retrieving logs for app some-app in org some-org-name / space some-space-name as some-user..."))
			})

			When("the logs actor returns an error", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeActor.GetRecentLogsForApplicationByNameAndSpaceReturns(
						[]sharedaction.LogMessage{
							*sharedaction.NewLogMessage(
								"all your base are belong to us",
								"1",
								time.Unix(0, 0),
								"app",
								"1",
							),
						},
						v7action.Warnings{"some-warning-1", "some-warning-2"},
						expectedErr)
				})

				It("displays the errors along with the logs and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(testUI.Out).To(Say("all your base are belong to us"))
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})

			When("the logs actor returns logs", func() {
				BeforeEach(func() {
					fakeActor.GetRecentLogsForApplicationByNameAndSpaceReturns(
						[]sharedaction.LogMessage{
							*sharedaction.NewLogMessage(
								"i am message 1",
								"1",
								time.Unix(0, 0),
								"app",
								"1",
							),
							*sharedaction.NewLogMessage(
								"i am message 2",
								"1",
								time.Unix(1, 0),
								"another-app",
								"2",
							),
						},
						v7action.Warnings{"some-warning-1", "some-warning-2"},
						nil)
				})

				It("displays the recent log messages and warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))

					Expect(testUI.Out).To(Say("i am message 1"))
					Expect(testUI.Out).To(Say("i am message 2"))

					Expect(fakeActor.GetRecentLogsForApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID, client := fakeActor.GetRecentLogsForApplicationByNameAndSpaceArgsForCall(0)

					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(client).To(Equal(logCacheClient))
				})
			})
		})

		When("the --recent flag is not provided", func() {
			BeforeEach(func() {
				cmd.Recent = false
				fakeActor.ScheduleTokenRefreshStub = func(
					after func(time.Duration) <-chan time.Time,
					stop chan struct{}, stoppedRefreshing chan struct{}) (<-chan error, error) {
					errCh := make(chan error, 1)
					go func() {
						<-stop
						close(stoppedRefreshing)
					}()
					return errCh, nil
				}
			})

			When("the logs setup returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeActor.GetStreamingLogsForApplicationByNameAndSpaceReturns(nil,
						nil,
						nil,
						v7action.Warnings{"some-warning-1",
							"some-warning-2"},
						expectedErr)
				})

				It("displays the error and all warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})

			When("the logs stream returns an error", func() {
				var (
					expectedErr                 error
					cancelFunctionHasBeenCalled bool
				)

				BeforeEach(func() {
					expectedErr = errors.New("banana")

					fakeActor.GetStreamingLogsForApplicationByNameAndSpaceStub =
						func(appName string, spaceGUID string, client sharedaction.LogCacheClient) (
							<-chan sharedaction.LogMessage,
							<-chan error,
							context.CancelFunc,
							v7action.Warnings, error) {
							logStream := make(chan sharedaction.LogMessage)
							errorStream := make(chan error)
							cancelFunctionHasBeenCalled = false

							cancelFunc := func() {
								if cancelFunctionHasBeenCalled {
									return
								}
								cancelFunctionHasBeenCalled = true
								close(logStream)
								close(errorStream)
							}
							go func() {
								errorStream <- expectedErr
							}()

							return logStream, errorStream, cancelFunc, v7action.Warnings{"steve for all I care"}, nil
						}
				})

				When("the token refresher returns an error", func() {
					BeforeEach(func() {
						cmd.Recent = false
						fakeActor.ScheduleTokenRefreshReturns(nil, errors.New("firs swimming"))
					})
					It("displays the errors", func() {
						Expect(executeErr).To(MatchError("firs swimming"))
						Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(0))
					})
				})

				It("displays the error and all warnings", func() {
					Expect(executeErr).To(MatchError("banana"))
					Expect(testUI.Err).To(Say("steve for all I care"))
					Expect(cancelFunctionHasBeenCalled).To(BeTrue())
				})
			})

			When("the logs actor returns logs", func() {
				BeforeEach(func() {
					fakeActor.GetStreamingLogsForApplicationByNameAndSpaceStub =
						func(_ string, _ string, _ sharedaction.LogCacheClient) (
							<-chan sharedaction.LogMessage,
							<-chan error, context.CancelFunc,
							v7action.Warnings,
							error) {

							logStream := make(chan sharedaction.LogMessage)
							errorStream := make(chan error)

							go func() {
								logStream <- *sharedaction.NewLogMessage("Here are some staging logs!", "OUT", time.Now(), sharedaction.StagingLog, "sourceInstance") //TODO: is it ok to leave staging logs here?
								logStream <- *sharedaction.NewLogMessage("Here are some other staging logs!", "OUT", time.Now(), sharedaction.StagingLog, "sourceInstance")
								close(logStream)
								close(errorStream)
							}()

							return logStream, errorStream, func() {}, v7action.Warnings{"some-warning-1", "some-warning-2"}, nil
						}
				})

				It("displays flavor text", func() {
					Expect(testUI.Out).To(Say("Retrieving logs for app some-app in org some-org-name / space some-space-name as some-user..."))
				})

				It("displays all streaming log messages and warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))

					Expect(testUI.Out).To(Say("Here are some staging logs!"))
					Expect(testUI.Out).To(Say("Here are some other staging logs!"))

					Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID, client := fakeActor.GetStreamingLogsForApplicationByNameAndSpaceArgsForCall(0)

					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(client).To(Equal(logCacheClient))
				})

				When("scheduling a token refresh errors immediately", func() {
					BeforeEach(func() {
						cmd.Recent = false
						fakeActor.ScheduleTokenRefreshReturns(nil, errors.New("fjords pining"))
					})
					It("displays the errors", func() {
						Expect(executeErr).To(MatchError("fjords pining"))
						Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(0))
					})
				})

				When("there is an error refreshing a token sometime later", func() {
					BeforeEach(func() {
						cmd.Recent = false
						fakeActor.ScheduleTokenRefreshStub = func(
							after func(time.Duration) <-chan time.Time,
							stop chan struct{}, stoppedRefreshing chan struct{}) (<-chan error, error) {
							errCh := make(chan error, 1)
							go func() {
								errCh <- errors.New("fjords pining")
								<-stop
								close(stoppedRefreshing)
							}()
							return errCh, nil
						}
					})
					It("displays the errors", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(1))
						Expect(testUI.Err).To(Say("fjords pining"))
					})
				})
			})
		})
	})
})
