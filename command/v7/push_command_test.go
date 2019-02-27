package v7_test

import (
	"errors"
	"fmt"
	. "github.com/onsi/gomega/gstruct"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cli/util/manifestparser"

	"gopkg.in/yaml.v2"

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
	"github.com/cloudfoundry/bosh-cli/director/template"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

type Step struct {
	Error    error
	Event    v7pushaction.Event
	Warnings v7pushaction.Warnings
}

func FillInEvents(tuples []Step) (<-chan v7pushaction.Event, <-chan v7pushaction.Warnings, <-chan error) {
	eventStream := make(chan v7pushaction.Event)
	warningsStream := make(chan v7pushaction.Warnings)
	errorStream := make(chan error)

	go func() {
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
	}()

	return eventStream, warningsStream, errorStream
}

func FillInValues(tuples []Step, state v7pushaction.PushPlan) func(v7pushaction.PushPlan, v7pushaction.ProgressBar) (<-chan v7pushaction.PushPlan, <-chan v7pushaction.Event, <-chan v7pushaction.Warnings, <-chan error) {
	return func(v7pushaction.PushPlan, v7pushaction.ProgressBar) (<-chan v7pushaction.PushPlan, <-chan v7pushaction.Event, <-chan v7pushaction.Warnings, <-chan error) {
		stateStream := make(chan v7pushaction.PushPlan)

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
		input            *Buffer
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v7fakes.FakePushActor
		fakeVersionActor *v7fakes.FakeV7ActorForPush
		fakeProgressBar  *v6fakes.FakeProgressBar
		fakeNOAAClient   *v7actionfakes.FakeNOAAClient
		binaryName       string
		executeErr       error

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
		fakeNOAAClient = new(v7actionfakes.FakeNOAAClient)

		appName1 = "first-app"
		appName2 = "second-app"
		userName = "some-user"
		spaceName = "some-space"
		orgName = "some-org"
		pwd = "/push/cmd/test"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.ExperimentalReturns(true) // TODO: Delete once we remove the experimental flag

		cmd = PushCommand{
			RequiredArgs: flag.AppName{AppName: "passed-as-command-arg"},
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

		BeforeEach(func() {
			appNamesChannel := make(chan []string, 1)
			appNamesChannel <- []string{appName1, appName2}
			close(appNamesChannel)
			events, warnings, errors := FillInEvents([]Step{
				{
					Warnings: v7pushaction.Warnings{"some-warning-1"},
					Event:    v7pushaction.ApplyManifest,
				},
				{
					Warnings: v7pushaction.Warnings{"some-warning-2"},
					Event:    v7pushaction.ApplyManifestComplete,
				},
			})

			fakeActor.PrepareSpaceReturns(appNamesChannel, events, warnings, errors)

			fakeActor.ActualizeStub = FillInValues([]Step{{}}, v7pushaction.PushPlan{})
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

			When("invalid flags are passed", func() {
				BeforeEach(func() {
					cmd.DockerUsername = "some-docker-username"
				})

				It("returns a validation error", func() {
					Expect(executeErr).To(MatchError(translatableerror.RequiredFlagsError{Arg1: "--docker-image, -o", Arg2: "--docker-username"}))
				})
			})

			Describe("manifest", func() {
				var tempDir string

				BeforeEach(func() {
					var err error
					tempDir, err = ioutil.TempDir("", "manifest-push-unit")
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
				})

				var yamlUnmarshalMarshal = func(b []byte) []byte {
					var obj interface{}
					yaml.Unmarshal(b, &obj)
					postMarshal, err := yaml.Marshal(obj)
					Expect(err).ToNot(HaveOccurred())
					return postMarshal
				}

				When("No path is provided", func() {
					BeforeEach(func() {
						cmd.PWD = tempDir
					})

					When("there is a manifest file in the current dir", func() {
						var yamlContents []byte
						var pathToYAMLFile string

						BeforeEach(func() {
							yamlContents = []byte("---\napplications:\n- name: banana")
							pathToYAMLFile = filepath.Join(tempDir, "manifest.yml")
							err := ioutil.WriteFile(pathToYAMLFile, yamlContents, 0644)
							Expect(err).ToNot(HaveOccurred())
						})

						It("creates a parser and passes through to PrepareSpace", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeActor.PrepareSpaceCallCount()).To(Equal(1))
							_, _, manifestParser, _ := fakeActor.PrepareSpaceArgsForCall(0)
							Expect(manifestParser.FullRawManifest()).To(Equal(yamlUnmarshalMarshal(yamlContents)))
							Expect(manifestParser.PathToManifest).To(Equal(pathToYAMLFile))
							Expect(manifestParser.AppNames()).To(ConsistOf("banana"))
						})

						When("the --no-manifest flag is specified", func() {
							BeforeEach(func() {
								cmd.NoManifest = true
							})
							It("passes an empty manifest to PrepareSpace", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(fakeActor.PrepareSpaceCallCount()).To(Equal(1))
								_, _, manifestParser, _ := fakeActor.PrepareSpaceArgsForCall(0)
								Expect(manifestParser).To(Equal(manifestparser.NewParser()))
							})
						})
					})

					When("there is not a manifest in the current dir", func() {
						It("does not pass a manifest to PrepareSpace", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeActor.PrepareSpaceCallCount()).To(Equal(1))
							_, _, manifestParser, _ := fakeActor.PrepareSpaceArgsForCall(0)
							Expect(manifestParser).To(Equal(manifestparser.NewParser()))
						})
					})
				})

				When("The -f flag is specified", func() {
					When("The manifest exists at the path", func() {
						var yamlContents []byte
						var pathToYAMLFile string

						BeforeEach(func() {
							yamlContents = []byte("---\napplications:\n- name: banana")
							pathToYAMLFile = filepath.Join(tempDir, "manifest.yml")
							err := ioutil.WriteFile(pathToYAMLFile, yamlContents, 0644)
							Expect(err).ToNot(HaveOccurred())
							cmd.PathToManifest = flag.PathWithExistenceCheck(pathToYAMLFile)
						})

						It("reads the manifest and passes through to PrepareSpace", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeActor.PrepareSpaceCallCount()).To(Equal(1))
							_, _, manifestParser, _ := fakeActor.PrepareSpaceArgsForCall(0)
							Expect(manifestParser.FullRawManifest()).To(Equal(yamlUnmarshalMarshal(yamlContents)))
							Expect(manifestParser.PathToManifest).To(Equal(pathToYAMLFile))
							Expect(manifestParser.AppNames()).To(ConsistOf("banana"))
						})
					})

					When("The manifest does not exist at the path", func() {
						BeforeEach(func() {
							cmd.PathToManifest = "/some/non-existant/path.yml"
						})

						It("throws an error", func() {
							Expect(os.IsNotExist(executeErr)).To(BeTrue(), fmt.Sprintf("expected to get an 'is not exists' error but got %#v", executeErr))
						})
					})
				})

				When("--vars-files are specified", func() {
					var yamlContents []byte
					var varFileContents []byte
					var expectedManifest []byte
					var pathToYAMLFile string

					BeforeEach(func() {
						expectedManifest = yamlUnmarshalMarshal([]byte("---\napplications:\n- name: turtle"))
						yamlContents = []byte("---\napplications:\n- name: ((put-var-here))")
						pathToYAMLFile = filepath.Join(tempDir, "manifest.yml")
						err := ioutil.WriteFile(pathToYAMLFile, yamlContents, 0644)
						Expect(err).ToNot(HaveOccurred())
						cmd.PathToManifest = flag.PathWithExistenceCheck(pathToYAMLFile)

						varFileContents = []byte("---\nput-var-here: turtle")
						pathToVarFile := filepath.Join(tempDir, "var.yml")
						err = ioutil.WriteFile(pathToVarFile, varFileContents, 0644)
						Expect(err).ToNot(HaveOccurred())
						cmd.PathsToVarsFiles = []flag.PathWithExistenceCheck{
							flag.PathWithExistenceCheck(pathToVarFile),
						}
					})

					It("reads the manifest, substitutes vars, and passes through to PrepareSpace", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeActor.PrepareSpaceCallCount()).To(Equal(1))
						_, _, manifestParser, _ := fakeActor.PrepareSpaceArgsForCall(0)
						Expect(manifestParser.FullRawManifest()).To(Equal(yamlUnmarshalMarshal(expectedManifest)))
						Expect(manifestParser.PathToManifest).To(Equal(pathToYAMLFile))
						Expect(manifestParser.AppNames()).To(ConsistOf("turtle"))
					})
				})

				When("The --var flag is provided", func() {
					var yamlContents []byte
					var expectedManifest []byte
					var pathToYAMLFile string

					BeforeEach(func() {
						expectedManifest = yamlUnmarshalMarshal([]byte("---\napplications:\n- name: turtle"))
						yamlContents = []byte("---\napplications:\n- name: ((put-var-here))")
						pathToYAMLFile = filepath.Join(tempDir, "manifest.yml")
						err := ioutil.WriteFile(pathToYAMLFile, yamlContents, 0644)
						Expect(err).ToNot(HaveOccurred())
						cmd.PathToManifest = flag.PathWithExistenceCheck(pathToYAMLFile)
						cmd.Vars = []template.VarKV{
							{Name: "put-var-here", Value: "turtle"},
						}
					})

					It("reads the manifest, substitutes vars, and passes through to PrepareSpace", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeActor.PrepareSpaceCallCount()).To(Equal(1))
						_, _, manifestParser, _ := fakeActor.PrepareSpaceArgsForCall(0)
						Expect(manifestParser.FullRawManifest()).To(Equal(yamlUnmarshalMarshal(expectedManifest)))
						Expect(manifestParser.PathToManifest).To(Equal(pathToYAMLFile))
						Expect(manifestParser.AppNames()).To(ConsistOf("turtle"))
					})
				})
			})

			Describe("delegating to Actor.PrepareSpace", func() {
				It("delegates to PrepareSpace", func() {
					expectedSpaceGUID, expectedAppName, expectedParser, _ := fakeActor.PrepareSpaceArgsForCall(0)
					Expect(expectedSpaceGUID).To(Equal("some-space-guid"))
					Expect(expectedAppName).To(Equal("passed-as-command-arg"))
					Expect(expectedParser).To(Equal(manifestparser.NewParser()))
				})

				When("Actor.PrepareSpace has no errors", func() {
					Describe("delegating to Actor.Conceptualize", func() {
						When("there are no flag overrides", func() {
							BeforeEach(func() {
								fakeActor.ConceptualizeReturns(
									[]v7pushaction.PushPlan{
										{Application: v7action.Application{Name: appName1}},
										{Application: v7action.Application{Name: appName2}},
									},
									v7pushaction.Warnings{"conceptualize-warning-1"}, nil)
							})

							It("generates a push plan with the specified app path", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(testUI.Out).To(Say(
									"Pushing apps %s, %s to org some-org / space some-space as some-user",
									appName1,
									appName2,
								))
								Expect(testUI.Out).To(Say(`Getting app info\.\.\.`))
								Expect(testUI.Err).To(Say("conceptualize-warning-1"))

								Expect(fakeActor.ConceptualizeCallCount()).To(Equal(1))
								names, spaceGUID, orgGUID, currentDirectory, _ := fakeActor.ConceptualizeArgsForCall(0)
								Expect(names).To(Equal([]string{appName1, appName2}))
								Expect(spaceGUID).To(Equal("some-space-guid"))
								Expect(orgGUID).To(Equal("some-org-guid"))
								Expect(currentDirectory).To(Equal(pwd))
							})

							Describe("delegating to Actor.Actualize", func() {
								When("Actualize returns success", func() {
									BeforeEach(func() {
										fakeActor.ActualizeStub = FillInValues([]Step{
											{},
										}, v7pushaction.PushPlan{Application: v7action.Application{GUID: "potato"}})
									})

									Describe("actualize events", func() {
										BeforeEach(func() {
											fakeActor.ActualizeStub = FillInValues([]Step{
												{
													Event:    v7pushaction.SkippingApplicationCreation,
													Warnings: v7pushaction.Warnings{"skipping app creation warnings"},
												},
												{
													Event:    v7pushaction.CreatingApplication,
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
											}, v7pushaction.PushPlan{})
										})

										It("actualizes the application and displays events/warnings", func() {
											Expect(executeErr).ToNot(HaveOccurred())

											Expect(fakeProgressBar.ReadyCallCount()).Should(Equal(2))
											Expect(fakeProgressBar.CompleteCallCount()).Should(Equal(2))

											Expect(testUI.Out).To(Say("Updating app first-app..."))
											Expect(testUI.Err).To(Say("skipping app creation warnings"))

											Expect(testUI.Out).To(Say("Creating app first-app..."))
											Expect(testUI.Err).To(Say("app creation warnings"))

											Expect(testUI.Out).To(Say("Mapping routes..."))
											Expect(testUI.Err).To(Say("routes warnings"))

											Expect(testUI.Out).To(Say("Packaging files to upload..."))

											Expect(testUI.Out).To(Say("Uploading files..."))
											Expect(testUI.Err).To(Say("upload app archive warning"))

											Expect(testUI.Out).To(Say("Retrying upload due to an error..."))
											Expect(testUI.Err).To(Say("retry upload warning"))

											Expect(testUI.Out).To(Say("Waiting for API to complete processing files..."))

											Expect(testUI.Out).To(Say("Waiting for app first-app to start..."))

											Expect(testUI.Out).To(Say("Updating app second-app..."))
											Expect(testUI.Err).To(Say("skipping app creation warnings"))

											Expect(testUI.Out).To(Say("Creating app second-app..."))
											Expect(testUI.Err).To(Say("app creation warnings"))

											Expect(testUI.Out).To(Say("Mapping routes..."))
											Expect(testUI.Err).To(Say("routes warnings"))

											Expect(testUI.Out).To(Say("Packaging files to upload..."))

											Expect(testUI.Out).To(Say("Uploading files..."))
											Expect(testUI.Err).To(Say("upload app archive warning"))

											Expect(testUI.Out).To(Say("Retrying upload due to an error..."))
											Expect(testUI.Err).To(Say("retry upload warning"))

											Expect(testUI.Out).To(Say("Waiting for API to complete processing files..."))

											Expect(testUI.Out).To(Say("Waiting for app second-app to start..."))
										})
									})

									Describe("staging logs", func() {
										BeforeEach(func() {
											fakeActor.ActualizeStub = FillInValues([]Step{
												{
													Event: v7pushaction.StartingStaging,
												},
											}, v7pushaction.PushPlan{})
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

									When("user does not request --no-start", func() {
										BeforeEach(func() {
											cmd.NoStart = false
										})

										When("restarting the app succeeds", func() {
											BeforeEach(func() {
												fakeVersionActor.RestartApplicationReturns(v7action.Warnings{"some-restart-warning"}, nil)
											})

											It("restarts the app and displays warnings", func() {
												Expect(executeErr).ToNot(HaveOccurred())

												Expect(testUI.Err).To(Say("some-restart-warning"))

												Expect(fakeVersionActor.RestartApplicationCallCount()).To(Equal(2))
												Expect(fakeVersionActor.RestartApplicationArgsForCall(0)).To(Equal("potato"))
												Expect(fakeVersionActor.RestartApplicationArgsForCall(1)).To(Equal("potato"))
											})

											When("when getting the application summary succeeds", func() {
												BeforeEach(func() {
													summary := v7action.ApplicationSummary{
														Application:      v7action.Application{},
														CurrentDroplet:   v7action.Droplet{},
														ProcessSummaries: v7action.ProcessSummaries{},
													}
													fakeVersionActor.GetApplicationSummaryByNameAndSpaceReturnsOnCall(0, summary, v7action.Warnings{"app-1-summary-warning-1", "app-1-summary-warning-2"}, nil)
													fakeVersionActor.GetApplicationSummaryByNameAndSpaceReturnsOnCall(1, summary, v7action.Warnings{"app-2-summary-warning-1", "app-2-summary-warning-2"}, nil)
												})

												// TODO: Don't test the shared.AppSummaryDisplayer.AppDisplay method.
												// Use DI to pass in a new AppSummaryDisplayer to the Command instead.
												It("displays the app summary", func() {
													Expect(executeErr).ToNot(HaveOccurred())
													Expect(fakeVersionActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(2))
												})

											})

											When("getting the application summary fails", func() {
												BeforeEach(func() {
													fakeVersionActor.GetApplicationSummaryByNameAndSpaceReturns(
														v7action.ApplicationSummary{},
														v7action.Warnings{"get-application-summary-warnings"},
														errors.New("get-application-summary-error"),
													)
												})

												It("does not display the app summary", func() {
													Expect(testUI.Out).ToNot(Say(`requested state:`))
												})

												It("returns the error from GetApplicationSummaryByNameAndSpace", func() {
													Expect(executeErr).To(MatchError("get-application-summary-error"))
												})

												It("prints the warnings", func() {
													Expect(testUI.Err).To(Say("get-application-summary-warnings"))
												})
											})

										})

										When("restarting the app fails", func() {
											When("restarting fails in a generic way", func() {
												BeforeEach(func() {
													fakeVersionActor.RestartApplicationReturns(v7action.Warnings{"some-restart-warning"}, errors.New("restart failure"))
												})

												It("returns an error and any warnings", func() {
													Expect(executeErr).To(MatchError("restart failure"))
													Expect(testUI.Err).To(Say("some-restart-warning"))
												})
											})

											When("the error is an AllInstancesCrashedError", func() {
												BeforeEach(func() {
													fakeVersionActor.RestartApplicationReturns(nil, actionerror.AllInstancesCrashedError{})
												})

												It("returns the ApplicationUnableToStartError", func() {
													Expect(executeErr).To(MatchError(translatableerror.ApplicationUnableToStartError{
														AppName:    "first-app",
														BinaryName: binaryName,
													}))
												})

											})

											When("restart times out", func() {
												BeforeEach(func() {
													fakeVersionActor.RestartApplicationReturns(v7action.Warnings{"some-restart-warning"}, actionerror.StartupTimeoutError{})
												})

												It("returns the StartupTimeoutError and prints warnings", func() {
													Expect(executeErr).To(MatchError(translatableerror.StartupTimeoutError{
														AppName:    "first-app",
														BinaryName: binaryName,
													}))

													Expect(testUI.Err).To(Say("some-restart-warning"))
												})
											})
										})
									})

									When("user requests --no-start", func() {
										BeforeEach(func() {
											cmd.NoStart = true
										})

										It("does not attempt to restart the app", func() {
											Expect(fakeVersionActor.RestartApplicationCallCount()).To(Equal(0))
										})
									})
								})

								When("Actualize returns an error", func() {
									BeforeEach(func() {
										fakeActor.ActualizeStub = FillInValues([]Step{
											{
												Error: errors.New("anti avant garde naming"),
											},
										}, v7pushaction.PushPlan{})
									})

									It("returns the error", func() {
										Expect(executeErr).To(MatchError("anti avant garde naming"))
									})
								})
							})
						})

						When("flag overrides are specified", func() {
							BeforeEach(func() {
								cmd.AppPath = "some/app/path"
							})

							It("generates a push plan with the specified flag overrides", func() {
								Expect(fakeActor.ConceptualizeCallCount()).To(Equal(1))
								_, _, _, _, overrides := fakeActor.ConceptualizeArgsForCall(0)
								Expect(overrides).To(MatchFields(IgnoreExtras, Fields{
									"ProvidedAppPath": Equal("some/app/path"),
								}))
							})
						})

						When("conceptualize returns an error", func() {
							var expectedErr error

							BeforeEach(func() {
								expectedErr = errors.New("some-error")
								fakeActor.ConceptualizeReturns(nil, v7pushaction.Warnings{"some-warning-1"}, expectedErr)
							})

							It("generates a push plan with the specified app path", func() {
								Expect(executeErr).To(MatchError(expectedErr))
								Expect(testUI.Err).To(Say("some-warning-1"))
							})
						})
					})
				})

				When("Actor.PrepareSpace has an error", func() {
					var appNamesChannel chan []string

					BeforeEach(func() {
						appNamesChannel = make(chan []string)
						close(appNamesChannel)
						events, warnings, errors := FillInEvents([]Step{
							{
								Warnings: v7pushaction.Warnings{"prepare-space-warning-1"},
								Error:    errors.New("prepare-space-error-1"),
							},
						})

						fakeActor.PrepareSpaceReturns(appNamesChannel, events, warnings, errors)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(errors.New("prepare-space-error-1")))
						Expect(testUI.Err).To(Say("prepare-space-warning-1"))
					})

					It("does not delegate to Conceptualize", func() {
						Expect(fakeActor.ConceptualizeCallCount()).To(Equal(0))
					})

					It("does not delegate to Actualize", func() {
						Expect(fakeActor.ActualizeCallCount()).To(Equal(0))
					})
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
			cmd.Stack = "validStack"
			cmd.HealthCheckType = flag.HealthCheckType{Type: constant.Port}
			cmd.HealthCheckHTTPEndpoint = "/health-check-http-endpoint"
			cmd.HealthCheckTimeout = flag.PositiveInteger{Value: 7}
			cmd.Memory = flag.Megabytes{NullUint64: types.NullUint64{Value: 100, IsSet: true}}
			cmd.Disk = flag.Megabytes{NullUint64: types.NullUint64{Value: 1024, IsSet: true}}
			cmd.StartCommand = flag.Command{FilteredString: types.FilteredString{IsSet: true, Value: "some-start-command"}}
			cmd.NoRoute = true
			cmd.NoStart = true
			cmd.Instances = flag.Instances{NullInt: types.NullInt{Value: 10, IsSet: true}}
		})

		JustBeforeEach(func() {
			overrides, overridesErr = cmd.GetFlagOverrides()
			Expect(overridesErr).ToNot(HaveOccurred())
		})

		It("sets them on the flag overrides", func() {
			Expect(overridesErr).ToNot(HaveOccurred())
			Expect(overrides.Buildpacks).To(ConsistOf("buildpack-1", "buildpack-2"))
			Expect(overrides.Stack).To(Equal("validStack"))
			Expect(overrides.HealthCheckType).To(Equal(constant.Port))
			Expect(overrides.HealthCheckEndpoint).To(Equal("/health-check-http-endpoint"))
			Expect(overrides.HealthCheckTimeout).To(BeEquivalentTo(7))
			Expect(overrides.Memory).To(Equal(types.NullUint64{Value: 100, IsSet: true}))
			Expect(overrides.Disk).To(Equal(types.NullUint64{Value: 1024, IsSet: true}))
			Expect(overrides.StartCommand).To(Equal(types.FilteredString{IsSet: true, Value: "some-start-command"}))
			Expect(overrides.SkipRouteCreation).To(BeTrue())
			Expect(overrides.NoStart).To(BeTrue())
			Expect(overrides.Instances).To(Equal(types.NullInt{Value: 10, IsSet: true}))
		})

		When("a docker image is provided", func() {
			BeforeEach(func() {
				cmd.DockerImage = flag.DockerImage{Path: "some-docker-image"}
			})

			It("sets docker image on the flag overrides", func() {
				Expect(overridesErr).ToNot(HaveOccurred())
				Expect(overrides.DockerImage).To(Equal("some-docker-image"))
			})

			When("docker username is provided", func() {
				When("a password is provided via environment variable", func() {
					BeforeEach(func() {
						cmd.DockerUsername = "some-docker-username"
						fakeConfig.DockerPasswordReturns("some-docker-password")
					})

					It("takes the password from the environment", func() {
						Expect(overridesErr).ToNot(HaveOccurred())

						Expect(testUI.Out).ToNot(Say("Environment variable CF_DOCKER_PASSWORD not set."))
						Expect(testUI.Out).ToNot(Say("Docker password"))

						Expect(testUI.Out).To(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))

						Expect(overrides.DockerUsername).To(Equal("some-docker-username"))
						Expect(overrides.DockerPassword).To(Equal("some-docker-password"))
					})
				})

				When("no password is provided", func() {
					BeforeEach(func() {
						cmd.DockerUsername = "some-docker-username"
						input.Write([]byte("some-docker-password\n"))
					})

					It("prompts for a password", func() {
						Expect(overridesErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Environment variable CF_DOCKER_PASSWORD not set."))
						Expect(testUI.Out).To(Say("Docker password"))

						Expect(overrides.DockerUsername).To(Equal("some-docker-username"))
						Expect(overrides.DockerPassword).To(Equal("some-docker-password"))
					})
				})
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

		Entry("when --endpoint does not have a matching -u",
			func() {
				cmd.HealthCheckHTTPEndpoint = "/health"
			},
			translatableerror.RequiredFlagsError{Arg1: "--health-check-type=http, -u=http", Arg2: "--endpoint"}),

		Entry("when --endpoint has a matching -u=process instead of a -u=http",
			func() {
				cmd.HealthCheckHTTPEndpoint = "/health"
				cmd.HealthCheckType.Type = constant.Process
			},
			translatableerror.RequiredFlagsError{Arg1: "--health-check-type=http, -u=http", Arg2: "--endpoint"}),

		Entry("when --endpoint has a matching -u=port instead of a -u=http",
			func() {
				cmd.HealthCheckHTTPEndpoint = "/health"
				cmd.HealthCheckType.Type = constant.Port
			},
			translatableerror.RequiredFlagsError{Arg1: "--health-check-type=http, -u=http", Arg2: "--endpoint"}),

		Entry("when -u http does have a matching --endpoint",
			func() {
				cmd.HealthCheckType.Type = constant.HTTP
				cmd.HealthCheckHTTPEndpoint = "/health"
			},
			nil),
	)

})
