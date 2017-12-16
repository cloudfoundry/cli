package v3_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/command/v3/shared/sharedfakes"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-scale Command", func() {
	var (
		cmd             v3.V3ScaleCommand
		input           *Buffer
		output          *Buffer
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3ScaleActor
		fakeV2Actor     *sharedfakes.FakeV2AppRouteActor
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
		fakeActor = new(v3fakes.FakeV3ScaleActor)
		fakeV2Actor = new(sharedfakes.FakeV2AppRouteActor)
		appName = "some-app"

		cmd = v3.V3ScaleCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			AppSummaryDisplayer: shared.AppSummaryDisplayer{
				UI:              testUI,
				Config:          fakeConfig,
				Actor:           fakeActor,
				V2AppRouteActor: fakeV2Actor,
				AppName:         appName,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd.RequiredArgs.AppName = appName
		cmd.ProcessType = constant.ProcessTypeWeb

		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionV3)
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
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
			fakeConfig.HasTargetedSpaceReturns(true)
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space"})
			fakeConfig.CurrentUserReturns(
				configv3.User{Name: "some-user"},
				nil)
		})

		Context("when getting the current user returns an error", func() {
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

		Context("when the application does not exist", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v3action.Application{},
					v3action.Warnings{"get-app-warning"},
					actionerror.ApplicationNotFoundError{Name: appName})
			})

			It("returns an ApplicationNotFoundError and all warnings", func() {
				Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: appName}))

				Expect(testUI.Out).ToNot(Say("Showing"))
				Expect(testUI.Out).ToNot(Say("Scaling"))
				Expect(testUI.Err).To(Say("get-app-warning"))

				Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appNameArg).To(Equal(appName))
				Expect(spaceGUIDArg).To(Equal("some-space-guid"))
			})
		})

		Context("when an error occurs getting the application", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get app error")
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v3action.Application{},
					v3action.Warnings{"get-app-warning"},
					expectedErr)
			})

			It("returns the error and displays all warnings", func() {
				Expect(executeErr).To(Equal(expectedErr))
				Expect(testUI.Err).To(Say("get-app-warning"))
			})
		})

		Context("when the application exists", func() {
			var appSummary v3action.ApplicationSummary

			BeforeEach(func() {
				appSummary = v3action.ApplicationSummary{
					ProcessSummaries: v3action.ProcessSummaries{
						{
							Process: v3action.Process{
								Type:       constant.ProcessTypeWeb,
								MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
								DiskInMB:   types.NullUint64{Value: 1024, IsSet: true},
							},
							InstanceDetails: []v3action.ProcessInstance{
								v3action.ProcessInstance{
									Index:       0,
									State:       constant.ProcessInstanceRunning,
									MemoryUsage: 1000000,
									DiskUsage:   1000000,
									MemoryQuota: 33554432,
									DiskQuota:   2000000,
									Uptime:      int(time.Now().Sub(time.Unix(267321600, 0)).Seconds()),
								},
								v3action.ProcessInstance{
									Index:       1,
									State:       constant.ProcessInstanceRunning,
									MemoryUsage: 2000000,
									DiskUsage:   2000000,
									MemoryQuota: 33554432,
									DiskQuota:   4000000,
									Uptime:      int(time.Now().Sub(time.Unix(330480000, 0)).Seconds()),
								},
								v3action.ProcessInstance{
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
							Process: v3action.Process{
								Type:       "console",
								MemoryInMB: types.NullUint64{Value: 16, IsSet: true},
								DiskInMB:   types.NullUint64{Value: 512, IsSet: true},
							},
							InstanceDetails: []v3action.ProcessInstance{
								v3action.ProcessInstance{
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
					v3action.Application{GUID: "some-app-guid"},
					v3action.Warnings{"get-app-warning"},
					nil)
			})

			Context("when no flag options are provided", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v3action.Warnings{"get-app-summary-warning"},
						nil)
				})

				It("displays current scale properties and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).ToNot(Say("Scaling"))
					Expect(testUI.Out).ToNot(Say("This will cause the app to restart"))
					Expect(testUI.Out).ToNot(Say("Stopping"))
					Expect(testUI.Out).ToNot(Say("Starting"))
					Expect(testUI.Out).ToNot(Say("Waiting"))
					Expect(testUI.Out).To(Say("Showing current scale of app some-app in org some-org / space some-space as some-user\\.\\.\\."))

					firstAppTable := helpers.ParseV3AppProcessTable(output.Contents())
					Expect(len(firstAppTable.Processes)).To(Equal(2))

					webProcessSummary := firstAppTable.Processes[0]
					Expect(webProcessSummary.Title).To(Equal("web:3/3"))

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
					Expect(consoleProcessSummary.Title).To(Equal("console:1/1"))

					Expect(consoleProcessSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
					Expect(consoleProcessSummary.Instances[0].Disk).To(Equal("976.6K of 7.6M"))
					Expect(consoleProcessSummary.Instances[0].CPU).To(Equal("0.0%"))

					Expect(testUI.Err).To(Say("get-app-warning"))
					Expect(testUI.Err).To(Say("get-app-summary-warning"))

					Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
					passedAppName, spaceName := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
					Expect(passedAppName).To(Equal("some-app"))
					Expect(spaceName).To(Equal("some-space-guid"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(0))
				})

				Context("when an error is encountered getting process information", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get process error")
						fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
							v3action.ApplicationSummary{},
							v3action.Warnings{"get-process-warning"},
							expectedErr,
						)
					})

					It("returns the error and displays all warnings", func() {
						Expect(executeErr).To(Equal(expectedErr))
						Expect(testUI.Err).To(Say("get-process-warning"))
					})
				})
			})

			Context("when all flag options are provided", func() {
				BeforeEach(func() {
					cmd.Instances.Value = 2
					cmd.Instances.IsSet = true
					cmd.DiskLimit.Value = 50
					cmd.DiskLimit.IsSet = true
					cmd.MemoryLimit.Value = 100
					cmd.MemoryLimit.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v3action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v3action.Warnings{"get-instances-warning"},
						nil)
				})

				Context("when force flag is not provided", func() {
					Context("when the user chooses default", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("does not scale the app", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).ToNot(Say("Showing"))
							Expect(testUI.Out).To(Say("Scaling app some-app in org some-org / space some-space as some-user\\.\\.\\."))
							Expect(testUI.Out).To(Say("This will cause the app to restart\\. Are you sure you want to scale some-app\\? \\[yN\\]:"))
							Expect(testUI.Out).To(Say("Scaling cancelled"))
							Expect(testUI.Out).ToNot(Say("Stopping"))
							Expect(testUI.Out).ToNot(Say("Starting"))
							Expect(testUI.Out).ToNot(Say("Waiting"))

							Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(0))
						})
					})

					Context("when the user chooses no", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("n\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("does not scale the app", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).ToNot(Say("Showing"))
							Expect(testUI.Out).To(Say("Scaling app some-app in org some-org / space some-space as some-user\\.\\.\\."))
							Expect(testUI.Out).To(Say("This will cause the app to restart\\. Are you sure you want to scale some-app\\? \\[yN\\]:"))
							Expect(testUI.Out).To(Say("Scaling cancelled"))
							Expect(testUI.Out).ToNot(Say("Stopping"))
							Expect(testUI.Out).ToNot(Say("Starting"))
							Expect(testUI.Out).ToNot(Say("Waiting"))

							Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(0))
						})
					})

					Context("when the user chooses yes", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("y\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						Context("when polling succeeds", func() {
							BeforeEach(func() {
								fakeActor.PollStartStub = func(appGUID string, warnings chan<- v3action.Warnings) error {
									warnings <- v3action.Warnings{"some-poll-warning-1", "some-poll-warning-2"}
									return nil
								}
							})

							It("scales, restarts, and displays scale properties", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Scaling app some-app in org some-org / space some-space as some-user\\.\\.\\."))
								Expect(testUI.Out).To(Say("This will cause the app to restart\\. Are you sure you want to scale some-app\\? \\[yN\\]:"))
								Expect(testUI.Out).To(Say("Stopping app some-app in org some-org / space some-space as some-user\\.\\.\\."))
								Expect(testUI.Out).To(Say("Starting app some-app in org some-org / space some-space as some-user\\.\\.\\."))

								// Note that this does test that the disk quota was scaled to 96M,
								// it is tested below when we check the arguments
								// passed to ScaleProcessByApplication
								firstAppTable := helpers.ParseV3AppProcessTable(output.Contents())
								Expect(len(firstAppTable.Processes)).To(Equal(2))

								webProcessSummary := firstAppTable.Processes[0]
								Expect(webProcessSummary.Title).To(Equal("web:3/3"))

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
								Expect(consoleProcessSummary.Title).To(Equal("console:1/1"))

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
								Expect(scaleProcess).To(Equal(v3action.Process{
									Type:       constant.ProcessTypeWeb,
									Instances:  types.NullInt{Value: 2, IsSet: true},
									DiskInMB:   types.NullUint64{Value: 50, IsSet: true},
									MemoryInMB: types.NullUint64{Value: 100, IsSet: true},
								}))

								Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
								Expect(fakeActor.StopApplicationArgsForCall(0)).To(Equal("some-app-guid"))

								Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
								Expect(fakeActor.StartApplicationArgsForCall(0)).To(Equal("some-app-guid"))

								Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
								passedAppName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
								Expect(passedAppName).To(Equal("some-app"))
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
					})
				})

				Context("when force flag is provided", func() {
					BeforeEach(func() {
						cmd.Force = true
					})

					It("does not prompt user to confirm app restart", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Scaling app some-app in org some-org / space some-space as some-user\\.\\.\\."))
						Expect(testUI.Out).NotTo(Say("This will cause the app to restart\\. Are you sure you want to scale some-app\\? \\[yN\\]:"))
						Expect(testUI.Out).To(Say("Stopping app some-app in org some-org / space some-space as some-user\\.\\.\\."))
						Expect(testUI.Out).To(Say("Starting app some-app in org some-org / space some-space as some-user\\.\\.\\."))

						Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
						Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(1))
						Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
						Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
						Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
					})
				})
			})

			Context("when only the instances flag option is provided", func() {
				BeforeEach(func() {
					cmd.Instances.Value = 3
					cmd.Instances.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v3action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v3action.Warnings{"get-instances-warning"},
						nil)
				})

				It("scales the number of instances, displays scale properties, and does not restart the application", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Scaling"))
					Expect(testUI.Out).NotTo(Say("This will cause the app to restart"))
					Expect(testUI.Out).NotTo(Say("Stopping"))
					Expect(testUI.Out).NotTo(Say("Starting"))

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
					Expect(scaleProcess).To(Equal(v3action.Process{
						Type:      constant.ProcessTypeWeb,
						Instances: types.NullInt{Value: 3, IsSet: true},
					}))

					Expect(fakeActor.StopApplicationCallCount()).To(Equal(0))
					Expect(fakeActor.StartApplicationCallCount()).To(Equal(0))

					Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
					passedAppName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
					Expect(passedAppName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when only the memory flag option is provided", func() {
				BeforeEach(func() {
					cmd.MemoryLimit.Value = 256
					cmd.MemoryLimit.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v3action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v3action.Warnings{"get-instances-warning"},
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
					Expect(scaleProcess).To(Equal(v3action.Process{
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
					passedAppName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
					Expect(passedAppName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when only the disk flag option is provided", func() {
				BeforeEach(func() {
					cmd.DiskLimit.Value = 1025
					cmd.DiskLimit.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v3action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v3action.Warnings{"get-instances-warning"},
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
					Expect(scaleProcess).To(Equal(v3action.Process{
						Type:     constant.ProcessTypeWeb,
						DiskInMB: types.NullUint64{Value: 1025, IsSet: true},
					}))

					Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
					appGUID := fakeActor.StopApplicationArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))

					Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
					appGUID = fakeActor.StartApplicationArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))

					Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
					passedAppName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
					Expect(passedAppName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when process flag is provided", func() {
				BeforeEach(func() {
					cmd.ProcessType = "some-process-type"
					cmd.Instances.Value = 2
					cmd.Instances.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v3action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
						appSummary,
						v3action.Warnings{"get-instances-warning"},
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
					Expect(scaleProcess).To(Equal(v3action.Process{
						Type:      "some-process-type",
						Instances: types.NullInt{Value: 2, IsSet: true},
					}))

					Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
					passedAppName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
					Expect(passedAppName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when an error is encountered scaling the application", func() {
				var expectedErr error

				BeforeEach(func() {
					cmd.Instances.Value = 3
					cmd.Instances.IsSet = true
					expectedErr = errors.New("scale process error")
					fakeActor.ScaleProcessByApplicationReturns(
						v3action.Warnings{"scale-process-warning"},
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
