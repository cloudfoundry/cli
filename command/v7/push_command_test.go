package v7_test

import (
	"errors"
	"os"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
)

type Step struct {
	Error    error
	Event    pushaction.Event
	Warnings pushaction.Warnings
}

func FillInValues(tuples []Step, state pushaction.PushState) func(pushaction.PushState, pushaction.ProgressBar) (<-chan pushaction.PushState, <-chan pushaction.Event, <-chan pushaction.Warnings, <-chan error) {
	return func(pushaction.PushState, pushaction.ProgressBar) (<-chan pushaction.PushState, <-chan pushaction.Event, <-chan pushaction.Warnings, <-chan error) {
		stateStream := make(chan pushaction.PushState)

		eventStream := make(chan pushaction.Event)
		warningsStream := make(chan pushaction.Warnings)
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
			eventStream <- pushaction.Complete
		}()

		return stateStream, eventStream, warningsStream, errorStream
	}
}

type LogEvent struct {
	Log   *v3action.LogMessage
	Error error
}

func ReturnLogs(logevents []LogEvent, passedWarnings v3action.Warnings, passedError error) func(appName string, spaceGUID string, client v3action.NOAAClient) (<-chan *v3action.LogMessage, <-chan error, v3action.Warnings, error) {
	return func(appName string, spaceGUID string, client v3action.NOAAClient) (<-chan *v3action.LogMessage, <-chan error, v3action.Warnings, error) {
		logStream := make(chan *v3action.LogMessage)
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
		fakeVersionActor *v7fakes.FakePushVersionActor
		fakeProgressBar  *v6fakes.FakeProgressBar
		fakeNOAAClient   *v3actionfakes.FakeNOAAClient
		binaryName       string
		executeErr       error

		appName   string
		userName  string
		spaceName string
		orgName   string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakePushActor)
		fakeVersionActor = new(v7fakes.FakePushVersionActor)
		fakeProgressBar = new(v6fakes.FakeProgressBar)
		fakeNOAAClient = new(v3actionfakes.FakeNOAAClient)

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
		}

		appName = "some-app"
		userName = "some-user"
		spaceName = "some-space"
		orgName = "some-org"
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
						[]pushaction.PushState{
							{
								Application: v3action.Application{Name: appName},
							},
						},
						pushaction.Warnings{"some-warning-1"}, nil)
				})

				Describe("actualizing non-logging events", func() {
					BeforeEach(func() {
						fakeActor.ActualizeStub = FillInValues([]Step{
							{
								Event:    pushaction.SkippingApplicationCreation,
								Warnings: pushaction.Warnings{"skipping app creation warnings"},
							},
							{
								Event:    pushaction.CreatedApplication,
								Warnings: pushaction.Warnings{"app creation warnings"},
							},
							{
								Event: pushaction.CreatingArchive,
							},
							{
								Event:    pushaction.UploadingApplicationWithArchive,
								Warnings: pushaction.Warnings{"upload app archive warning"},
							},
							{
								Event:    pushaction.RetryUpload,
								Warnings: pushaction.Warnings{"retry upload warning"},
							},
							{
								Event: pushaction.UploadWithArchiveComplete,
							},
							{
								Event: pushaction.StagingComplete,
							},
						}, pushaction.PushState{})
					})

					It("generates a push state with the specified app path", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Pushing app %s to org some-org / space some-space as some-user", appName))
						Expect(testUI.Out).To(Say(`Getting app info\.\.\.`))
						Expect(testUI.Err).To(Say("some-warning-1"))

						Expect(fakeActor.ConceptualizeCallCount()).To(Equal(1))
						settings, spaceGUID := fakeActor.ConceptualizeArgsForCall(0)
						Expect(settings).To(MatchFields(IgnoreExtras, Fields{
							"Name": Equal("some-app"),
						}))
						Expect(spaceGUID).To(Equal("some-space-guid"))
					})

					It("actualizes the application and displays events/warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Updating app some-app..."))
						Expect(testUI.Err).To(Say("skipping app creation warnings"))

						Expect(testUI.Out).To(Say("Creating app some-app..."))
						Expect(testUI.Err).To(Say("app creation warnings"))

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
								Event: pushaction.StartingStaging,
							},
						}, pushaction.PushState{})
					})

					When("there are no logging errors", func() {
						BeforeEach(func() {
							fakeVersionActor.GetStreamingLogsForApplicationByNameAndSpaceStub = ReturnLogs(
								[]LogEvent{
									{Log: v3action.NewLogMessage("log-message-1", 1, time.Now(), v3action.StagingLog, "source-instance")},
									{Log: v3action.NewLogMessage("log-message-2", 1, time.Now(), v3action.StagingLog, "source-instance")},
									{Log: v3action.NewLogMessage("log-message-3", 1, time.Now(), "potato", "source-instance")},
								},
								v3action.Warnings{"log-warning-1", "log-warning-2"},
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
									{Log: v3action.NewLogMessage("log-message-1", 1, time.Now(), v3action.StagingLog, "source-instance")},
								},
								v3action.Warnings{"log-warning-1", "log-warning-2"},
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
						}, pushaction.PushState{Application: v3action.Application{GUID: "potato"}})
					})

					// It("outputs flavor text prior to generating app configuration", func() {
					// })

					When("restarting the app succeeds", func() {
						BeforeEach(func() {
							fakeVersionActor.RestartApplicationReturns(v3action.Warnings{"some-restart-warning"}, nil)
						})

						It("restarts the app and displays warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeVersionActor.RestartApplicationCallCount()).To(Equal(1))
							Expect(fakeVersionActor.RestartApplicationArgsForCall(0)).To(Equal("potato"))
							Expect(testUI.Err).To(Say("some-restart-warning"))
						})

						When("polling the restart succeeds", func() {
							BeforeEach(func() {
								fakeVersionActor.PollStartStub = func(appGUID string, warnings chan<- v3action.Warnings) error {
									warnings <- v3action.Warnings{"some-poll-warning-1", "some-poll-warning-2"}
									return nil
								}
							})

							It("displays all warnings", func() {
								Expect(testUI.Err).To(Say("some-poll-warning-1"))
								Expect(testUI.Err).To(Say("some-poll-warning-2"))

								Expect(executeErr).ToNot(HaveOccurred())
							})
						})

						When("polling the start fails", func() {
							BeforeEach(func() {
								fakeVersionActor.PollStartStub = func(appGUID string, warnings chan<- v3action.Warnings) error {
									warnings <- v3action.Warnings{"some-poll-warning-1", "some-poll-warning-2"}
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
							fakeVersionActor.RestartApplicationReturns(v3action.Warnings{"some-restart-warning"}, errors.New("restart failure"))
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
						}, pushaction.PushState{})
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
					fakeActor.ConceptualizeReturns(nil, pushaction.Warnings{"some-warning-1"}, expectedErr)
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
					settings, spaceGUID := fakeActor.ConceptualizeArgsForCall(0)
					Expect(settings).To(MatchFields(IgnoreExtras, Fields{
						"Name":            Equal("some-app"),
						"ProvidedAppPath": Equal("some/app/path"),
					}))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			When("buildpack is specified", func() {
				BeforeEach(func() {
					cmd.Buildpacks = []string{"some-buildpack-1", "some-buildpack-2"}
				})

				It("generates a push state with the specified buildpacks", func() {
					Expect(fakeActor.ConceptualizeCallCount()).To(Equal(1))
					settings, spaceGUID := fakeActor.ConceptualizeArgsForCall(0)
					Expect(settings).To(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("some-app"),
						"Buildpacks": Equal([]string{"some-buildpack-1", "some-buildpack-2"}),
					}))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})
		})
	})

	Describe("GetCommandLineSettings", func() {
		Context("valid flag combinations", func() {
			var (
				settings               pushaction.CommandLineSettings
				commandLineSettingsErr error
			)

			JustBeforeEach(func() {
				settings, commandLineSettingsErr = cmd.GetCommandLineSettings()
				Expect(commandLineSettingsErr).ToNot(HaveOccurred())
			})

			// When("general app settings are given", func() {
			// 	BeforeEach(func() {
			// 		cmd.Buildpacks = []string{"some-buildpack"}
			// 		cmd.Command = flag.Command{FilteredString: types.FilteredString{IsSet: true, Value: "echo foo bar baz"}}
			// 		cmd.DiskQuota = flag.Megabytes{NullUint64: types.NullUint64{Value: 1024, IsSet: true}}
			// 		cmd.HealthCheckTimeout = 14
			// 		cmd.HealthCheckType = flag.HealthCheckType{Type: "http"}
			// 		cmd.Instances = flag.Instances{NullInt: types.NullInt{Value: 12, IsSet: true}}
			// 		cmd.Memory = flag.Megabytes{NullUint64: types.NullUint64{Value: 100, IsSet: true}}
			// 		cmd.StackName = "some-stack"
			// 	})

			// 	It("sets them on the command line settings", func() {
			// 		Expect(commandLineSettingsErr).ToNot(HaveOccurred())
			// 		Expect(settings.Buildpacks).To(ConsistOf("some-buildpack"))
			// 		Expect(settings.Command).To(Equal(types.FilteredString{IsSet: true, Value: "echo foo bar baz"}))
			// 		Expect(settings.DiskQuota).To(Equal(uint64(1024)))
			// 		Expect(settings.HealthCheckTimeout).To(Equal(14))
			// 		Expect(settings.HealthCheckType).To(Equal("http"))
			// 		Expect(settings.Instances).To(Equal(types.NullInt{Value: 12, IsSet: true}))
			// 		Expect(settings.Memory).To(Equal(uint64(100)))
			// 		Expect(settings.StackName).To(Equal("some-stack"))
			// 	})
			// })

			// Context("route related flags", func() {
			// 	When("given customed route settings", func() {
			// 		BeforeEach(func() {
			// 			cmd.Domain = "some-domain"
			// 		})

			// 		It("sets NoHostname on the command line settings", func() {
			// 			Expect(settings.DefaultRouteDomain).To(Equal("some-domain"))
			// 		})
			// 	})

			// 	When("--hostname is given", func() {
			// 		BeforeEach(func() {
			// 			cmd.Hostname = "some-hostname"
			// 		})

			// 		It("sets DefaultRouteHostname on the command line settings", func() {
			// 			Expect(settings.DefaultRouteHostname).To(Equal("some-hostname"))
			// 		})
			// 	})

			// 	When("--no-hostname is given", func() {
			// 		BeforeEach(func() {
			// 			cmd.NoHostname = true
			// 		})

			// 		It("sets NoHostname on the command line settings", func() {
			// 			Expect(settings.NoHostname).To(BeTrue())
			// 		})
			// 	})

			// 	When("--random-route is given", func() {
			// 		BeforeEach(func() {
			// 			cmd.RandomRoute = true
			// 		})

			// 		It("sets --random-route on the command line settings", func() {
			// 			Expect(commandLineSettingsErr).ToNot(HaveOccurred())
			// 			Expect(settings.RandomRoute).To(BeTrue())
			// 		})
			// 	})

			// 	When("--route-path is given", func() {
			// 		BeforeEach(func() {
			// 			cmd.RoutePath = flag.RoutePath{Path: "/some-path"}
			// 		})

			// 		It("sets --route-path on the command line settings", func() {
			// 			Expect(commandLineSettingsErr).ToNot(HaveOccurred())
			// 			Expect(settings.RoutePath).To(Equal("/some-path"))
			// 		})
			// 	})

			// 	When("--no-route is given", func() {
			// 		BeforeEach(func() {
			// 			cmd.NoRoute = true
			// 		})

			// 		It("sets NoRoute on the command line settings", func() {
			// 			Expect(settings.NoRoute).To(BeTrue())
			// 		})
			// 	})
			// })

			Context("app bits", func() {
				When("-p flag is given", func() {
					BeforeEach(func() {
						cmd.AppPath = "some-directory-path"
					})

					It("sets ProvidedAppPath", func() {
						Expect(settings.ProvidedAppPath).To(Equal("some-directory-path"))
					})
				})

				It("sets the current directory in the command config", func() {
					pwd, err := os.Getwd()
					Expect(err).ToNot(HaveOccurred())
					Expect(settings.CurrentDirectory).To(Equal(pwd))
				})

				// When("the -o flag is given", func() {
				// 	BeforeEach(func() {
				// 		cmd.DockerImage.Path = "some-docker-image-path"
				// 	})

				// 	It("creates command line setting from command line arguments", func() {
				// 		Expect(settings.DockerImage).To(Equal("some-docker-image-path"))
				// 	})

				// 	Context("--docker-username flags is given", func() {
				// 		BeforeEach(func() {
				// 			cmd.DockerUsername = "some-docker-username"
				// 		})

				// 		Context("the docker password environment variable is set", func() {
				// 			BeforeEach(func() {
				// 				fakeConfig.DockerPasswordReturns("some-docker-password")
				// 			})

				// 			It("creates command line setting from command line arguments and config", func() {
				// 				Expect(testUI.Out).To(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))

				// 				Expect(settings.Name).To(Equal(appName))
				// 				Expect(settings.DockerImage).To(Equal("some-docker-image-path"))
				// 				Expect(settings.DockerUsername).To(Equal("some-docker-username"))
				// 				Expect(settings.DockerPassword).To(Equal("some-docker-password"))
				// 			})
				// 		})

				// 		Context("the docker password environment variable is *not* set", func() {
				// 			BeforeEach(func() {
				// 				input.Write([]byte("some-docker-password\n"))
				// 			})

				// 			It("prompts the user for a password", func() {
				// 				Expect(testUI.Out).To(Say("Environment variable CF_DOCKER_PASSWORD not set."))
				// 				Expect(testUI.Out).To(Say("Docker password"))

				// 				Expect(settings.Name).To(Equal(appName))
				// 				Expect(settings.DockerImage).To(Equal("some-docker-image-path"))
				// 				Expect(settings.DockerUsername).To(Equal("some-docker-username"))
				// 				Expect(settings.DockerPassword).To(Equal("some-docker-password"))
				// 			})
				// 		})
				// 	})
				// })
			})
		})

		// DescribeTable("validation errors when flags are passed",
		// 	func(setup func(), expectedErr error) {
		// 		setup()
		// 		_, commandLineSettingsErr := cmd.GetCommandLineSettings()
		// 		Expect(commandLineSettingsErr).To(MatchError(expectedErr))
		// 	},

		// 	Entry("--droplet and --docker-username",
		// 		func() {
		// 			cmd.DropletPath = "some-droplet-path"
		// 			cmd.DockerUsername = "some-docker-username"
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--droplet", "--docker-username", "-p"}}),

		// 	Entry("--droplet and --docker-image",
		// 		func() {
		// 			cmd.DropletPath = "some-droplet-path"
		// 			cmd.DockerImage.Path = "some-docker-image"
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--droplet", "--docker-image", "-o"}}),

		// 	Entry("--droplet and -p",
		// 		func() {
		// 			cmd.DropletPath = "some-droplet-path"
		// 			cmd.AppPath = "some-directory-path"
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--droplet", "-p"}}),

		// 	Entry("-o and -p",
		// 		func() {
		// 			cmd.DockerImage.Path = "some-docker-image"
		// 			cmd.AppPath = "some-directory-path"
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--docker-image, -o", "-p"}}),

		// 	Entry("-b and --docker-image",
		// 		func() {
		// 			cmd.DockerImage.Path = "some-docker-image"
		// 			cmd.Buildpacks = []string{"some-buildpack"}
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"-b", "--docker-image, -o"}}),

		// 	Entry("--docker-username (without DOCKER_PASSWORD env set)",
		// 		func() {
		// 			cmd.DockerUsername = "some-docker-username"
		// 		},
		// 		translatableerror.RequiredFlagsError{Arg1: "--docker-image, -o", Arg2: "--docker-username"}),

		// 	Entry("-d and --no-route",
		// 		func() {
		// 			cmd.Domain = "some-domain"
		// 			cmd.NoRoute = true
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"-d", "--no-route"}}),

		// 	Entry("--hostname and --no-hostname",
		// 		func() {
		// 			cmd.Hostname = "po-tate-toe"
		// 			cmd.NoHostname = true
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--hostname", "-n", "--no-hostname"}}),

		// 	Entry("--hostname and --no-route",
		// 		func() {
		// 			cmd.Hostname = "po-tate-toe"
		// 			cmd.NoRoute = true
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--hostname", "-n", "--no-route"}}),

		// 	Entry("--no-hostname and --no-route",
		// 		func() {
		// 			cmd.NoHostname = true
		// 			cmd.NoRoute = true
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--no-hostname", "--no-route"}}),

		// 	Entry("-f and --no-manifest",
		// 		func() {
		// 			cmd.PathToManifest = "/some/path.yml"
		// 			cmd.NoManifest = true
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"-f", "--no-manifest"}}),

		// 	Entry("--random-route and --hostname",
		// 		func() {
		// 			cmd.Hostname = "po-tate-toe"
		// 			cmd.RandomRoute = true
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--hostname", "-n", "--random-route"}}),

		// 	Entry("--random-route and --no-hostname",
		// 		func() {
		// 			cmd.RandomRoute = true
		// 			cmd.NoHostname = true
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--no-hostname", "--random-route"}}),

		// 	Entry("--random-route and --no-route",
		// 		func() {
		// 			cmd.RandomRoute = true
		// 			cmd.NoRoute = true
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--no-route", "--random-route"}}),

		// 	Entry("--random-route and --route-path",
		// 		func() {
		// 			cmd.RoutePath = flag.RoutePath{Path: "/bananas"}
		// 			cmd.RandomRoute = true
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--random-route", "--route-path"}}),

		// 	Entry("--route-path and --no-route",
		// 		func() {
		// 			cmd.RoutePath = flag.RoutePath{Path: "/bananas"}
		// 			cmd.NoRoute = true
		// 		},
		// 		translatableerror.ArgumentCombinationError{Args: []string{"--route-path", "--no-route"}}),
		// )
	})
})
