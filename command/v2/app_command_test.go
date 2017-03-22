package v2_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
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
		cmd             AppCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeAppActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeAppActor)

		cmd = AppCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.AppName = "some-app"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		// TODO: remove when experimental flag is removed
		fakeConfig.ExperimentalReturns(true)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	// TODO: remove when experimental flag is removed
	It("Displays the experimental warning message", func() {
		Expect(testUI.Out).To(Say(command.ExperimentalWarning))
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

		It("displays flavor text", func() {
			Expect(testUI.Out).To(Say("Showing health and status for app some-app in org some-org / space some-space as some-user..."))
		})

		Context("when the --guid flag is provided", func() {
			BeforeEach(func() {
				cmd.GUID = true
			})

			Context("when no errors occur", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v2action.Application{GUID: "some-guid"},
						v2action.Warnings{"warning-1", "warning-2"},
						nil)
				})

				It("displays the application guid and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("some-guid"))
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			Context("when an error is encountered getting the app", func() {
				Context("when the error is translatable", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationByNameAndSpaceReturns(
							v2action.Application{},
							v2action.Warnings{"warning-1", "warning-2"},
							v2action.ApplicationNotFoundError{Name: "some-app"})
					})

					It("returns a translatable error and all warnings", func() {
						Expect(executeErr).To(MatchError(command.ApplicationNotFoundError{Name: "some-app"}))

						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
					})
				})

				Context("when the error is not translatable", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get app summary error")
						fakeActor.GetApplicationByNameAndSpaceReturns(
							v2action.Application{},
							v2action.Warnings{"warning-1", "warning-2"},
							expectedErr)
					})

					It("returns the error and all warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
					})
				})
			})
		})

		Context("when the --guid flag is not provided", func() {
			Context("when no errors occur", func() {
				var (
					applicationSummary v2action.ApplicationSummary
					warnings           []string
				)

				BeforeEach(func() {
					applicationSummary = v2action.ApplicationSummary{
						Application: v2action.Application{
							Name:              "some-app",
							GUID:              "some-app-guid",
							Instances:         3,
							Memory:            128,
							PackageUpdatedAt:  time.Unix(0, 0),
							DetectedBuildpack: "some-buildpack",
							State:             "STARTED",
						},
						Stack: v2action.Stack{
							Name: "potatos",
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
					warnings = []string{"app-summary-warning"}
				})

				Context("when the app does not have running instances", func() {
					BeforeEach(func() {
						applicationSummary.RunningInstances = []v2action.ApplicationInstanceWithStats{}
						fakeActor.GetApplicationSummaryByNameAndSpaceReturns(applicationSummary, warnings, nil)
					})

					It("displays the app summary, 'no running instances' message, and all warnings", func() {
						Expect(testUI.Out).To(Say("Showing health and status for app some-app in org some-org / space some-space as some-user..."))
						Expect(testUI.Out).To(Say(""))
						Expect(testUI.Out).To(Say("name:\\s+some-app"))
						Expect(testUI.Out).To(Say("requested state:\\s+started"))
						Expect(testUI.Out).To(Say("instances:\\s+0\\/3"))
						Expect(testUI.Out).To(Say("usage:\\s+128M x 3 instances"))
						Expect(testUI.Out).To(Say("routes:\\s+banana.fruit.com/hi, foobar.com:13"))
						Expect(testUI.Out).To(Say("last uploaded:\\s+\\w{3} [0-3]\\d \\w{3} [0-2]\\d:[0-5]\\d:[0-5]\\d \\w+ \\d{4}"))
						Expect(testUI.Out).To(Say("stack:\\s+potatos"))
						Expect(testUI.Out).To(Say("buildpack:\\s+some-buildpack"))
						Expect(testUI.Out).To(Say(""))
						Expect(testUI.Out).To(Say("There are no running instances of this app"))

						Expect(testUI.Err).To(Say("app-summary-warning"))
					})

					It("should not display the instance table", func() {
						Expect(testUI.Out).NotTo(Say("State\\s+Since\\s+CPU\\s+Memory\\s+Disk"))
					})
				})

				Context("when the app has running instances", func() {
					BeforeEach(func() {
						applicationSummary.RunningInstances = []v2action.ApplicationInstanceWithStats{
							{
								ID:          0,
								State:       v2action.ApplicationInstanceState(ccv2.ApplicationInstanceRunning),
								Since:       1403140717.984577,
								CPU:         0.73,
								Disk:        50 * bytefmt.MEGABYTE,
								DiskQuota:   2048 * bytefmt.MEGABYTE,
								Memory:      100 * bytefmt.MEGABYTE,
								MemoryQuota: 128 * bytefmt.MEGABYTE,
								Details:     "info from the backend",
							},
							{
								ID:          1,
								State:       v2action.ApplicationInstanceState(ccv2.ApplicationInstanceCrashed),
								Since:       1403100000.900000,
								CPU:         0.37,
								Disk:        50 * bytefmt.MEGABYTE,
								DiskQuota:   2048 * bytefmt.MEGABYTE,
								Memory:      100 * bytefmt.MEGABYTE,
								MemoryQuota: 128 * bytefmt.MEGABYTE,
								Details:     "potato",
							},
						}
						fakeActor.GetApplicationSummaryByNameAndSpaceReturns(applicationSummary, warnings, nil)
					})

					It("displays app summary, instance table, and all warnings", func() {
						Expect(testUI.Out).To(Say("Showing health and status for app some-app in org some-org / space some-space as some-user..."))
						Expect(testUI.Out).To(Say(""))
						Expect(testUI.Out).To(Say("name:\\s+some-app"))
						Expect(testUI.Out).To(Say("requested state:\\s+started"))
						Expect(testUI.Out).To(Say("instances:\\s+2\\/3"))
						Expect(testUI.Out).To(Say("usage:\\s+128M x 3 instances"))
						Expect(testUI.Out).To(Say("routes:\\s+banana.fruit.com/hi, foobar.com:13"))
						Expect(testUI.Out).To(Say("last uploaded:\\s+\\w{3} [0-3]\\d \\w{3} [0-2]\\d:[0-5]\\d:[0-5]\\d \\w+ \\d{4}"))
						Expect(testUI.Out).To(Say("stack:\\s+potatos"))
						Expect(testUI.Out).To(Say("buildpack:\\s+some-buildpack"))
						Expect(testUI.Out).To(Say(""))
						Expect(testUI.Out).To(Say("State\\s+Since\\s+CPU\\s+Memory\\s+Disk\\s+Details"))
						Expect(testUI.Out).To(Say(`#0\s+running\s+2014-06-19T01:18:37Z\s+73.0%\s+100M of 128M\s+50M of 2G\s+info from the backend`))
						Expect(testUI.Out).To(Say(`#1\s+crashed\s+2014-06-18T14:00:00Z\s+37.0%\s+100M of 128M\s+50M of 2G\s+potato`))

						Expect(testUI.Err).To(Say("app-summary-warning"))

						Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
						appName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
						Expect(appName).To(Equal("some-app"))
						Expect(spaceGUID).To(Equal("some-space-guid"))
					})
				})
			})

			Context("when an error is encountered getting app summary", func() {
				Context("when the error is not translatable", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get app summary error")
						fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
							v2action.ApplicationSummary{},
							nil,
							expectedErr)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(expectedErr))
					})
				})

				Context("when the error is translatable", func() {
					BeforeEach(func() {
						fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
							v2action.ApplicationSummary{},
							nil,
							v2action.ApplicationNotFoundError{Name: "some-app"})
					})

					It("returns a translatable error", func() {
						Expect(executeErr).To(MatchError(command.ApplicationNotFoundError{Name: "some-app"}))
					})
				})
			})
		})
	})
})
