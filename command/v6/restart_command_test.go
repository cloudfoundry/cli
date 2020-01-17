package v6_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
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

var _ = Describe("Restart Command", func() {
	var (
		cmd                         RestartCommand
		testUI                      *ui.UI
		fakeConfig                  *commandfakes.FakeConfig
		fakeSharedActor             *commandfakes.FakeSharedActor
		fakeApplicationSummaryActor *sharedfakes.FakeApplicationSummaryActor
		fakeActor                   *v6fakes.FakeRestartActor
		binaryName                  string
		executeErr                  error
	allLogsWritten chan bool
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeRestartActor)
		fakeApplicationSummaryActor = new(sharedfakes.FakeApplicationSummaryActor)

		cmd = RestartCommand{
			UI:                      testUI,
			Config:                  fakeConfig,
			SharedActor:             fakeSharedActor,
			Actor:                   fakeActor,
			ApplicationSummaryActor: fakeApplicationSummaryActor,
		}

		cmd.RequiredArgs.AppName = "some-app"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		var err error
		testUI.TimezoneLocation, err = time.LoadLocation("America/Los_Angeles")
		Expect(err).NotTo(HaveOccurred())

		fakeActor.RestartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
			appState := make(chan v2action.ApplicationStateChange)
			warnings := make(chan string)
			errs := make(chan error)

			go func() {
				<-allLogsWritten
				appState <- v2action.ApplicationStateStopping
				appState <- v2action.ApplicationStateStaging
				appState <- v2action.ApplicationStateStarting
				close(appState)
				close(warnings)
				close(errs)
			}()

			return appState, warnings, errs
		}
		allLogsWritten, fakeActor.GetStreamingLogsStub = GetStreamingLogsStub([]string{}, []string{})
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
			Expect(testUI.Out).To(Say("Restarting app some-app in org some-org / space some-space as some-user..."))
		})

		When("the app exists", func() {
			When("the app is started", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v2action.Application{State: constant.ApplicationStarted},
						v2action.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("stops and starts the app", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).Should(Say("Stopping app..."))
					Expect(testUI.Out).Should(Say("Staging app and tracing logs..."))
					Expect(testUI.Out).Should(Say("Waiting for app to start..."))
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					Expect(fakeActor.RestartApplicationCallCount()).To(Equal(1))
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

					Expect(fakeActor.RestartApplicationCallCount()).To(Equal(1))
					app := fakeActor.RestartApplicationArgsForCall(0)
					Expect(app.GUID).To(Equal("app-guid"))
				})

				When("passed an appStarting message", func() {
					BeforeEach(func() {
						allLogsWritten, fakeActor.GetStreamingLogsStub = GetStreamingLogsStub([]string{"log message 1", "log message 2"}, []string{})
						fakeActor.RestartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
							appState := make(chan v2action.ApplicationStateChange)
							warnings := make(chan string)
							errs := make(chan error)

							go func() {
								<- allLogsWritten
								appState <- v2action.ApplicationStateStopping
								appState <- v2action.ApplicationStateStaging
								appState <- v2action.ApplicationStateStarting
								close(appState)
								close(warnings)
								close(errs)
							}()

							return appState, warnings, errs
						}
					})

					It("displays the streaming logs", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("log message 1"))
						Expect(testUI.Out).To(Say("log message 2"))
					})
					It("displays the application stage-change logs", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Waiting for app to start..."))
					})
				})

				When("passed a warning", func() {
					Context("while logcache is still logging", func() {
						BeforeEach(func() {
							fakeActor.RestartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
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

					Context("while logcache is no longer logging", func() {
						BeforeEach(func() {
							fakeActor.RestartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
								appState := make(chan v2action.ApplicationStateChange)
								warnings := make(chan string)
								errs := make(chan error)

								go func() {
									<- allLogsWritten
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
						fakeActor.RestartApplicationStub = func(app v2action.Application) (<-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
							appState := make(chan v2action.ApplicationStateChange)
							warnings := make(chan string)
							errs := make(chan error)

							go func() {
								<- allLogsWritten
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
							apiErr = actionerror.StagingTimeoutError{AppName: "some-app", Timeout: time.Nanosecond}
						})

						It("stops logging and returns StagingTimeoutError", func() {
							Expect(executeErr).To(MatchError(translatableerror.StagingTimeoutError{AppName: "some-app", Timeout: time.Nanosecond}))
						})
					})

					When("the app instance crashes", func() {
						BeforeEach(func() {
							apiErr = actionerror.ApplicationInstanceCrashedError{Name: "some-app"}
						})

						It("stops logging and returns ApplicationUnableToStartError", func() {
							Expect(executeErr).To(MatchError(translatableerror.ApplicationUnableToStartError{AppName: "some-app", BinaryName: "faceman"}))
						})
					})

					When("the app instance flaps", func() {
						BeforeEach(func() {
							apiErr = actionerror.ApplicationInstanceFlappingError{Name: "some-app"}
						})

						It("stops logging and returns ApplicationUnableToStartError", func() {
							Expect(executeErr).To(MatchError(translatableerror.ApplicationUnableToStartError{AppName: "some-app", BinaryName: "faceman"}))
						})
					})

					Context("starting timeout", func() {
						BeforeEach(func() {
							apiErr = actionerror.StartupTimeoutError{Name: "some-app"}
						})

						It("stops logging and returns StartupTimeoutError", func() {
							Expect(executeErr).To(MatchError(translatableerror.StartupTimeoutError{AppName: "some-app", BinaryName: "faceman"}))
						})
					})
				})

				When("the app finishes starting", func() {
					BeforeEach(func() {
						fakeApplicationSummaryActor.GetApplicationSummaryByNameAndSpaceReturns(
							v2v3action.ApplicationSummary{
								ApplicationSummary: v3action.ApplicationSummary{
									Application: v3action.Application{
										Name: "some-app",
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
								},
							},
							v2v3action.Warnings{"combo-summary-warning"},
							nil)
					})

					It("displays process information", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`name:\s+%s`, "some-app"))
						Expect(testUI.Out).To(Say(`type:\s+aba`))
						Expect(testUI.Out).To(Say(`instances:\s+0/0`))
						Expect(testUI.Out).To(Say(`memory usage:\s+32M`))
						Expect(testUI.Out).To(Say(`start command:\s+some-command-1`))
						Expect(testUI.Out).To(Say(`type:\s+console`))
						Expect(testUI.Out).To(Say(`instances:\s+0/0`))
						Expect(testUI.Out).To(Say(`memory usage:\s+16M`))
						Expect(testUI.Out).To(Say(`start command:\s+some-command-2`))

						Expect(testUI.Err).To(Say("combo-summary-warning"))

						Expect(fakeApplicationSummaryActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
						passedAppName, spaceGUID, withObfuscatedValues := fakeApplicationSummaryActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
						Expect(passedAppName).To(Equal("some-app"))
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(withObfuscatedValues).To(BeTrue())
					})
				})
			})
		})

		When("the app does *not* exists", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v2action.Application{},
					v2action.Warnings{"warning-1", "warning-2"},
					actionerror.ApplicationNotFoundError{Name: "some-app"},
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
