package v2_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
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
		cmd              V2PushCommand
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v2fakes.FakeV2PushActor
		fakeRestartActor *v2fakes.FakeRestartActor
		fakeProgressBar  *v2fakes.FakeProgressBar
		input            *Buffer
		binaryName       string

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
		fakeRestartActor = new(v2fakes.FakeRestartActor)
		fakeProgressBar = new(v2fakes.FakeProgressBar)

		cmd = V2PushCommand{
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			RestartActor: fakeRestartActor,
			ProgressBar:  fakeProgressBar,
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
			Expect(executeErr).To(MatchError(translatableerror.NotLoggedInError{BinaryName: binaryName}))

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
							CurrentApplication: v2action.Application{Name: appName, State: ccv2.ApplicationStarted},
							DesiredApplication: v2action.Application{Name: appName},
							CurrentRoutes: []v2action.Route{
								{
									Host: "route1",
									Domain: v2action.Domain{
										Name: "example.com",
									},
								},
								{
									Host: "route2",
									Domain: v2action.Domain{
										Name: "example.com",
									},
								},
							},
							DesiredRoutes: []v2action.Route{
								{
									Host: "route3",
									Domain: v2action.Domain{
										Name: "example.com",
									},
								},
								{
									Host: "route4",
									Domain: v2action.Domain{
										Name: "example.com",
									},
								},
							},
							TargetedSpaceGUID: "some-space-guid",
							Path:              pwd,
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

								Eventually(eventStream).Should(BeSent(pushaction.SettingUpApplication))
								Eventually(eventStream).Should(BeSent(pushaction.CreatedApplication))
								Eventually(eventStream).Should(BeSent(pushaction.UpdatedApplication))
								Eventually(eventStream).Should(BeSent(pushaction.ConfiguringRoutes))
								Eventually(eventStream).Should(BeSent(pushaction.CreatedRoutes))
								Eventually(eventStream).Should(BeSent(pushaction.BoundRoutes))
								Eventually(eventStream).Should(BeSent(pushaction.ResourceMatching))
								Eventually(eventStream).Should(BeSent(pushaction.CreatingArchive))
								Eventually(eventStream).Should(BeSent(pushaction.UploadingApplication))
								Eventually(fakeProgressBar.ReadyCallCount).Should(Equal(1))
								Eventually(eventStream).Should(BeSent(pushaction.RetryUpload))
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

						fakeRestartActor.RestartApplicationStub = func(app v2action.Application, client v2action.NOAAClient, config v2action.Config) (<-chan *v2action.LogMessage, <-chan error, <-chan v2action.ApplicationState, <-chan string, <-chan error) {
							messages := make(chan *v2action.LogMessage)
							logErrs := make(chan error)
							appState := make(chan v2action.ApplicationState)
							warnings := make(chan string)
							errs := make(chan error)

							go func() {
								messages <- v2action.NewLogMessage("log message 1", 1, time.Unix(0, 0), "STG", "1")
								messages <- v2action.NewLogMessage("log message 2", 1, time.Unix(0, 0), "STG", "1")
								appState <- v2action.ApplicationStateStopping
								appState <- v2action.ApplicationStateStaging
								appState <- v2action.ApplicationStateStarting
								close(messages)
								close(logErrs)
								close(appState)
								close(warnings)
								close(errs)
							}()

							return messages, logErrs, appState, warnings, errs
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

						fakeRestartActor.GetApplicationSummaryByNameAndSpaceReturns(applicationSummary, warnings, nil)
					})

					Context("when no manifest is provided", func() {
						It("passes through the command line flags", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(fakeActor.MergeAndValidateSettingsAndManifestsCallCount()).To(Equal(1))
							cmdSettings, _ := fakeActor.MergeAndValidateSettingsAndManifestsArgsForCall(0)
							Expect(cmdSettings).To(Equal(pushaction.CommandLineSettings{
								Name:             appName,
								CurrentDirectory: pwd,
							}))
						})
					})

					Context("when a manifest is provided", func() {
						var (
							tmpDir         string
							pathToManifest string

							originalDir string
						)

						BeforeEach(func() {
							var err error
							tmpDir, err = ioutil.TempDir("", "v2-push-command-test")
							Expect(err).ToNot(HaveOccurred())

							// OS X uses weird symlinks that causes problems for some tests
							tmpDir, err = filepath.EvalSymlinks(tmpDir)
							Expect(err).ToNot(HaveOccurred())

							originalDir, err = os.Getwd()
							Expect(err).ToNot(HaveOccurred())

							cmd.OptionalArgs.AppName = ""
						})

						AfterEach(func() {
							Expect(os.Chdir(originalDir)).ToNot(HaveOccurred())
							Expect(os.RemoveAll(tmpDir)).ToNot(HaveOccurred())
						})

						Context("via a manfiest.yml in the current directory", func() {
							var expectedApps []manifest.Application

							BeforeEach(func() {
								err := os.Chdir(tmpDir)
								Expect(err).ToNot(HaveOccurred())

								pathToManifest = filepath.Join(tmpDir, "manifest.yml")
								err = ioutil.WriteFile(pathToManifest, []byte("some manfiest file"), 0666)
								Expect(err).ToNot(HaveOccurred())

								expectedApps = []manifest.Application{{Name: "some-app"}, {Name: "some-other-app"}}
								fakeActor.ReadManifestReturns(expectedApps, nil)
							})

							Context("when reading the manifest file is successful", func() {
								It("merges app manifest and flags", func() {
									Expect(executeErr).ToNot(HaveOccurred())

									Expect(fakeActor.ReadManifestCallCount()).To(Equal(1))
									Expect(fakeActor.ReadManifestArgsForCall(0)).To(Equal(pathToManifest))

									Expect(fakeActor.MergeAndValidateSettingsAndManifestsCallCount()).To(Equal(1))
									cmdSettings, manifestApps := fakeActor.MergeAndValidateSettingsAndManifestsArgsForCall(0)
									Expect(cmdSettings).To(Equal(pushaction.CommandLineSettings{
										CurrentDirectory: tmpDir,
									}))
									Expect(manifestApps).To(Equal(expectedApps))
								})
							})

							Context("when reading manifest file errors", func() {
								var expectedErr error

								BeforeEach(func() {
									expectedErr = errors.New("I am an error!!!")

									fakeActor.ReadManifestReturns(nil, expectedErr)
								})

								It("returns the error", func() {
									Expect(executeErr).To(MatchError(expectedErr))
								})
							})

							Context("when --no-manifest is specified", func() {
								BeforeEach(func() {
									cmd.NoManifest = true
								})

								It("ignores the manifest file", func() {
									Expect(executeErr).ToNot(HaveOccurred())

									Expect(fakeActor.MergeAndValidateSettingsAndManifestsCallCount()).To(Equal(1))
									cmdSettings, manifestApps := fakeActor.MergeAndValidateSettingsAndManifestsArgsForCall(0)
									Expect(cmdSettings).To(Equal(pushaction.CommandLineSettings{
										CurrentDirectory: tmpDir,
									}))
									Expect(manifestApps).To(BeNil())
								})
							})
						})

						Context("via a manfiest.yaml in the current directory", func() {
							BeforeEach(func() {
								err := os.Chdir(tmpDir)
								Expect(err).ToNot(HaveOccurred())

								pathToManifest = filepath.Join(tmpDir, "manifest.yaml")
								err = ioutil.WriteFile(pathToManifest, []byte("some manfiest file"), 0666)
								Expect(err).ToNot(HaveOccurred())
							})

							It("should read the manifest.yml", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(fakeActor.ReadManifestCallCount()).To(Equal(1))
								Expect(fakeActor.ReadManifestArgsForCall(0)).To(Equal(pathToManifest))
							})
						})

						Context("via the -f flag", func() {
							BeforeEach(func() {
								pathToManifest = filepath.Join(tmpDir, "manifest.yaml")
								err := ioutil.WriteFile(pathToManifest, []byte("some manfiest file"), 0666)
								Expect(err).ToNot(HaveOccurred())

								cmd.PathToManifest = flag.PathWithExistenceCheck(pathToManifest)
							})

							It("should read the manifest.yml", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(fakeActor.ReadManifestCallCount()).To(Equal(1))
								Expect(fakeActor.ReadManifestArgsForCall(0)).To(Equal(pathToManifest))
							})
						})
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
						Expect(testUI.Out).To(Say("Getting app info\\.\\.\\."))
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

						Expect(testUI.Out).To(Say("Creating app with these attributes\\.\\.\\."))
						Expect(testUI.Out).To(Say("Mapping routes\\.\\.\\."))
						Expect(testUI.Out).To(Say("Checking for existing files on server\\.\\.\\."))
						Expect(testUI.Out).To(Say("Packaging files to upload\\.\\.\\."))
						Expect(testUI.Out).To(Say("Uploading files\\.\\.\\."))
						Expect(testUI.Out).To(Say("Retrying upload due to an error\\.\\.\\."))
						Expect(testUI.Out).To(Say("Waiting for API to complete processing files\\.\\.\\."))
						Expect(testUI.Out).To(Say("Stopping app\\.\\.\\."))

						Expect(testUI.Err).To(Say("some-config-warnings"))
						Expect(testUI.Err).To(Say("apply-1"))
						Expect(testUI.Err).To(Say("apply-2"))
					})

					It("displays app staging logs", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("log message 1"))
						Expect(testUI.Out).To(Say("log message 2"))

						Expect(fakeRestartActor.RestartApplicationCallCount()).To(Equal(1))
						appConfig, _, _ := fakeRestartActor.RestartApplicationArgsForCall(0)
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

					Context("when pushing app bits", func() {
						It("display diff of changes with path", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("\\s+name:\\s+%s", appName))
							Expect(testUI.Out).To(Say("\\s+path:\\s+%s", regexp.QuoteMeta(appConfigs[0].Path)))
							Expect(testUI.Out).To(Say("\\s+routes:"))
							for _, route := range appConfigs[0].CurrentRoutes {
								Expect(testUI.Out).To(Say(route.String()))
							}
							for _, route := range appConfigs[0].DesiredRoutes {
								Expect(testUI.Out).To(Say(route.String()))
							}
						})
					})

					Context("when pushing a docker image", func() {
						var docker string
						BeforeEach(func() {
							docker = "some-path"
							appConfigs[0].DesiredApplication.DockerImage = docker
						})

						It("display diff of changes with docker image", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("\\s+name:\\s+%s", appName))
							Expect(testUI.Out).To(Say("\\s+docker image:\\s+%s", docker))
							Expect(testUI.Out).To(Say("\\s+routes:"))
							for _, route := range appConfigs[0].CurrentRoutes {
								Expect(testUI.Out).To(Say(route.String()))
							}
							for _, route := range appConfigs[0].DesiredRoutes {
								Expect(testUI.Out).To(Say(route.String()))
							}
						})
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

	Describe("GetCommandLineSettings", func() {
		Context("when the -o and -p flags are both given", func() {
			BeforeEach(func() {
				cmd.DockerImage.Path = "some-docker-image"
				cmd.AppPath = "some-directory-path"
			})

			It("returns an error", func() {
				_, err := cmd.GetCommandLineSettings()
				Expect(err).To(MatchError(translatableerror.ArgumentCombinationError{
					Arg1: "--docker-image, -o",
					Arg2: "-p",
				}))
			})
		})

		Context("when only -f and --no-manifest flags are passed", func() {
			BeforeEach(func() {
				cmd.PathToManifest = "/some/path.yml"
				cmd.NoManifest = true
			})

			It("returns an ArgumentCombinationError", func() {
				_, err := cmd.GetCommandLineSettings()
				Expect(err).To(MatchError(translatableerror.ArgumentCombinationError{
					Arg1: "-f",
					Arg2: "--no-manifest",
				}))
			})
		})

		Context("when only -o flag is passed", func() {
			BeforeEach(func() {
				cmd.DockerImage.Path = "some-docker-image-path"
			})

			It("creates command line setting from command line arguments", func() {
				settings, err := cmd.GetCommandLineSettings()
				Expect(err).ToNot(HaveOccurred())
				Expect(settings.Name).To(Equal(appName))
				Expect(settings.DockerImage).To(Equal("some-docker-image-path"))
			})
		})

		Context("when only -p flag is passed", func() {
			BeforeEach(func() {
				cmd.AppPath = "some-directory-path"
			})

			It("creates command line setting from command line arguments", func() {
				settings, err := cmd.GetCommandLineSettings()
				Expect(err).ToNot(HaveOccurred())
				Expect(settings.Name).To(Equal(appName))
				Expect(settings.ProvidedAppPath).To(Equal("some-directory-path"))
			})
		})
	})
})
