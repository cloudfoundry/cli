package v2_test

import (
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	"github.com/cloudfoundry/bytefmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("App Command", func() {
	var (
		cmd             v2.AppCommand
		testUI          *ui.UI
		fakeSharedActor *v2fakes.FakeSharedActor
		fakeActor       *v2fakes.FakeAppActor
		fakeConfig      *commandfakes.FakeConfig
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
		fakeSharedActor = new(v2fakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeAppActor)
		fakeConfig = new(commandfakes.FakeConfig)

		cmd = v2.AppCommand{
			UI:          testUI,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			Config:      fakeConfig,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		fakeConfig.ExperimentalReturns(true)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error if the check fails", func() {
			Expect(executeErr).To(MatchError(command.NotLoggedInError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when a cloud controller API endpoint is set", func() {
		Context("when the user is logged in, and org and space are targeted", func() {
			BeforeEach(func() {
				fakeConfig.HasTargetedOrganizationReturns(true)
				fakeConfig.HasTargetedSpaceReturns(true)
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					GUID: "some-space-guid",
					Name: "some-space",
				})
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
				fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
				cmd.RequiredArgs.AppName = "some-app"
			})

			It("displays app flavor text", func() {
				Expect(testUI.Out).To(Say("Showing health and status for app some-app in org some-org / space some-space as some-user..."))
			})

			Context("when the passed the --guid flag", func() {
				BeforeEach(func() {
					cmd.GUID = true
				})

				Context("when retrieving the app information fails", func() {
					BeforeEach(func() {
						warnings := v2action.Warnings{"warning-1", "warning-2"}
						fakeActor.GetApplicationByNameAndSpaceReturns(v2action.Application{}, warnings, v2action.ApplicationNotFoundError{Name: "some-app"})
					})

					It("returns the err", func() {
						Expect(executeErr).To(Equal(command.ApplicationNotFoundError{Name: "some-app"}))
					})

					It("sends all warnings to stderr", func() {
						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
					})
				})

				Context("when no errors occur", func() {
					BeforeEach(func() {
						warnings := v2action.Warnings{"warning-1", "warning-2"}
						fakeActor.GetApplicationByNameAndSpaceReturns(v2action.Application{GUID: "some-guid"}, warnings, nil)
					})

					It("displays the application guid", func() {
						Expect(testUI.Out).To(Say("some-guid"))
					})

					It("sends all warnings to stderr", func() {
						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
					})
				})
			})

			Context("when no flags are passed", func() {
				Context("when retrieving application summary fails", func() {
					BeforeEach(func() {
						warnings := v2action.Warnings{"warning-1", "warning-2"}
						fakeActor.GetApplicationSummaryByNameAndSpaceReturns(v2action.ApplicationSummary{}, warnings, v2action.ApplicationNotFoundError{Name: "some-app"})
					})

					It("returns the err", func() {
						Expect(executeErr).To(Equal(command.ApplicationNotFoundError{Name: "some-app"}))
					})

					It("sends all warnings to stderr", func() {
						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
					})
				})

				Context("when no errors occur", func() {
					BeforeEach(func() {
						warnings := v2action.Warnings{"warning-1", "warning-2"}
						applicationSummary := v2action.ApplicationSummary{
							Application: v2action.Application{
								Name:              "some-app",
								GUID:              "some-guid",
								Instances:         3,
								Memory:            128,
								PackageUpdatedAt:  time.Unix(0, 0),
								DetectedBuildpack: "some-buildpack",
							},
							Stack: v2action.Stack{
								Name: "potatos",
							},
							RunningInstances: []v2action.ApplicationInstance{
								{
									CPU:         0.73,
									DiskQuota:   2048 * bytefmt.MEGABYTE,
									Disk:        50 * bytefmt.MEGABYTE,
									ID:          0,
									Memory:      100 * bytefmt.MEGABYTE,
									MemoryQuota: 128 * bytefmt.MEGABYTE,
									State:       ccv2.ApplicationInstanceRunning,
									Uptime:      0,
								},
								{
									CPU:         0.37,
									DiskQuota:   2048 * bytefmt.MEGABYTE,
									Disk:        50 * bytefmt.MEGABYTE,
									ID:          1,
									Memory:      100 * bytefmt.MEGABYTE,
									MemoryQuota: 128 * bytefmt.MEGABYTE,
									State:       ccv2.ApplicationInstanceCrashed,
									Uptime:      0,
								},
							},
							Routes: []v2action.Route{
								{
									Host:   "banana",
									Domain: "fruit.com",
									Path:   "/hi",
								},
								{
									Domain: "foobar.com",
									Port:   13,
								},
							},
						}
						fakeActor.GetApplicationSummaryByNameAndSpaceReturns(applicationSummary, warnings, nil)
					})

					It("displays the health and status header", func() {
						Expect(testUI.Out).To(Say(
							"Showing health and status for app some-app in org some-org / space some-space as some-user..."))
					})

					It("sends all warnings to stderr", func() {
						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
					})

					It("shows the app summary", func() {
						Expect(testUI.Out).To(Say("Name:\\s+some-app"))
						Expect(testUI.Out).To(Say("Instances:\\s+2\\/3"))
						Expect(testUI.Out).To(Say("Usage:\\s+128M x 3 instances"))
						Expect(testUI.Out).To(Say("Routes:\\s+banana.fruit.com/hi, foobar.com:13"))
						Expect(testUI.Out).To(Say("Last uploaded:\\s+1970-01-01T00:00:00Z"))
						Expect(testUI.Out).To(Say("Stack:\\s+potatos"))
						Expect(testUI.Out).To(Say("Buildpack:\\s+some-buildpack"))

						timeFormat := testUI.UserFriendlyDate(time.Now())
						Expect(testUI.Out).To(Say("State\\s+Since\\s+CPU\\s+Memory\\s+Disk"))
						Expect(testUI.Out).To(Say("#0\\s+running\\s+%s\\s+73.0%%\\s+100M of 128M\\s+50M of 2G", timeFormat))
						Expect(testUI.Out).To(Say("#1\\s+crashed\\s+%s\\s+37.0%%\\s+100M of 128M\\s+50M of 2G", timeFormat))

						Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
						appName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
						Expect(appName).To(Equal("some-app"))
						Expect(spaceGUID).To(Equal("some-space-guid"))
					})

					//TODO: 0/1 instances
					//TODO: 0/0 instances
					//TODO: unknown buildpack
				})
			})
		})
	})
})
