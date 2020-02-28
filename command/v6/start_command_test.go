package v6_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/shared/sharedfakes"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

// Used for command/v6 tests that need a `GetStreamingLogs' stub
func GetStreamingLogsStub(logMessages []string, errorStrings []string) (chan bool, func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc)) {
	var closedTheStreams bool
	allLogsWritten := make(chan bool)
	return allLogsWritten, func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc) {

		outgoingLogStream := make(chan sharedaction.LogMessage, 1)
		outgoingErrStream := make(chan error, 1)
		cancelFunc := func() {
			if closedTheStreams {
				return
			}
			closedTheStreams = true
			close(outgoingLogStream)
			close(outgoingErrStream)
		}

		closedTheStreams = false
		go func() {
			for _, s := range logMessages {
				outgoingLogStream <- *sharedaction.NewLogMessage(s, "OUT", time.Now(), sharedaction.StagingLog, "source-instance")
			}
			for _, s := range errorStrings {
				outgoingErrStream <- errors.New(s)
			}
			allLogsWritten <- true
		}()
		return outgoingLogStream, outgoingErrStream, cancelFunc
	}
}

var _ = Describe("Start Command", func() {
	var (
		cmd                         StartCommand
		testUI                      *ui.UI
		fakeConfig                  *commandfakes.FakeConfig
		fakeSharedActor             *commandfakes.FakeSharedActor
		fakeActor                   *v6fakes.FakeStartActor
		fakeApplicationSummaryActor *sharedfakes.FakeApplicationSummaryActor
		binaryName                  string
		appName                     string
		executeErr                  error
		allLogsWritten              chan bool
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeStartActor)
		fakeApplicationSummaryActor = new(sharedfakes.FakeApplicationSummaryActor)

		cmd = StartCommand{
			UI:                      testUI,
			Config:                  fakeConfig,
			SharedActor:             fakeSharedActor,
			Actor:                   fakeActor,
			ApplicationSummaryActor: fakeApplicationSummaryActor,
		}

		appName = "some-app"
		cmd.RequiredArgs.AppName = appName

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		var err error
		testUI.TimezoneLocation, err = time.LoadLocation("America/Los_Angeles")
		Expect(err).NotTo(HaveOccurred())

		allLogsWritten, fakeActor.GetStreamingLogsStub = GetStreamingLogsStub([]string{}, []string{})
		fakeActor.StartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
			appState := make(chan v2action.ApplicationStateChange)
			warnings := make(chan string)
			errs := make(chan error)

			go func() {
				<-allLogsWritten
				close(appState)
				close(warnings)
				close(errs)
			}()

			return appState, warnings, errs
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error if the check fails", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is logged in, and org and space are targeted", func() {
		BeforeEach(func() {
			fakeConfig.HasTargetedOrganizationReturns(true)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
			fakeConfig.HasTargetedSpaceReturns(true)
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space"})
			fakeConfig.CurrentUserReturns(
				configv3.User{Name: "some-user"},
				nil)
		})

		When("getting the current user returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("getting current user error")
				fakeConfig.CurrentUserReturns(
					configv3.User{},
					expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		It("displays flavor text", func() {
			Expect(testUI.Out).To(Say("Starting app %s in org some-org / space some-space as some-user...", appName))
		})

		When("the app exists", func() {
			When("the app is already started", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v2action.Application{State: constant.ApplicationStarted},
						v2action.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("short circuits and displays message", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("App %s is already started", appName))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					Expect(fakeActor.StartApplicationCallCount()).To(Equal(0))
				})
			})

			When("the app is not already started", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v2action.Application{GUID: "app-guid", State: constant.ApplicationStopped},
						v2action.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("starts the app", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
					app := fakeActor.StartApplicationArgsForCall(0)
					Expect(app.GUID).To(Equal("app-guid"))
				})

				When("passed an ApplicationStateStarting message", func() {
					BeforeEach(func() {
						fakeActor.StartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
							appState := make(chan v2action.ApplicationStateChange)
							warnings := make(chan string)
							errs := make(chan error)

							go func() {
								<-allLogsWritten
								appState <- v2action.ApplicationStateStarting
								close(appState)
								close(warnings)
								close(errs)
							}()

							return appState, warnings, errs
						}
					})

					It("displays the log", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Waiting for app to start..."))
					})

					When("passed both an ApplicationStateStarting message and a log message", func() {
						BeforeEach(func() {
							allLogsWritten, fakeActor.GetStreamingLogsStub = GetStreamingLogsStub([]string{"log message 1", "log message 2"}, []string{})
						})
						It("displays the log messages first", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).To(Say("log message 1"))
							Expect(testUI.Out).To(Say("log message 2"))
							Expect(testUI.Out).To(Say("Waiting for app to start..."))

						})
					})
				})

				When("passed a log message", func() {
					BeforeEach(func() {
						allLogsWritten, fakeActor.GetStreamingLogsStub = GetStreamingLogsStub([]string{"log message 1", "log message 2"}, []string{})
						fakeActor.StartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
							appState := make(chan v2action.ApplicationStateChange)
							warnings := make(chan string)
							errs := make(chan error)

							go func() {
								<-allLogsWritten
								close(appState)
								close(warnings)
								close(errs)
							}()

							return appState, warnings, errs
						}
					})

					It("displays the log", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("log message 1"))
						Expect(testUI.Out).To(Say("log message 2"))
						Expect(testUI.Out).ToNot(Say("log message 3"))
					})
				})

				When("passed an log err", func() {
					Context("NOAA connection times out/closes", func() {
						BeforeEach(func() {
							allLogsWritten, fakeActor.GetStreamingLogsStub = GetStreamingLogsStub([]string{"message 1", "message 2", "message 3"},
								[]string{"timeout connecting to log server, no log will be shown"},
							)
							fakeActor.StartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
								appState := make(chan v2action.ApplicationStateChange)
								warnings := make(chan string)
								errs := make(chan error)

								go func() {
									<-allLogsWritten
									close(appState)
									close(warnings)
									close(errs)
								}()

								return appState, warnings, errs
							}
							v3ApplicationSummary := v3action.ApplicationSummary{
								Application: v3action.Application{
									Name: appName,
								},
								ProcessSummaries: v3action.ProcessSummaries{
									{
										Process: v3action.Process{
											Type:       "aba",
											Command:    *types.NewFilteredString("some-command-1"),
											MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
											DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
										},
									},
									{
										Process: v3action.Process{
											Type:       "console",
											Command:    *types.NewFilteredString("some-command-2"),
											MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
											DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
										},
									},
								},
							}

							applicationSummary := v2v3action.ApplicationSummary{
								ApplicationSummary: v3ApplicationSummary,
							}

							warnings := []string{"app-summary-warning"}

							fakeApplicationSummaryActor.GetApplicationSummaryByNameAndSpaceReturns(applicationSummary, warnings, nil)
						})

						It("displays a warning and continues until app has started", func() {
							Expect(executeErr).To(BeNil())
							Expect(testUI.Out).To(Say("message 1"))
							Expect(testUI.Out).To(Say("message 2"))
							Expect(testUI.Out).To(Say("message 3"))
							Expect(testUI.Err).To(Say("timeout connecting to log server, no log will be shown"))
							Expect(testUI.Out).To(Say(`name:\s+%s`, appName))
						})
					})

					Context("an unexpected error occurs", func() {

						BeforeEach(func() {
							allLogsWritten, fakeActor.GetStreamingLogsStub = GetStreamingLogsStub([]string{}, []string{"unexpected status code 404"})

							fakeActor.StartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
								appState := make(chan v2action.ApplicationStateChange)
								warnings := make(chan string)
								errs := make(chan error)

								go func() {
									<-allLogsWritten
									close(appState)
									close(warnings)
									close(errs)
								}()

								return appState, warnings, errs
							}
						})

						It("displays the error and continues to poll", func() {
							Expect(executeErr).NotTo(HaveOccurred())
							Expect(testUI.Err).To(Say("Failed to retrieve logs from Log Cache: unexpected status code 404"))
						})
					})
				})

				When("passed a warning", func() {
					Context("while NOAA is still logging", func() {
						BeforeEach(func() {
							fakeActor.StartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
								appState := make(chan v2action.ApplicationStateChange)
								warnings := make(chan string)
								errs := make(chan error)

								go func() {
									<-allLogsWritten
									warnings <- "warning 1"
									warnings <- "warning 2"
									close(appState)
									close(warnings)
									close(errs)
								}()

								return appState, warnings, errs
							}
						})

						It("displays the warnings to STDERR", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Err).To(Say("warning 1"))
							Expect(testUI.Err).To(Say("warning 2"))
						})
					})

					Context("while NOAA is no longer logging", func() {
						BeforeEach(func() {
							fakeActor.StartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
								appState := make(chan v2action.ApplicationStateChange)
								warnings := make(chan string)
								errs := make(chan error)

								go func() {
									<-allLogsWritten
									warnings <- "warning 1"
									warnings <- "warning 2"
									warnings <- "warning 3"
									warnings <- "warning 4"
									close(appState)
									close(warnings)
									close(errs)
								}()

								return appState, warnings, errs
							}
						})

						It("displays the warnings to STDERR", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Err).To(Say("warning 1"))
							Expect(testUI.Err).To(Say("warning 2"))
							Expect(testUI.Err).To(Say("warning 3"))
							Expect(testUI.Err).To(Say("warning 4"))
						})
					})
				})

				When("passed an API err", func() {
					var apiErr error

					BeforeEach(func() {
						fakeActor.StartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
							appState := make(chan v2action.ApplicationStateChange)
							warnings := make(chan string)
							errs := make(chan error)

							go func() {
								<-allLogsWritten
								errs <- apiErr
								close(appState)
								close(warnings)
								close(errs)
							}()

							return appState, warnings, errs
						}
					})

					Context("an unexpected error", func() {
						BeforeEach(func() {
							apiErr = errors.New("err log message")
						})

						It("stops logging and returns the error", func() {
							Expect(executeErr).To(MatchError(apiErr))
						})
					})

					Context("staging failed", func() {
						BeforeEach(func() {
							apiErr = actionerror.StagingFailedError{Reason: "Something, but not nothing"}
						})

						It("stops logging and returns StagingFailedError", func() {
							Expect(executeErr).To(MatchError(translatableerror.StagingFailedError{Message: "Something, but not nothing"}))
						})
					})

					Context("staging timed out", func() {
						BeforeEach(func() {
							apiErr = actionerror.StagingTimeoutError{AppName: appName, Timeout: time.Nanosecond}
						})

						It("stops logging and returns StagingTimeoutError", func() {
							Expect(executeErr).To(MatchError(translatableerror.StagingTimeoutError{AppName: appName, Timeout: time.Nanosecond}))
						})
					})

					When("the app instance crashes", func() {
						BeforeEach(func() {
							apiErr = actionerror.ApplicationInstanceCrashedError{Name: appName}
						})

						It("stops logging and returns ApplicationUnableToStartError", func() {
							Expect(executeErr).To(MatchError(translatableerror.ApplicationUnableToStartError{AppName: appName, BinaryName: "faceman"}))
						})
					})

					When("the app instance flaps", func() {
						BeforeEach(func() {
							apiErr = actionerror.ApplicationInstanceFlappingError{Name: appName}
						})

						It("stops logging and returns ApplicationUnableToStartError", func() {
							Expect(executeErr).To(MatchError(translatableerror.ApplicationUnableToStartError{AppName: appName, BinaryName: "faceman"}))
						})
					})

					Context("starting timeout", func() {
						BeforeEach(func() {
							apiErr = actionerror.StartupTimeoutError{Name: appName}
						})

						It("stops logging and returns StartupTimeoutError", func() {
							Expect(executeErr).To(MatchError(translatableerror.StartupTimeoutError{AppName: appName, BinaryName: "faceman"}))
						})
					})
				})

				When("the app finishes starting", func() {
					Describe("version-dependent display", func() {
						When("CC API >= 3.27.0", func() {
							var (
								applicationSummary v2v3action.ApplicationSummary
							)

							BeforeEach(func() {
								fakeApplicationSummaryActor.CloudControllerV3APIVersionReturns("3.50.0")
								v3ApplicationSummary := v3action.ApplicationSummary{
									Application: v3action.Application{
										Name: appName,
									},
									ProcessSummaries: v3action.ProcessSummaries{
										{
											Process: v3action.Process{
												Type:       "aba",
												Command:    *types.NewFilteredString("some-command-1"),
												MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
												DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
											},
										},
										{
											Process: v3action.Process{
												Type:       "console",
												Command:    *types.NewFilteredString("some-command-2"),
												MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
												DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
											},
										},
									},
								}
								applicationSummary = v2v3action.ApplicationSummary{
									ApplicationSummary: v3ApplicationSummary,
								}

								fakeApplicationSummaryActor.GetApplicationSummaryByNameAndSpaceReturns(
									applicationSummary,
									v2v3action.Warnings{"combo-summary-warning"},
									nil)

							})

							It("uses the multiprocess display", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(fakeApplicationSummaryActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
								passedAppName, spaceGUID, withObfuscatedValues := fakeApplicationSummaryActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
								Expect(passedAppName).To(Equal(appName))
								Expect(spaceGUID).To(Equal("some-space-guid"))
								Expect(withObfuscatedValues).To(BeTrue())

								Expect(testUI.Out).To(Say(`name:\s+%s`, appName))
								Expect(testUI.Out).To(Say(`type:\s+aba`))
								Expect(testUI.Out).To(Say(`instances:\s+0/0`))
								Expect(testUI.Out).To(Say(`memory usage:\s+32M`))
								Expect(testUI.Out).To(Say(`start command:\s+some-command-1`))
								Expect(testUI.Out).To(Say(`type:\s+console`))
								Expect(testUI.Out).To(Say(`instances:\s+0/0`))
								Expect(testUI.Out).To(Say(`memory usage:\s+16M`))
								Expect(testUI.Out).To(Say(`start command:\s+some-command-2`))

								Expect(testUI.Err).To(Say("combo-summary-warning"))
							})
						})
					})
				})
			})
		})

		When("the app does *not* exists", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v2action.Application{},
					v2action.Warnings{"warning-1", "warning-2"},
					actionerror.ApplicationNotFoundError{Name: appName},
				)
			})

			It("returns back an error", func() {
				Expect(executeErr).To(MatchError(actionerror.ApplicationNotFoundError{Name: "some-app"}))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
			})
		})
	})
})
