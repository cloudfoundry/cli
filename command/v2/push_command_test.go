package v2_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/manifest"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

	Context("Execute", func() {
		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

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
								CurrentApplication: pushaction.Application{Application: v2action.Application{Name: appName, State: ccv2.ApplicationStarted}},
								DesiredApplication: pushaction.Application{Application: v2action.Application{Name: appName}},
								CurrentRoutes: []v2action.Route{
									{Host: "route1", Domain: v2action.Domain{Name: "example.com"}},
									{Host: "route2", Domain: v2action.Domain{Name: "example.com"}},
								},
								DesiredRoutes: []v2action.Route{
									{Host: "route3", Domain: v2action.Domain{Name: "example.com"}},
									{Host: "route4", Domain: v2action.Domain{Name: "example.com"}},
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
									CurrentApplication: pushaction.Application{Application: v2action.Application{Name: appName, GUID: "some-app-guid"}},
									DesiredApplication: pushaction.Application{Application: v2action.Application{Name: appName, GUID: "some-app-guid"}},
									TargetedSpaceGUID:  "some-space-guid",
									Path:               pwd,
								}

								go func() {
									defer GinkgoRecover()

									Eventually(eventStream).Should(BeSent(pushaction.SettingUpApplication))
									Eventually(eventStream).Should(BeSent(pushaction.CreatedApplication))
									Eventually(eventStream).Should(BeSent(pushaction.UpdatedApplication))
									Eventually(eventStream).Should(BeSent(pushaction.CreatingAndMappingRoutes))
									Eventually(eventStream).Should(BeSent(pushaction.CreatedRoutes))
									Eventually(eventStream).Should(BeSent(pushaction.BoundRoutes))
									Eventually(eventStream).Should(BeSent(pushaction.UnmappingRoutes))
									Eventually(eventStream).Should(BeSent(pushaction.ConfiguringServices))
									Eventually(eventStream).Should(BeSent(pushaction.BoundServices))
									Eventually(eventStream).Should(BeSent(pushaction.ResourceMatching))
									Eventually(eventStream).Should(BeSent(pushaction.UploadingApplication))
									Eventually(eventStream).Should(BeSent(pushaction.CreatingArchive))
									Eventually(eventStream).Should(BeSent(pushaction.UploadingApplicationWithArchive))
									Eventually(fakeProgressBar.ReadyCallCount).Should(Equal(1))
									Eventually(eventStream).Should(BeSent(pushaction.RetryUpload))
									Eventually(eventStream).Should(BeSent(pushaction.UploadWithArchiveComplete))
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

							fakeRestartActor.RestartApplicationStub = func(app v2action.Application, client v2action.NOAAClient, config v2action.Config) (<-chan *v2action.LogMessage, <-chan error, <-chan v2action.ApplicationStateChange, <-chan string, <-chan error) {
								messages := make(chan *v2action.LogMessage)
								logErrs := make(chan error)
								appState := make(chan v2action.ApplicationStateChange)
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
									DetectedBuildpack:    types.FilteredString{IsSet: true, Value: "some-buildpack"},
									DetectedStartCommand: types.FilteredString{IsSet: true, Value: "some start command"},
									GUID:                 "some-app-guid",
									Instances:            types.NullInt{Value: 3, IsSet: true},
									Memory:               types.NullByteSizeInMb{IsSet: true, Value: 128},
									Name:                 appName,
									PackageUpdatedAt:     time.Unix(0, 0),
									State:                "STARTED",
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
										Port: types.NullInt{IsSet: true, Value: 13},
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
								tmpDir       string
								providedPath string

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

							Context("via a manifest.yml in the current directory", func() {
								var expectedApps []manifest.Application

								BeforeEach(func() {
									err := os.Chdir(tmpDir)
									Expect(err).ToNot(HaveOccurred())

									providedPath = filepath.Join(tmpDir, "manifest.yml")
									err = ioutil.WriteFile(providedPath, []byte("some manifest file"), 0666)
									Expect(err).ToNot(HaveOccurred())

									expectedApps = []manifest.Application{{Name: "some-app"}, {Name: "some-other-app"}}
									fakeActor.ReadManifestReturns(expectedApps, nil)
								})

								Context("when reading the manifest file is successful", func() {
									It("merges app manifest and flags", func() {
										Expect(executeErr).ToNot(HaveOccurred())

										Expect(fakeActor.ReadManifestCallCount()).To(Equal(1))
										Expect(fakeActor.ReadManifestArgsForCall(0)).To(Equal(providedPath))

										Expect(fakeActor.MergeAndValidateSettingsAndManifestsCallCount()).To(Equal(1))
										cmdSettings, manifestApps := fakeActor.MergeAndValidateSettingsAndManifestsArgsForCall(0)
										Expect(cmdSettings).To(Equal(pushaction.CommandLineSettings{
											CurrentDirectory: tmpDir,
										}))
										Expect(manifestApps).To(Equal(expectedApps))
									})

									It("outputs corresponding flavor text", func() {
										Expect(executeErr).ToNot(HaveOccurred())

										Expect(testUI.Out).To(Say("Pushing from manifest to org some-org / space some-space as some-user\\.\\.\\."))
										Expect(testUI.Out).To(Say("Using manifest file %s", regexp.QuoteMeta(providedPath)))
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

							Context("via a manifest.yaml in the current directory", func() {
								BeforeEach(func() {
									err := os.Chdir(tmpDir)
									Expect(err).ToNot(HaveOccurred())

									providedPath = filepath.Join(tmpDir, "manifest.yaml")
									err = ioutil.WriteFile(providedPath, []byte("some manifest file"), 0666)
									Expect(err).ToNot(HaveOccurred())
								})

								It("should read the manifest.yml", func() {
									Expect(executeErr).ToNot(HaveOccurred())

									Expect(fakeActor.ReadManifestCallCount()).To(Equal(1))
									Expect(fakeActor.ReadManifestArgsForCall(0)).To(Equal(providedPath))
								})
							})

							Context("via the -f flag", func() {
								Context("given a path with filename 'manifest.yml'", func() {
									BeforeEach(func() {
										providedPath = filepath.Join(tmpDir, "manifest.yml")
									})

									Context("when the manifest.yml file does not exist", func() {
										BeforeEach(func() {
											cmd.PathToManifest = flag.PathWithExistenceCheck(providedPath)
										})

										It("returns an error", func() {
											Expect(os.IsNotExist(executeErr)).To(BeTrue())

											Expect(testUI.Out).ToNot(Say("Pushing from manifest"))
											Expect(testUI.Out).ToNot(Say("Using manifest file"))

											Expect(fakeActor.ReadManifestCallCount()).To(Equal(0))
										})
									})

									Context("when the manifest.yml file exists", func() {
										BeforeEach(func() {
											err := ioutil.WriteFile(providedPath, []byte(`key: "value"`), 0666)
											Expect(err).ToNot(HaveOccurred())

											cmd.PathToManifest = flag.PathWithExistenceCheck(providedPath)
										})

										It("should read the manifest.yml file and outputs corresponding flavor text", func() {
											Expect(executeErr).ToNot(HaveOccurred())

											Expect(testUI.Out).To(Say("Pushing from manifest to org some-org / space some-space as some-user\\.\\.\\."))
											Expect(testUI.Out).To(Say("Using manifest file %s", regexp.QuoteMeta(providedPath)))

											Expect(fakeActor.ReadManifestCallCount()).To(Equal(1))
											Expect(fakeActor.ReadManifestArgsForCall(0)).To(Equal(providedPath))
										})
									})
								})

								Context("given a path that is a directory", func() {

									var (
										ymlFile  string
										yamlFile string
									)

									BeforeEach(func() {
										providedPath = tmpDir
										cmd.PathToManifest = flag.PathWithExistenceCheck(providedPath)
									})

									Context("when the directory does not contain a 'manifest.y{a}ml' file", func() {
										It("returns an error", func() {
											Expect(executeErr).To(MatchError(translatableerror.ManifestFileNotFoundInDirectoryError{PathToManifest: providedPath}))
											Expect(testUI.Out).ToNot(Say("Pushing from manifest"))
											Expect(testUI.Out).ToNot(Say("Using manifest file"))

											Expect(fakeActor.ReadManifestCallCount()).To(Equal(0))
										})
									})

									Context("when the directory contains a 'manifest.yml' file", func() {
										BeforeEach(func() {
											ymlFile = filepath.Join(providedPath, "manifest.yml")
											err := ioutil.WriteFile(ymlFile, []byte(`key: "value"`), 0666)
											Expect(err).ToNot(HaveOccurred())
										})

										It("should read the manifest.yml file and outputs corresponding flavor text", func() {
											Expect(executeErr).ToNot(HaveOccurred())

											Expect(testUI.Out).To(Say("Pushing from manifest to org some-org / space some-space as some-user\\.\\.\\."))
											Expect(testUI.Out).To(Say("Using manifest file %s", regexp.QuoteMeta(ymlFile)))

											Expect(fakeActor.ReadManifestCallCount()).To(Equal(1))
											Expect(fakeActor.ReadManifestArgsForCall(0)).To(Equal(ymlFile))
										})
									})

									Context("when the directory contains a 'manifest.yaml' file", func() {
										BeforeEach(func() {
											yamlFile = filepath.Join(providedPath, "manifest.yaml")
											err := ioutil.WriteFile(yamlFile, []byte(`key: "value"`), 0666)
											Expect(err).ToNot(HaveOccurred())
										})

										It("should read the manifest.yaml file and outputs corresponding flavor text", func() {
											Expect(executeErr).ToNot(HaveOccurred())

											Expect(testUI.Out).To(Say("Pushing from manifest to org some-org / space some-space as some-user\\.\\.\\."))
											Expect(testUI.Out).To(Say("Using manifest file %s", regexp.QuoteMeta(yamlFile)))

											Expect(fakeActor.ReadManifestCallCount()).To(Equal(1))
											Expect(fakeActor.ReadManifestArgsForCall(0)).To(Equal(yamlFile))
										})
									})

									Context("when the directory contains both a 'manifest.yml' and 'manifest.yaml' file", func() {
										BeforeEach(func() {
											ymlFile = filepath.Join(providedPath, "manifest.yml")
											err := ioutil.WriteFile(ymlFile, []byte(`key: "value"`), 0666)
											Expect(err).ToNot(HaveOccurred())

											yamlFile = filepath.Join(providedPath, "manifest.yaml")
											err = ioutil.WriteFile(yamlFile, []byte(`key: "value"`), 0666)
											Expect(err).ToNot(HaveOccurred())
										})

										It("should read the manifest.yml file and outputs corresponding flavor text", func() {
											Expect(executeErr).ToNot(HaveOccurred())

											Expect(testUI.Out).To(Say("Pushing from manifest to org some-org / space some-space as some-user\\.\\.\\."))
											Expect(testUI.Out).To(Say("Using manifest file %s", regexp.QuoteMeta(ymlFile)))

											Expect(fakeActor.ReadManifestCallCount()).To(Equal(1))
											Expect(fakeActor.ReadManifestArgsForCall(0)).To(Equal(ymlFile))
										})
									})
								})
							})
						})

						Context("when an app name and manifest are provided", func() {
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

								pathToManifest = filepath.Join(tmpDir, "manifest.yml")
								err = ioutil.WriteFile(pathToManifest, []byte("some manfiest file"), 0666)
								Expect(err).ToNot(HaveOccurred())

								originalDir, err = os.Getwd()
								Expect(err).ToNot(HaveOccurred())

								err = os.Chdir(tmpDir)
								Expect(err).ToNot(HaveOccurred())
							})

							AfterEach(func() {
								Expect(os.Chdir(originalDir)).ToNot(HaveOccurred())
								Expect(os.RemoveAll(tmpDir)).ToNot(HaveOccurred())
							})

							It("outputs corresponding flavor text", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Pushing from manifest to org some-org / space some-space as some-user\\.\\.\\."))
								Expect(testUI.Out).To(Say("Using manifest file %s", regexp.QuoteMeta(pathToManifest)))
							})
						})

						It("converts the manifests to app configs and outputs config warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Err).To(Say("some-config-warnings"))

							Expect(fakeActor.ConvertToApplicationConfigsCallCount()).To(Equal(1))
							orgGUID, spaceGUID, noStart, manifests := fakeActor.ConvertToApplicationConfigsArgsForCall(0)
							Expect(orgGUID).To(Equal("some-org-guid"))
							Expect(spaceGUID).To(Equal("some-space-guid"))
							Expect(noStart).To(BeFalse())
							Expect(manifests).To(Equal(appManifests))
						})

						It("outputs flavor text prior to generating app configuration", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).To(Say("Pushing app %s to org some-org / space some-space as some-user", appName))
							Expect(testUI.Out).To(Say("Getting app info\\.\\.\\."))
						})

						It("applies each of the application configurations", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(fakeActor.ApplyCallCount()).To(Equal(1))
							config, progressBar := fakeActor.ApplyArgsForCall(0)
							Expect(config).To(Equal(appConfigs[0]))
							Expect(progressBar).To(Equal(fakeProgressBar))
						})

						It("display diff of changes", func() {
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

						Context("when the app starts", func() {
							It("displays app events and warnings", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Creating app with these attributes\\.\\.\\."))
								Expect(testUI.Out).To(Say("Mapping routes\\.\\.\\."))
								Expect(testUI.Out).To(Say("Unmapping routes\\.\\.\\."))
								Expect(testUI.Out).To(Say("Binding services\\.\\.\\."))
								Expect(testUI.Out).To(Say("Comparing local files to remote cache\\.\\.\\."))
								Expect(testUI.Out).To(Say("All files found in remote cache; nothing to upload."))
								Expect(testUI.Out).To(Say("Waiting for API to complete processing files\\.\\.\\."))
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
								Expect(appConfig).To(Equal(updatedConfig.CurrentApplication.Application))
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

							Context("when the start command is explicitly set", func() {
								BeforeEach(func() {
									applicationSummary := v2action.ApplicationSummary{
										Application: v2action.Application{
											Command:              types.FilteredString{IsSet: true, Value: "a-different-start-command"},
											DetectedBuildpack:    types.FilteredString{IsSet: true, Value: "some-buildpack"},
											DetectedStartCommand: types.FilteredString{IsSet: true, Value: "some start command"},
											GUID:                 "some-app-guid",
											Instances:            types.NullInt{Value: 3, IsSet: true},
											Memory:               types.NullByteSizeInMb{IsSet: true, Value: 128},
											Name:                 appName,
											PackageUpdatedAt:     time.Unix(0, 0),
											State:                "STARTED",
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
												Port: types.NullInt{IsSet: true, Value: 13},
											},
										},
									}
									warnings := []string{"app-summary-warning"}

									applicationSummary.RunningInstances = []v2action.ApplicationInstanceWithStats{{State: "RUNNING"}}

									fakeRestartActor.GetApplicationSummaryByNameAndSpaceReturns(applicationSummary, warnings, nil)
								})

								It("displays the correct start command", func() {
									Expect(executeErr).ToNot(HaveOccurred())
									Expect(testUI.Out).To(Say("name:\\s+%s", appName))
									Expect(testUI.Out).To(Say("start command:\\s+a-different-start-command"))
								})
							})
						})

						Context("when no-start is set", func() {
							BeforeEach(func() {
								cmd.NoStart = true

								applicationSummary := v2action.ApplicationSummary{
									Application: v2action.Application{
										Command:              types.FilteredString{IsSet: true, Value: "a-different-start-command"},
										DetectedBuildpack:    types.FilteredString{IsSet: true, Value: "some-buildpack"},
										DetectedStartCommand: types.FilteredString{IsSet: true, Value: "some start command"},
										GUID:                 "some-app-guid",
										Instances:            types.NullInt{Value: 3, IsSet: true},
										Memory:               types.NullByteSizeInMb{IsSet: true, Value: 128},
										Name:                 appName,
										PackageUpdatedAt:     time.Unix(0, 0),
										State:                "STOPPED",
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
											Port: types.NullInt{IsSet: true, Value: 13},
										},
									},
								}
								warnings := []string{"app-summary-warning"}

								fakeRestartActor.GetApplicationSummaryByNameAndSpaceReturns(applicationSummary, warnings, nil)
							})

							Context("when the app is not running", func() {
								It("does not start the app", func() {
									Expect(executeErr).ToNot(HaveOccurred())
									Expect(testUI.Out).To(Say("Waiting for API to complete processing files\\.\\.\\."))
									Expect(testUI.Out).To(Say("name:\\s+%s", appName))
									Expect(testUI.Out).To(Say("requested state:\\s+stopped"))

									Expect(fakeRestartActor.RestartApplicationCallCount()).To(Equal(0))
								})
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

			Context("when general app settings are given", func() {
				BeforeEach(func() {
					cmd.Buildpack = flag.Buildpack{FilteredString: types.FilteredString{Value: "some-buildpack", IsSet: true}}
					cmd.Command = flag.Command{FilteredString: types.FilteredString{IsSet: true, Value: "echo foo bar baz"}}
					cmd.DiskQuota = flag.Megabytes{NullUint64: types.NullUint64{Value: 1024, IsSet: true}}
					cmd.HealthCheckTimeout = 14
					cmd.HealthCheckType = flag.HealthCheckType{Type: "http"}
					cmd.Instances = flag.Instances{NullInt: types.NullInt{Value: 12, IsSet: true}}
					cmd.Memory = flag.Megabytes{NullUint64: types.NullUint64{Value: 100, IsSet: true}}
					cmd.StackName = "some-stack"
				})

				It("sets them on the command line settings", func() {
					Expect(commandLineSettingsErr).ToNot(HaveOccurred())
					Expect(settings.Buildpack).To(Equal(types.FilteredString{Value: "some-buildpack", IsSet: true}))
					Expect(settings.Command).To(Equal(types.FilteredString{IsSet: true, Value: "echo foo bar baz"}))
					Expect(settings.DiskQuota).To(Equal(uint64(1024)))
					Expect(settings.HealthCheckTimeout).To(Equal(14))
					Expect(settings.HealthCheckType).To(Equal("http"))
					Expect(settings.Instances).To(Equal(types.NullInt{Value: 12, IsSet: true}))
					Expect(settings.Memory).To(Equal(uint64(100)))
					Expect(settings.StackName).To(Equal("some-stack"))
				})
			})

			Context("route related flags", func() {
				Context("when given customed route settings", func() {
					BeforeEach(func() {
						cmd.Domain = "some-domain"
					})

					It("sets NoHostname on the command line settings", func() {
						Expect(settings.DefaultRouteDomain).To(Equal("some-domain"))
					})
				})

				Context("when --hostname is given", func() {
					BeforeEach(func() {
						cmd.Hostname = "some-hostname"
					})

					It("sets DefaultRouteHostname on the command line settings", func() {
						Expect(settings.DefaultRouteHostname).To(Equal("some-hostname"))
					})
				})

				Context("when --no-hostname is given", func() {
					BeforeEach(func() {
						cmd.NoHostname = true
					})

					It("sets NoHostname on the command line settings", func() {
						Expect(settings.NoHostname).To(BeTrue())
					})
				})

				Context("when --random-route is given", func() {
					BeforeEach(func() {
						cmd.RandomRoute = true
					})

					It("sets --random-route on the command line settings", func() {
						Expect(commandLineSettingsErr).ToNot(HaveOccurred())
						Expect(settings.RandomRoute).To(BeTrue())
					})
				})

				Context("when --route-path is given", func() {
					BeforeEach(func() {
						cmd.RoutePath = flag.RoutePath{Path: "/some-path"}
					})

					It("sets --route-path on the command line settings", func() {
						Expect(commandLineSettingsErr).ToNot(HaveOccurred())
						Expect(settings.RoutePath).To(Equal("/some-path"))
					})
				})

				Context("when --no-route is given", func() {
					BeforeEach(func() {
						cmd.NoRoute = true
					})

					It("sets NoRoute on the command line settings", func() {
						Expect(settings.NoRoute).To(BeTrue())
					})
				})
			})

			Context("app bits", func() {
				Context("when -p flag is given", func() {
					BeforeEach(func() {
						cmd.AppPath = "some-directory-path"
					})

					It("sets ProvidedAppPath", func() {
						Expect(settings.ProvidedAppPath).To(Equal("some-directory-path"))
					})
				})

				Context("when the -o flag is given", func() {
					BeforeEach(func() {
						cmd.DockerImage.Path = "some-docker-image-path"
					})

					It("creates command line setting from command line arguments", func() {
						Expect(settings.DockerImage).To(Equal("some-docker-image-path"))
					})

					Context("--docker-username flags is given", func() {
						BeforeEach(func() {
							cmd.DockerUsername = "some-docker-username"
						})

						Context("the docker password environment variable is set", func() {
							BeforeEach(func() {
								fakeConfig.DockerPasswordReturns("some-docker-password")
							})

							It("creates command line setting from command line arguments and config", func() {
								Expect(testUI.Out).To(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))

								Expect(settings.Name).To(Equal(appName))
								Expect(settings.DockerImage).To(Equal("some-docker-image-path"))
								Expect(settings.DockerUsername).To(Equal("some-docker-username"))
								Expect(settings.DockerPassword).To(Equal("some-docker-password"))
							})
						})

						Context("the docker password environment variable is *not* set", func() {
							BeforeEach(func() {
								input.Write([]byte("some-docker-password\n"))
							})

							It("prompts the user for a password", func() {
								Expect(testUI.Out).To(Say("Environment variable CF_DOCKER_PASSWORD not set."))
								Expect(testUI.Out).To(Say("Docker password"))

								Expect(settings.Name).To(Equal(appName))
								Expect(settings.DockerImage).To(Equal("some-docker-image-path"))
								Expect(settings.DockerUsername).To(Equal("some-docker-username"))
								Expect(settings.DockerPassword).To(Equal("some-docker-password"))
							})
						})
					})
				})
			})
		})

		DescribeTable("validation errors when flags are passed",
			func(setup func(), expectedErr error) {
				setup()
				_, commandLineSettingsErr := cmd.GetCommandLineSettings()
				Expect(commandLineSettingsErr).To(MatchError(expectedErr))
			},

			Entry("-o and -p",
				func() {
					cmd.DockerImage.Path = "some-docker-image"
					cmd.AppPath = "some-directory-path"
				},
				translatableerror.ArgumentCombinationError{Args: []string{"--docker-image, -o", "-p"}}),

			Entry("-b and --docker-image",
				func() {
					cmd.DockerImage.Path = "some-docker-image"
					cmd.Buildpack = flag.Buildpack{FilteredString: types.FilteredString{Value: "some-buildpack", IsSet: true}}
				},
				translatableerror.ArgumentCombinationError{Args: []string{"-b", "--docker-image, -o"}}),

			Entry("--docker-username (without DOCKER_PASSWORD env set)",
				func() {
					cmd.DockerUsername = "some-docker-username"
				},
				translatableerror.RequiredFlagsError{Arg1: "--docker-image, -o", Arg2: "--docker-username"}),

			Entry("-d and --no-route",
				func() {
					cmd.Domain = "some-domain"
					cmd.NoRoute = true
				},
				translatableerror.ArgumentCombinationError{Args: []string{"-d", "--no-route"}}),

			Entry("--hostname and --no-hostname",
				func() {
					cmd.Hostname = "po-tate-toe"
					cmd.NoHostname = true
				},
				translatableerror.ArgumentCombinationError{Args: []string{"--hostname", "-n", "--no-hostname"}}),

			Entry("--hostname and --no-route",
				func() {
					cmd.Hostname = "po-tate-toe"
					cmd.NoRoute = true
				},
				translatableerror.ArgumentCombinationError{Args: []string{"--hostname", "-n", "--no-route"}}),

			Entry("--no-hostname and --no-route",
				func() {
					cmd.NoHostname = true
					cmd.NoRoute = true
				},
				translatableerror.ArgumentCombinationError{Args: []string{"--no-hostname", "--no-route"}}),

			Entry("-f and --no-manifest",
				func() {
					cmd.PathToManifest = "/some/path.yml"
					cmd.NoManifest = true
				},
				translatableerror.ArgumentCombinationError{Args: []string{"-f", "--no-manifest"}}),

			Entry("--random-route and --hostname",
				func() {
					cmd.Hostname = "po-tate-toe"
					cmd.RandomRoute = true
				},
				translatableerror.ArgumentCombinationError{Args: []string{"--hostname", "-n", "--random-route"}}),

			Entry("--random-route and --no-hostname",
				func() {
					cmd.RandomRoute = true
					cmd.NoHostname = true
				},
				translatableerror.ArgumentCombinationError{Args: []string{"--no-hostname", "--random-route"}}),

			Entry("--random-route and --no-route",
				func() {
					cmd.RandomRoute = true
					cmd.NoRoute = true
				},
				translatableerror.ArgumentCombinationError{Args: []string{"--no-route", "--random-route"}}),

			Entry("--random-route and --route-path",
				func() {
					cmd.RoutePath = flag.RoutePath{Path: "/bananas"}
					cmd.RandomRoute = true
				},
				translatableerror.ArgumentCombinationError{Args: []string{"--random-route", "--route-path"}}),

			Entry("--route-path and --no-route",
				func() {
					cmd.RoutePath = flag.RoutePath{Path: "/bananas"}
					cmd.NoRoute = true
				},
				translatableerror.ArgumentCombinationError{Args: []string{"--route-path", "--no-route"}}),
		)
	})
})
