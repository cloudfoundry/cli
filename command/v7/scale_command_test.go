package v7_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("scale Command", func() {
	var (
		cmd             ScaleCommand
		input           *Buffer
		output          *Buffer
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeScaleActor
		appName         string
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		output = NewBuffer()
		testUI = ui.NewTestUI(input, output, NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeScaleActor)
		appName = "some-app"

		cmd = ScaleCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd.RequiredArgs.AppName = appName
		cmd.ProcessType = constant.ProcessTypeWeb
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
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

		When("the application does not exist", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v7action.Application{},
					v7action.Warnings{"get-app-warning"},
					actionerror.ApplicationNotFoundError{Name: appName})
			})

			It("returns an ApplicationNotFoundError and all warnings", func() {
				Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: appName}))

				Expect(testUI.Out).ToNot(Say("Showing | Scaling"))
				Expect(testUI.Err).To(Say("get-app-warning"))

				Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appNameArg).To(Equal(appName))
				Expect(spaceGUIDArg).To(Equal("some-space-guid"))
			})
		})

		When("an error occurs getting the application", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get app error")
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v7action.Application{},
					v7action.Warnings{"get-app-warning"},
					expectedErr)
			})

			It("returns the error and displays all warnings", func() {
				Expect(executeErr).To(Equal(expectedErr))
				Expect(testUI.Err).To(Say("get-app-warning"))
			})
		})

		When("the application exists", func() {
			var appSummary v7action.ApplicationSummary

			BeforeEach(func() {
				appSummary = v7action.ApplicationSummary{
					ProcessSummaries: v7action.ProcessSummaries{
						{
							Process: v7action.Process{
								Type:       constant.ProcessTypeWeb,
								MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
								DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
							},
							InstanceDetails: []v7action.ProcessInstance{
								v7action.ProcessInstance{
									Index:       0,
									State:       constant.ProcessInstanceRunning,
									MemoryUsage: 1000000,
									DiskUsage:   1000000,
									MemoryQuota: 33554432,
									DiskQuota:   2000000,
									Uptime:      int(time.Now().Sub(time.Unix(267321600, 0)).Seconds()),
								},
								v7action.ProcessInstance{
									Index:       1,
									State:       constant.ProcessInstanceRunning,
									MemoryUsage: 2000000,
									DiskUsage:   2000000,
									MemoryQuota: 33554432,
									DiskQuota:   4000000,
									Uptime:      int(time.Now().Sub(time.Unix(330480000, 0)).Seconds()),
								},
								v7action.ProcessInstance{
									Index:       2,
									State:       constant.ProcessInstanceRunning,
									MemoryUsage: 3000000,
									DiskUsage:   3000000,
									MemoryQuota: 33554432,
									DiskQuota:   6000000,
									Uptime:      int(time.Now().Sub(time.Unix(1277164800, 0)).Seconds()),
								},
							},
						},
						{
							Process: v7action.Process{
								Type:       "console",
								MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
								DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
							},
							InstanceDetails: []v7action.ProcessInstance{
								v7action.ProcessInstance{
									Index:       0,
									State:       constant.ProcessInstanceRunning,
									MemoryUsage: 1000000,
									DiskUsage:   1000000,
									MemoryQuota: 33554432,
									DiskQuota:   8000000,
									Uptime:      int(time.Now().Sub(time.Unix(167572800, 0)).Seconds()),
								},
							},
						},
					},
				}

				fakeActor.GetApplicationByNameAndSpaceReturns(
					v7action.Application{GUID: "some-app-guid"},
					v7action.Warnings{"get-app-warning"},
					nil)
			})

			When("no flag options are provided", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v7action.Warnings{"get-app-summary-warning"},
						nil)
				})

				It("displays current scale properties and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say(`Showing current scale of app some-app in org some-org / space some-space as some-user\.\.\.`))
					Expect(testUI.Out).ToNot(Say("Scaling | This will cause the app to restart | Stopping | Starting | Waiting"))

					firstAppTable := helpers.ParseV3AppProcessTable(output.Contents())
					Expect(len(firstAppTable.Processes)).To(Equal(2))

					webProcessSummary := firstAppTable.Processes[0]
					Expect(webProcessSummary.Type).To(Equal("web"))
					Expect(webProcessSummary.InstanceCount).To(Equal("3/3"))
					Expect(webProcessSummary.MemUsage).To(Equal("32M"))

					Expect(webProcessSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
					Expect(webProcessSummary.Instances[0].Disk).To(Equal("976.6K of 1.9M"))
					Expect(webProcessSummary.Instances[0].CPU).To(Equal("0.0%"))

					Expect(webProcessSummary.Instances[1].Memory).To(Equal("1.9M of 32M"))
					Expect(webProcessSummary.Instances[1].Disk).To(Equal("1.9M of 3.8M"))
					Expect(webProcessSummary.Instances[1].CPU).To(Equal("0.0%"))

					Expect(webProcessSummary.Instances[2].Memory).To(Equal("2.9M of 32M"))
					Expect(webProcessSummary.Instances[2].Disk).To(Equal("2.9M of 5.7M"))
					Expect(webProcessSummary.Instances[2].CPU).To(Equal("0.0%"))

					consoleProcessSummary := firstAppTable.Processes[1]
					Expect(consoleProcessSummary.Type).To(Equal("console"))
					Expect(consoleProcessSummary.InstanceCount).To(Equal("1/1"))
					Expect(consoleProcessSummary.MemUsage).To(Equal("16M"))

					Expect(consoleProcessSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
					Expect(consoleProcessSummary.Instances[0].Disk).To(Equal("976.6K of 7.6M"))
					Expect(consoleProcessSummary.Instances[0].CPU).To(Equal("0.0%"))

					Expect(testUI.Err).To(Say("get-app-warning"))
					Expect(testUI.Err).To(Say("get-app-summary-warning"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(0))
				})

				When("an error is encountered getting process information", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get process error")
						fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
							v7action.ApplicationSummary{},
							v7action.Warnings{"get-process-warning"},
							expectedErr,
						)
					})

					It("returns the error and displays all warnings", func() {
						Expect(executeErr).To(Equal(expectedErr))
						Expect(testUI.Err).To(Say("get-process-warning"))
					})
				})
			})

			When("all flag options are provided", func() {
				BeforeEach(func() {
					cmd.Instances.Value = 2
					cmd.Instances.IsSet = true
					cmd.DiskLimit.Value = 50
					cmd.DiskLimit.IsSet = true
					cmd.MemoryLimit.Value = 100
					cmd.MemoryLimit.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v7action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v7action.Warnings{"get-instances-warning"},
						nil)
				})

				When("force flag is not provided", func() {
					When("the user chooses default", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("does not scale the app", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say(`Scaling app some-app in org some-org / space some-space as some-user\.\.\.`))
							Expect(testUI.Out).To(Say(`This will cause the app to restart\. Are you sure you want to scale some-app\? \[yN\]:`))
							Expect(testUI.Out).To(Say("Scaling cancelled"))
							Expect(testUI.Out).ToNot(Say("Showing | Stopping | Starting | Waiting"))

							Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(0))
						})
					})

					When("the user chooses no", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("n\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("does not scale the app", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say(`Scaling app some-app in org some-org / space some-space as some-user\.\.\.`))
							Expect(testUI.Out).To(Say(`This will cause the app to restart\. Are you sure you want to scale some-app\? \[yN\]:`))
							Expect(testUI.Out).To(Say("Scaling cancelled"))
							Expect(testUI.Out).ToNot(Say("Showing | Stopping | Starting | Waiting"))

							Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(0))
						})
					})

					When("the user chooses yes", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("y\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						When("polling succeeds and the app has running instances", func() {
							BeforeEach(func() {
								fakeActor.PollStartReturns(v7action.Warnings{"some-poll-warning-1", "some-poll-warning-2"}, nil)
							})

							It("delegates the right appGUID", func() {
								Expect(fakeActor.PollStartArgsForCall(0)).To(Equal("some-app-guid"))
							})

							When("Restarting the app fails to stop the app", func() {
								BeforeEach(func() {
									fakeActor.StopApplicationReturns(v7action.Warnings{"some-restart-warning"}, errors.New("stop-error"))
								})

								It("Prints warnings and returns an error", func() {
									Expect(executeErr).To(MatchError("stop-error"))

									Expect(testUI.Err).To(Say("some-restart-warning"))
								})
							})

							When("Restarting the app fails to start the app", func() {
								BeforeEach(func() {
									fakeActor.StartApplicationReturns(v7action.Application{}, v7action.Warnings{"some-start-warning"}, errors.New("start-error"))
								})

								It("Delegates the correct appGUID", func() {
									actualGUID := fakeActor.StartApplicationArgsForCall(0)
									Expect(actualGUID).To(Equal("some-app-guid"))
								})

								It("Prints warnings and returns an error", func() {
									Expect(executeErr).To(MatchError("start-error"))

									Expect(testUI.Err).To(Say("some-start-warning"))
								})
							})

							It("scales, restarts, and displays scale properties", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say(`Scaling app some-app in org some-org / space some-space as some-user\.\.\.`))
								Expect(testUI.Out).To(Say(`This will cause the app to restart\. Are you sure you want to scale some-app\? \[yN\]:`))
								Expect(testUI.Out).To(Say(`Stopping app some-app in org some-org / space some-space as some-user\.\.\.`))
								Expect(testUI.Out).To(Say(`Starting app some-app in org some-org / space some-space as some-user\.\.\.`))

								// Note that this does test that the disk quota was scaled to 96M,
								// it is tested below when we check the arguments
								// passed to ScaleProcessByApplication
								firstAppTable := helpers.ParseV3AppProcessTable(output.Contents())
								Expect(len(firstAppTable.Processes)).To(Equal(2))

								webProcessSummary := firstAppTable.Processes[0]
								Expect(webProcessSummary.Type).To(Equal("web"))
								Expect(webProcessSummary.InstanceCount).To(Equal("3/3"))
								Expect(webProcessSummary.MemUsage).To(Equal("32M"))

								Expect(webProcessSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
								Expect(webProcessSummary.Instances[0].Disk).To(Equal("976.6K of 1.9M"))
								Expect(webProcessSummary.Instances[0].CPU).To(Equal("0.0%"))

								Expect(webProcessSummary.Instances[1].Memory).To(Equal("1.9M of 32M"))
								Expect(webProcessSummary.Instances[1].Disk).To(Equal("1.9M of 3.8M"))
								Expect(webProcessSummary.Instances[1].CPU).To(Equal("0.0%"))

								Expect(webProcessSummary.Instances[2].Memory).To(Equal("2.9M of 32M"))
								Expect(webProcessSummary.Instances[2].Disk).To(Equal("2.9M of 5.7M"))
								Expect(webProcessSummary.Instances[2].CPU).To(Equal("0.0%"))

								consoleProcessSummary := firstAppTable.Processes[1]
								Expect(consoleProcessSummary.Type).To(Equal("console"))
								Expect(consoleProcessSummary.InstanceCount).To(Equal("1/1"))
								Expect(consoleProcessSummary.MemUsage).To(Equal("16M"))

								Expect(consoleProcessSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
								Expect(consoleProcessSummary.Instances[0].Disk).To(Equal("976.6K of 7.6M"))
								Expect(consoleProcessSummary.Instances[0].CPU).To(Equal("0.0%"))

								Expect(testUI.Err).To(Say("get-app-warning"))
								Expect(testUI.Err).To(Say("scale-warning"))
								Expect(testUI.Err).To(Say("some-poll-warning-1"))
								Expect(testUI.Err).To(Say("some-poll-warning-2"))
								Expect(testUI.Err).To(Say("get-instances-warning"))

								Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
								appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
								Expect(appNameArg).To(Equal(appName))
								Expect(spaceGUIDArg).To(Equal("some-space-guid"))

								Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(1))
								appGUIDArg, scaleProcess := fakeActor.ScaleProcessByApplicationArgsForCall(0)
								Expect(appGUIDArg).To(Equal("some-app-guid"))
								Expect(scaleProcess).To(Equal(v7action.Process{
									Type:       constant.ProcessTypeWeb,
									Instances:  types.NullInt{Value: 2, IsSet: true},
									DiskInMB:   types.NullUint64{Value: 50, IsSet: true},
									MemoryInMB: types.NullUint64{Value: 100, IsSet: true},
								}))

								Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
								Expect(fakeActor.StopApplicationArgsForCall(0)).To(Equal("some-app-guid"))

								Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
								Expect(fakeActor.StartApplicationArgsForCall(0)).To(Equal("some-app-guid"))
							})
						})

						When("polling succeeds but all the app's instances have crashed", func() {
							BeforeEach(func() {
								fakeActor.PollStartReturns(v7action.Warnings{"some-poll-warning-1", "some-poll-warning-2"}, actionerror.AllInstancesCrashedError{})
							})

							It("delegates the right appGUID", func() {
								Expect(fakeActor.PollStartArgsForCall(0)).To(Equal("some-app-guid"))
							})

							It("displays the process table", func () {
								Expect(testUI.Out).To(Say("Showing current scale of app " + appName))
							})

							It("displays all warnings and fails", func() {
								Expect(testUI.Err).To(Say("some-poll-warning-1"))
								Expect(testUI.Err).To(Say("some-poll-warning-2"))

								Expect(executeErr).To(MatchError(translatableerror.ApplicationUnableToStartError{
									AppName: appName,
									BinaryName: binaryName,
								}))
							})
						})

						When("polling the start fails", func() {
							BeforeEach(func() {
								fakeActor.PollStartReturns(v7action.Warnings{"some-poll-warning-1", "some-poll-warning-2"}, errors.New("some-error"))
							})

							It("delegates the right appGUID", func() {
								Expect(fakeActor.PollStartArgsForCall(0)).To(Equal("some-app-guid"))
							})

							It("displays all warnings and fails", func() {
								Expect(testUI.Err).To(Say("some-poll-warning-1"))
								Expect(testUI.Err).To(Say("some-poll-warning-2"))

								Expect(executeErr).To(MatchError("some-error"))
							})
						})

						When("polling times out", func() {
							BeforeEach(func() {
								fakeActor.PollStartReturns(nil, actionerror.StartupTimeoutError{})
							})

							It("delegates the right appGUID", func() {
								Expect(fakeActor.PollStartArgsForCall(0)).To(Equal("some-app-guid"))
							})

							It("returns the StartupTimeoutError", func() {
								Expect(executeErr).To(MatchError(translatableerror.StartupTimeoutError{
									AppName:    "some-app",
									BinaryName: binaryName,
								}))
							})
						})
					})
				})

				When("force flag is provided", func() {
					BeforeEach(func() {
						cmd.Force = true
					})

					It("does not prompt user to confirm app restart", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`Scaling app some-app in org some-org / space some-space as some-user\.\.\.`))
						Expect(testUI.Out).To(Say(`Stopping app some-app in org some-org / space some-space as some-user\.\.\.`))
						Expect(testUI.Out).To(Say(`Starting app some-app in org some-org / space some-space as some-user\.\.\.`))
						Expect(testUI.Out).NotTo(Say(`This will cause the app to restart\. Are you sure you want to scale some-app\? \[yN\]:`))

						Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
						Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(1))
						Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
						Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
						Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
					})
				})
			})

			When("only the instances flag option is provided", func() {
				BeforeEach(func() {
					cmd.Instances.Value = 3
					cmd.Instances.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v7action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v7action.Warnings{"get-instances-warning"},
						nil)
				})

				It("scales the number of instances, displays scale properties, and does not restart the application", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Scaling"))
					Expect(testUI.Out).NotTo(Say("This will cause the app to restart | Stopping | Starting"))

					Expect(testUI.Err).To(Say("get-app-warning"))
					Expect(testUI.Err).To(Say("scale-warning"))
					Expect(testUI.Err).To(Say("get-instances-warning"))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appNameArg).To(Equal(appName))
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(1))
					appGUIDArg, scaleProcess := fakeActor.ScaleProcessByApplicationArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(scaleProcess).To(Equal(v7action.Process{
						Type:      constant.ProcessTypeWeb,
						Instances: types.NullInt{Value: 3, IsSet: true},
					}))

					Expect(fakeActor.StopApplicationCallCount()).To(Equal(0))
					Expect(fakeActor.StartApplicationCallCount()).To(Equal(0))
				})
			})

			When("only the memory flag option is provided", func() {
				BeforeEach(func() {
					cmd.MemoryLimit.Value = 256
					cmd.MemoryLimit.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v7action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v7action.Warnings{"get-instances-warning"},
						nil)

					_, err := input.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("scales, restarts, and displays scale properties", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Scaling"))
					Expect(testUI.Out).To(Say("This will cause the app to restart"))
					Expect(testUI.Out).To(Say("Stopping"))
					Expect(testUI.Out).To(Say("Starting"))

					Expect(testUI.Err).To(Say("get-app-warning"))
					Expect(testUI.Err).To(Say("scale-warning"))
					Expect(testUI.Err).To(Say("get-instances-warning"))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appNameArg).To(Equal(appName))
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(1))
					appGUIDArg, scaleProcess := fakeActor.ScaleProcessByApplicationArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(scaleProcess).To(Equal(v7action.Process{
						Type:       constant.ProcessTypeWeb,
						MemoryInMB: types.NullUint64{Value: 256, IsSet: true},
					}))

					Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
					appGUID := fakeActor.StopApplicationArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))

					Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
					appGUID = fakeActor.StartApplicationArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))

					Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
				})
			})

			When("only the disk flag option is provided", func() {
				BeforeEach(func() {
					cmd.DiskLimit.Value = 1025
					cmd.DiskLimit.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v7action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v7action.Warnings{"get-instances-warning"},
						nil)
					_, err := input.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("scales the number of instances, displays scale properties, and restarts the application", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Scaling"))
					Expect(testUI.Out).To(Say("This will cause the app to restart"))
					Expect(testUI.Out).To(Say("Stopping"))
					Expect(testUI.Out).To(Say("Starting"))

					Expect(testUI.Err).To(Say("get-app-warning"))
					Expect(testUI.Err).To(Say("scale-warning"))
					Expect(testUI.Err).To(Say("get-instances-warning"))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appNameArg).To(Equal(appName))
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(1))
					appGUIDArg, scaleProcess := fakeActor.ScaleProcessByApplicationArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(scaleProcess).To(Equal(v7action.Process{
						Type:     constant.ProcessTypeWeb,
						DiskInMB: types.NullUint64{Value: 1025, IsSet: true},
					}))

					Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
					appGUID := fakeActor.StopApplicationArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))

					Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
					appGUID = fakeActor.StartApplicationArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))
				})
			})

			When("process flag is provided", func() {
				BeforeEach(func() {
					cmd.ProcessType = "some-process-type"
					cmd.Instances.Value = 2
					cmd.Instances.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v7action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v7action.Warnings{"get-instances-warning"},
						nil)
					_, err := input.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("scales the specified process", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Scaling"))

					Expect(testUI.Err).To(Say("get-app-warning"))
					Expect(testUI.Err).To(Say("scale-warning"))
					Expect(testUI.Err).To(Say("get-instances-warning"))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appNameArg).To(Equal(appName))
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(1))
					appGUIDArg, scaleProcess := fakeActor.ScaleProcessByApplicationArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(scaleProcess).To(Equal(v7action.Process{
						Type:      "some-process-type",
						Instances: types.NullInt{Value: 2, IsSet: true},
					}))
				})
			})

			When("an error is encountered scaling the application", func() {
				var expectedErr error

				BeforeEach(func() {
					cmd.Instances.Value = 3
					cmd.Instances.IsSet = true
					expectedErr = errors.New("scale process error")
					fakeActor.ScaleProcessByApplicationReturns(
						v7action.Warnings{"scale-process-warning"},
						expectedErr,
					)
				})

				It("returns the error and displays all warnings", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(testUI.Err).To(Say("scale-process-warning"))
				})
			})
		})
	})
})
