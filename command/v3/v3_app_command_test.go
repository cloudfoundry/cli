package v3_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/command/v3/shared/sharedfakes"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-app Command", func() {
	var (
		cmd             v3.V3AppCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3AppActor
		fakeV2Actor     *sharedfakes.FakeV2AppRouteActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3AppActor)
		fakeV2Actor = new(sharedfakes.FakeV2AppRouteActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		appSummaryDisplayer := shared.AppSummaryDisplayer{
			UI:              testUI,
			Config:          fakeConfig,
			Actor:           fakeActor,
			V2AppRouteActor: fakeV2Actor,
			AppName:         app,
		}

		cmd = v3.V3AppCommand{
			RequiredArgs: flag.AppName{AppName: app},

			UI:                  testUI,
			Config:              fakeConfig,
			SharedActor:         fakeSharedActor,
			Actor:               fakeActor,
			AppSummaryDisplayer: appSummaryDisplayer,
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})

		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
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

	Context("when the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	Context("when getting the application summary returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = actionerror.ApplicationNotFoundError{Name: app}
			fakeActor.GetApplicationSummaryByNameAndSpaceReturns(v3action.ApplicationSummary{}, v3action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))

			Expect(testUI.Out).To(Say("Showing health and status for app some-app in org some-org / space some-space as steve\\.\\.\\."))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})

	Context("when the --guid flag is provided", func() {
		BeforeEach(func() {
			cmd.GUID = true
		})

		Context("when no errors occur", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v3action.Application{GUID: "some-guid"},
					v3action.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("displays the application guid and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("some-guid"))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal("some-app"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
			})
		})

		Context("when an error is encountered getting the app", func() {
			Context("when the error is translatable", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v3action.Application{},
						v3action.Warnings{"warning-1", "warning-2"},
						actionerror.ApplicationNotFoundError{Name: "some-app"})
				})

				It("returns a translatable error and all warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.ApplicationNotFoundError{Name: "some-app"}))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			Context("when the error is not translatable", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get app summary error")
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v3action.Application{},
						v3action.Warnings{"warning-1", "warning-2"},
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

	Context("when the app is a docker app", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
				v3action.ApplicationSummary{
					Application: v3action.Application{
						GUID:      "some-guid",
						Name:      "some-app",
						State:     constant.ApplicationStarted,
						Lifecycle: v3action.AppLifecycle{Type: constant.DockerAppLifecycleType},
					},
					CurrentDroplet: v3action.Droplet{
						Image: "docker/some-image",
					},
				},
				v3action.Warnings{"warning-1", "warning-2"},
				nil)
		})

		It("displays app information without routes", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("(?m)Showing health and status for app some-app in org some-org / space some-space as steve\\.\\.\\.\n\n"))
			Expect(testUI.Out).To(Say("name:\\s+some-app"))
			Expect(testUI.Out).To(Say("requested state:\\s+started"))
			Expect(testUI.Out).To(Say("processes:\\s+\\n"))
			Expect(testUI.Out).To(Say("memory usage:\\s+\\n"))
			Expect(testUI.Out).To(Say("routes:\\s+\\n"))
			Expect(testUI.Out).To(Say("stack:\\s+\\n"))
			Expect(testUI.Out).To(Say("(?m)docker image:\\s+docker/some-image$\\n"))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))

			Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))

			Expect(fakeV2Actor.GetApplicationRoutesCallCount()).To(Equal(0))
		})
	})

	Context("when app has no processes", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
				v3action.ApplicationSummary{
					Application: v3action.Application{
						GUID:  "some-guid",
						Name:  "some-app",
						State: constant.ApplicationStarted,
					},
				},
				v3action.Warnings{"warning-1", "warning-2"},
				nil)
		})

		It("displays app information without routes", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("(?m)Showing health and status for app some-app in org some-org / space some-space as steve\\.\\.\\.\n\n"))
			Expect(testUI.Out).To(Say("name:\\s+some-app"))
			Expect(testUI.Out).To(Say("requested state:\\s+started"))
			Expect(testUI.Out).To(Say("processes:\\s+\\n"))
			Expect(testUI.Out).To(Say("memory usage:\\s+\\n"))
			Expect(testUI.Out).To(Say("routes:\\s+\\n"))
			Expect(testUI.Out).To(Say("stack:\\s+\\n"))
			Expect(testUI.Out).To(Say("(?m)buildpacks:\\s+$\\n"))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))

			Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
			appName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
			Expect(appName).To(Equal("some-app"))
			Expect(spaceGUID).To(Equal("some-space-guid"))

			Expect(fakeV2Actor.GetApplicationRoutesCallCount()).To(Equal(0))
		})
	})

	Context("when app has processes", func() {
		Context("when getting routes returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = ccerror.RequestError{}
				summary := v3action.ApplicationSummary{
					Application: v3action.Application{
						Name:  "some-app",
						State: constant.ApplicationStarted,
					},
					ProcessSummaries: []v3action.ProcessSummary{
						{Process: v3action.Process{Type: constant.ProcessTypeWeb}},
					},
				}
				fakeActor.GetApplicationSummaryByNameAndSpaceReturns(
					summary,
					v3action.Warnings{"warning-1", "warning-2"},
					nil,
				)

				fakeV2Actor.GetApplicationRoutesReturns([]v2action.Route{}, v2action.Warnings{"route-warning-1", "route-warning-2"}, expectedErr)
			})

			It("returns the error and prints warnings", func() {
				Expect(executeErr).To(Equal(ccerror.RequestError{}))

				Expect(testUI.Out).To(Say("Showing health and status for app some-app in org some-org / space some-space as steve\\.\\.\\."))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
				Expect(testUI.Err).To(Say("route-warning-1"))
				Expect(testUI.Err).To(Say("route-warning-2"))
			})
		})

		Context("when getting routes does not return any errors", func() {
			BeforeEach(func() {
				fakeV2Actor.GetApplicationRoutesReturns([]v2action.Route{
					{Domain: v2action.Domain{Name: "some-other-domain"}}, {
						Domain: v2action.Domain{Name: "some-domain"}}},
					v2action.Warnings{"route-warning-1", "route-warning-2"}, nil)
			})

			Context("when there are no instances of any process in the app", func() {
				BeforeEach(func() {
					summary := v3action.ApplicationSummary{
						Application: v3action.Application{
							Name:  "some-app",
							State: constant.ApplicationStarted,
						},
						CurrentDroplet: v3action.Droplet{
							Stack: "cflinuxfs2",
							Buildpacks: []v3action.Buildpack{
								{
									Name:         "ruby_buildpack",
									DetectOutput: "some-detect-output",
								},
								{
									Name:         "some-buildpack",
									DetectOutput: "",
								},
							},
						},
						ProcessSummaries: []v3action.ProcessSummary{
							{
								Process: v3action.Process{
									Type:       "console",
									MemoryInMB: types.NullUint64{Value: 128, IsSet: true},
								},
							},
							{
								Process: v3action.Process{
									Type:       "worker",
									MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
								},
							},
							{
								Process: v3action.Process{
									Type:       constant.ProcessTypeWeb,
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
								},
							},
						},
					}
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(summary, nil, nil)
				})

				It("says no instances are running", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("There are no running instances of this app."))
				})

				It("does not display the instance table", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).NotTo(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				})
			})

			Context("when all the instances in all processes are down", func() {
				BeforeEach(func() {
					summary := v3action.ApplicationSummary{
						Application: v3action.Application{
							Name:  "some-app",
							State: constant.ApplicationStarted,
						},
						CurrentDroplet: v3action.Droplet{
							Stack: "cflinuxfs2",
							Buildpacks: []v3action.Buildpack{
								{
									Name:         "ruby_buildpack",
									DetectOutput: "some-detect-output",
								},
								{
									Name:         "some-buildpack",
									DetectOutput: "",
								},
							},
						},
						ProcessSummaries: []v3action.ProcessSummary{
							{
								Process: v3action.Process{
									Type:       "console",
									MemoryInMB: types.NullUint64{Value: 128, IsSet: true},
								},
								InstanceDetails: []v3action.ProcessInstance{{State: constant.ProcessInstanceDown}},
							},
							{
								Process: v3action.Process{
									Type:       "worker",
									MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
								},
								InstanceDetails: []v3action.ProcessInstance{{State: constant.ProcessInstanceDown}},
							},
							{
								Process: v3action.Process{
									Type:       constant.ProcessTypeWeb,
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
								},
								InstanceDetails: []v3action.ProcessInstance{{State: constant.ProcessInstanceDown}},
							},
						},
					}
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(summary, nil, nil)
				})

				It("says no instances are running", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("buildpacks:"))
					Expect(testUI.Out).To(Say("\n\n"))

					Expect(testUI.Out).To(Say("There are no running instances of this app."))
				})

				It("does not display the instance table", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).NotTo(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				})
			})

			Context("when there are running instances of the app", func() {
				BeforeEach(func() {
					summary := v3action.ApplicationSummary{
						Application: v3action.Application{
							Name:  "some-app",
							State: constant.ApplicationStarted,
						},
						CurrentDroplet: v3action.Droplet{
							Stack: "cflinuxfs2",
							Buildpacks: []v3action.Buildpack{
								{
									Name:         "ruby_buildpack",
									DetectOutput: "some-detect-output",
								},
								{
									Name:         "some-buildpack",
									DetectOutput: "",
								},
							},
						},
						ProcessSummaries: []v3action.ProcessSummary{
							{
								Process: v3action.Process{
									Type:       "console",
									MemoryInMB: types.NullUint64{Value: 128, IsSet: true},
								},
								InstanceDetails: []v3action.ProcessInstance{},
							},
							{
								Process: v3action.Process{
									Type:       "worker",
									MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
								},
								InstanceDetails: []v3action.ProcessInstance{
									v3action.ProcessInstance{
										Index:       0,
										State:       constant.ProcessInstanceDown,
										MemoryUsage: 4000000,
										DiskUsage:   4000000,
										MemoryQuota: 67108864,
										DiskQuota:   8000000,
										Uptime:      int(time.Now().Sub(time.Unix(1371859200, 0)).Seconds()),
									},
								},
							},
							{
								Process: v3action.Process{
									Type:       constant.ProcessTypeWeb,
									MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
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
						},
					}
					fakeActor.GetApplicationSummaryByNameAndSpaceReturns(summary, v3action.Warnings{"warning-1", "warning-2"}, nil)
				})

				It("prints the application summary and outputs warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("(?m)Showing health and status for app some-app in org some-org / space some-space as steve\\.\\.\\.\n\n"))
					Expect(testUI.Out).To(Say("name:\\s+some-app"))
					Expect(testUI.Out).To(Say("requested state:\\s+started"))
					Expect(testUI.Out).To(Say("processes:\\s+web:3/3, console:0/0, worker:0/1"))
					Expect(testUI.Out).To(Say("memory usage:\\s+32M x 3, 64M x 1"))
					Expect(testUI.Out).To(Say("routes:\\s+some-other-domain, some-domain"))
					Expect(testUI.Out).To(Say("stack:\\s+cflinuxfs2"))
					Expect(testUI.Out).To(Say("(?m)buildpacks:\\s+some-detect-output, some-buildpack\n\n"))
					Expect(testUI.Out).To(Say("web:3/3"))
					Expect(testUI.Out).To(Say("\\s+state\\s+since\\s+cpu\\s+memory\\s+disk"))
					Expect(testUI.Out).To(Say("#0\\s+running\\s+1978-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M\\s+0.0%\\s+976.6K of 32M\\s+976.6K of 1.9M"))
					Expect(testUI.Out).To(Say("#1\\s+running\\s+1980-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M\\s+0.0%\\s+1.9M of 32M\\s+1.9M of 3.8M"))
					Expect(testUI.Out).To(Say("#2\\s+running\\s+2010-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M\\s+0.0%\\s+2.9M of 32M\\s+2.9M of 5.7M"))

					Expect(testUI.Out).To(Say("console:0/0"))

					Expect(testUI.Out).To(Say("worker:0/1"))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("some-app"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})
		})
	})
})
