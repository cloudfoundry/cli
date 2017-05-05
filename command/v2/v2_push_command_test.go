package v2_test

import (
	"errors"
	"os"
	"time"

	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v2-push Command", func() {
	var (
		cmd             V2PushCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeV2PushActor
		fakeStartActor  *v2fakes.FakeStartActor
		fakeProgressBar *v2fakes.FakeProgressBar
		input           *Buffer
		binaryName      string

		appName    string
		executeErr error
		pwd        string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeV2PushActor)
		fakeStartActor = new(v2fakes.FakeStartActor)
		fakeProgressBar = new(v2fakes.FakeProgressBar)

		cmd = V2PushCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			StartActor:  fakeStartActor,
			ProgressBar: fakeProgressBar,
		}

		appName = "some-app"
		cmd.OptionalArgs.AppName = appName
		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		var err error
		pwd, err = os.Getwd()
		Expect(err).ToNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(command.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is logged in, and org and space are targeted", func() {
		BeforeEach(func() {
			fakeConfig.HasTargetedOrganizationReturns(true)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: "some-org-guid", Name: "some-org"})
			fakeConfig.HasTargetedSpaceReturns(true)
			fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: "some-space-guid", Name: "some-space"})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
		})

		Context("when the push settings are valid", func() {
			var appManifests []manifest.Application

			BeforeEach(func() {
				appManifests = []manifest.Application{
					{
						Name: appName,
						Path: pwd,
					},
				}
				fakeActor.MergeAndValidateSettingsAndManifestsReturns(appManifests, nil)
			})

			Context("when the settings can be converted to a valid config", func() {
				var appConfigs []pushaction.ApplicationConfig

				BeforeEach(func() {
					appConfigs = []pushaction.ApplicationConfig{
						{
							DesiredApplication: v2action.Application{Name: appName},
							TargetedSpaceGUID:  "some-space-guid",
							Path:               pwd,
						},
					}
					fakeActor.ConvertToApplicationConfigsReturns(appConfigs, pushaction.Warnings{"some-config-warnings"}, nil)
				})

				Context("when the apply is successful", func() {
					var updatedConfig pushaction.ApplicationConfig

					BeforeEach(func() {
						fakeActor.ApplyStub = func(_ pushaction.ApplicationConfig, _ pushaction.ProgressBar) (<-chan pushaction.ApplicationConfig, <-chan pushaction.Event, <-chan pushaction.Warnings, <-chan error) {
							configStream := make(chan pushaction.ApplicationConfig, 1)
							eventStream := make(chan pushaction.Event)
							warningsStream := make(chan pushaction.Warnings)
							errorStream := make(chan error)

							updatedConfig = pushaction.ApplicationConfig{
								CurrentApplication: v2action.Application{Name: appName, GUID: "some-app-guid"},
								DesiredApplication: v2action.Application{Name: appName, GUID: "some-app-guid"},
								TargetedSpaceGUID:  "some-space-guid",
								Path:               pwd,
							}

							go func() {
								defer GinkgoRecover()

								Eventually(eventStream).Should(BeSent(pushaction.ApplicationCreated))
								Eventually(eventStream).Should(BeSent(pushaction.ApplicationUpdated))
								Eventually(eventStream).Should(BeSent(pushaction.RouteCreated))
								Eventually(eventStream).Should(BeSent(pushaction.RouteBound))
								Eventually(eventStream).Should(BeSent(pushaction.UploadingApplication))
								Eventually(fakeProgressBar.ReadyCallCount).Should(Equal(1))
								Eventually(eventStream).Should(BeSent(pushaction.UploadComplete))
								Eventually(fakeProgressBar.CompleteCallCount).Should(Equal(1))
								Eventually(configStream).Should(BeSent(updatedConfig))
								Eventually(eventStream).Should(BeSent(pushaction.Complete))
								Eventually(warningsStream).Should(BeSent(pushaction.Warnings{"apply-1", "apply-2"}))
								close(configStream)
								close(eventStream)
								close(warningsStream)
								close(errorStream)
							}()

							return configStream, eventStream, warningsStream, errorStream
						}

						fakeStartActor.RestartApplicationStub = func(app v2action.Application, client v2action.NOAAClient, config v2action.Config) (<-chan *v2action.LogMessage, <-chan error, <-chan bool, <-chan string, <-chan error) {
							messages := make(chan *v2action.LogMessage)
							logErrs := make(chan error)
							appStart := make(chan bool)
							warnings := make(chan string)
							errs := make(chan error)

							go func() {
								messages <- v2action.NewLogMessage("log message 1", 1, time.Unix(0, 0), "STG", "1")
								messages <- v2action.NewLogMessage("log message 2", 1, time.Unix(0, 0), "STG", "1")
								appStart <- true
								close(messages)
								close(logErrs)
								close(appStart)
								close(warnings)
								close(errs)
							}()

							return messages, logErrs, appStart, warnings, errs
						}
						applicationSummary := v2action.ApplicationSummary{
							Application: v2action.Application{
								Name:                 appName,
								GUID:                 "some-app-guid",
								Instances:            3,
								Memory:               128,
								PackageUpdatedAt:     time.Unix(0, 0),
								DetectedBuildpack:    "some-buildpack",
								State:                "STARTED",
								DetectedStartCommand: "some start command",
							},
							Stack: v2action.Stack{
								Name: "potatos",
							},
							Routes: []v2action.Route{
								{
									Host: "banana",
									Domain: v2action.Domain{
										Name: "fruit.com",
									},
									Path: "/hi",
								},
								{
									Domain: v2action.Domain{
										Name: "foobar.com",
									},
									Port: 13,
								},
							},
						}
						warnings := []string{"app-summary-warning"}

						applicationSummary.RunningInstances = []v2action.ApplicationInstanceWithStats{{State: "RUNNING"}}

						fakeStartActor.GetApplicationSummaryByNameAndSpaceReturns(applicationSummary, warnings, nil)
					})

					It("merges app manifest and flags", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeActor.MergeAndValidateSettingsAndManifestsCallCount()).To(Equal(1))
						cmdSettings, _ := fakeActor.MergeAndValidateSettingsAndManifestsArgsForCall(0)
						Expect(cmdSettings).To(Equal(pushaction.CommandLineSettings{
							Name: appName,
							Path: pwd,
						}))
					})

					It("converts the manifests to app configs and outputs config warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Err).To(Say("some-config-warnings"))

						Expect(fakeActor.ConvertToApplicationConfigsCallCount()).To(Equal(1))
						orgGUID, spaceGUID, manifests := fakeActor.ConvertToApplicationConfigsArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(manifests).To(Equal(appManifests))
					})

					It("outputs flavor text prior to generating app configuration", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Getting app info..."))
					})

					It("applies each of the application configurations", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeActor.ApplyCallCount()).To(Equal(1))
						config, progressBar := fakeActor.ApplyArgsForCall(0)
						Expect(config).To(Equal(appConfigs[0]))
						Expect(progressBar).To(Equal(fakeProgressBar))
					})

					It("displays app events and warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Creating app %s in org %s / space %s as %s...", appName, "some-org", "some-space", "some-user"))
						Expect(testUI.Out).To(Say("Updating app %s in org %s / space %s as %s...", appName, "some-org", "some-space", "some-user"))
						Expect(testUI.Out).To(Say("Creating routes..."))
						Expect(testUI.Out).To(Say("Binding routes..."))
						Expect(testUI.Out).To(Say("Uploading application..."))
						Expect(testUI.Out).To(Say("Upload complete"))

						Expect(testUI.Err).To(Say("some-config-warnings"))
						Expect(testUI.Err).To(Say("apply-1"))
						Expect(testUI.Err).To(Say("apply-2"))
					})

					It("displays app staging logs", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("log message 1"))
						Expect(testUI.Out).To(Say("log message 2"))

						Expect(fakeStartActor.RestartApplicationCallCount()).To(Equal(1))
						appConfig, _, _ := fakeStartActor.RestartApplicationArgsForCall(0)
						Expect(appConfig).To(Equal(updatedConfig.CurrentApplication))
					})

					It("displays the app summary with isolation segments as well as warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("name:\\s+%s", appName))
						Expect(testUI.Out).To(Say("requested state:\\s+started"))
						Expect(testUI.Out).To(Say("instances:\\s+1\\/3"))
						Expect(testUI.Out).To(Say("usage:\\s+128M x 3 instances"))
						Expect(testUI.Out).To(Say("routes:\\s+banana.fruit.com/hi, foobar.com:13"))
						Expect(testUI.Out).To(Say("last uploaded:\\s+\\w{3} [0-3]\\d \\w{3} [0-2]\\d:[0-5]\\d:[0-5]\\d \\w+ \\d{4}"))
						Expect(testUI.Out).To(Say("stack:\\s+potatos"))
						Expect(testUI.Out).To(Say("buildpack:\\s+some-buildpack"))
						Expect(testUI.Out).To(Say("start command:\\s+some start command"))

						Expect(testUI.Err).To(Say("app-summary-warning"))
					})
				})

				Context("when the apply errors", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("no wayz dude")
						fakeActor.ApplyStub = func(_ pushaction.ApplicationConfig, _ pushaction.ProgressBar) (<-chan pushaction.ApplicationConfig, <-chan pushaction.Event, <-chan pushaction.Warnings, <-chan error) {
							configStream := make(chan pushaction.ApplicationConfig)
							eventStream := make(chan pushaction.Event)
							warningsStream := make(chan pushaction.Warnings)
							errorStream := make(chan error)

							go func() {
								defer GinkgoRecover()

								Eventually(warningsStream).Should(BeSent(pushaction.Warnings{"apply-1", "apply-2"}))
								Eventually(errorStream).Should(BeSent(expectedErr))
								close(configStream)
								close(eventStream)
								close(warningsStream)
								close(errorStream)
							}()

							return configStream, eventStream, warningsStream, errorStream
						}
					})

					It("outputs the warnings and returns the error", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Err).To(Say("some-config-warnings"))
						Expect(testUI.Err).To(Say("apply-1"))
						Expect(testUI.Err).To(Say("apply-2"))
					})
				})
			})

			Context("when there is an error converting the app setting into a config", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("no wayz dude")
					fakeActor.ConvertToApplicationConfigsReturns(nil, pushaction.Warnings{"some-config-warnings"}, expectedErr)
				})

				It("outputs the warnings and returns the error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Err).To(Say("some-config-warnings"))
				})
			})
		})

		Context("when the push settings are invalid", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("no wayz dude")
				fakeActor.MergeAndValidateSettingsAndManifestsReturns(nil, expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})
	})
})
