package v3_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = XDescribe("v3-push Command", func() {
	var (
		cmd             v3.V3PushCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeNOAAClient  *v3actionfakes.FakeNOAAClient
		fakeActor       *v3fakes.FakeV3PushActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3PushActor)
		fakeNOAAClient = new(v3actionfakes.FakeNOAAClient)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = v3.V3PushCommand{
			AppName: app,

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			NOAAClient:  fakeNOAAClient,
		}

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

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})

			cmd.AppName = app

			// we stub out StagePackage out here so partial-happy paths below don't hang
			fakeActor.StagePackageStub = func(_ string) (<-chan v3action.Build, <-chan v3action.Warnings, <-chan error) {
				buildStream := make(chan v3action.Build)
				warningsStream := make(chan v3action.Warnings)
				errorStream := make(chan error)

				go func() {
					defer close(buildStream)
					defer close(warningsStream)
					defer close(errorStream)
					warningsStream <- v3action.Warnings{"some-staging-warning", "some-other-staging-warning"}
				}()

				return buildStream, warningsStream, errorStream
			}
		})

		Context("when creating the application is unsuccessful", func() {
			Context("due to an unexpected error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am an error")
					fakeActor.CreateApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"I am a warning", "I am also a warning"}, expectedErr)
				})

				It("displays the warnings and error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
					Expect(testUI.Out).ToNot(Say("V3 app some-app in org some-org / space some-space as banana..."))

				})
			})

			Context("due to an ApplicationAlreadyExistsError", func() {
				BeforeEach(func() {
					fakeActor.CreateApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"I am a warning", "I am also a warning"}, v3action.ApplicationAlreadyExistsError{})
				})

				Context("when getting the existing application returns an error", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationByNameAndSpaceReturns(
							v3action.Application{},
							v3action.Warnings{"this is a warning", "this is a second warning"},
							errors.New("something went wrong"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("something went wrong"))
					})

					It("displays all warnings", func() {
						Expect(testUI.Err).To(Say("this is a warning"))
						Expect(testUI.Err).To(Say("this is a second warning"))
					})
				})

				Context("when getting the existing application succeeds", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationByNameAndSpaceReturns(v3action.Application{GUID: "app-guid"}, nil, nil)
					})

					It("calls PollStart with the correct guid", func() {
						guid, _ := fakeActor.PollStartArgsForCall(0)
						Expect(guid).To(Equal("app-guid"))
					})

					It("displays the updating application message, OK, and proceeds to create the package", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Updating V3 app some-app in org some-org / space some-space as banana..."))
						Expect(testUI.Out).To(Say("OK"))

						Expect(testUI.Err).To(Say("I am a warning"))
						Expect(testUI.Err).To(Say("I am also a warning"))

						Expect(fakeActor.CreateAndUploadPackageByApplicationNameAndSpaceCallCount()).To(Equal(1))
					})
				})
			})
		})

		Context("when the application doesn't already exist", func() {
			BeforeEach(func() {
				fakeActor.CreateApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"I am a warning", "I am also a warning"}, nil)
			})

			Context("when creating the package fails", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am an error")
					fakeActor.CreateAndUploadPackageByApplicationNameAndSpaceReturns(v3action.Package{}, v3action.Warnings{"I am a package warning", "I am also a package warning"}, expectedErr)
				})

				It("displays the header and error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Out).To(Say("Uploading V3 app some-app in org some-org / space some-space as banana..."))

					Expect(testUI.Err).To(Say("I am a package warning"))
					Expect(testUI.Err).To(Say("I am also a package warning"))

					Expect(testUI.Out).ToNot(Say("Staging package for %s in org some-org / space some-space as banana...", app))
				})
			})

			Context("when creating the package succeeds", func() {
				BeforeEach(func() {
					fakeActor.CreateAndUploadPackageByApplicationNameAndSpaceReturns(v3action.Package{GUID: "some-guid"}, v3action.Warnings{"I am a package warning", "I am also a package warning"}, nil)
				})

				It("displays the header and OK", func() {
					Expect(testUI.Out).To(Say("Uploading V3 app some-app in org some-org / space some-space as banana..."))

					Expect(testUI.Err).To(Say("I am a package warning"))
					Expect(testUI.Err).To(Say("I am also a package warning"))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("Staging package for %s in org some-org / space some-space as banana...", app))
				})

				Context("when getting streaming logs fails", func() {
					var expectedErr error
					BeforeEach(func() {
						expectedErr = errors.New("something is wrong!")
						fakeActor.GetStreamingLogsForApplicationByNameAndSpaceReturns(nil, nil, v3action.Warnings{"some-logging-warning", "some-other-logging-warning"}, expectedErr)
					})

					It("returns the error and displays warnings", func() {
						Expect(executeErr).To(Equal(expectedErr))

						Expect(testUI.Out).To(Say("Staging package for %s in org some-org / space some-space as banana...", app))

						Expect(testUI.Err).To(Say("some-logging-warning"))
						Expect(testUI.Err).To(Say("some-other-logging-warning"))

					})
				})

				Context("when the logging does not error", func() {
					allLogsWritten := make(chan bool)

					BeforeEach(func() {
						fakeActor.GetStreamingLogsForApplicationByNameAndSpaceStub = func(appName string, spaceGUID string, client v3action.NOAAClient) (<-chan *v3action.LogMessage, <-chan error, v3action.Warnings, error) {
							logStream := make(chan *v3action.LogMessage)
							errorStream := make(chan error)

							go func() {
								logStream <- v3action.NewLogMessage("Here are some staging logs!", 1, time.Now(), v3action.StagingLog, "sourceInstance")
								logStream <- v3action.NewLogMessage("Here are some other staging logs!", 1, time.Now(), v3action.StagingLog, "sourceInstance")
								allLogsWritten <- true
							}()

							return logStream, errorStream, v3action.Warnings{"steve for all I care"}, nil
						}
					})

					Context("when the staging returns an error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("any gibberish")
							fakeActor.StagePackageStub = func(packageGUID string) (<-chan v3action.Build, <-chan v3action.Warnings, <-chan error) {
								buildStream := make(chan v3action.Build)
								warningsStream := make(chan v3action.Warnings)
								errorStream := make(chan error)

								go func() {
									defer close(buildStream)
									defer close(warningsStream)
									defer close(errorStream)
									warningsStream <- v3action.Warnings{"some-staging-warning", "some-other-staging-warning"}
									errorStream <- expectedErr
								}()

								return buildStream, warningsStream, errorStream
							}
						})

						It("returns the error and displays warnings", func() {
							Expect(executeErr).To(Equal(expectedErr))

							Expect(testUI.Out).To(Say("Staging package for %s in org some-org / space some-space as banana...", app))

							Expect(testUI.Err).To(Say("some-staging-warning"))
							Expect(testUI.Err).To(Say("some-other-staging-warning"))

							Expect(testUI.Out).ToNot(Say("Setting app some-app to droplet some-droplet-guid in org some-org / space some-space as banana..."))
						})
					})

					Context("when the staging is successful", func() {
						BeforeEach(func() {
							fakeActor.StagePackageStub = func(packageGUID string) (<-chan v3action.Build, <-chan v3action.Warnings, <-chan error) {
								buildStream := make(chan v3action.Build)
								warningsStream := make(chan v3action.Warnings)
								errorStream := make(chan error)

								go func() {
									<-allLogsWritten
									defer close(buildStream)
									defer close(warningsStream)
									defer close(errorStream)
									warningsStream <- v3action.Warnings{"some-staging-warning", "some-other-staging-warning"}
									buildStream <- v3action.Build{Droplet: ccv3.Droplet{GUID: "some-droplet-guid"}}
								}()

								return buildStream, warningsStream, errorStream
							}
						})

						It("outputs the droplet GUID", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Staging package for %s in org some-org / space some-space as banana...", app))
							Expect(testUI.Out).To(Say("droplet: some-droplet-guid"))
							Expect(testUI.Out).To(Say("OK"))

							Expect(testUI.Err).To(Say("some-staging-warning"))
							Expect(testUI.Err).To(Say("some-other-staging-warning"))
						})

						It("stages the package", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
							Expect(fakeActor.StagePackageArgsForCall(0)).To(Equal("some-guid"))
						})

						It("displays staging logs and their warnings", func() {
							Expect(testUI.Out).To(Say("Here are some staging logs!"))
							Expect(testUI.Out).To(Say("Here are some other staging logs!"))

							Expect(testUI.Err).To(Say("steve for all I care"))

							Expect(fakeActor.GetStreamingLogsForApplicationByNameAndSpaceCallCount()).To(Equal(1))
							appName, spaceGUID, noaaClient := fakeActor.GetStreamingLogsForApplicationByNameAndSpaceArgsForCall(0)
							Expect(appName).To(Equal(app))
							Expect(spaceGUID).To(Equal("some-space-guid"))
							Expect(noaaClient).To(Equal(fakeNOAAClient))

							Expect(fakeActor.StagePackageCallCount()).To(Equal(1))
							Expect(fakeActor.StagePackageArgsForCall(0)).To(Equal("some-guid"))
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
								Expect(testUI.Out).To(Say("Staging package for %s in org some-org / space some-space as banana...", app))
								Expect(testUI.Out).To(Say("droplet: some-droplet-guid"))
								Expect(testUI.Out).To(Say("OK"))

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
									fakeActor.StartApplicationReturns(v3action.Application{}, v3action.Warnings{"start-warning-1", "start-warning-2"}, nil)
								})

								It("says that the app was started and outputs warnings", func() {
									Expect(testUI.Out).To(Say("Starting app some-app in org some-org / space some-space as banana\\.\\.\\."))

									Expect(testUI.Err).To(Say("start-warning-1"))
									Expect(testUI.Err).To(Say("start-warning-2"))
									Expect(testUI.Out).To(Say("OK"))

									Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
									appName, spaceGUID := fakeActor.StartApplicationArgsForCall(0)
									Expect(appName).To(Equal("some-app"))
									Expect(spaceGUID).To(Equal("some-space-guid"))
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
									fakeActor.PollStartReturns(v3action.StartupTimeoutError{})
								})

								It("returns the StartupTimeoutError", func() {
									Expect(executeErr).To(MatchError(shared.StartupTimeoutError{AppName: "some-app"}))
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
										expectedErr = v3action.ApplicationNotFoundError{Name: app}
										fakeActor.GetApplicationSummaryByNameAndSpaceReturns(v3action.ApplicationSummary{}, v3action.Warnings{"display-warning-1", "display-warning-2"}, expectedErr)
									})

									It("returns the error and prints warnings", func() {
										Expect(executeErr).To(Equal(command.ApplicationNotFoundError{Name: app}))

										Expect(testUI.Out).To(Say("Showing health and status for app some-app in org some-org / space some-space as banana\\.\\.\\."))

										Expect(testUI.Err).To(Say("display-warning-1"))
										Expect(testUI.Err).To(Say("display-warning-2"))

										Expect(testUI.Out).ToNot(Say("name:\\s+some-app"))
									})
								})

								Context("when displaying the application info is successful", func() {
									BeforeEach(func() {
										summary := v3action.ApplicationSummary{
											Application: v3action.Application{
												Name:  "some-app",
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
											Processes: []v3action.Process{
												v3action.Process{
													Type:       "worker",
													MemoryInMB: 64,
													Instances: []v3action.Instance{
														v3action.Instance{
															Index:       0,
															State:       "RUNNING",
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

									It("prints the application summary and outputs warnings", func() {
										Expect(executeErr).ToNot(HaveOccurred())

										Expect(testUI.Out).To(Say("(?m)Showing health and status for app some-app in org some-org / space some-space as banana\\.\\.\\.\n\n"))
										Expect(testUI.Out).To(Say("name:\\s+some-app"))
										Expect(testUI.Out).To(Say("requested state:\\s+started"))
										Expect(testUI.Out).To(Say("processes:\\s+worker:1/1"))
										Expect(testUI.Out).To(Say("memory usage:\\s+64M x 1"))
										Expect(testUI.Out).To(Say("stack:\\s+cflinuxfs2"))
										Expect(testUI.Out).To(Say("(?m)buildpacks:\\s+some-detect-output\n\n"))

										Expect(testUI.Out).To(Say("worker"))
										Expect(testUI.Out).To(Say("\\s+state\\s+since\\s+cpu\\s+memory\\s+disk"))
										Expect(testUI.Out).To(Say("#0\\s+running\\s+2013-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M\\s+0.0%\\s+3.8M of 64M\\s+3.8M of 7.6M"))

										Expect(testUI.Err).To(Say("display-warning-1"))
										Expect(testUI.Err).To(Say("display-warning-2"))

										Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
										appName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
										Expect(appName).To(Equal("some-app"))
										Expect(spaceGUID).To(Equal("some-space-guid"))
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
