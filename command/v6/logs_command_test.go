package v6_test

import (
	"code.cloudfoundry.org/cli/actor/loggingaction"
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
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
		fakeActor       *v6fakes.FakeLogsActor
		logCacheClient  *logcache.Client
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeLogsActor)
		logCacheClient = new(logcache.Client)

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
						nil,
						v2action.Warnings{"some-warning-1", "some-warning-2"},
						expectedErr)
				})

				It("displays the error", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})

			When("the logs actor returns logs", func() {
				BeforeEach(func() {
					fakeActor.GetRecentLogsForApplicationByNameAndSpaceReturns(
						[]loggingaction.LogMessage{
							{
								Message:        "i am message 1",
								MessageType:    "OUT",
								Timestamp:      time.Unix(0, 0),
								SourceType:     "app",
								SourceInstance: "1",
							},
							{
								Message:        "i am message 2",
								MessageType:    "OUT",
								Timestamp:      time.Unix(1, 0),
								SourceType:     "another-app",
								SourceInstance: "2",
							},
						},
						v2action.Warnings{"some-warning-1", "some-warning-2"},
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
			})

			When("the logs setup returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeActor.GetStreamingLogsForApplicationByNameAndSpaceReturns(nil, nil, v2action.Warnings{"some-warning-1", "some-warning-2"}, expectedErr, func() {})
				})

				It("displays the error and all warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})

			When("the logs stream returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-error")

					fakeActor.GetStreamingLogsForApplicationByNameAndSpaceStub = func(_ string, _ string, _ loggingaction.LogCacheClient) (<-chan loggingaction.LogMessage, <-chan error, v2action.Warnings, error, context.CancelFunc) {
						messages := make(chan loggingaction.LogMessage)
						logErrs := make(chan error)

						go func() {
							logErrs <- expectedErr
							close(messages)
							close(logErrs)
						}()

						return messages, logErrs, v2action.Warnings{"some-warning-1", "some-warning-2"}, nil, func() {}
					}
				})

				It("displays the error and all warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})

			When("the logs actor returns logs", func() {
				BeforeEach(func() {
					fakeActor.GetStreamingLogsForApplicationByNameAndSpaceStub = func(_ string, _ string, _ loggingaction.LogCacheClient) (<-chan loggingaction.LogMessage, <-chan error, v2action.Warnings, error, context.CancelFunc) {
						messages := make(chan loggingaction.LogMessage)
						logErrs := make(chan error)
						message1 := loggingaction.LogMessage{
							Message:        "i am message 1",
							MessageType:    "OUT",
							Timestamp:      time.Unix(0, 0),
							SourceType:     "app",
							SourceInstance: "1",
						}
						message2 := loggingaction.LogMessage{
							Message:        "i am message 2",
							MessageType:    "OUT",
							Timestamp:      time.Unix(1, 0),
							SourceType:     "another-app",
							SourceInstance: "2",
						}

						go func() {
							messages <- message1
							messages <- message2
							close(messages)
							close(logErrs)
						}()

						return messages, logErrs, v2action.Warnings{"some-warning-1", "some-warning-2"}, nil, func() {}
					}
				})

				It("displays flavor text", func() {
					Expect(testUI.Out).To(Say("Retrieving logs for app some-app in org some-org-name / space some-space-name as some-user..."))
				})

				It("displays all streaming log messages and warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))

					Expect(testUI.Out).To(Say("i am message 1"))
					Expect(testUI.Out).To(Say("i am message 2"))

					Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID, client := fakeActor.GetStreamingLogsForApplicationByNameAndSpaceArgsForCall(0)

					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(client).To(Equal(logCacheClient))
				})
			})
		})
	})
})
