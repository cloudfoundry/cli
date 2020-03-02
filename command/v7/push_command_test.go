package v7_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	"code.cloudfoundry.org/cli/actor/v7action"
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
	"code.cloudfoundry.org/cli/util/manifestparser"
	"code.cloudfoundry.org/cli/util/ui"
	"github.com/cloudfoundry/bosh-cli/director/template"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

type Step struct {
	Plan     v7pushaction.PushPlan
	Error    error
	Event    v7pushaction.Event
	Warnings v7pushaction.Warnings
}

func FillInEvents(steps []Step) <-chan *v7pushaction.PushEvent {
	eventStream := make(chan *v7pushaction.PushEvent)

	go func() {
		defer close(eventStream)

		for _, step := range steps {
			eventStream <- &v7pushaction.PushEvent{Plan: step.Plan, Warnings: step.Warnings, Err: step.Error, Event: step.Event}
		}
	}()

	return eventStream
}

type LogEvent struct {
	Log   *sharedaction.LogMessage
	Error error
}

func ReturnLogs(logevents []LogEvent, passedWarnings v7action.Warnings, passedError error) func(appName string, spaceGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc, v7action.Warnings, error) {
	return func(appName string, spaceGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc, v7action.Warnings, error) {
		logStream := make(chan sharedaction.LogMessage)
		errStream := make(chan error)
		go func() {
			defer close(logStream)
			defer close(errStream)

			for _, log := range logevents {
				if log.Log != nil {
					logStream <- *log.Log
				}
				if log.Error != nil {
					errStream <- log.Error
				}
			}
		}()

		return logStream, errStream, func() {}, passedWarnings, passedError
	}
}

var _ = Describe("push Command", func() {
	var (
		cmd                 PushCommand
		input               *Buffer
		testUI              *ui.UI
		fakeConfig          *commandfakes.FakeConfig
		fakeSharedActor     *commandfakes.FakeSharedActor
		fakeActor           *v7fakes.FakePushActor
		fakeVersionActor    *v7fakes.FakeV7ActorForPush
		fakeProgressBar     *v6fakes.FakeProgressBar
		fakeLogCacheClient  *sharedactionfakes.FakeLogCacheClient
		fakeManifestLocator *v7fakes.FakeManifestLocator
		fakeManifestParser  *v7fakes.FakeManifestParser
		binaryName          string
		executeErr          error

		appName1  string
		appName2  string
		userName  string
		spaceName string
		orgName   string
		pwd       string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakePushActor)
		fakeVersionActor = new(v7fakes.FakeV7ActorForPush)
		fakeProgressBar = new(v6fakes.FakeProgressBar)
		fakeLogCacheClient = new(sharedactionfakes.FakeLogCacheClient)

		appName1 = "first-app"
		appName2 = "second-app"
		userName = "some-user"
		spaceName = "some-space"
		orgName = "some-org"
		pwd = "/push/cmd/test"
		fakeManifestLocator = new(v7fakes.FakeManifestLocator)
		fakeManifestParser = new(v7fakes.FakeManifestParser)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = PushCommand{
			UI:              testUI,
			Config:          fakeConfig,
			Actor:           fakeActor,
			VersionActor:    fakeVersionActor,
			SharedActor:     fakeSharedActor,
			ProgressBar:     fakeProgressBar,
			LogCacheClient:  fakeLogCacheClient,
			CWD:             pwd,
			ManifestLocator: fakeManifestLocator,
			ManifestParser:  fakeManifestParser,
		}
	})

	Describe("Execute", func() {
		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		BeforeEach(func() {
			fakeActor.ActualizeStub = func(v7pushaction.PushPlan, v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent {
				return FillInEvents([]Step{})
			}
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

			When("invalid flags are passed", func() {
				BeforeEach(func() {
					cmd.DockerUsername = "some-docker-username"
				})

				It("returns a validation error", func() {
					Expect(executeErr).To(MatchError(translatableerror.RequiredFlagsError{Arg1: "--docker-image, -o", Arg2: "--docker-username"}))
				})
			})

			When("the flags are all valid", func() {
				It("delegating to the GetBaseManifest", func() {
					// This tells us GetBaseManifest is being called because we dont have a fake
					Expect(fakeManifestLocator.PathCallCount()).To(Equal(1))
				})

				When("getting the base manifest fails", func() {
					BeforeEach(func() {
						fakeManifestLocator.PathReturns("", false, errors.New("locate-error"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(errors.New("locate-error")))
					})
				})

				When("getting the base manifest succeeds", func() {
					BeforeEach(func() {
						// essentially fakes GetBaseManifest
						fakeManifestLocator.PathReturns("", true, nil)
						fakeManifestParser.InterpolateAndParseReturns(
							manifestparser.Manifest{
								Applications: []manifestparser.Application{
									{
										Name: "some-app-name",
									},
								},
							},
							nil,
						)
					})

					It("delegates to the flag override handler", func() {
						Expect(fakeActor.HandleFlagOverridesCallCount()).To(Equal(1))
						actualManifest, actualFlagOverrides := fakeActor.HandleFlagOverridesArgsForCall(0)
						Expect(actualManifest).To(Equal(
							manifestparser.Manifest{
								Applications: []manifestparser.Application{
									{Name: "some-app-name"},
								},
							},
						))
						Expect(actualFlagOverrides).To(Equal(v7pushaction.FlagOverrides{}))
					})

					When("handling the flag overrides fails", func() {
						BeforeEach(func() {
							fakeActor.HandleFlagOverridesReturns(manifestparser.Manifest{}, errors.New("override-handler-error"))
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError("override-handler-error"))
						})
					})

					When("handling the flag overrides succeeds", func() {
						BeforeEach(func() {
							fakeActor.HandleFlagOverridesReturns(
								manifestparser.Manifest{
									Applications: []manifestparser.Application{
										{Name: "some-app-name"},
									},
								},
								nil,
							)
						})

						When("the docker password is needed", func() {
							// TODO remove this in favor of a fake manifest
							BeforeEach(func() {
								fakeActor.HandleFlagOverridesReturns(
									manifestparser.Manifest{
										Applications: []manifestparser.Application{
											{
												Name:   "some-app-name",
												Docker: &manifestparser.Docker{Username: "username", Image: "image"},
											},
										},
									},
									nil,
								)
							})

							It("delegates to the GetDockerPassword", func() {
								Expect(fakeConfig.DockerPasswordCallCount()).To(Equal(1))
							})
						})

						It("delegates to the manifest parser", func() {
							Expect(fakeManifestParser.MarshalManifestCallCount()).To(Equal(1))
							Expect(fakeManifestParser.MarshalManifestArgsForCall(0)).To(Equal(
								manifestparser.Manifest{
									Applications: []manifestparser.Application{
										{Name: "some-app-name"},
									},
								},
							))
						})

						When("marshalling the manifest fails", func() {
							BeforeEach(func() {
								fakeManifestParser.MarshalManifestReturns([]byte{}, errors.New("marshal error"))
							})

							It("returns the error", func() {
								Expect(executeErr).To(MatchError("marshal error"))
							})
						})

						When("marsahlling the manifest succeeds", func() {
							BeforeEach(func() {
								fakeManifestParser.MarshalManifestReturns([]byte("our-manifest"), nil)
							})

							It("delegates to the version actor", func() {
								Expect(fakeVersionActor.SetSpaceManifestCallCount()).To(Equal(1))
								actualSpaceGUID, actualManifestBytes := fakeVersionActor.SetSpaceManifestArgsForCall(0)
								Expect(actualSpaceGUID).To(Equal("some-space-guid"))
								Expect(actualManifestBytes).To(Equal([]byte("our-manifest")))
							})

							When("applying the manifest fails", func() {
								BeforeEach(func() {
									fakeVersionActor.SetSpaceManifestReturns(v7action.Warnings{"apply-manifest-warnings"}, errors.New("apply-manifest-error"))
								})

								It("returns an error and prints warnings", func() {
									Expect(executeErr).To(MatchError("apply-manifest-error"))
									Expect(testUI.Err).To(Say("apply-manifest-warnings"))
								})

							})

							When("applying the manifest succeeds", func() {
								BeforeEach(func() {
									fakeVersionActor.SetSpaceManifestReturns(v7action.Warnings{"apply-manifest-warnings"}, nil)
								})

								It("delegates to the push actor", func() {
									Expect(fakeActor.CreatePushPlansCallCount()).To(Equal(1))
									spaceGUID, orgGUID, manifest, overrides := fakeActor.CreatePushPlansArgsForCall(0)
									Expect(spaceGUID).To(Equal("some-space-guid"))
									Expect(orgGUID).To(Equal("some-org-guid"))
									Expect(manifest).To(Equal(
										manifestparser.Manifest{
											Applications: []manifestparser.Application{
												{Name: "some-app-name"},
											},
										},
									))
									Expect(overrides).To(Equal(v7pushaction.FlagOverrides{}))
								})

								When("creating the push plans fails", func() {
									BeforeEach(func() {
										fakeActor.CreatePushPlansReturns(
											nil,
											v7action.Warnings{"create-push-plans-warnings"},
											errors.New("create-push-plans-error"),
										)
									})

									It("returns errors and warnings", func() {
										Expect(executeErr).To(MatchError("create-push-plans-error"))
										Expect(testUI.Err).To(Say("create-push-plans-warnings"))
									})

								})

								When("creating the push plans succeeds", func() {
									BeforeEach(func() {
										fakeActor.CreatePushPlansReturns(
											[]v7pushaction.PushPlan{
												v7pushaction.PushPlan{Application: v7action.Application{Name: "first-app", GUID: "potato"}},
												v7pushaction.PushPlan{Application: v7action.Application{Name: "second-app", GUID: "potato"}},
											},
											v7action.Warnings{"create-push-plans-warnings"},
											nil,
										)
									})

									It("it displays the warnings from create push plans", func() {
										Expect(testUI.Err).To(Say("create-push-plans-warnings"))
									})

									Describe("delegating to Actor.Actualize", func() {
										When("Actualize returns success", func() {
											BeforeEach(func() {
												fakeActor.ActualizeStub = func(v7pushaction.PushPlan, v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent {
													return FillInEvents([]Step{
														{Plan: v7pushaction.PushPlan{Application: v7action.Application{GUID: "potato"}}},
													})
												}
											})

											Describe("actualize events", func() {
												BeforeEach(func() {
													fakeActor.ActualizeStub = func(pushPlan v7pushaction.PushPlan, _ v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent {
														return FillInEvents([]Step{
															{
																Plan:  v7pushaction.PushPlan{Application: v7action.Application{GUID: pushPlan.Application.GUID, Name: pushPlan.Application.Name}},
																Event: v7pushaction.CreatingArchive,
															},
															{
																Plan:     v7pushaction.PushPlan{Application: v7action.Application{GUID: pushPlan.Application.GUID, Name: pushPlan.Application.Name}},
																Event:    v7pushaction.UploadingApplicationWithArchive,
																Warnings: v7pushaction.Warnings{"upload app archive warning"},
															},
															{
																Plan:     v7pushaction.PushPlan{Application: v7action.Application{GUID: pushPlan.Application.GUID, Name: pushPlan.Application.Name}},
																Event:    v7pushaction.RetryUpload,
																Warnings: v7pushaction.Warnings{"retry upload warning"},
															},
															{
																Plan:  v7pushaction.PushPlan{Application: v7action.Application{GUID: pushPlan.Application.GUID, Name: pushPlan.Application.Name}},
																Event: v7pushaction.UploadWithArchiveComplete,
															},
															{
																Plan:  v7pushaction.PushPlan{Application: v7action.Application{GUID: pushPlan.Application.GUID, Name: pushPlan.Application.Name}},
																Event: v7pushaction.RestartingApplication,
															},
															{
																Plan:  v7pushaction.PushPlan{Application: v7action.Application{GUID: pushPlan.Application.GUID, Name: pushPlan.Application.Name}},
																Event: v7pushaction.StartingDeployment,
															},
															{
																Plan:  v7pushaction.PushPlan{Application: v7action.Application{GUID: pushPlan.Application.GUID, Name: pushPlan.Application.Name}},
																Event: v7pushaction.WaitingForDeployment,
															},
														})
													}
												})

												It("actualizes the application and displays events/warnings", func() {
													Expect(executeErr).ToNot(HaveOccurred())

													Expect(fakeProgressBar.ReadyCallCount()).Should(Equal(2))
													Expect(fakeProgressBar.CompleteCallCount()).Should(Equal(2))

													Expect(testUI.Out).To(Say("Packaging files to upload..."))

													Expect(testUI.Out).To(Say("Uploading files..."))
													Expect(testUI.Err).To(Say("upload app archive warning"))

													Expect(testUI.Out).To(Say("Retrying upload due to an error..."))
													Expect(testUI.Err).To(Say("retry upload warning"))

													Expect(testUI.Out).To(Say("Waiting for API to complete processing files..."))

													Expect(testUI.Out).To(Say("Waiting for app first-app to start..."))

													Expect(testUI.Out).To(Say("Packaging files to upload..."))

													Expect(testUI.Out).To(Say("Uploading files..."))
													Expect(testUI.Err).To(Say("upload app archive warning"))

													Expect(testUI.Out).To(Say("Retrying upload due to an error..."))
													Expect(testUI.Err).To(Say("retry upload warning"))

													Expect(testUI.Out).To(Say("Waiting for API to complete processing files..."))

													Expect(testUI.Out).To(Say("Waiting for app second-app to start..."))

													Expect(testUI.Out).To(Say("Starting deployment for app second-app..."))

													Expect(testUI.Out).To(Say("Waiting for app to deploy..."))
												})
											})

											Describe("staging logs", func() {
												BeforeEach(func() {
													fakeActor.ActualizeStub = func(pushPlan v7pushaction.PushPlan, _ v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent {
														return FillInEvents([]Step{
															{Plan: pushPlan, Event: v7pushaction.StartingStaging},
														})
													}
												})

												When("there are no logging errors", func() {
													BeforeEach(func() {
														fakeVersionActor.GetStreamingLogsForApplicationByNameAndSpaceStub = ReturnLogs(
															[]LogEvent{
																{Log: sharedaction.NewLogMessage("log-message-1", "OUT", time.Now(), sharedaction.StagingLog, "source-instance")},
																{Log: sharedaction.NewLogMessage("log-message-2", "OUT", time.Now(), sharedaction.StagingLog, "source-instance")},
																{Log: sharedaction.NewLogMessage("log-message-3", "OUT", time.Now(), "potato", "source-instance")},
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

														Expect(fakeVersionActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(2))
														passedAppName, spaceGUID, _ := fakeVersionActor.GetStreamingLogsForApplicationByNameAndSpaceArgsForCall(0)
														Expect(passedAppName).To(Equal(appName1))
														Expect(spaceGUID).To(Equal("some-space-guid"))
														passedAppName, spaceGUID, _ = fakeVersionActor.GetStreamingLogsForApplicationByNameAndSpaceArgsForCall(1)
														Expect(passedAppName).To(Equal(appName2))
														Expect(spaceGUID).To(Equal("some-space-guid"))

													})
												})

												When("there are logging errors", func() {
													BeforeEach(func() {
														fakeVersionActor.GetStreamingLogsForApplicationByNameAndSpaceStub = ReturnLogs(
															[]LogEvent{
																{Error: errors.New("some-random-err")},
																{Error: actionerror.LogCacheTimeoutError{}},
																{Log: sharedaction.NewLogMessage("log-message-1", "OUT", time.Now(), sharedaction.StagingLog, "source-instance")},
															},
															v7action.Warnings{"log-warning-1", "log-warning-2"},
															nil,
														)
													})

													It("displays the errors as warnings", func() {
														Expect(testUI.Out).To(Say("Staging app and tracing logs..."))

														Expect(testUI.Err).To(Say("log-warning-1"))
														Expect(testUI.Err).To(Say("log-warning-2"))
														Eventually(testUI.Err).Should(Say("Failed to retrieve logs from Log Cache: some-random-err"))
														Eventually(testUI.Err).Should(Say("timeout connecting to log server, no log will be shown"))

														Eventually(testUI.Out).Should(Say("log-message-1"))
													})
												})
											})

											When("when getting the application summary succeeds", func() {
												BeforeEach(func() {
													summary := v7action.DetailedApplicationSummary{
														ApplicationSummary: v7action.ApplicationSummary{
															Application:      v7action.Application{},
															ProcessSummaries: v7action.ProcessSummaries{},
														},
														CurrentDroplet: v7action.Droplet{},
													}
													fakeVersionActor.GetDetailedAppSummaryReturnsOnCall(0, summary, v7action.Warnings{"app-1-summary-warning-1", "app-1-summary-warning-2"}, nil)
													fakeVersionActor.GetDetailedAppSummaryReturnsOnCall(1, summary, v7action.Warnings{"app-2-summary-warning-1", "app-2-summary-warning-2"}, nil)
												})

												// TODO: Don't test the shared.AppSummaryDisplayer.AppDisplay method.
												// Use DI to pass in a new AppSummaryDisplayer to the Command instead.
												It("displays the app summary", func() {
													Expect(executeErr).ToNot(HaveOccurred())
													Expect(fakeVersionActor.GetDetailedAppSummaryCallCount()).To(Equal(2))
												})
											})

											When("getting the application summary fails", func() {
												BeforeEach(func() {
													fakeVersionActor.GetDetailedAppSummaryReturns(
														v7action.DetailedApplicationSummary{},
														v7action.Warnings{"get-application-summary-warnings"},
														errors.New("get-application-summary-error"),
													)
												})

												It("does not display the app summary", func() {
													Expect(testUI.Out).ToNot(Say(`requested state:`))
												})

												It("returns the error from GetDetailedAppSummary", func() {
													Expect(executeErr).To(MatchError("get-application-summary-error"))
												})

												It("prints the warnings", func() {
													Expect(testUI.Err).To(Say("get-application-summary-warnings"))
												})
											})

										})

										When("actualize returns an error", func() {
											When("the error is generic", func() {
												BeforeEach(func() {
													fakeActor.ActualizeStub = func(v7pushaction.PushPlan, v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent {
														return FillInEvents([]Step{
															{Error: errors.New("anti avant garde naming")},
														})
													}
												})

												It("returns the error", func() {
													Expect(executeErr).To(MatchError("anti avant garde naming"))
												})
											})

											When("the error is a startup timeout error", func() {
												BeforeEach(func() {
													fakeActor.ActualizeStub = func(v7pushaction.PushPlan, v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent {
														return FillInEvents([]Step{
															{Error: actionerror.StartupTimeoutError{}},
														})
													}
												})

												It("returns the StartupTimeoutError and prints warnings", func() {
													Expect(executeErr).To(MatchError(translatableerror.StartupTimeoutError{
														AppName:    "first-app",
														BinaryName: binaryName,
													}))
												})
											})

											When("the error is a process crashed error", func() {
												BeforeEach(func() {
													fakeActor.ActualizeStub = func(v7pushaction.PushPlan, v7pushaction.ProgressBar) <-chan *v7pushaction.PushEvent {
														return FillInEvents([]Step{
															{Error: actionerror.AllInstancesCrashedError{}},
														})
													}
												})

												It("returns the ApplicationUnableToStartError", func() {
													Expect(executeErr).To(MatchError(translatableerror.ApplicationUnableToStartError{
														AppName:    "first-app",
														BinaryName: binaryName,
													}))
												})

												It("displays the app summary", func() {
													Expect(executeErr).To(HaveOccurred())
													Expect(fakeVersionActor.GetDetailedAppSummaryCallCount()).To(Equal(1))
												})
											})
										})
									})
								})
							})
						})
					})
				})
			})
		})
	})

	Describe("GetDockerPassword", func() {
		var (
			cmd        PushCommand
			fakeConfig *commandfakes.FakeConfig
			testUI     *ui.UI

			dockerUsername        string
			containsPrivateDocker bool

			executeErr     error
			dockerPassword string

			input *Buffer
		)

		BeforeEach(func() {
			input = NewBuffer()
			testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
			fakeConfig = new(commandfakes.FakeConfig)

			cmd = PushCommand{
				Config: fakeConfig,
				UI:     testUI,
			}
		})

		Describe("Get", func() {
			JustBeforeEach(func() {
				dockerPassword, executeErr = cmd.GetDockerPassword(dockerUsername, containsPrivateDocker)
			})

			When("docker image is private", func() {
				When("there is a manifest", func() {
					BeforeEach(func() {
						dockerUsername = ""
						containsPrivateDocker = true
					})

					When("a password is provided via environment variable", func() {
						BeforeEach(func() {
							fakeConfig.DockerPasswordReturns("some-docker-password")
						})

						It("takes the password from the environment", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).ToNot(Say("Environment variable CF_DOCKER_PASSWORD not set."))
							Expect(testUI.Out).ToNot(Say("Docker password"))

							Expect(testUI.Out).To(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))

							Expect(dockerPassword).To(Equal("some-docker-password"))
						})
					})

					When("no password is provided", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("some-docker-password\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("prompts for a password", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Environment variable CF_DOCKER_PASSWORD not set."))
							Expect(testUI.Out).To(Say("Docker password"))

							Expect(dockerPassword).To(Equal("some-docker-password"))
						})
					})
				})

				When("there is no manifest", func() {
					BeforeEach(func() {
						dockerUsername = "some-docker-username"
						containsPrivateDocker = false
					})

					When("a password is provided via environment variable", func() {
						BeforeEach(func() {
							fakeConfig.DockerPasswordReturns("some-docker-password")
						})

						It("takes the password from the environment", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).ToNot(Say("Environment variable CF_DOCKER_PASSWORD not set."))
							Expect(testUI.Out).ToNot(Say("Docker password"))

							Expect(testUI.Out).To(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))

							Expect(dockerPassword).To(Equal("some-docker-password"))
						})
					})

					When("no password is provided", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("some-docker-password\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("prompts for a password", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Environment variable CF_DOCKER_PASSWORD not set."))
							Expect(testUI.Out).To(Say("Docker password"))

							Expect(dockerPassword).To(Equal("some-docker-password"))
						})
					})
				})
			})
			When("docker image is public", func() {
				BeforeEach(func() {
					dockerUsername = ""
					containsPrivateDocker = false
				})

				It("does not prompt for a password", func() {
					Expect(testUI.Out).ToNot(Say("Environment variable CF_DOCKER_PASSWORD not set."))
					Expect(testUI.Out).ToNot(Say("Docker password"))
					Expect(testUI.Out).ToNot(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))
				})

				It("returns an empty password", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(dockerPassword).To(Equal(""))
				})
			})
		})
	})

	Describe("GetBaseManifest", func() {
		var (
			somePath      string
			flagOverrides v7pushaction.FlagOverrides
			manifest      manifestparser.Manifest
			executeErr    error
		)

		JustBeforeEach(func() {
			manifest, executeErr = cmd.GetBaseManifest(flagOverrides)
		})

		When("no flags are specified", func() {
			BeforeEach(func() {
				cmd.CWD = somePath
			})

			When("a manifest exists in the current dir", func() {
				BeforeEach(func() {
					fakeManifestLocator.PathReturns("/manifest/path", true, nil)
					fakeManifestParser.InterpolateAndParseReturns(
						manifestparser.Manifest{
							Applications: []manifestparser.Application{
								{Name: "new-app"},
							},
						},
						nil,
					)
				})

				It("uses the manifest in the current directory", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(manifest).To(Equal(
						manifestparser.Manifest{
							Applications: []manifestparser.Application{
								{Name: "new-app"},
							},
						}),
					)

					Expect(fakeManifestLocator.PathCallCount()).To(Equal(1))
					Expect(fakeManifestLocator.PathArgsForCall(0)).To(Equal(cmd.CWD))

					Expect(fakeManifestParser.InterpolateAndParseCallCount()).To(Equal(1))
					actualManifestPath, _, _ := fakeManifestParser.InterpolateAndParseArgsForCall(0)
					Expect(actualManifestPath).To(Equal("/manifest/path"))
				})
			})

			When("there is not a manifest in the current dir", func() {
				BeforeEach(func() {
					flagOverrides.AppName = "new-app"
					fakeManifestLocator.PathReturns("", false, nil)
				})

				It("ignores the file not found error", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeManifestParser.InterpolateAndParseCallCount()).To(Equal(0))
				})

				It("returns a default empty manifest", func() {
					Expect(manifest).To(Equal(
						manifestparser.Manifest{
							Applications: []manifestparser.Application{
								{Name: "new-app"},
							},
						}),
					)
				})
			})

			When("when there is an error locating the manifest in the current directory", func() {
				BeforeEach(func() {
					fakeManifestLocator.PathReturns("", false, errors.New("err-location"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("err-location"))
					Expect(fakeManifestParser.InterpolateAndParseCallCount()).To(Equal(0))
				})
			})

			When("parsing the manifest fails", func() {
				BeforeEach(func() {
					fakeManifestLocator.PathReturns("/manifest/path", true, nil)
					fakeManifestParser.InterpolateAndParseReturns(
						manifestparser.Manifest{},
						errors.New("bad yaml"),
					)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("bad yaml"))
					Expect(fakeManifestParser.InterpolateAndParseCallCount()).To(Equal(1))
				})
			})
		})

		When("The -f flag is specified", func() {
			BeforeEach(func() {
				somePath = "some-path"
				flagOverrides.ManifestPath = somePath
				fakeManifestLocator.PathReturns("/manifest/path", true, nil)
				fakeManifestParser.InterpolateAndParseReturns(
					manifestparser.Manifest{
						Applications: []manifestparser.Application{
							{Name: "new-app"},
						},
					},
					nil,
				)
			})

			It("parses the specified manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeManifestLocator.PathCallCount()).To(Equal(1))
				Expect(fakeManifestLocator.PathArgsForCall(0)).To(Equal(somePath))

				Expect(fakeManifestParser.InterpolateAndParseCallCount()).To(Equal(1))
				actualManifestPath, _, _ := fakeManifestParser.InterpolateAndParseArgsForCall(0)
				Expect(actualManifestPath).To(Equal("/manifest/path"))
				Expect(manifest).To(Equal(
					manifestparser.Manifest{
						Applications: []manifestparser.Application{
							{Name: "new-app"},
						},
					}),
				)
			})
		})

		When("--vars-files are specified", func() {
			var varsFiles []string

			BeforeEach(func() {
				fakeManifestLocator.PathReturns("/manifest/path", true, nil)
				varsFiles = []string{"path1", "path2"}
				flagOverrides.PathsToVarsFiles = append(flagOverrides.PathsToVarsFiles, varsFiles...)
			})

			It("passes vars files to the manifest parser", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeManifestParser.InterpolateAndParseCallCount()).To(Equal(1))
				_, actualVarsFiles, _ := fakeManifestParser.InterpolateAndParseArgsForCall(0)
				Expect(actualVarsFiles).To(Equal(varsFiles))
			})
		})

		When("The --var flag is provided", func() {
			var vars []template.VarKV

			BeforeEach(func() {
				fakeManifestLocator.PathReturns("/manifest/path", true, nil)
				vars = []template.VarKV{
					{Name: "put-var-here", Value: "turtle"},
				}
				flagOverrides.Vars = vars
			})

			It("passes vars files to the manifest parser", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeManifestParser.InterpolateAndParseCallCount()).To(Equal(1))
				_, _, actualVars := fakeManifestParser.InterpolateAndParseArgsForCall(0)
				Expect(actualVars).To(Equal(vars))
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
			cmd.Stack = "validStack"
			cmd.HealthCheckType = flag.HealthCheckType{Type: constant.Port}
			cmd.HealthCheckHTTPEndpoint = "/health-check-http-endpoint"
			cmd.HealthCheckTimeout = flag.PositiveInteger{Value: 7}
			cmd.Memory = "64M"
			cmd.Disk = "256M"
			cmd.DropletPath = flag.PathWithExistenceCheck("some-droplet.tgz")
			cmd.StartCommand = flag.Command{FilteredString: types.FilteredString{IsSet: true, Value: "some-start-command"}}
			cmd.NoRoute = true
			cmd.RandomRoute = false
			cmd.NoStart = true
			cmd.NoWait = true
			cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyRolling}
			cmd.Instances = flag.Instances{NullInt: types.NullInt{Value: 10, IsSet: true}}
			cmd.PathToManifest = "/manifest/path"
			cmd.PathsToVarsFiles = []flag.PathWithExistenceCheck{"/vars1", "/vars2"}
			cmd.Vars = []template.VarKV{{Name: "key", Value: "val"}}
			cmd.Task = true
		})

		JustBeforeEach(func() {
			overrides, overridesErr = cmd.GetFlagOverrides()
			Expect(overridesErr).ToNot(HaveOccurred())
		})

		It("sets them on the flag overrides", func() {
			Expect(overridesErr).ToNot(HaveOccurred())
			Expect(overrides.Buildpacks).To(ConsistOf("buildpack-1", "buildpack-2"))
			Expect(overrides.DropletPath).To(Equal("some-droplet.tgz"))
			Expect(overrides.Stack).To(Equal("validStack"))
			Expect(overrides.HealthCheckType).To(Equal(constant.Port))
			Expect(overrides.HealthCheckEndpoint).To(Equal("/health-check-http-endpoint"))
			Expect(overrides.HealthCheckTimeout).To(BeEquivalentTo(7))
			Expect(overrides.Memory).To(Equal("64M"))
			Expect(overrides.Disk).To(Equal("256M"))
			Expect(overrides.StartCommand).To(Equal(types.FilteredString{IsSet: true, Value: "some-start-command"}))
			Expect(overrides.NoRoute).To(BeTrue())
			Expect(overrides.NoStart).To(BeTrue())
			Expect(overrides.NoWait).To(BeTrue())
			Expect(overrides.RandomRoute).To(BeFalse())
			Expect(overrides.Strategy).To(Equal(constant.DeploymentStrategyRolling))
			Expect(overrides.Instances).To(Equal(types.NullInt{Value: 10, IsSet: true}))
			Expect(overrides.ManifestPath).To(Equal("/manifest/path"))
			Expect(overrides.PathsToVarsFiles).To(Equal([]string{"/vars1", "/vars2"}))
			Expect(overrides.Vars).To(Equal([]template.VarKV{{Name: "key", Value: "val"}}))
			Expect(overrides.Task).To(BeTrue())
		})

		When("a docker image is provided", func() {
			BeforeEach(func() {
				cmd.DockerImage = flag.DockerImage{Path: "some-docker-image"}
			})

			It("sets docker image on the flag overrides", func() {
				Expect(overridesErr).ToNot(HaveOccurred())
				Expect(overrides.DockerImage).To(Equal("some-docker-image"))
			})
		})
	})

	DescribeTable("ValidateFlags returns an error",
		func(setup func(), expectedErr error) {
			setup()
			err := cmd.ValidateFlags()
			if expectedErr == nil {
				Expect(err).To(BeNil())
			} else {
				Expect(err).To(MatchError(expectedErr))
			}
		},

		Entry("when docker username flag is passed *without* docker flag",
			func() {
				cmd.DockerUsername = "some-docker-username"
			},
			translatableerror.RequiredFlagsError{Arg1: "--docker-image, -o", Arg2: "--docker-username"}),

		Entry("when docker and buildpacks flags are passed",
			func() {
				cmd.DockerImage.Path = "some-docker-image"
				cmd.Buildpacks = []string{"some-buildpack"}
			},
			translatableerror.ArgumentCombinationError{Args: []string{"--buildpack, -b", "--docker-image, -o"}}),

		Entry("when docker and stack flags are passed",
			func() {
				cmd.DockerImage.Path = "some-docker-image"
				cmd.Stack = "validStack"
			},
			translatableerror.ArgumentCombinationError{Args: []string{"--stack, -s", "--docker-image, -o"}}),

		Entry("when docker and path flags are passed",
			func() {
				cmd.DockerImage.Path = "some-docker-image"
				cmd.AppPath = "some-directory-path"
			},
			translatableerror.ArgumentCombinationError{Args: []string{"--docker-image, -o", "--path, -p"}}),

		Entry("when -u http does not have a matching --endpoint",
			func() {
				cmd.HealthCheckType.Type = constant.HTTP
			},
			translatableerror.RequiredFlagsError{Arg1: "--endpoint", Arg2: "--health-check-type=http, -u=http"}),

		Entry("when -u http does have a matching --endpoint",
			func() {
				cmd.HealthCheckType.Type = constant.HTTP
				cmd.HealthCheckHTTPEndpoint = "/health"
			},
			nil),

		Entry("when droplet and path flags are passed",
			func() {
				cmd.DropletPath = "some-droplet.tgz"
				cmd.AppPath = "/my/app"
			},
			translatableerror.ArgumentCombinationError{
				Args: []string{
					"--droplet", "--docker-image, -o", "--docker-username", "-p",
				},
			}),

		Entry("when droplet and docker image flags are passed",
			func() {
				cmd.DropletPath = "some-droplet.tgz"
				cmd.DockerImage.Path = "docker-image"
			},
			translatableerror.ArgumentCombinationError{
				Args: []string{
					"--droplet", "--docker-image, -o", "--docker-username", "-p",
				},
			}),

		Entry("when droplet, docker image, and docker username flags are passed",
			func() {
				cmd.DropletPath = "some-droplet.tgz"
				cmd.DockerImage.Path = "docker-image"
				cmd.DockerUsername = "docker-username"
			},
			translatableerror.ArgumentCombinationError{
				Args: []string{
					"--droplet", "--docker-image, -o", "--docker-username", "-p",
				},
			}),

		Entry("when strategy 'rolling' and no-start flags are passed",
			func() {
				cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyRolling}
				cmd.NoStart = true
			},
			translatableerror.ArgumentCombinationError{
				Args: []string{
					"--no-start", "--strategy=rolling",
				},
			}),

		Entry("when strategy is not set and no-start flags are passed",
			func() {
				cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyDefault}
				cmd.NoStart = true
			},
			nil),

		Entry("when no-start and no-wait flags are passed",
			func() {
				cmd.NoStart = true
				cmd.NoWait = true
			},
			translatableerror.ArgumentCombinationError{
				Args: []string{
					"--no-start", "--no-wait",
				},
			}),
		Entry("when no-route and random-route flags are passed",
			func() {
				cmd.NoRoute = true
				cmd.RandomRoute = true
			},
			translatableerror.ArgumentCombinationError{
				Args: []string{
					"--no-route", "--random-route",
				},
			}),

		Entry("default is combined with non default buildpacks",
			func() {
				cmd.Buildpacks = []string{"some-docker-username", "default"}
			},
			translatableerror.InvalidBuildpacksError{}),

		Entry("default is combined with non default buildpacks",
			func() {
				cmd.Buildpacks = []string{"some-docker-username", "null"}
			},
			translatableerror.InvalidBuildpacksError{}),

		Entry("task and strategy flags are passed",
			func() {
				cmd.Task = true
				cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyRolling}
			},
			translatableerror.ArgumentCombinationError{
				Args: []string{
					"--task", "--strategy=rolling",
				},
			}),
	)
})
