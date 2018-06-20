package v3_test

import (
	"errors"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/v3action"
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
	. "github.com/onsi/gomega/gstruct"
)

type Step struct {
	Error    error
	Event    pushaction.Event
	Warnings pushaction.Warnings
}

func FillInValues(tuples []Step, state pushaction.PushState) func(state pushaction.PushState, progressBar pushaction.ProgressBar) (<-chan pushaction.PushState, <-chan pushaction.Event, <-chan pushaction.Warnings, <-chan error) {
	return func(state pushaction.PushState, progressBar pushaction.ProgressBar) (<-chan pushaction.PushState, <-chan pushaction.Event, <-chan pushaction.Warnings, <-chan error) {
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

			eventStream <- pushaction.Complete
			stateStream <- state
		}()

		return stateStream, eventStream, warningsStream, errorStream
	}
}

var _ = Describe("v3-push Command", func() {
	var (
		cmd              v3.V3PushCommand
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v3fakes.FakeV3PushActor
		fakeVersionActor *v3fakes.FakeV3PushVersionActor
		fakeProgressBar  *v3fakes.FakeProgressBar
		binaryName       string
		executeErr       error

		userName  string
		spaceName string
		orgName   string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3PushActor)
		fakeVersionActor = new(v3fakes.FakeV3PushVersionActor)
		fakeProgressBar = new(v3fakes.FakeProgressBar)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.ExperimentalReturns(true) // TODO: Delete once we remove the experimental flag

		cmd = v3.V3PushCommand{
			RequiredArgs: flag.AppName{AppName: "some-app"},
			UI:           testUI,
			Config:       fakeConfig,
			Actor:        fakeActor,
			VersionActor: fakeVersionActor,
			SharedActor:  fakeSharedActor,
			ProgressBar:  fakeProgressBar,
		}

		userName = "banana"
		spaceName = "some-space"
		orgName = "some-org"
	})

	Describe("Execute", func() {
		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		Context("when the API version is below the minimum", func() {
			BeforeEach(func() {
				fakeVersionActor.CloudControllerAPIVersionReturns("0.0.0")
			})

			It("returns a MinimumAPIVersionNotMetError", func() {
				Expect(executeErr).To(MatchError(translatableerror.MinimumAPIVersionNotMetError{
					CurrentVersion: "0.0.0",
					MinimumVersion: ccversion.MinVersionV3,
				}))
			})
		})

		Context("when the API version is met", func() {
			BeforeEach(func() {
				fakeVersionActor.CloudControllerAPIVersionReturns(ccversion.MinVersionV3)
			})

			Context("when checking target fails", func() {
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

			Context("when checking target fails because the user is not logged in", func() {
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

				Context("when getting app settings is successful", func() {
					BeforeEach(func() {
						fakeActor.ConceptualizeReturns(
							[]pushaction.PushState{
								{
									Application: v3action.Application{Name: "some-app"},
								},
							},
							pushaction.Warnings{"some-warning-1"}, nil)
					})

					Context("when the app is successfully actualized", func() {
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
							}, pushaction.PushState{})
						})

						It("generates a push state with the specified app path", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).To(Say("Getting app info..."))
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
							Expect(fakeProgressBar.CompleteCallCount()).Should(Equal(1))
						})
					})

					Context("when actualizing fails", func() {
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

				Context("when getting app settings returns an error", func() {
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

				Context("when app path is specified", func() {
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

				Context("when buildpack is specified", func() {
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

			// Context("when general app settings are given", func() {
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
			// 	Context("when given customed route settings", func() {
			// 		BeforeEach(func() {
			// 			cmd.Domain = "some-domain"
			// 		})

			// 		It("sets NoHostname on the command line settings", func() {
			// 			Expect(settings.DefaultRouteDomain).To(Equal("some-domain"))
			// 		})
			// 	})

			// 	Context("when --hostname is given", func() {
			// 		BeforeEach(func() {
			// 			cmd.Hostname = "some-hostname"
			// 		})

			// 		It("sets DefaultRouteHostname on the command line settings", func() {
			// 			Expect(settings.DefaultRouteHostname).To(Equal("some-hostname"))
			// 		})
			// 	})

			// 	Context("when --no-hostname is given", func() {
			// 		BeforeEach(func() {
			// 			cmd.NoHostname = true
			// 		})

			// 		It("sets NoHostname on the command line settings", func() {
			// 			Expect(settings.NoHostname).To(BeTrue())
			// 		})
			// 	})

			// 	Context("when --random-route is given", func() {
			// 		BeforeEach(func() {
			// 			cmd.RandomRoute = true
			// 		})

			// 		It("sets --random-route on the command line settings", func() {
			// 			Expect(commandLineSettingsErr).ToNot(HaveOccurred())
			// 			Expect(settings.RandomRoute).To(BeTrue())
			// 		})
			// 	})

			// 	Context("when --route-path is given", func() {
			// 		BeforeEach(func() {
			// 			cmd.RoutePath = flag.RoutePath{Path: "/some-path"}
			// 		})

			// 		It("sets --route-path on the command line settings", func() {
			// 			Expect(commandLineSettingsErr).ToNot(HaveOccurred())
			// 			Expect(settings.RoutePath).To(Equal("/some-path"))
			// 		})
			// 	})

			// 	Context("when --no-route is given", func() {
			// 		BeforeEach(func() {
			// 			cmd.NoRoute = true
			// 		})

			// 		It("sets NoRoute on the command line settings", func() {
			// 			Expect(settings.NoRoute).To(BeTrue())
			// 		})
			// 	})
			// })

			Context("app bits", func() {
				Context("when -p flag is given", func() {
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

				// Context("when the -o flag is given", func() {
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
