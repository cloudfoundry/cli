package v7_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
)

type Step struct {
	Error    error
	Event    v7pushaction.Event
	Warnings v7pushaction.Warnings
}

func FillInValues(tuples []Step, state v7pushaction.PushState) func(v7pushaction.PushState, v7pushaction.ProgressBar) (<-chan v7pushaction.PushState, <-chan v7pushaction.Event, <-chan v7pushaction.Warnings, <-chan error) {
	return func(v7pushaction.PushState, v7pushaction.ProgressBar) (<-chan v7pushaction.PushState, <-chan v7pushaction.Event, <-chan v7pushaction.Warnings, <-chan error) {
		stateStream := make(chan v7pushaction.PushState)

		eventStream := make(chan v7pushaction.Event)
		warningsStream := make(chan v7pushaction.Warnings)
		errorStream := make(chan error)

		go func() {
			defer close(stateStream)
			defer close(eventStream)
			defer close(warningsStream)
			defer close(errorStream)

			for _, tuple := range tuples {
				warningsStream <- tuple.Warnings
				if tuple.Error != nil {
					errorStream <- tuple.Error
					return
				} else {
					eventStream <- tuple.Event
				}
			}

			stateStream <- state
			eventStream <- v7pushaction.Complete
		}()

		return stateStream, eventStream, warningsStream, errorStream
	}
}

type LogEvent struct {
	Log   *v7action.LogMessage
	Error error
}

func ReturnLogs(logevents []LogEvent, passedWarnings v7action.Warnings, passedError error) func(appName string, spaceGUID string, client v7action.NOAAClient) (<-chan *v7action.LogMessage, <-chan error, v7action.Warnings, error) {
	return func(appName string, spaceGUID string, client v7action.NOAAClient) (<-chan *v7action.LogMessage, <-chan error, v7action.Warnings, error) {
		logStream := make(chan *v7action.LogMessage)
		errStream := make(chan error)
		go func() {
			defer close(logStream)
			defer close(errStream)

			for _, log := range logevents {
				if log.Log != nil {
					logStream <- log.Log
				}
				if log.Error != nil {
					errStream <- log.Error
				}
			}
		}()

		return logStream, errStream, passedWarnings, passedError
	}
}

var _ = Describe("push Command", func() {
	var (
		cmd              PushCommand
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v7fakes.FakePushActor
		fakeVersionActor *v7fakes.FakeV7ActorForPush
		fakeProgressBar  *v6fakes.FakeProgressBar
		fakeNOAAClient   *v7actionfakes.FakeNOAAClient
		binaryName       string
		executeErr       error

		appName   string
		userName  string
		spaceName string
		orgName   string
		pwd       string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakePushActor)
		fakeVersionActor = new(v7fakes.FakeV7ActorForPush)
		fakeProgressBar = new(v6fakes.FakeProgressBar)
		fakeNOAAClient = new(v7actionfakes.FakeNOAAClient)

		appName = "some-app"
		userName = "some-user"
		spaceName = "some-space"
		orgName = "some-org"
		pwd = "/push/cmd/test"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.ExperimentalReturns(true) // TODO: Delete once we remove the experimental flag

		cmd = PushCommand{
			RequiredArgs: flag.AppName{AppName: "some-app"},
			UI:           testUI,
			Config:       fakeConfig,
			Actor:        fakeActor,
			VersionActor: fakeVersionActor,
			SharedActor:  fakeSharedActor,
			ProgressBar:  fakeProgressBar,
			NOAAClient:   fakeNOAAClient,
			PWD:          pwd,
		}
	})

	Describe("Execute", func() {
		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
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

		When("checking target fails because the user is not logged in", func() {
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

		When("the user is logged in, and org and space are targeted", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: userName}, nil)

				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					Name: orgName,
					GUID: "some-org-guid",
				})
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					Name: spaceName,
					GUID: "some-space-guid",
				})
			})

			It("displays the experimental warning", func() {
				Expect(testUI.Err).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
			})

			When("getting app settings is successful", func() {
				BeforeEach(func() {
					fakeActor.ConceptualizeReturns(
						[]v7pushaction.PushState{
							{
								Application: v7action.Application{Name: appName},
							},
						},
						v7pushaction.Warnings{"some-warning-1"}, nil)
				})

				Describe("actualizing non-logging events", func() {
					BeforeEach(func() {
						fakeActor.ActualizeStub = FillInValues([]Step{
							{
								Event:    v7pushaction.SkippingApplicationCreation,
								Warnings: v7pushaction.Warnings{"skipping app creation warnings"},
							},
							{
								Event:    v7pushaction.CreatedApplication,
								Warnings: v7pushaction.Warnings{"app creation warnings"},
							},
							{
								Event: v7pushaction.CreatingAndMappingRoutes,
							},
							{
								Event:    v7pushaction.CreatedRoutes,
								Warnings: v7pushaction.Warnings{"routes warnings"},
							},
							{
								Event: v7pushaction.CreatingArchive,
							},
							{
								Event:    v7pushaction.UploadingApplicationWithArchive,
								Warnings: v7pushaction.Warnings{"upload app archive warning"},
							},
							{
								Event:    v7pushaction.RetryUpload,
								Warnings: v7pushaction.Warnings{"retry upload warning"},
							},
							{
								Event: v7pushaction.UploadWithArchiveComplete,
							},
							{
								Event: v7pushaction.StagingComplete,
							},
						}, v7pushaction.PushState{})
					})

					It("generates a push state with the specified app path", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Pushing app %s to org some-org / space some-space as some-user", appName))
						Expect(testUI.Out).To(Say(`Getting app info\.\.\.`))
						Expect(testUI.Err).To(Say("some-warning-1"))

						Expect(fakeActor.ConceptualizeCallCount()).To(Equal(1))
						name, spaceGUID, orgGUID, currentDirectory, _ := fakeActor.ConceptualizeArgsForCall(0)
						Expect(name).To(Equal(appName))
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(orgGUID).To(Equal("some-org-guid"))
						Expect(currentDirectory).To(Equal(pwd))
					})

					It("actualizes the application and displays events/warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Updating app some-app..."))
						Expect(testUI.Err).To(Say("skipping app creation warnings"))

						Expect(testUI.Out).To(Say("Creating app some-app..."))
						Expect(testUI.Err).To(Say("app creation warnings"))

						Expect(testUI.Out).To(Say("Mapping routes..."))
						Expect(testUI.Err).To(Say("routes warnings"))

						Expect(testUI.Out).To(Say("Packaging files to upload..."))

						Expect(testUI.Out).To(Say("Uploading files..."))
						Expect(testUI.Err).To(Say("upload app archive warning"))
						Expect(fakeProgressBar.ReadyCallCount()).Should(Equal(1))

						Expect(testUI.Out).To(Say("Retrying upload due to an error..."))
						Expect(testUI.Err).To(Say("retry upload warning"))

						Expect(testUI.Out).To(Say("Waiting for API to complete processing files..."))

						Expect(testUI.Out).To(Say("Waiting for app to start..."))
						Expect(fakeProgressBar.CompleteCallCount()).Should(Equal(1))
					})
				})

				Describe("actualizing logging events", func() {
					BeforeEach(func() {
						fakeActor.ActualizeStub = FillInValues([]Step{
							{
								Event: v7pushaction.StartingStaging,
							},
						}, v7pushaction.PushState{})
					})

					When("there are no logging errors", func() {
						BeforeEach(func() {
							fakeVersionActor.GetStreamingLogsForApplicationByNameAndSpaceStub = ReturnLogs(
								[]LogEvent{
									{Log: v7action.NewLogMessage("log-message-1", 1, time.Now(), v7action.StagingLog, "source-instance")},
									{Log: v7action.NewLogMessage("log-message-2", 1, time.Now(), v7action.StagingLog, "source-instance")},
									{Log: v7action.NewLogMessage("log-message-3", 1, time.Now(), "potato", "source-instance")},
								},
								v7action.Warnings{"log-warning-1", "log-warning-2"},
								nil,
							)
						})

						It("displays the staging logs and warnings", func() {
							Expect(testUI.Out).To(Say("Staging app and tracing logs..."))

							Expect(testUI.Err).To(Say("log-warning-1"))
							Expect(testUI.Err).To(Say("log-warning-2"))

							Eventually(testUI.Out).Should(Say("log-message-1"))
							Eventually(testUI.Out).Should(Say("log-message-2"))
							Eventually(testUI.Out).ShouldNot(Say("log-message-3"))

							Expect(fakeVersionActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(1))
							passedAppName, spaceGUID, _ := fakeVersionActor.GetStreamingLogsForApplicationByNameAndSpaceArgsForCall(0)
							Expect(passedAppName).To(Equal(appName))
							Expect(spaceGUID).To(Equal("some-space-guid"))
						})
					})

					When("there are logging errors", func() {
						BeforeEach(func() {
							fakeVersionActor.GetStreamingLogsForApplicationByNameAndSpaceStub = ReturnLogs(
								[]LogEvent{
									{Error: errors.New("some-random-err")},
									{Error: actionerror.NOAATimeoutError{}},
									{Log: v7action.NewLogMessage("log-message-1", 1, time.Now(), v7action.StagingLog, "source-instance")},
								},
								v7action.Warnings{"log-warning-1", "log-warning-2"},
								nil,
							)
						})

						It("displays the errors as warnings", func() {
							Expect(testUI.Out).To(Say("Staging app and tracing logs..."))

							Expect(testUI.Err).To(Say("log-warning-1"))
							Expect(testUI.Err).To(Say("log-warning-2"))
							Eventually(testUI.Err).Should(Say("some-random-err"))
							Eventually(testUI.Err).Should(Say("timeout connecting to log server, no log will be shown"))

							Eventually(testUI.Out).Should(Say("log-message-1"))
						})
					})
				})

				When("the app is successfully actualized", func() {
					BeforeEach(func() {
						fakeActor.ActualizeStub = FillInValues([]Step{
							{},
						}, v7pushaction.PushState{Application: v7action.Application{GUID: "potato"}})
					})

					When("restarting the app succeeds", func() {
						BeforeEach(func() {
							fakeVersionActor.RestartApplicationReturns(v7action.Warnings{"some-restart-warning"}, nil)
						})

						It("restarts the app and displays warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeVersionActor.RestartApplicationCallCount()).To(Equal(1))
							Expect(fakeVersionActor.RestartApplicationArgsForCall(0)).To(Equal("potato"))
							Expect(testUI.Err).To(Say("some-restart-warning"))
						})

						When("polling the restart succeeds", func() {
							BeforeEach(func() {
								fakeVersionActor.PollStartStub = func(appGUID string, warnings chan<- v7action.Warnings) error {
									warnings <- v7action.Warnings{"some-poll-warning-1", "some-poll-warning-2"}
									return nil
								}

								summary := v7action.ApplicationSummary{
									Application: v7action.Application{
										Name:  appName,
										State: constant.ApplicationStarted,
									},
									CurrentDroplet: v7action.Droplet{
										Stack: "cflinuxfs2",
										Buildpacks: []v7action.Buildpack{
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
												Command: "some-command-1",
											},
										},
										{
											Process: v7action.Process{
												Type:    "console",
												Command: "some-command-2",
											},
										},
									},
								}
								fakeVersionActor.GetApplicationSummaryByNameAndSpaceReturns(summary, v7action.Warnings{"app-summary-warning-1", "app-summary-warning-2"}, nil)
							})

							It("displays all warnings", func() {
								Expect(testUI.Err).To(Say("some-poll-warning-1"))
								Expect(testUI.Err).To(Say("some-poll-warning-2"))

								Expect(executeErr).ToNot(HaveOccurred())
							})

							It("displays the app summary", func() {
								Expect(testUI.Out).To(Say(`name:\s+some-app`))
								Expect(testUI.Out).To(Say(`requested state:\s+started`))
								Expect(testUI.Out).To(Say("type:\\s+web"))
								Expect(testUI.Out).To(Say("start command:\\s+some-command-1"))
								Expect(testUI.Out).To(Say("type:\\s+console"))
								Expect(testUI.Out).To(Say("start command:\\s+some-command-2"))

								Expect(testUI.Err).To(Say("warning-1"))
								Expect(testUI.Err).To(Say("warning-2"))

								Expect(fakeVersionActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
								name, spaceGUID, withObfuscatedValues, _ := fakeVersionActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
								Expect(name).To(Equal("some-app"))
								Expect(spaceGUID).To(Equal("some-space-guid"))
								Expect(withObfuscatedValues).To(BeTrue())
							})
						})

						When("polling the start fails", func() {
							BeforeEach(func() {
								fakeVersionActor.PollStartStub = func(appGUID string, warnings chan<- v7action.Warnings) error {
									warnings <- v7action.Warnings{"some-poll-warning-1", "some-poll-warning-2"}
									return errors.New("some-error")
								}
							})

							It("displays all warnings and fails", func() {
								Expect(testUI.Err).To(Say("some-poll-warning-1"))
								Expect(testUI.Err).To(Say("some-poll-warning-2"))

								Expect(executeErr).To(MatchError("some-error"))
							})
						})

						When("polling times out", func() {
							BeforeEach(func() {
								fakeVersionActor.PollStartReturns(actionerror.StartupTimeoutError{})
							})

							It("returns the StartupTimeoutError", func() {
								Expect(executeErr).To(MatchError(translatableerror.StartupTimeoutError{
									AppName:    "some-app",
									BinaryName: binaryName,
								}))
							})
						})
					})

					When("restarting the app fails", func() {
						BeforeEach(func() {
							fakeVersionActor.RestartApplicationReturns(v7action.Warnings{"some-restart-warning"}, errors.New("restart failure"))
						})

						It("returns an error and any warnings", func() {
							Expect(executeErr).To(MatchError("restart failure"))
							Expect(testUI.Err).To(Say("some-restart-warning"))
						})
					})
				})

				When("actualizing fails", func() {
					BeforeEach(func() {
						fakeActor.ActualizeStub = FillInValues([]Step{
							{
								Error: errors.New("anti avant garde naming"),
							},
						}, v7pushaction.PushState{})
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("anti avant garde naming"))
					})
				})
			})

			When("getting app settings returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeActor.ConceptualizeReturns(nil, v7pushaction.Warnings{"some-warning-1"}, expectedErr)
				})

				It("generates a push state with the specified app path", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(testUI.Err).To(Say("some-warning-1"))
				})
			})

			When("app path is specified", func() {
				BeforeEach(func() {
					cmd.AppPath = "some/app/path"
				})

				It("generates a push state with the specified app path", func() {
					Expect(fakeActor.ConceptualizeCallCount()).To(Equal(1))
					name, spaceGUID, orgGUID, currentDirectory, overrides := fakeActor.ConceptualizeArgsForCall(0)
					Expect(name).To(Equal(appName))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(currentDirectory).To(Equal(pwd))
					Expect(overrides).To(MatchFields(IgnoreExtras, Fields{
						"ProvidedAppPath": Equal("some/app/path"),
					}))
				})
			})
		})
	})

	Describe("GetFlagOverrides", func() {
		var (
			overrides    v7pushaction.FlagOverrides
			overridesErr error
		)

		BeforeEach(func() {
			cmd.Buildpacks = []string{"buildpack-1", "buildpack-2"}
			cmd.HealthCheckType = flag.HealthCheckType{Type: "port"}
			cmd.Memory = flag.Megabytes{NullUint64: types.NullUint64{Value: 100, IsSet: true}}
			cmd.StartCommand = flag.Command{FilteredString: types.FilteredString{IsSet: true, Value: "some-start-command"}}
		})

		JustBeforeEach(func() {
			overrides, overridesErr = cmd.GetFlagOverrides()
			Expect(overridesErr).ToNot(HaveOccurred())
		})

		It("sets them on the command line settings", func() {
			Expect(overridesErr).ToNot(HaveOccurred())
			Expect(overrides.Buildpacks).To(ConsistOf("buildpack-1", "buildpack-2"))
			Expect(overrides.HealthCheckType).To(Equal("port"))
			Expect(overrides.Memory).To(Equal(types.NullUint64{Value: 100, IsSet: true}))
			Expect(overrides.StartCommand).To(Equal(types.FilteredString{IsSet: true, Value: "some-start-command"}))
		})
	})
})
