package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
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
)

var _ = FDescribe("Scale Command", func() {
	var (
		cmd             v3.V3ScaleCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3ScaleActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3ScaleActor)

		cmd = v3.V3ScaleCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.AppName = "some-app"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
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
					v3action.Warnings{"get-app-warning-1", "get-app-warning-2"},
					v3action.ApplicationNotFoundError{Name: "some-app"})
			})

			It("returns an ApplicationNotFoundError and all warnings", func() {
				Expect(executeErr).To(Equal(translatableerror.ApplicationNotFoundError{Name: "some-app"}))

				Expect(testUI.Out).ToNot(Say("Showing"))
				Expect(testUI.Out).ToNot(Say("Scaling"))
				Expect(testUI.Err).To(Say("get-app-warning-1"))
				Expect(testUI.Err).To(Say("get-app-warning-2"))

				Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appNameArg).To(Equal("some-app"))
				Expect(spaceGUIDArg).To(Equal("some-space-guid"))
			})
		})

		Context("when an error occurs getting the application", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get app error")
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v3action.Application{},
					v3action.Warnings{"get-app-warning-1", "get-app-warning-2"},
					expectedErr)
			})

			It("returns the error and displays all warnings", func() {
				Expect(executeErr).To(Equal(expectedErr))

				Expect(testUI.Err).To(Say("get-app-warning-1"))
				Expect(testUI.Err).To(Say("get-app-warning-2"))
			})
		})

		Context("when the application exists", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v3action.Application{GUID: "some-app-guid"},
					v3action.Warnings{"get-app-warning-1", "get-app-warning-2"},
					nil)
			})

			Context("when no flag options are provided", func() {
				BeforeEach(func() {
					fakeActor.GetProcessByApplicationReturns(
						ccv3.Process{
							Type:       "web",
							Instances:  2,
							MemoryInMB: 128,
							DiskInMB:   2048,
						},
						v3action.Warnings{"get-process-warning-1", "get-process-warning-2"},
						nil)
				})

				It("displays current scale properties and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).ToNot(Say("Scaling"))
					Expect(testUI.Out).To(Say("Showing current scale of app some-app in org some-org / space some-space as some-user\\.\\.\\."))
					Expect(testUI.Out).To(Say("instances:\\s+2"))
					Expect(testUI.Out).To(Say("memory:\\s+128M"))
					Expect(testUI.Out).To(Say("disk:\\s+2G"))

					Expect(testUI.Err).To(Say("get-app-warning-1"))
					Expect(testUI.Err).To(Say("get-app-warning-2"))
					Expect(testUI.Err).To(Say("get-process-warning-1"))
					Expect(testUI.Err).To(Say("get-process-warning-2"))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appNameArg).To(Equal("some-app"))
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))

					Expect(fakeActor.GetProcessByApplicationCallCount()).To(Equal(1))
					Expect(fakeActor.GetProcessByApplicationArgsForCall(0)).To(Equal("some-app-guid"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(0))
				})

				Context("when an error is encountered getting process information", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get process error")
						fakeActor.GetProcessByApplicationReturns(
							ccv3.Process{},
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
					cmd.Instances = 3
					cmd.MemoryLimit = flag.Megabytes{Size: 256}
					cmd.DiskLimit = flag.Megabytes{Size: 64}

					fakeActor.ScaleProcessByApplicationReturns(
						ccv3.Process{
							Instances:  3,
							MemoryInMB: 256,
							DiskInMB:   64,
						},
						v3action.Warnings{"scale-warning-1", "scale-warning-2"},
						nil,
					)
				})

				It("scales the application", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).ToNot(Say("Showing"))
					Expect(testUI.Out).To(Say("Scaling app some-app in org some-org / space some-space as some-user\\.\\.\\."))
					Expect(testUI.Out).To(Say("instances:\\s+3"))
					Expect(testUI.Out).To(Say("memory:\\s+256M"))
					Expect(testUI.Out).To(Say("disk:\\s+64M"))

					Expect(testUI.Err).To(Say("get-app-warning-1"))
					Expect(testUI.Err).To(Say("get-app-warning-2"))
					Expect(testUI.Err).To(Say("scale-warning-1"))
					Expect(testUI.Err).To(Say("scale-warning-2"))

					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appNameArg, spaceGUIDArg := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appNameArg).To(Equal("some-app"))
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))

					Expect(fakeActor.ScaleProcessByApplicationCallCount()).To(Equal(1))
					appGUIDArg, processArg := fakeActor.ScaleProcessByApplicationArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(processArg).To(Equal(ccv3.Process{
						Type:       "web",
						Instances:  3,
						MemoryInMB: 256,
						DiskInMB:   64,
					}))

					Expect(fakeActor.GetProcessByApplicationCallCount()).To(Equal(0))
				})

				Context("when an error is encountered scaling the application", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("scale process error")
						fakeActor.ScaleProcessByApplicationReturns(
							ccv3.Process{},
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

			FContext("when the '-i' flag option is not provided", func() {
				It("should not set the default value of 0 for number of instances", func() {})
			})

			XContext("when only the instances flag option is provided", func() {
				It("scales the number of instances and does not restart the application", func() {
				})
			})

			XContext("when only the instances flag option is not provided", func() {
				It("scales the disk and memory of all instances and restarts the application", func() {
				})
			})
		})
	})
})
