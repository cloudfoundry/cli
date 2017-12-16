package v3_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/command/v3/shared/sharedfakes"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-push Command", func() {
	var (
		cmd             v3.V3PushCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeNOAAClient  *v3actionfakes.FakeNOAAClient
		fakeActor       *v3fakes.FakeV3PushActor
		fakeV2PushActor *v3fakes.FakeV2PushActor
		fakeV2AppActor  *sharedfakes.FakeV2AppRouteActor
		binaryName      string
		executeErr      error
		app             string
		userName        string
		spaceName       string
		orgName         string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3PushActor)
		fakeV2PushActor = new(v3fakes.FakeV2PushActor)
		fakeV2AppActor = new(sharedfakes.FakeV2AppRouteActor)
		fakeNOAAClient = new(v3actionfakes.FakeNOAAClient)

		fakeConfig.StagingTimeoutReturns(10 * time.Minute)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"
		userName = "banana"
		spaceName = "some-space"
		orgName = "some-org"

		appSummaryDisplayer := shared.AppSummaryDisplayer{
			UI:              testUI,
			Config:          fakeConfig,
			Actor:           fakeActor,
			V2AppRouteActor: fakeV2AppActor,
			AppName:         app,
		}
		packageDisplayer := shared.NewPackageDisplayer(
			testUI,
			fakeConfig,
		)

		cmd = v3.V3PushCommand{
			RequiredArgs: flag.AppName{AppName: app},

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			V2PushActor: fakeV2PushActor,

			NOAAClient:          fakeNOAAClient,
			AppSummaryDisplayer: appSummaryDisplayer,
			PackageDisplayer:    packageDisplayer,
		}
		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionV3)

		// we stub out StagePackage out here so the happy paths below don't hang
		fakeActor.StagePackageStub = func(_ string, _ string) (<-chan v3action.Droplet, <-chan v3action.Warnings, <-chan error) {
			dropletStream := make(chan v3action.Droplet)
			warningsStream := make(chan v3action.Warnings)
			errorStream := make(chan error)

			go func() {
				defer close(dropletStream)
				defer close(warningsStream)
				defer close(errorStream)
			}()

			return dropletStream, warningsStream, errorStream
		}
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

	DescribeTable("argument combinations",
		func(dockerImage string, dockerUsername string, dockerPassword string,
			buildpacks []string, appPath string,
			expectedErr error) {
			cmd.DockerImage.Path = dockerImage
			cmd.DockerUsername = dockerUsername
			fakeConfig.DockerPasswordReturns(dockerPassword)
			cmd.Buildpacks = buildpacks
			cmd.AppPath = flag.PathWithExistenceCheck(appPath)
			Expect(cmd.Execute(nil)).To(MatchError(expectedErr))
		},
		Entry("docker username",
			"", "some-docker-username", "", []string{}, "",
			translatableerror.RequiredFlagsError{
				Arg1: "--docker-image, -o",
				Arg2: "--docker-username",
			}),
		Entry("docker username, password",
			"", "some-docker-username", "my-password", []string{}, "",
			translatableerror.RequiredFlagsError{
				Arg1: "--docker-image, -o",
				Arg2: "--docker-username",
			}),
		Entry("docker username, app path",
			"", "some-docker-username", "", []string{}, "some/app/path",
			translatableerror.RequiredFlagsError{
				Arg1: "--docker-image, -o",
				Arg2: "--docker-username",
			}),
		Entry("docker username, buildpacks",
			"", "some-docker-username", "", []string{"ruby_buildpack"}, "",
			translatableerror.RequiredFlagsError{
				Arg1: "--docker-image, -o",
				Arg2: "--docker-username",
			}),
		Entry("docker image, docker username",
			"some-docker-image", "some-docker-username", "", []string{}, "",
			translatableerror.DockerPasswordNotSetError{}),
		Entry("docker image, app path",
			"some-docker-image", "", "", []string{}, "some/app/path",
			translatableerror.ArgumentCombinationError{
				Args: []string{"--docker-image", "-o", "-p"},
			}),
		Entry("docker image, buildpacks",
			"some-docker-image", "", "", []string{"ruby_buildpack"}, "",
			translatableerror.ArgumentCombinationError{
				Args: []string{"-b", "--docker-image", "-o"},
			}),
	)

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
			fakeConfig.CurrentUserReturns(configv3.User{Name: userName}, nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: spaceName, GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: orgName, GUID: "some-org-guid"})

		})

		Context("when looking up the application returns some api error", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"get-warning"}, errors.New("some-error"))
			})

			It("returns the error and displays all warnings", func() {
				Expect(executeErr).To(MatchError("some-error"))

				Expect(testUI.Err).To(Say("get-warning"))
			})
		})

		Context("when the application doesn't exist", func() {
			It("doesn't stop the application", func() {
				Expect(fakeActor.StopApplicationCallCount()).To(Equal(0))
			})

			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"get-warning"}, actionerror.ApplicationNotFoundError{Name: "some-app"})
			})

			Context("when creating the application returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am an error")
					fakeActor.CreateApplicationInSpaceReturns(v3action.Application{}, v3action.Warnings{"I am a warning", "I am also a warning"}, expectedErr)
				})

				It("displays the warnings and error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
					Expect(testUI.Out).ToNot(Say("app some-app in org some-org / space some-space as banana..."))
				})
			})

			Context("when creating the application does not error", func() {
				BeforeEach(func() {
					fakeActor.CreateApplicationInSpaceReturns(v3action.Application{Name: "some-app", GUID: "some-app-guid"}, v3action.Warnings{"I am a warning", "I am also a warning"}, nil)
				})

				It("calls CreateApplication", func() {
					Expect(fakeActor.CreateApplicationInSpaceCallCount()).To(Equal(1), "Expected CreateApplicationInSpace to be called once")
					createApp, createSpaceGUID := fakeActor.CreateApplicationInSpaceArgsForCall(0)
					Expect(createApp).To(Equal(v3action.Application{
						Name: "some-app",
						Lifecycle: v3action.AppLifecycle{
							Type: constant.BuildpackAppLifecycleType,
						},
					}))
					Expect(createSpaceGUID).To(Equal("some-space-guid"))
				})

				Context("when creating the package fails", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("I am an error")
						fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceReturns(v3action.Package{}, v3action.Warnings{"I am a package warning", "I am also a package warning"}, expectedErr)
					})

					It("displays the header and error", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Out).To(Say("Uploading and creating bits package for app some-app in org some-org / space some-space as banana..."))

						Expect(testUI.Err).To(Say("I am a package warning"))
						Expect(testUI.Err).To(Say("I am also a package warning"))

						Expect(testUI.Out).ToNot(Say("Staging package for %s in org some-org / space some-space as banana...", app))
					})
				})

				Context("when creating the package succeeds", func() {
					BeforeEach(func() {
						fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceReturns(v3action.Package{GUID: "some-guid"}, v3action.Warnings{"I am a package warning", "I am also a package warning"}, nil)
					})

					Context("when the -p flag is provided", func() {
						BeforeEach(func() {
							cmd.AppPath = "some-app-path"
						})

						It("creates the package with the provided path", func() {
							Expect(testUI.Out).To(Say("Uploading and creating bits package for app %s in org %s / space %s as %s", app, orgName, spaceName, userName))
							Expect(testUI.Err).To(Say("I am a package warning"))
							Expect(testUI.Err).To(Say("I am also a package warning"))
							Expect(testUI.Out).To(Say("OK"))
							Expect(testUI.Out).To(Say("Staging package for app %s in org some-org / space some-space as banana...", app))

							Expect(fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceCallCount()).To(Equal(1))
							_, _, appPath := fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceArgsForCall(0)

							Expect(appPath).To(Equal("some-app-path"))
						})
					})

					Context("when the -o flag is provided", func() {
						BeforeEach(func() {
							cmd.DockerImage.Path = "example.com/docker/docker/docker:docker"
							fakeActor.CreateDockerPackageByApplicationNameAndSpaceReturns(v3action.Package{GUID: "some-guid"}, v3action.Warnings{"I am a docker package warning", "I am also a docker package warning"}, nil)
						})

						It("creates a docker package with the provided image path", func() {

							Expect(testUI.Out).To(Say("Creating docker package for app %s in org %s / space %s as %s", app, orgName, spaceName, userName))
							Expect(testUI.Err).To(Say("I am a docker package warning"))
							Expect(testUI.Err).To(Say("I am also a docker package warning"))
							Expect(testUI.Out).To(Say("OK"))
							Expect(testUI.Out).To(Say("Staging package for app %s in org some-org / space some-space as banana...", app))

							Expect(fakeActor.CreateDockerPackageByApplicationNameAndSpaceCallCount()).To(Equal(1))
							_, _, dockerImageCredentials := fakeActor.CreateDockerPackageByApplicationNameAndSpaceArgsForCall(0)

							Expect(dockerImageCredentials.Path).To(Equal("example.com/docker/docker/docker:docker"))
						})
					})

					Context("when neither -p nor -o flags are provided", func() {
						It("calls CreateAndUploadBitsPackageByApplicationNameAndSpace with empty string", func() {
							Expect(testUI.Out).To(Say("Uploading and creating bits package for app %s in org %s / space %s as %s", app, orgName, spaceName, userName))

							Expect(fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceCallCount()).To(Equal(1))
							_, _, appPath := fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceArgsForCall(0)

							Expect(appPath).To(BeEmpty())
						})
					})

					Context("when getting streaming logs fails", func() {
						var expectedErr error
						BeforeEach(func() {
							expectedErr = errors.New("something is wrong!")
							fakeActor.GetStreamingLogsForApplicationByNameAndSpaceReturns(nil, nil, v3action.Warnings{"some-logging-warning", "some-other-logging-warning"}, expectedErr)
						})

						It("returns the error and displays warnings", func() {
							Expect(executeErr).To(Equal(expectedErr))

							Expect(testUI.Out).To(Say("Staging package for app %s in org some-org / space some-space as banana...", app))

							Expect(testUI.Err).To(Say("some-logging-warning"))
							Expect(testUI.Err).To(Say("some-other-logging-warning"))

						})
					})

					Context("when --no-start is provided", func() {
						BeforeEach(func() {
							cmd.NoStart = true
						})

						It("does not stage the package and returns", func() {
							Expect(testUI.Out).To(Say("Uploading and creating bits package for app %s in org %s / space %s as %s", app, orgName, spaceName, userName))
							Expect(fakeActor.CreateAndUploadBitsPackageByApplicationNameAndSpaceCallCount()).To(Equal(1))
							Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(0))

							Expect(executeErr).ToNot(HaveOccurred())
						})
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
									logStream <- v3action.NewLogMessage("not from staging", 1, time.Now(), "potato", "sourceInstance")
									allLogsWritten <- true
								}()

								return logStream, errorStream, v3action.Warnings{"steve for all I care"}, nil
							}
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
										warningsStream <- v3action.Warnings{"some-staging-warning", "some-other-staging-warning"}
										errorStream <- expectedErr
									}()

									return dropletStream, warningsStream, errorStream
								}
							})

							It("returns the error and displays warnings", func() {
								Expect(executeErr).To(Equal(expectedErr))

								Expect(testUI.Out).To(Say("Staging package for app %s in org some-org / space some-space as banana...", app))

								Expect(testUI.Err).To(Say("some-staging-warning"))
								Expect(testUI.Err).To(Say("some-other-staging-warning"))

								Expect(testUI.Out).ToNot(Say("Setting app some-app to droplet some-droplet-guid in org some-org / space some-space as banana..."))
							})
						})

						Context("when the staging is successful", func() {
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
										warningsStream <- v3action.Warnings{"some-staging-warning", "some-other-staging-warning"}
										dropletStream <- v3action.Droplet{GUID: "some-droplet-guid"}
									}()

									return dropletStream, warningsStream, errorStream
								}
							})

							It("outputs the staging message and warnings", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Staging package for app %s in org some-org / space some-space as banana...", app))
								Expect(testUI.Out).To(Say("OK"))

								Expect(testUI.Err).To(Say("some-staging-warning"))
								Expect(testUI.Err).To(Say("some-other-staging-warning"))
							})

							It("stages the package", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
								guidArg, _ := fakeActor.StagePackageArgsForCall(0)
								Expect(guidArg).To(Equal("some-guid"))
							})

							It("displays staging logs and their warnings", func() {
								Expect(testUI.Out).To(Say("Here are some staging logs!"))
								Expect(testUI.Out).To(Say("Here are some other staging logs!"))
								Expect(testUI.Out).ToNot(Say("not from staging"))

								Expect(testUI.Err).To(Say("steve for all I care"))

								Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(1))
								appName, spaceGUID, noaaClient := fakeActor.GetStreamingLogsForApplicationByNameAndSpaceArgsForCall(0)
								Expect(appName).To(Equal(app))
								Expect(spaceGUID).To(Equal("some-space-guid"))
								Expect(noaaClient).To(Equal(fakeNOAAClient))

								guidArg, _ := fakeActor.StagePackageArgsForCall(0)
								Expect(guidArg).To(Equal("some-guid"))
							})

							Context("when setting the droplet fails", func() {
								BeforeEach(func() {
									fakeActor.SetApplicationDropletReturns(v3action.Warnings{"droplet-warning-1", "droplet-warning-2"}, errors.New("some-error"))
								})

								It("returns the error", func() {
									Expect(executeErr).To(Equal(errors.New("some-error")))

									Expect(testUI.Out).To(Say("Setting app some-app to droplet some-droplet-guid in org some-org / space some-space as banana..."))

									Expect(testUI.Err).To(Say("droplet-warning-1"))
									Expect(testUI.Err).To(Say("droplet-warning-2"))

									Expect(testUI.Out).ToNot(Say("Starting app some-app in org some-org / space some-space as banana\\.\\.\\."))
								})
							})

							Context("when setting the application droplet is successful", func() {
								BeforeEach(func() {
									fakeActor.SetApplicationDropletReturns(v3action.Warnings{"droplet-warning-1", "droplet-warning-2"}, nil)
								})

								It("displays that the droplet was assigned", func() {
									Expect(testUI.Out).To(Say("Staging package for app %s in org some-org / space some-space as banana...", app))
									Expect(testUI.Out).To(Say("OK"))

									Expect(testUI.Out).ToNot(Say("Stopping .*"))

									Expect(testUI.Out).To(Say("Setting app some-app to droplet some-droplet-guid in org some-org / space some-space as banana..."))

									Expect(testUI.Err).To(Say("droplet-warning-1"))
									Expect(testUI.Err).To(Say("droplet-warning-2"))
									Expect(testUI.Out).To(Say("OK"))

									Expect(fakeActor.SetApplicationDropletCallCount()).To(Equal(1))
									appName, spaceGUID, dropletGUID := fakeActor.SetApplicationDropletArgsForCall(0)
									Expect(appName).To(Equal("some-app"))
									Expect(spaceGUID).To(Equal("some-space-guid"))
									Expect(dropletGUID).To(Equal("some-droplet-guid"))
								})

								Context("when --no-route flag is set to true", func() {
									BeforeEach(func() {
										cmd.NoRoute = true
									})

									It("does not create any routes", func() {
										Expect(fakeV2PushActor.CreateAndMapDefaultApplicationRouteCallCount()).To(Equal(0))

										Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
									})
								})

								Context("when buildpack(s) are provided via -b flag", func() {
									BeforeEach(func() {
										cmd.Buildpacks = []string{"some-buildpack"}
									})

									It("creates the app with the specified buildpack and prints the buildpack name in the summary", func() {
										Expect(fakeActor.CreateApplicationInSpaceCallCount()).To(Equal(1), "Expected CreateApplicationInSpace to be called once")
										createApp, createSpaceGUID := fakeActor.CreateApplicationInSpaceArgsForCall(0)
										Expect(createApp).To(Equal(v3action.Application{
											Name: "some-app",
											Lifecycle: v3action.AppLifecycle{
												Type: constant.BuildpackAppLifecycleType,
												Data: v3action.AppLifecycleData{
													Buildpacks: []string{"some-buildpack"},
												},
											},
										}))
										Expect(createSpaceGUID).To(Equal("some-space-guid"))
									})
								})

								Context("when a docker image is specified", func() {
									BeforeEach(func() {
										cmd.DockerImage.Path = "example.com/docker/docker/docker:docker"
									})

									It("creates the app with a docker lifecycle", func() {
										Expect(fakeActor.CreateApplicationInSpaceCallCount()).To(Equal(1), "Expected CreateApplicationInSpace to be called once")
										createApp, createSpaceGUID := fakeActor.CreateApplicationInSpaceArgsForCall(0)
										Expect(createApp).To(Equal(v3action.Application{
											Name: "some-app",
											Lifecycle: v3action.AppLifecycle{
												Type: constant.DockerAppLifecycleType,
												Data: v3action.AppLifecycleData{},
											},
										}))
										Expect(createSpaceGUID).To(Equal("some-space-guid"))
									})
								})

								Context("when mapping routes fails", func() {
									BeforeEach(func() {
										fakeV2PushActor.CreateAndMapDefaultApplicationRouteReturns(pushaction.Warnings{"route-warning"}, errors.New("some-error"))
									})

									It("returns the error", func() {
										Expect(executeErr).To(MatchError("some-error"))
										Expect(testUI.Out).To(Say("Mapping routes\\.\\.\\."))
										Expect(testUI.Err).To(Say("route-warning"))

										Expect(fakeActor.StartApplicationCallCount()).To(Equal(0))
									})
								})

								Context("when mapping routes succeeds", func() {
									BeforeEach(func() {
										fakeV2PushActor.CreateAndMapDefaultApplicationRouteReturns(pushaction.Warnings{"route-warning"}, nil)
									})

									It("displays the header and OK", func() {
										Expect(testUI.Out).To(Say("Mapping routes\\.\\.\\."))
										Expect(testUI.Out).To(Say("OK"))

										Expect(testUI.Err).To(Say("route-warning"))

										Expect(fakeV2PushActor.CreateAndMapDefaultApplicationRouteCallCount()).To(Equal(1), "Expected CreateAndMapDefaultApplicationRoute to be called")
										orgArg, spaceArg, appArg := fakeV2PushActor.CreateAndMapDefaultApplicationRouteArgsForCall(0)
										Expect(orgArg).To(Equal("some-org-guid"))
										Expect(spaceArg).To(Equal("some-space-guid"))
										Expect(appArg).To(Equal(v2action.Application{Name: "some-app", GUID: "some-app-guid"}))

										Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
									})

									Context("when starting the application fails", func() {
										BeforeEach(func() {
											fakeActor.StartApplicationReturns(v3action.Application{}, v3action.Warnings{"start-warning-1", "start-warning-2"}, errors.New("some-error"))
										})

										It("says that the app failed to start", func() {
											Expect(executeErr).To(Equal(errors.New("some-error")))
											Expect(testUI.Out).To(Say("Starting app some-app in org some-org / space some-space as banana\\.\\.\\."))

											Expect(testUI.Err).To(Say("start-warning-1"))
											Expect(testUI.Err).To(Say("start-warning-2"))

											Expect(testUI.Out).ToNot(Say("Showing health and status for app some-app in org some-org / space some-space as banana\\.\\.\\."))
										})
									})

									Context("when starting the application succeeds", func() {
										BeforeEach(func() {
											fakeActor.StartApplicationReturns(v3action.Application{GUID: "some-app-guid"}, v3action.Warnings{"start-warning-1", "start-warning-2"}, nil)
										})

										It("says that the app was started and outputs warnings", func() {
											Expect(testUI.Out).To(Say("Starting app some-app in org some-org / space some-space as banana\\.\\.\\."))

											Expect(testUI.Err).To(Say("start-warning-1"))
											Expect(testUI.Err).To(Say("start-warning-2"))
											Expect(testUI.Out).To(Say("OK"))

											Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
											appGUID := fakeActor.StartApplicationArgsForCall(0)
											Expect(appGUID).To(Equal("some-app-guid"))
										})
									})

									Context("when polling the start fails", func() {
										BeforeEach(func() {
											fakeActor.PollStartStub = func(appGUID string, warnings chan<- v3action.Warnings) error {
												warnings <- v3action.Warnings{"some-poll-warning-1", "some-poll-warning-2"}
												return errors.New("some-error")
											}
										})

										It("displays all warnings and fails", func() {
											Expect(testUI.Out).To(Say("Waiting for app to start\\.\\.\\."))

											Expect(testUI.Err).To(Say("some-poll-warning-1"))
											Expect(testUI.Err).To(Say("some-poll-warning-2"))

											Expect(executeErr).To(MatchError("some-error"))
										})
									})

									Context("when polling times out", func() {
										BeforeEach(func() {
											fakeActor.PollStartReturns(actionerror.StartupTimeoutError{})
										})

										It("returns the StartupTimeoutError", func() {
											Expect(executeErr).To(MatchError(translatableerror.StartupTimeoutError{
												AppName:    "some-app",
												BinaryName: binaryName,
											}))
										})
									})

									Context("when polling the start succeeds", func() {
										BeforeEach(func() {
											fakeActor.PollStartStub = func(appGUID string, warnings chan<- v3action.Warnings) error {
												warnings <- v3action.Warnings{"some-poll-warning-1", "some-poll-warning-2"}
												return nil
											}
										})

										It("displays all warnings", func() {
											Expect(testUI.Out).To(Say("Waiting for app to start\\.\\.\\."))

											Expect(testUI.Err).To(Say("some-poll-warning-1"))
											Expect(testUI.Err).To(Say("some-poll-warning-2"))

											Expect(executeErr).ToNot(HaveOccurred())
										})

										Context("when displaying the application info fails", func() {
											BeforeEach(func() {
												var expectedErr error
												expectedErr = actionerror.ApplicationNotFoundError{Name: app}
												fakeActor.GetApplicationSummaryByNameAndSpaceReturns(v3action.ApplicationSummary{}, v3action.Warnings{"display-warning-1", "display-warning-2"}, expectedErr)
											})

											It("returns the error and prints warnings", func() {
												Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))

												Expect(testUI.Out).To(Say("Showing health and status for app some-app in org some-org / space some-space as banana\\.\\.\\."))

												Expect(testUI.Err).To(Say("display-warning-1"))
												Expect(testUI.Err).To(Say("display-warning-2"))

												Expect(testUI.Out).ToNot(Say("name:\\s+some-app"))
											})
										})

										Context("when getting the application summary is successful", func() {
											BeforeEach(func() {
												summary := v3action.ApplicationSummary{
													Application: v3action.Application{
														Name:  "some-app",
														GUID:  "some-app-guid",
														State: "started",
													},
													CurrentDroplet: v3action.Droplet{
														Stack: "cflinuxfs2",
														Buildpacks: []v3action.Buildpack{
															{
																Name:         "ruby_buildpack",
																DetectOutput: "some-detect-output",
															},
														},
													},
													ProcessSummaries: []v3action.ProcessSummary{
														{
															Process: v3action.Process{
																Type:       "worker",
																MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
															},
															InstanceDetails: []v3action.ProcessInstance{
																v3action.ProcessInstance{
																	Index:       0,
																	State:       constant.ProcessInstanceRunning,
																	MemoryUsage: 4000000,
																	DiskUsage:   4000000,
																	MemoryQuota: 67108864,
																	DiskQuota:   8000000,
																	Uptime:      int(time.Now().Sub(time.Unix(1371859200, 0)).Seconds()),
																},
															},
														},
													},
												}

												fakeActor.GetApplicationSummaryByNameAndSpaceReturns(summary, v3action.Warnings{"display-warning-1", "display-warning-2"}, nil)
											})

											Context("when getting the application routes fails", func() {
												BeforeEach(func() {
													fakeV2AppActor.GetApplicationRoutesReturns([]v2action.Route{},
														v2action.Warnings{"route-warning-1", "route-warning-2"}, errors.New("some-error"))
												})

												It("displays all warnings and returns the error", func() {
													Expect(executeErr).To(MatchError("some-error"))

													Expect(testUI.Out).To(Say("Showing health and status for app some-app in org some-org / space some-space as banana\\.\\.\\."))

													Expect(testUI.Err).To(Say("display-warning-1"))
													Expect(testUI.Err).To(Say("display-warning-2"))
													Expect(testUI.Err).To(Say("route-warning-1"))
													Expect(testUI.Err).To(Say("route-warning-2"))

													Expect(testUI.Out).ToNot(Say("name:\\s+some-app"))
												})
											})

											Context("when getting the application routes is successful", func() {
												BeforeEach(func() {
													fakeV2AppActor.GetApplicationRoutesReturns([]v2action.Route{
														{Domain: v2action.Domain{Name: "some-other-domain"}}, {
															Domain: v2action.Domain{Name: "some-domain"}}},
														v2action.Warnings{"route-warning-1", "route-warning-2"}, nil)
												})

												It("prints the application summary and outputs warnings", func() {
													Expect(executeErr).ToNot(HaveOccurred())

													Expect(testUI.Out).To(Say("(?m)Showing health and status for app some-app in org some-org / space some-space as banana\\.\\.\\.\n\n"))
													Expect(testUI.Out).To(Say("name:\\s+some-app"))
													Expect(testUI.Out).To(Say("requested state:\\s+started"))
													Expect(testUI.Out).To(Say("processes:\\s+worker:1/1"))
													Expect(testUI.Out).To(Say("memory usage:\\s+64M x 1"))
													Expect(testUI.Out).To(Say("routes:\\s+some-other-domain, some-domain"))
													Expect(testUI.Out).To(Say("stack:\\s+cflinuxfs2"))
													Expect(testUI.Out).To(Say("(?m)buildpacks:\\s+some-detect-output\n\n"))

													Expect(testUI.Out).To(Say("worker:1/1"))
													Expect(testUI.Out).To(Say("\\s+state\\s+since\\s+cpu\\s+memory\\s+disk"))
													Expect(testUI.Out).To(Say("#0\\s+running\\s+2013-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M\\s+0.0%\\s+3.8M of 64M\\s+3.8M of 7.6M"))

													Expect(testUI.Err).To(Say("display-warning-1"))
													Expect(testUI.Err).To(Say("display-warning-2"))
													Expect(testUI.Err).To(Say("route-warning-1"))
													Expect(testUI.Err).To(Say("route-warning-2"))

													Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
													appName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
													Expect(appName).To(Equal("some-app"))
													Expect(spaceGUID).To(Equal("some-space-guid"))

													Expect(fakeV2AppActor.GetApplicationRoutesCallCount()).To(Equal(1))
													Expect(fakeV2AppActor.GetApplicationRoutesArgsForCall(0)).To(Equal("some-app-guid"))
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

		Context("when looking up the application succeeds", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{
					Name: "some-app",
					GUID: "some-app-guid",
				}, v3action.Warnings{"get-warning"}, nil)
			})

			It("updates the application", func() {
				Expect(fakeActor.CreateApplicationInSpaceCallCount()).To(Equal(0))
				Expect(fakeActor.UpdateApplicationCallCount()).To(Equal(1))
			})

			Context("when updating the application fails", func() {
				BeforeEach(func() {
					fakeActor.UpdateApplicationReturns(v3action.Application{}, v3action.Warnings{"update-warning-1"}, errors.New("some-error"))
				})

				It("returns the error and displays warnings", func() {
					Expect(executeErr).To(MatchError("some-error"))

					Expect(testUI.Err).To(Say("get-warning"))
					Expect(testUI.Err).To(Say("update-warning"))
				})
			})

			Context("when a docker image is provided", func() {
				BeforeEach(func() {
					cmd.DockerImage.Path = "example.com/docker/docker/docker:docker"
					cmd.DockerUsername = "username"
					fakeConfig.DockerPasswordReturns("password")
				})

				Context("when a username/password are provided", func() {
					It("updates the app with the provided credentials", func() {
						appName, spaceGuid, dockerImageCredentials := fakeActor.CreateDockerPackageByApplicationNameAndSpaceArgsForCall(0)
						Expect(appName).To(Equal("some-app"))
						Expect(spaceGuid).To(Equal("some-space-guid"))
						Expect(dockerImageCredentials.Path).To(Equal(cmd.DockerImage.Path))
						Expect(dockerImageCredentials.Username).To(Equal("username"))
						Expect(dockerImageCredentials.Password).To(Equal("password"))
					})
				})

				It("updates the app with a docker lifecycle", func() {
					Expect(fakeActor.UpdateApplicationCallCount()).To(Equal(1), "Expected UpdateApplication to be called once")
					updateApp := fakeActor.UpdateApplicationArgsForCall(0)
					Expect(updateApp).To(Equal(v3action.Application{
						GUID: "some-app-guid",
						Lifecycle: v3action.AppLifecycle{
							Type: constant.DockerAppLifecycleType,
							Data: v3action.AppLifecycleData{},
						},
					}))
				})
			})

			Context("when the app has a buildpack lifecycle", func() {
				Context("when a buildpack was not provided", func() {
					BeforeEach(func() {
						cmd.Buildpacks = []string{}
					})

					It("does not update the buildpack", func() {
						appArg := fakeActor.UpdateApplicationArgsForCall(0)
						Expect(appArg).To(Equal(v3action.Application{
							GUID: "some-app-guid",
							Lifecycle: v3action.AppLifecycle{
								Type: constant.BuildpackAppLifecycleType,
								Data: v3action.AppLifecycleData{
									Buildpacks: []string{},
								},
							},
						}))
					})
				})

				Context("when a buildpack was provided", func() {
					BeforeEach(func() {
						cmd.Buildpacks = []string{"some-buildpack"}
					})

					It("updates the buildpack", func() {
						appArg := fakeActor.UpdateApplicationArgsForCall(0)
						Expect(appArg).To(Equal(v3action.Application{
							GUID: "some-app-guid",
							Lifecycle: v3action.AppLifecycle{
								Type: constant.BuildpackAppLifecycleType,
								Data: v3action.AppLifecycleData{
									Buildpacks: []string{"some-buildpack"},
								},
							},
						}))
					})
				})

				Context("when multiple buildpacks are provided", func() {
					BeforeEach(func() {
						cmd.Buildpacks = []string{"some-buildpack-1", "some-buildpack-2"}
					})

					It("updates the buildpacks", func() {
						appArg := fakeActor.UpdateApplicationArgsForCall(0)
						Expect(appArg).To(Equal(v3action.Application{
							GUID: "some-app-guid",
							Lifecycle: v3action.AppLifecycle{
								Type: constant.BuildpackAppLifecycleType,
								Data: v3action.AppLifecycleData{
									Buildpacks: []string{"some-buildpack-1", "some-buildpack-2"},
								},
							},
						}))
					})

					Context("when default was also provided", func() {
						BeforeEach(func() {
							cmd.Buildpacks = []string{"default", "some-buildpack-2"}
						})

						It("returns the ConflictingBuildpacksError", func() {
							Expect(executeErr).To(Equal(translatableerror.ConflictingBuildpacksError{}))
							Expect(fakeActor.UpdateApplicationCallCount()).To(Equal(0))
						})
					})

					Context("when null was also provided", func() {
						BeforeEach(func() {
							cmd.Buildpacks = []string{"null", "some-buildpack-2"}
						})

						It("returns the ConflictingBuildpacksError", func() {
							Expect(executeErr).To(Equal(translatableerror.ConflictingBuildpacksError{}))
							Expect(fakeActor.UpdateApplicationCallCount()).To(Equal(0))
						})
					})
				})
			})

			Context("when updating the application succeeds", func() {
				Context("when the application is stopped", func() {
					BeforeEach(func() {
						fakeActor.UpdateApplicationReturns(v3action.Application{GUID: "some-app-guid", State: constant.ApplicationStopped}, v3action.Warnings{"update-warning"}, nil)
					})

					It("skips stopping the application and pushes it", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Err).To(Say("get-warning"))
						Expect(testUI.Err).To(Say("update-warning"))

						Expect(testUI.Out).ToNot(Say("Stopping"))

						Expect(fakeActor.StopApplicationCallCount()).To(Equal(0), "Expected StopApplication to not be called")

						Expect(fakeActor.StartApplicationCallCount()).To(Equal(1), "Expected StartApplication to be called")
					})
				})

				Context("when the application is started", func() {
					BeforeEach(func() {
						fakeActor.UpdateApplicationReturns(v3action.Application{GUID: "some-app-guid", State: constant.ApplicationStarted}, nil, nil)
					})

					It("stops the application and pushes it", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Stopping app some-app in org some-org / space some-space as banana..."))

						Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))

						Expect(fakeActor.StartApplicationCallCount()).To(Equal(1), "Expected StartApplication to be called")
					})

					Context("when no-start is provided", func() {
						BeforeEach(func() {
							cmd.NoStart = true
						})

						It("stops the application and returns", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).To(Say("Stopping app some-app in org some-org / space some-space as banana..."))

							Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))

							Expect(fakeActor.StartApplicationCallCount()).To(Equal(0))
						})
					})
				})
			})
		})
	})
})
