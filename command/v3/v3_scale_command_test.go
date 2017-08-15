package v3_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Scale Command", func() {
	var (
		cmd             v3.V3ScaleCommand
		input           *Buffer
		output          *Buffer
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3ScaleActor
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
				V2AppRouteActor: nil,
				AppName:         appName,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd.RequiredArgs.AppName = appName
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
					v3action.ApplicationNotFoundError{Name: appName})
			})

			It("returns an ApplicationNotFoundError and all warnings", func() {
				Expect(executeErr).To(Equal(translatableerror.ApplicationNotFoundError{Name: appName}))

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
			var process v3action.Process

			BeforeEach(func() {
				process = v3action.Process{
					Type:       "web",
					MemoryInMB: 32,
					Instances: []v3action.Instance{
						v3action.Instance{
							Index:       0,
							State:       "RUNNING",
							MemoryUsage: 1000000,
							DiskUsage:   1000000,
							MemoryQuota: 33554432,
							DiskQuota:   2000000,
							Uptime:      int(time.Now().Sub(time.Unix(267321600, 0)).Seconds()),
						},
						v3action.Instance{
							Index:       1,
							State:       "RUNNING",
							MemoryUsage: 2000000,
							DiskUsage:   2000000,
							MemoryQuota: 33554432,
							DiskQuota:   4000000,
							Uptime:      int(time.Now().Sub(time.Unix(330480000, 0)).Seconds()),
						},
						v3action.Instance{
							Index:       2,
							State:       "RUNNING",
							MemoryUsage: 3000000,
							DiskUsage:   3000000,
							MemoryQuota: 33554432,
							DiskQuota:   6000000,
							Uptime:      int(time.Now().Sub(time.Unix(1277164800, 0)).Seconds()),
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
					fakeActor.GetInstancesByApplicationAndProcessTypeReturns(
						process,
						v3action.Warnings{"get-instance-warning"},
						nil)
				})

				It("displays current scale properties and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).ToNot(Say("Scaling"))
					Expect(testUI.Out).To(Say("Showing current scale of app some-app in org some-org / space some-space as some-user\\.\\.\\."))

					firstAppTable := helpers.ParseV3AppTable(output.Contents())
					Expect(len(firstAppTable.Processes)).To(Equal(1))

					processSummary := firstAppTable.Processes[0]
					Expect(processSummary.Title).To(Equal("web:3/3"))

					Expect(processSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
					Expect(processSummary.Instances[0].Disk).To(Equal("976.6K of 1.9M"))
					Expect(processSummary.Instances[0].CPU).To(Equal("0.0%"))

					Expect(processSummary.Instances[1].Memory).To(Equal("1.9M of 32M"))
					Expect(processSummary.Instances[1].Disk).To(Equal("1.9M of 3.8M"))
					Expect(processSummary.Instances[1].CPU).To(Equal("0.0%"))

					Expect(processSummary.Instances[2].Memory).To(Equal("2.9M of 32M"))
					Expect(processSummary.Instances[2].Disk).To(Equal("2.9M of 5.7M"))
					Expect(processSummary.Instances[2].CPU).To(Equal("0.0%"))

					Expect(testUI.Err).To(Say("get-app-warning"))
					Expect(testUI.Err).To(Say("get-instance-warning"))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appNameArg).To(Equal(appName))
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))

					Expect(fakeActor.GetInstancesByApplicationAndProcessTypeCallCount()).To(Equal(1))
					appGUIDArg, processTypeArg := fakeActor.GetInstancesByApplicationAndProcessTypeArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(processTypeArg).To(Equal("web"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(0))
				})

				Context("when an error is encountered getting process information", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get process error")
						fakeActor.GetInstancesByApplicationAndProcessTypeReturns(
							v3action.Process{},
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
					fakeActor.GetInstancesByApplicationAndProcessTypeReturns(
						process,
						v3action.Warnings{"get-instances-warning"},
						nil)
				})

				Context("when the user chooses the default option", func() {
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

						Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(0))
					})
				})

				Context("when the user is prompted to confirm scaling and chooses yes", func() {
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

						It("scales the app, restarts the app, and displays scale properties", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).ToNot(Say("Showing"))
							Expect(testUI.Out).To(Say("Scaling app some-app in org some-org / space some-space as some-user\\.\\.\\."))
							Expect(testUI.Out).To(Say("This will cause the app to restart\\. Are you sure you want to scale some-app\\? \\[yN\\]:"))
							Expect(testUI.Out).To(Say("Stopping app some-app in org some-org / space some-space as some-user\\.\\.\\."))
							Expect(testUI.Out).To(Say("Starting app some-app in org some-org / space some-space as some-user\\.\\.\\."))
							Expect(testUI.Out).To(Say("Waiting for app to start\\.\\.\\."))

							// Note that this does test that the disk quota was scaled to 96M,
							// it is tested below when we check the arguments
							// passed to ScaleProcessByApplication
							firstAppTable := helpers.ParseV3AppTable(output.Contents())
							Expect(len(firstAppTable.Processes)).To(Equal(1))

							processSummary := firstAppTable.Processes[0]
							Expect(processSummary.Title).To(Equal("web:3/3"))

							Expect(processSummary.Instances[0].Memory).To(Equal("976.6K of 32M"))
							Expect(processSummary.Instances[0].Disk).To(Equal("976.6K of 1.9M"))
							Expect(processSummary.Instances[0].CPU).To(Equal("0.0%"))

							Expect(processSummary.Instances[1].Memory).To(Equal("1.9M of 32M"))
							Expect(processSummary.Instances[1].Disk).To(Equal("1.9M of 3.8M"))
							Expect(processSummary.Instances[1].CPU).To(Equal("0.0%"))

							Expect(processSummary.Instances[2].Memory).To(Equal("2.9M of 32M"))
							Expect(processSummary.Instances[2].Disk).To(Equal("2.9M of 5.7M"))
							Expect(processSummary.Instances[2].CPU).To(Equal("0.0%"))

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
							appGUIDArg, processTypeArg, scaleOptionsArg := fakeActor.ScaleProcessByApplicationArgsForCall(0)
							Expect(appGUIDArg).To(Equal("some-app-guid"))
							Expect(processTypeArg).To(Equal("web"))
							Expect(scaleOptionsArg).To(Equal(v3action.ProcessScaleOptions{
								Instances:  types.NullInt{Value: 2, IsSet: true},
								DiskInMB:   types.NullUint64{Value: 50, IsSet: true},
								MemoryInMB: types.NullUint64{Value: 100, IsSet: true},
							}))

							Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
							Expect(fakeActor.StopApplicationArgsForCall(0)).To(Equal("some-app-guid"))

							Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
							Expect(fakeActor.StartApplicationArgsForCall(0)).To(Equal("some-app-guid"))

							Expect(fakeActor.GetInstancesByApplicationAndProcessTypeCallCount()).To(Equal(1))
							appGUID, processType := fakeActor.GetInstancesByApplicationAndProcessTypeArgsForCall(0)
							Expect(appGUID).To(Equal("some-app-guid"))
							Expect(processType).To(Equal("web"))
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
							Expect(executeErr).To(MatchError(translatableerror.StartupTimeoutError{
								AppName:    "some-app",
								BinaryName: binaryName,
							}))
						})
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
					fakeActor.GetInstancesByApplicationAndProcessTypeReturns(
						process,
						v3action.Warnings{"get-instances-warning"},
						nil)
				})

				It("scales the number of instances, displays scale properties, and does not restart the application", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).ToNot(Say("Showing"))
					Expect(testUI.Out).To(Say("Scaling app some-app in org some-org / space some-space as some-user\\.\\.\\."))

					Expect(testUI.Err).To(Say("get-app-warning"))
					Expect(testUI.Err).To(Say("scale-warning"))
					Expect(testUI.Err).To(Say("get-instances-warning"))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appNameArg).To(Equal(appName))
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(1))
					appGUIDArg, processTypeArg, processScaleOptions := fakeActor.ScaleProcessByApplicationArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(processTypeArg).To(Equal("web"))
					Expect(processScaleOptions).To(Equal(v3action.ProcessScaleOptions{
						Instances: types.NullInt{Value: 3, IsSet: true},
					}))

					Expect(fakeActor.StopApplicationCallCount()).To(Equal(0))
					Expect(fakeActor.StartApplicationCallCount()).To(Equal(0))

					Expect(fakeActor.GetInstancesByApplicationAndProcessTypeCallCount()).To(Equal(1))
					appGUID, processType := fakeActor.GetInstancesByApplicationAndProcessTypeArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))
					Expect(processType).To(Equal("web"))
				})
			})

			Context("when only the memory flag option is provided", func() {
				BeforeEach(func() {
					cmd.MemoryLimit.Value = 256
					cmd.MemoryLimit.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v3action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetInstancesByApplicationAndProcessTypeReturns(
						process,
						v3action.Warnings{"get-instances-warning"},
						nil)

					_, err := input.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("scales the number of instances, displays scale properties, and restarts the application", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).ToNot(Say("Showing"))
					Expect(testUI.Out).To(Say("Scaling app some-app in org some-org / space some-space as some-user\\.\\.\\."))

					Expect(testUI.Err).To(Say("get-app-warning"))
					Expect(testUI.Err).To(Say("scale-warning"))
					Expect(testUI.Err).To(Say("get-instances-warning"))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appNameArg).To(Equal(appName))
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(1))
					appGUIDArg, processTypeArg, processScaleOptions := fakeActor.ScaleProcessByApplicationArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(processTypeArg).To(Equal("web"))
					Expect(processScaleOptions).To(Equal(v3action.ProcessScaleOptions{
						MemoryInMB: types.NullUint64{Value: 256, IsSet: true},
					}))

					Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
					appGUID := fakeActor.StopApplicationArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))

					Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
					appGUID = fakeActor.StartApplicationArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))

					Expect(fakeActor.GetInstancesByApplicationAndProcessTypeCallCount()).To(Equal(1))
					appGUID, processType := fakeActor.GetInstancesByApplicationAndProcessTypeArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))
					Expect(processType).To(Equal("web"))
				})
			})

			Context("when only the memory flag option is provided", func() {
				BeforeEach(func() {
					cmd.DiskLimit.Value = 1024
					cmd.DiskLimit.IsSet = true
					fakeActor.ScaleProcessByApplicationReturns(
						v3action.Warnings{"scale-warning"},
						nil)
					fakeActor.GetInstancesByApplicationAndProcessTypeReturns(
						process,
						v3action.Warnings{"get-instances-warning"},
						nil)
					_, err := input.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("scales the number of instances, displays scale properties, and restarts the application", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).ToNot(Say("Showing"))
					Expect(testUI.Out).To(Say("Scaling app some-app in org some-org / space some-space as some-user\\.\\.\\."))

					Expect(testUI.Err).To(Say("get-app-warning"))
					Expect(testUI.Err).To(Say("scale-warning"))
					Expect(testUI.Err).To(Say("get-instances-warning"))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appNameArg).To(Equal(appName))
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(1))
					appGUIDArg, processTypeArg, processScaleOptions := fakeActor.ScaleProcessByApplicationArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(processTypeArg).To(Equal("web"))
					Expect(processScaleOptions).To(Equal(v3action.ProcessScaleOptions{
						DiskInMB: types.NullUint64{Value: 1024, IsSet: true},
					}))

					Expect(fakeActor.StopApplicationCallCount()).To(Equal(1))
					appGUID := fakeActor.StopApplicationArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))

					Expect(fakeActor.StartApplicationCallCount()).To(Equal(1))
					appGUID = fakeActor.StartApplicationArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))

					Expect(fakeActor.GetInstancesByApplicationAndProcessTypeCallCount()).To(Equal(1))
					appGUID, processType := fakeActor.GetInstancesByApplicationAndProcessTypeArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))
					Expect(processType).To(Equal("web"))
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
